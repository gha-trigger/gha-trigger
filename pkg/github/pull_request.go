package github

import (
	"context"
)

type PullRequestsService interface {
	ListFiles(ctx context.Context, owner, repo string, number int, opts *ListOptions) ([]*CommitFile, *Response, error)
	Get(ctx context.Context, owner, repo string, number int) (*PullRequest, *Response, error)
}

const maxPerPage = 100

type ParamsListPRFiles struct {
	Owner  string
	Repo   string
	Number int
	Count  int
}

func (client *Client) GetPR(ctx context.Context, owner, repo string, number int) (*PullRequest, *Response, error) {
	return client.pr.Get(ctx, owner, repo, number)
}

func (client *Client) ListPRFiles(ctx context.Context, param *ParamsListPRFiles) ([]*CommitFile, *Response, error) {
	ret := []*CommitFile{}
	if param.Count == 0 {
		return nil, nil, nil
	}
	n := (param.Count / maxPerPage) + 1
	var gResp *Response
	for i := 1; i <= n; i++ {
		opts := &ListOptions{
			Page:    i,
			PerPage: maxPerPage,
		}
		files, resp, err := client.pr.ListFiles(ctx, param.Owner, param.Repo, param.Number, opts)
		if err != nil {
			return nil, resp, err
		}
		gResp = resp
		ret = append(ret, files...)
		if len(files) != maxPerPage {
			return ret, gResp, nil
		}
	}

	return ret, gResp, nil
}
