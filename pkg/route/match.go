package route

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

type matchFunc func(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error)

func Match(ctx context.Context, event *domain.Event, repo *config.Repo) ([]*config.Workflow, *domain.Response, error) {
	numEvents := len(repo.Events)
	var wfs []*config.Workflow
	for i := 0; i < numEvents; i++ {
		ev := repo.Events[i]
		f, resp, err := matchEvent(ctx, ev, event)
		if err != nil {
			return nil, resp, err
		}
		if f {
			wfs = append(wfs, ev.Workflow)
		}
	}
	return wfs, nil, nil
}

func matchEvent(ctx context.Context, ev *config.Event, event *domain.Event) (bool, *domain.Response, error) {
	if len(ev.Matches) == 0 {
		return true, nil, nil
	}
	for _, matchConfig := range ev.Matches {
		f, resp, err := matchMatchConfig(ctx, matchConfig, event)
		if err != nil {
			return false, resp, err
		}
		// OR Condition
		if f {
			return true, nil, nil
		}
	}
	return false, nil, nil
}

func matchMatchConfig(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	funcs := []matchFunc{
		matchEventType,
		matchBranches,
		matchTags,
		matchBranches,
		matchBranchesIgnore,
		matchTagsIgnore,
		// check paths lastly because api call is required
		matchPaths,
		matchPathsIgnore,
		matchIf,
	}
	for _, fn := range funcs {
		f, resp, err := fn(ctx, matchConfig, event)
		if err != nil {
			return false, resp, err
		}
		// AND condition
		if !f {
			return false, nil, nil
		}
	}
	return true, nil, nil
}

func matchEventType(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Events) == 0 {
		return true, nil, nil
	}
	for _, ev := range matchConfig.Events {
		if ev.Name != event.Type {
			continue
		}
		if len(ev.Types) == 0 {
			return true, nil, nil
		}
		for _, typ := range ev.Types {
			if typ == event.Payload.Action {
				return true, nil, nil
			}
		}
	}
	return false, nil, nil
}

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

func matchTags(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil, nil
	}
	if event.Payload.Ref != "" {
		ref := strings.TrimPrefix(event.Payload.Ref, "refs/tags/")
		for _, tag := range matchConfig.Tags {
			f, err := tag.Match(ref)
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

func matchPaths(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Paths) == 0 {
		return true, nil, nil
	}
	changedFiles, err := event.GetChangedFiles(ctx)
	if err != nil {
		return false, nil, err
	}
	for _, changedFile := range changedFiles {
		for _, p := range matchConfig.Paths {
			f, err := p.Match(changedFile)
			if err != nil {
				return false, nil, err
			}
			if f {
				return true, nil, nil
			}
		}
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

func matchTagsIgnore(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil, nil
	}
	if event.Payload.Ref != "" {
		ref := strings.TrimPrefix(event.Payload.Ref, "refs/tags/")
		for _, tag := range matchConfig.TagsIgnore {
			f, err := tag.Match(ref)
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

func matchPathsIgnore(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.PathsIgnore) == 0 {
		return true, nil, nil
	}
	changedFiles, err := event.GetChangedFiles(ctx)
	if err != nil {
		return false, nil, err
	}
	for _, changedFile := range changedFiles {
		f, err := matchPath(changedFile, matchConfig.PathsIgnore)
		if err != nil {
			return false, nil, err
		}
		if !f {
			return true, nil, nil
		}
	}
	return false, nil, nil
}

func matchPath(changedFile string, paths []*config.StringMatch) (bool, error) {
	for _, p := range paths {
		f, err := p.Match(changedFile)
		if err != nil {
			return false, err
		}
		if f {
			return true, nil
		}
	}
	return false, nil
}

func matchIf(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	// TODO
	return true, nil, nil
}
