package reposerver

import (
	"time"

	"github.com/google/go-github/v72/github"
)

type RepoChannelObject struct {
	Repository      github.Repository
	LatestCommitSha string
}

type RepoServer struct {
	Client          *github.Client
	RefreshInterval time.Duration // time the server should check
	RepoChannel     chan<- RepoChannelObject
}

type RequestRepo struct {
	User         string   `json:"user"`
	Repositories []string `json:"repositories"`
	IsOrg        bool     `json:"isOrg,omitempty"`
}
