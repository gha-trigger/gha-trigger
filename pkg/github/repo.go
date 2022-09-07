package github

import (
	"context"
)

type RepositoriesService interface {
	GetCommit(ctx context.Context, owner, repo, sha string, opts *ListOptions) (*RepositoryCommit, *Response, error)
}

func (client *Client) GetCommit(ctx context.Context, owner, repo, sha string) (*RepositoryCommit, *Response, error) {
	opt := &ListOptions{
		PerPage: 100,
	}
	baseCommit, resp, err := client.repo.GetCommit(ctx, owner, repo, sha, opt)
	if err != nil {
		return nil, resp, err
	}
	if resp.NextPage == 0 {
		return baseCommit, resp, err
	}
	// https://docs.github.com/en/rest/commits/commits#get-a-commit
	// Note: If there are more than 300 files in the commit diff,
	// the response will include pagination link headers for the remaining files, up to a limit of 3000 files
	for i := 0; i < 30; i++ {
		commit, resp, err := client.repo.GetCommit(ctx, owner, repo, sha, opt)
		if err != nil {
			return nil, resp, err
		}
		baseCommit.Files = append(baseCommit.Files, commit.Files...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return baseCommit, resp, err
}
