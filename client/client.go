package client

import "github.com/google/go-github/github"

type Client interface {
	Releases(repo string) ([]Release, error)
	Assets(repo string, id int64) ([]Asset, error)
}

type Release struct {
	ID            int64
	Name          string
	Tag           string
	Prerelease    bool
	PublishedTime github.Timestamp
	Description   string
}

type Asset struct {
	Name      string
	Downloads int
}
