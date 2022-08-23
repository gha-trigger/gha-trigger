package lambda

import (
	"regexp"
	"strings"

	"github.com/suzuki-shunsuke/gha-trigger/pkg/config"
	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
)

type matchFunc func(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error)

func (handler *Handler) match(body interface{}, event *Event) ([]*config.WorkflowConfig, *Response, error) {
	cfg := handler.cfg
	numEvents := len(cfg.Events)
	var wfs []*config.WorkflowConfig
	for i := 0; i < numEvents; i++ {
		ev := cfg.Events[i]
		f, resp, err := handler.matchEvent(ev, body, event)
		if err != nil {
			return nil, resp, err
		}
		if f {
			wfs = append(wfs, ev.Workflows...)
		}
	}
	return wfs, nil, nil
}

func (handler *Handler) matchEvent(ev *config.EventConfig, payload interface{}, event *Event) (bool, *Response, error) {
	if len(ev.Matches) == 0 {
		return true, nil, nil
	}
	for _, matchConfig := range ev.Matches {
		f, resp, err := handler.matchMatchConfig(matchConfig, payload, event)
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

func (handler *Handler) matchMatchConfig(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) {
	funcs := []matchFunc{
		handler.matchRepo,
		handler.matchBranches,
		handler.matchTags,
		handler.matchPaths,
		handler.matchBranches,
		handler.matchBranchesIgnore,
		handler.matchTagsIgnore,
		handler.matchPathsIgnore,
		handler.matchIf,
	}
	for _, fn := range funcs {
		f, resp, err := fn(matchConfig, payload, event)
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

type HasPR interface {
	GetPullRequest() *github.PullRequest
}

type HasRef interface {
	GetRef() string
}

func (handler *Handler) matchRepo(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) {
	repo := event.Repo
	return repo.GetFullName() == matchConfig.RepoOwner+"/"+matchConfig.RepoName, nil, nil
}

func (handler *Handler) matchBranches(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) { //nolint:cyclop
	if len(matchConfig.Branches) == 0 {
		return true, nil, nil
	}
	if hasPR, ok := payload.(HasPR); ok {
		pr := hasPR.GetPullRequest()
		base := pr.GetBase()
		ref := base.GetRef()
		for _, branch := range matchConfig.Branches {
			if ref == branch {
				return true, nil, nil
			}
		}
		for _, branch := range matchConfig.CompiledBranches {
			if branch.MatchString(ref) {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	if hasRef, ok := payload.(HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/heads/")
		for _, branch := range matchConfig.Branches {
			if ref == branch {
				return true, nil, nil
			}
		}
		for _, branch := range matchConfig.CompiledBranches {
			if branch.MatchString(ref) {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func (handler *Handler) matchTags(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil, nil
	}
	if hasRef, ok := payload.(HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/tags/")
		for _, tag := range matchConfig.Tags {
			if ref == tag {
				return true, nil, nil
			}
		}
		for _, tag := range matchConfig.CompiledTags {
			if tag.MatchString(ref) {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func (handler *Handler) matchPaths(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) {
	if len(matchConfig.Paths) == 0 {
		return true, nil, nil
	}
	for _, changedFile := range event.ChangedFiles {
		for _, p := range matchConfig.Paths {
			if changedFile == p {
				return true, nil, nil
			}
		}
		for _, p := range matchConfig.CompiledPaths {
			if p.MatchString(changedFile) {
				return true, nil, nil
			}
		}
	}
	return false, nil, nil
}

func (handler *Handler) matchBranchesIgnore(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) { //nolint:cyclop
	if len(matchConfig.BranchesIgnore) == 0 {
		return true, nil, nil
	}
	if hasPR, ok := payload.(HasPR); ok {
		pr := hasPR.GetPullRequest()
		base := pr.GetBase()
		ref := base.GetRef()
		for _, branch := range matchConfig.Branches {
			if ref == branch {
				return false, nil, nil
			}
		}
		for _, branch := range matchConfig.CompiledBranches {
			if branch.MatchString(ref) {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	if hasRef, ok := payload.(HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/heads/")
		for _, branch := range matchConfig.Branches {
			if ref == branch {
				return false, nil, nil
			}
		}
		for _, branch := range matchConfig.CompiledBranches {
			if branch.MatchString(ref) {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	return true, nil, nil
}

func (handler *Handler) matchTagsIgnore(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil, nil
	}
	if hasRef, ok := payload.(HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/tags/")
		for _, tag := range matchConfig.Tags {
			if ref == tag {
				return false, nil, nil
			}
		}
		for _, tag := range matchConfig.CompiledTags {
			if tag.MatchString(ref) {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	return true, nil, nil
}

func (handler *Handler) matchPathsIgnore(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) {
	if len(matchConfig.PathsIgnore) == 0 {
		return true, nil, nil
	}
	for _, changedFile := range event.ChangedFiles {
		if !matchPath(changedFile, matchConfig.PathsIgnore, matchConfig.CompiledPaths) {
			return true, nil, nil
		}
	}
	return false, nil, nil
}

func matchPath(changedFile string, paths []string, compiledPaths []*regexp.Regexp) bool {
	for _, p := range paths {
		if changedFile == p {
			return true
		}
	}
	for _, p := range compiledPaths {
		if p.MatchString(changedFile) {
			return true
		}
	}
	return false
}

func (handler *Handler) matchIf(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) {
	return true, nil, nil
}

// func (handler *Handler) _matchRepo(repoConfig *config.RepoConfig, payload interface{}, event *Event) ([]*config.WorkflowConfig, *Response, error) {
// 	repoEvent, ok := payload.(RepoEvent)
// 	if !ok {
// 		return nil, nil, nil
// 	}
// 	repo := repoEvent.GetRepo()
// 	if fullName := repo.GetFullName(); repoConfig.Name != fullName {
// 		return nil, nil, nil
// 	}
// 	numWorkflows := len(repoConfig.Workflows)
// 	workflows := make([]*config.WorkflowConfig, 0, numWorkflows)
// 	for j := 0; j < numWorkflows; j++ {
// 		workflowConfig := repoConfig.Workflows[j]
// 		f, resp, err := handler.matchWorkflow(workflowConfig, payload, event)
// 		if err != nil {
// 			return nil, resp, err
// 		}
// 		if f {
// 			workflows = append(workflows, workflowConfig)
// 		}
// 	}
// 	return workflows, nil, nil
// }

// func (handler *Handler) matchWorkflow(workflowConfig *config.WorkflowConfig, payload interface{}, event *Event) (bool, *Response, error) {
// 	numConditions := len(workflowConfig.Conditions)
// 	for k := 0; k < numConditions; k++ {
// 		workflowCondition := workflowConfig.Conditions[k]
// 		f, resp, err := handler.matchCondition(workflowCondition, payload, event)
// 		if err != nil {
// 			return false, resp, err
// 		}
// 		if f {
// 			// OR Condition
// 			return true, nil, nil
// 		}
// 	}
// 	return false, nil, nil
// }

// func (handler *Handler) matchCondition(matchConfig *config.MatchConfig, payload interface{}, event *Event) (bool, *Response, error) {
// 	// And Condition
// 	funcs := []matchFunc{
// 		handler.matchEvent,
// 		handler.matchBranches,
// 		handler.matchTags,
// 		handler.matchPaths,
// 		handler.matchBranches,
// 		handler.matchBranchesIgnore,
// 		handler.matchTagsIgnore,
// 		handler.matchPathsIgnore,
// 		handler.matchIf,
// 	}
// 	for _, fn := range funcs {
// 		f, resp, err := fn(wfCondition, payload, event)
// 		if err != nil {
// 			return false, resp, err
// 		}
// 		if !f {
// 			return false, nil, nil
// 		}
// 	}
// 	return true, nil, nil
// }
