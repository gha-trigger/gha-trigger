package route

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

type matchFunc func(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, error)

func Match(ctx context.Context, event *domain.Event, repo *config.Repo) ([]*config.Workflow, error) {
	numEvents := len(repo.Events)
	var wfs []*config.Workflow
	for i := 0; i < numEvents; i++ {
		ev := repo.Events[i]
		f, err := matchEvent(ctx, ev, event)
		if err != nil {
			return nil, err
		}
		if f {
			wfs = append(wfs, ev.Workflow)
		}
	}
	return wfs, nil
}

func matchEvent(ctx context.Context, ev *config.Event, event *domain.Event) (bool, error) {
	if len(ev.Matches) == 0 {
		return true, nil
	}
	for _, matchConfig := range ev.Matches {
		f, err := matchMatchConfig(ctx, matchConfig, event)
		if err != nil {
			return false, err
		}
		// OR Condition
		if f {
			return true, nil
		}
	}
	return false, nil
}

func matchMatchConfig(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, error) {
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
		f, err := fn(ctx, matchConfig, event)
		if err != nil {
			return false, err
		}
		// AND condition
		if !f {
			return false, nil
		}
	}
	return true, nil
}
