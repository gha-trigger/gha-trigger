package route

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func matchPaths(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, error) {
	if len(matchConfig.Paths) == 0 {
		return true, nil
	}
	changedFiles, err := event.GetChangedFiles(ctx)
	if err != nil {
		return false, err
	}
	for _, changedFile := range changedFiles {
		for _, p := range matchConfig.Paths {
			f, err := p.Match(changedFile)
			if err != nil {
				return false, err
			}
			if f {
				return true, nil
			}
		}
	}
	return false, nil
}

func matchPathsIgnore(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, error) {
	if len(matchConfig.PathsIgnore) == 0 {
		return true, nil
	}
	changedFiles, err := event.GetChangedFiles(ctx)
	if err != nil {
		return false, err
	}
	for _, changedFile := range changedFiles {
		f, err := matchPath(changedFile, matchConfig.PathsIgnore)
		if err != nil {
			return false, err
		}
		if !f {
			return true, nil
		}
	}
	return false, nil
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
