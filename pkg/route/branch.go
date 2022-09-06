package route

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func matchBranches(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Branches) == 0 {
		return true, nil, nil
	}
	if event.Payload.PullRequest != nil {
		base := event.Payload.PullRequest.GetBase()
		ref := base.GetRef()
		for _, branch := range matchConfig.Branches {
			f, err := branch.Match(ref)
			if err != nil {
				return false, nil, err
			}
			if f {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	if event.Payload.Ref != "" {
		ref := strings.TrimPrefix(event.Payload.Ref, "refs/heads/")
		for _, branch := range matchConfig.Branches {
			f, err := branch.Match(ref)
			if err != nil {
				return false, nil, err
			}
			if f {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func matchBranchesIgnore(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.BranchesIgnore) == 0 {
		return true, nil, nil
	}
	if event.Payload.PullRequest != nil {
		base := event.Payload.PullRequest.GetBase()
		ref := base.GetRef()
		for _, branch := range matchConfig.BranchesIgnore {
			f, err := branch.Match(ref)
			if err != nil {
				return false, nil, err
			}
			if f {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	if event.Payload.Ref != "" {
		ref := strings.TrimPrefix(event.Payload.Ref, "refs/heads/")
		for _, branch := range matchConfig.BranchesIgnore {
			f, err := branch.Match(ref)
			if err != nil {
				return false, nil, err
			}
			if f {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	return true, nil, nil
}
