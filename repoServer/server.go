package reposerver

import (
	"context"
	"fmt"
	"paldab/home-agent-operator/config"
	"time"

	"github.com/google/go-github/v72/github"
	"go.uber.org/zap"
)

var (
	isAuthenticatedUser bool
)

func (r *RepoServer) StartServer(ctx context.Context) error {
	if config.GITHUB_TOKEN == "" {
		return fmt.Errorf("could not start RepoServer. missing variable GITHUB_TOKEN")
	}

	ticker := time.NewTicker(r.RefreshInterval)
	manualRefreshChan := make(chan struct{})
	quit := make(chan struct{})

	zap.L().Info("Starting reposerver", zap.String("client", "github"), zap.Bool("Using GITHUB_TOKEN", isAuthenticatedUser), zap.Any("refreshInterval", r.RefreshInterval))

	for {
		select {
		case <-ticker.C:
			r.refreshRepos()
		case <-manualRefreshChan:
			r.refreshRepos()
		case <-ctx.Done():
			r.serverGracefulShutdown(ticker)
			return nil
		case <-quit:
			r.serverGracefulShutdown(ticker)
			return nil
		}
	}
}

func (r *RepoServer) refreshRepos() {
	var repos []*github.Repository
	var err error

	zap.L().Info("refreshing repositories", zap.String("organization", config.GITHUB_ORG), zap.String("user", ""), zap.Bool("isAuthenticatedUser", isAuthenticatedUser))

	if config.GITHUB_ORG != "" {
		repos, err = r.getReposFromOrganization(config.GITHUB_ORG)

		if err != nil {
			zap.L().Error("could not fetch repos from organization", zap.String("organization", config.GITHUB_ORG), zap.Bool("isAuthenticatedUser", isAuthenticatedUser), zap.Error(err))
		}
	} else {
		// eventually, add dynamic adding of repos. Now only get user from authenticated GITHUB token
		repos, err = r.getReposFromUser("")
		if err != nil {
			zap.L().Error("could not fetch repos from user", zap.String("user", ""), zap.Bool("isAuthenticatedUser", isAuthenticatedUser), zap.Error(err))
		}
	}

	// Sending repos to channel
	for _, repo := range repos {
		latestCommitSha, err := r.getLatestCommitSHA(repo)

		if err != nil {
			zap.L().Error("could not find latest commit from repository", zap.String("repository", repo.GetName()), zap.Error(err))
			continue
		}

		repoObject := RepoChannelObject{
			Repository:      *repo,
			LatestCommitSha: latestCommitSha,
		}

		r.RepoChannel <- repoObject
	}
}

func (r *RepoServer) serverGracefulShutdown(ticker *time.Ticker) {
	zap.L().Info("stopping repo server")
	ticker.Stop()
	close(r.RepoChannel)
}

func NewRepoServer(githubToken string, refreshInterval time.Duration, repoChannel chan<- RepoChannelObject) *RepoServer {
	isAuthenticatedUser = githubToken != ""

	return &RepoServer{
		Client:          newGithubClient(githubToken),
		RefreshInterval: refreshInterval,
		RepoChannel:     repoChannel,
	}
}
