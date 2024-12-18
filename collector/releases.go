package collector

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/github_releases_exporter/client"
	"github.com/caarlos0/github_releases_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type releasesCollector struct {
	mutex  sync.Mutex
	config *config.Config
	client client.Client

	up             *prometheus.Desc
	downloads      *prometheus.Desc
	publishedTime  *prometheus.Desc
	scrapeDuration *prometheus.Desc
}

// NewReleasesCollector returns a releases collector
func NewReleasesCollector(config *config.Config, client client.Client) prometheus.Collector {
	const namespace = "github"
	const subsystem = "release"
	return &releasesCollector{
		config: config,
		client: client,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "up"),
			"Exporter is being able to talk with GitHub API",
			nil,
			nil,
		),
		downloads: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "asset_download_count"),
			"Download count of each asset of a github release",
			[]string{"repository", "tag", "name", "extension"},
			nil,
		),
		publishedTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "published_time"),
			"published time of a github release",
			[]string{"repository", "releaseid", "releasename", "publishedtime", "tag", "prerelease", "description"},
			nil,
		),
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "scrape_duration_seconds"),
			"Returns how long the probe took to complete in seconds",
			nil,
			nil,
		),
	}
}

// Describe all metrics
func (c *releasesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.downloads
	ch <- c.publishedTime
	ch <- c.scrapeDuration
}

// Collect all metrics
func (c *releasesCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	log.Infof("collecting %d repositories", len(c.config.Repositories))

	var start = time.Now()
	var success = 1
	for _, repository := range c.config.Repositories {
		log.Infof("collecting %s", repository)
		releases, err := c.client.Releases(repository)
		if err != nil {
			success = 0
			log.Errorf("failed to collect: %s", err.Error())
			continue
		}

		for _, release := range releases {
			ch <- prometheus.MustNewConstMetric(
				c.publishedTime,
				prometheus.GaugeValue,
				float64(release.PublishedTime.Unix()),
				repository,
				strconv.FormatInt(release.ID, 10),
				sanitizeLabelValue(release.Name),
				release.PublishedTime.String(),
				release.Tag,
				strconv.FormatBool(release.Prerelease),
				sanitizeLabelValue(release.Description),
			)

			assets, err := c.client.Assets(repository, release.ID)
			if err != nil {
				success = 0
				log.Errorf(
					"failed to collect repo %s, release %s: %s",
					repository,
					release.Tag,
					err.Error(),
				)
				continue
			}
			for _, asset := range assets {
				ext := strings.TrimPrefix(filepath.Ext(asset.Name), ".")
				log.Debugf(
					"collecting %s@%s / %s (%s)",
					repository,
					release.Tag,
					asset.Name,
					ext,
				)
				ch <- prometheus.MustNewConstMetric(
					c.downloads,
					prometheus.CounterValue,
					float64(asset.Downloads),
					repository,
					release.Tag,
					asset.Name,
					ext,
				)
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, float64(success))
}

// sanitizeLabelValue 清理字符串以符合 Prometheus 标签的要求
func sanitizeLabelValue(value string) string {
	// 去除前后的空白字符
	value = strings.TrimSpace(value)
	// 使用正则表达式替换特殊字符
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	value = re.ReplaceAllString(value, "_")
	// 限制长度
	if len(value) > 100 {
		value = value[:100]
	}
	return value
}
