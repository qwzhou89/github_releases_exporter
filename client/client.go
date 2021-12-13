package client

type Client interface {
	Releases(repo string) ([]Release, error)
	Assets(repo string, id int64) ([]Asset, error)
}

type Release struct {
	ID            int64
	Tag           string
	Prerelease    bool
	PublishedTime int64
}

type Asset struct {
	Name      string
	Downloads int
}
