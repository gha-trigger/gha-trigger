package route

import (
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

type matchFunc func(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error)

func Match(event *domain.Event, repo *config.Repo) ([]*config.Workflow, *domain.Response, error) {
	numEvents := len(repo.Events)
	var wfs []*config.Workflow
	for i := 0; i < numEvents; i++ {
		ev := repo.Events[i]
		f, resp, err := matchEvent(ev, event)
		if err != nil {
			return nil, resp, err
		}
		if f {
			wfs = append(wfs, ev.Workflow)
		}
	}
	return wfs, nil, nil
}

func matchEvent(ev *config.Event, event *domain.Event) (bool, *domain.Response, error) {
	if len(ev.Matches) == 0 {
		return true, nil, nil
	}
	for _, matchConfig := range ev.Matches {
		f, resp, err := matchMatchConfig(matchConfig, event)
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

func matchMatchConfig(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
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
		f, resp, err := fn(matchConfig, event)
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

func matchEventType(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
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
			if typ == event.Action {
				return true, nil, nil
			}
		}
	}
	return false, nil, nil
}

func matchBranches(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Branches) == 0 {
		return true, nil, nil
	}
	payload := event.Body
	if hasPR, ok := payload.(domain.HasPR); ok {
		pr := hasPR.GetPullRequest()
		base := pr.GetBase()
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
	if hasRef, ok := payload.(domain.HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/heads/")
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

func matchTags(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil, nil
	}
	payload := event.Body
	if hasRef, ok := payload.(domain.HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/tags/")
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

func matchPaths(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Paths) == 0 {
		return true, nil, nil
	}
	for _, changedFile := range event.ChangedFiles {
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

func matchBranchesIgnore(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.BranchesIgnore) == 0 {
		return true, nil, nil
	}
	payload := event.Body
	if hasPR, ok := payload.(domain.HasPR); ok {
		pr := hasPR.GetPullRequest()
		base := pr.GetBase()
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
	if hasRef, ok := payload.(domain.HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/heads/")
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

func matchTagsIgnore(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil, nil
	}
	payload := event.Body
	if hasRef, ok := payload.(domain.HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/tags/")
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

func matchPathsIgnore(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.PathsIgnore) == 0 {
		return true, nil, nil
	}
	for _, changedFile := range event.ChangedFiles {
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

func matchIf(matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	// TODO
	return true, nil, nil
}
