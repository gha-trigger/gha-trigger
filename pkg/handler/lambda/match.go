package lambda

import (
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

type matchFunc func(matchConfig *config.Match, event *Event) (bool, *Response, error)

func (handler *Handler) match(event *Event, repo *config.Repo) ([]*config.Workflow, *Response, error) {
	numEvents := len(repo.Events)
	var wfs []*config.Workflow
	for i := 0; i < numEvents; i++ {
		ev := repo.Events[i]
		f, resp, err := handler.matchEvent(ev, event)
		if err != nil {
			return nil, resp, err
		}
		if f {
			wfs = append(wfs, ev.Workflow)
		}
	}
	return wfs, nil, nil
}

func (handler *Handler) matchEvent(ev *config.Event, event *Event) (bool, *Response, error) {
	if len(ev.Matches) == 0 {
		return true, nil, nil
	}
	for _, matchConfig := range ev.Matches {
		f, resp, err := handler.matchMatchConfig(matchConfig, event)
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

func (handler *Handler) matchMatchConfig(matchConfig *config.Match, event *Event) (bool, *Response, error) {
	funcs := []matchFunc{
		handler.matchEventType,
		handler.matchBranches,
		handler.matchTags,
		handler.matchBranches,
		handler.matchBranchesIgnore,
		handler.matchTagsIgnore,
		// check paths lastly because api call is required
		handler.matchPaths,
		handler.matchPathsIgnore,
		handler.matchIf,
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

func (handler *Handler) matchEventType(matchConfig *config.Match, event *Event) (bool, *Response, error) {
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

func (handler *Handler) matchBranches(matchConfig *config.Match, event *Event) (bool, *Response, error) {
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

func (handler *Handler) matchTags(matchConfig *config.Match, event *Event) (bool, *Response, error) {
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

func (handler *Handler) matchPaths(matchConfig *config.Match, event *Event) (bool, *Response, error) {
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

func (handler *Handler) matchBranchesIgnore(matchConfig *config.Match, event *Event) (bool, *Response, error) {
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

func (handler *Handler) matchTagsIgnore(matchConfig *config.Match, event *Event) (bool, *Response, error) {
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

func (handler *Handler) matchPathsIgnore(matchConfig *config.Match, event *Event) (bool, *Response, error) {
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

func (handler *Handler) matchIf(matchConfig *config.Match, event *Event) (bool, *Response, error) {
	// TODO
	return true, nil, nil
}
