package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/gha-trigger/gha-trigger/pkg/github"
)

type Event struct {
	Body            interface{}
	Raw             map[string]interface{}
	Payload         *Payload
	ChangedFiles    []string
	ChangedFileObjs []*github.CommitFile
	Type            string
	Request         *Request
	GitHub          GitHubInEvent
}

type GitHubInEvent interface {
	ListPRFiles(ctx context.Context, param *github.ParamsListPRFiles) ([]*github.CommitFile, *github.Response, error)
	GetCommit(ctx context.Context, owner, repo, sha string) (*github.RepositoryCommit, *github.Response, error)
}

func getChangedFiles(files []*github.CommitFile) []string {
	changedFileMap := make(map[string]struct{}, len(files))
	for _, file := range files {
		if f := file.GetFilename(); f != "" {
			changedFileMap[f] = struct{}{}
		}
		if f := file.GetPreviousFilename(); f != "" {
			changedFileMap[f] = struct{}{}
		}
	}
	changedFiles := make([]string, 0, len(changedFileMap))
	for k := range changedFileMap {
		changedFiles = append(changedFiles, k)
	}
	return changedFiles
}

func (ev *Event) GetChangedFiles(ctx context.Context) ([]string, error) {
	if ev.ChangedFileObjs != nil {
		return ev.ChangedFiles, nil
	}
	switch ev.Type {
	case "pull_request", "pull_request_target":
		if ev.Payload.PullRequest == nil {
			return nil, errors.New("body must have a pull request")
		}
		pr := ev.Payload.PullRequest
		files, _, err := ev.GitHub.ListPRFiles(ctx, &github.ParamsListPRFiles{
			Owner:  ev.Payload.Repo.GetOwner().GetLogin(),
			Repo:   ev.Payload.Repo.GetName(),
			Number: pr.GetNumber(),
			Count:  pr.GetChangedFiles(),
		})
		if err != nil {
			return nil, fmt.Errorf("list pull request files: %w", err)
		}
		ev.ChangedFileObjs = files
		ev.ChangedFiles = getChangedFiles(files)
	case "push":
		commit, _, err := ev.GitHub.GetCommit(ctx, ev.Payload.Repo.GetOwner().GetLogin(), ev.Payload.Repo.GetName(), ev.Payload.HeadCommit.GetID())
		if err != nil {
			return nil, fmt.Errorf("list commit files: %w", err)
		}
		ev.ChangedFileObjs = commit.Files
		ev.ChangedFiles = getChangedFiles(commit.Files)
	default:
	}
	return ev.ChangedFiles, nil
}
