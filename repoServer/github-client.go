package reposerver

import (
	"context"
	"fmt"

	"github.com/google/go-github/v72/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func newGithubClient(githubToken string) *github.Client {
	var client *github.Client

	if githubToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
		tc := oauth2.NewClient(context.Background(), ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return client
}

var defaultListOptions github.ListOptions = github.ListOptions{
	PerPage: 50,
}

func (r *RepoServer) getReposFromUser(user string) ([]*github.Repository, error) {
	ctx := context.Background()
	if !isAuthenticatedUser {
		repos, _, err := r.Client.Repositories.ListByUser(ctx, user, &github.RepositoryListByUserOptions{
			Type:        "all",
			ListOptions: defaultListOptions,
		})

		if err != nil {
			return []*github.Repository{}, err
		}

		return repos, nil
	}

	// authenticated user
	repos, _, err := r.Client.Repositories.ListByAuthenticatedUser(ctx, &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: defaultListOptions,
	})

	if err != nil {
		return []*github.Repository{}, err
	}

	return repos, nil
}

func (r *RepoServer) getReposFromOrganization(org string) ([]*github.Repository, error) {
	repos, _, err := r.Client.Repositories.ListByOrg(context.Background(), org, &github.RepositoryListByOrgOptions{
		ListOptions: defaultListOptions,
	})
	if err != nil {
		return []*github.Repository{}, err
	}

	return repos, nil
}

func (r *RepoServer) getLatestCommitSHA(repository *github.Repository) (string, error) {
	masterBranch := repository.GetDefaultBranch()

	if masterBranch == "" {
		return "", fmt.Errorf("could not get masterbranch")
	}

	owner := repository.GetOwner()

	if owner == nil {
		return "", fmt.Errorf("could not fetch any owner information")
	}

	branch, _, err := r.Client.Repositories.GetBranch(context.TODO(), owner.GetLogin(), repository.GetName(), masterBranch, 5)

	if err != nil {
		return "", err
	}

	return *branch.GetCommit().SHA, nil
}

// dev
func PrintRepos(repos []*github.Repository, client *github.Client) {
	for _, repo := range repos {
		branch := *repo.DefaultBranch

		// smth worng here
		commit, _, err := client.Repositories.GetCommit(context.TODO(), repo.GetOwner().String(), repo.GetName(), branch, nil)

		if err != nil {
			fmt.Println(err)
		}

		zap.L().Info(
			"repo", zap.String("repo", *repo.Name), zap.Bool("Private", *repo.Private), zap.String("Default Branch", branch), zap.Any("Commit", commit))
	}
}
