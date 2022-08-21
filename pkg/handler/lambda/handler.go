package lambda

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/config"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/domain"
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/github"
	"go.uber.org/zap"
)

type Handler struct {
	secret  *Secret
	actions domain.ActionsService
	// gh      domain.GitHub
	cfg    *config.Config
	logger *zap.Logger
}

func New() (*Handler, error) {
	// read config
	// read env
	// read secret
	// initialize handler
	return &Handler{}, nil
}

type Secret struct {
	WebhookSecret string
}

type Event struct {
	Body         string             `json:"body"`
	Headers      *Headers           `json:"headers"`
	ChangedFiles []string           `json:"-"`
	Repo         *github.Repository `json:"-"`
}

type Headers struct {
	Signature string `json:"signature"`
	Event     string `json:"event"`
}

type Response struct {
	StatusCode int              `json:"statusCode"`
	Headers    *ResponseHeaders `json:"headers"`
	Body       interface{}      `json:"body"`
}

type ResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

func (handler *Handler) Do(ctx context.Context, event *Event) (*Response, error) {
	logger := handler.logger
	if err := github.ValidateSignature(event.Headers.Signature, []byte(event.Body), []byte(handler.secret.WebhookSecret)); err != nil {
		logger.Debug("validate the webhook signature", zap.Error(err))
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "signature is invalid",
			},
		}, nil
	}

	body, err := github.ParseWebHook(event.Headers.Event, []byte(event.Body))
	if err != nil {
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body: map[string]interface{}{
				"error": "failed to parse a webhook payload",
			},
		}, nil
	}

	repoEvent, ok := body.(RepoEvent)
	if !ok {
		return &Response{
			StatusCode: http.StatusOK,
			Body: map[string]interface{}{
				"message": "event is ignored because a repository isn't found in the payload",
			},
		}, nil
	}
	repo := repoEvent.GetRepo()
	event.Repo = repo
	repoOwner := repo.GetOwner()
	logger = logger.With(
		zap.String("repo_owner", repoOwner.GetLogin()),
		zap.String("repo_name", repo.GetName()),
	)

	// route and filter request
	// list labels and changed files
	workflows, resp, err := handler.match(body, event)
	if err != nil {
		return resp, err
	}
	if len(workflows) == 0 {
		return nil, nil //nolint:nilnil
	}

	numWorkflows := len(workflows)
	for i := 0; i < numWorkflows; i++ {
		workflow := workflows[i]
		// Run GitHub Actions Workflow
		_, err := handler.actions.CreateWorkflowDispatchEventByFileName(ctx, repo.GetOwner().GetLogin(), repo.GetName(), workflow.FileName, github.CreateWorkflowDispatchEventRequest{})
		if err != nil {
			logger.Error(
				"create a workflow dispatch event by file name",
				zap.Error(err), zap.String("workflow_file_name", workflow.FileName))
			continue
		}
	}
	return nil, nil //nolint:nilnil
}

type RepoEvent interface {
	GetRepo() *github.Repository
}

func (handler *Handler) match(body interface{}, event *Event) ([]*config.WorkflowConfig, *Response, error) {
	cfg := handler.cfg
	numRepos := len(cfg.Repos)
	for i := 0; i < numRepos; i++ {
		repoConfig := cfg.Repos[i]
		workflows, resp, err := handler.matchRepo(repoConfig, body, event)
		if err != nil {
			return nil, resp, err
		}
		if len(workflows) > 0 {
			return workflows, resp, nil
		}
	}
	return nil, nil, nil
}

func (handler *Handler) matchRepo(repoConfig *config.RepoConfig, payload interface{}, event *Event) ([]*config.WorkflowConfig, *Response, error) {
	repoEvent, ok := payload.(RepoEvent)
	if !ok {
		return nil, nil, nil
	}
	repo := repoEvent.GetRepo()
	if fullName := repo.GetFullName(); repoConfig.Name != fullName {
		return nil, nil, nil
	}
	numWorkflows := len(repoConfig.Workflows)
	workflows := make([]*config.WorkflowConfig, 0, numWorkflows)
	for j := 0; j < numWorkflows; j++ {
		workflowConfig := repoConfig.Workflows[j]
		f, resp, err := handler.matchWorkflow(workflowConfig, payload, event)
		if err != nil {
			return nil, resp, err
		}
		if f {
			workflows = append(workflows, workflowConfig)
		}
	}
	return workflows, nil, nil
}

func (handler *Handler) matchWorkflow(workflowConfig *config.WorkflowConfig, payload interface{}, event *Event) (bool, *Response, error) {
	numConditions := len(workflowConfig.Conditions)
	for k := 0; k < numConditions; k++ {
		workflowCondition := workflowConfig.Conditions[k]
		f, resp, err := handler.matchCondition(workflowCondition, payload, event)
		if err != nil {
			return false, resp, err
		}
		if f {
			// OR Condition
			return true, nil, nil
		}
	}
	return false, nil, nil
}

type matchFunc func(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error)

func (handler *Handler) matchCondition(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) {
	// And Condition
	funcs := []matchFunc{
		handler.matchEvent,
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
		f, resp, err := fn(wfCondition, payload, event)
		if err != nil {
			return false, resp, err
		}
		if !f {
			return false, nil, nil
		}
	}
	return true, nil, nil
}

type HasEventType interface {
	GetAction() string
}

func (handler *Handler) matchEvent(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) {
	if len(wfCondition.Events) == 0 {
		return true, nil, nil
	}
	hasEventType, ok := payload.(HasEventType)
	if !ok {
		hasEventType = nil
	}
	for _, eventConfig := range wfCondition.Events {
		if eventConfig.Name != event.Headers.Event {
			continue
		}
		if len(eventConfig.Types) == 0 {
			return true, nil, nil
		}
		if hasEventType == nil {
			continue
		}
		for _, eventType := range eventConfig.Types {
			if eventType == hasEventType.GetAction() {
				return true, nil, nil
			}
		}
	}
	return false, nil, nil
}

type HasPR interface {
	GetPullRequest() *github.PullRequest
}

type HasRef interface {
	GetRef() string
}

func (handler *Handler) matchBranches(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) { //nolint:cyclop
	if len(wfCondition.Branches) == 0 {
		return true, nil, nil
	}
	if hasPR, ok := payload.(HasPR); ok {
		pr := hasPR.GetPullRequest()
		base := pr.GetBase()
		ref := base.GetRef()
		for _, branch := range wfCondition.Branches {
			if ref == branch {
				return true, nil, nil
			}
		}
		for _, branch := range wfCondition.CompiledBranches {
			if branch.MatchString(ref) {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	if hasRef, ok := payload.(HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/heads/")
		for _, branch := range wfCondition.Branches {
			if ref == branch {
				return true, nil, nil
			}
		}
		for _, branch := range wfCondition.CompiledBranches {
			if branch.MatchString(ref) {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func (handler *Handler) matchTags(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) {
	if len(wfCondition.Tags) == 0 {
		return true, nil, nil
	}
	if hasRef, ok := payload.(HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/tags/")
		for _, tag := range wfCondition.Tags {
			if ref == tag {
				return true, nil, nil
			}
		}
		for _, tag := range wfCondition.CompiledTags {
			if tag.MatchString(ref) {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func (handler *Handler) matchPaths(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) {
	if len(wfCondition.Paths) == 0 {
		return true, nil, nil
	}
	for _, changedFile := range event.ChangedFiles {
		for _, p := range wfCondition.Paths {
			if changedFile == p {
				return true, nil, nil
			}
		}
		for _, p := range wfCondition.CompiledPaths {
			if p.MatchString(changedFile) {
				return true, nil, nil
			}
		}
	}
	return false, nil, nil
}

func (handler *Handler) matchBranchesIgnore(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) { //nolint:cyclop
	if len(wfCondition.BranchesIgnore) == 0 {
		return true, nil, nil
	}
	if hasPR, ok := payload.(HasPR); ok {
		pr := hasPR.GetPullRequest()
		base := pr.GetBase()
		ref := base.GetRef()
		for _, branch := range wfCondition.Branches {
			if ref == branch {
				return false, nil, nil
			}
		}
		for _, branch := range wfCondition.CompiledBranches {
			if branch.MatchString(ref) {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	if hasRef, ok := payload.(HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/heads/")
		for _, branch := range wfCondition.Branches {
			if ref == branch {
				return false, nil, nil
			}
		}
		for _, branch := range wfCondition.CompiledBranches {
			if branch.MatchString(ref) {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	return true, nil, nil
}

func (handler *Handler) matchTagsIgnore(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) {
	if len(wfCondition.Tags) == 0 {
		return true, nil, nil
	}
	if hasRef, ok := payload.(HasRef); ok {
		ref := strings.TrimPrefix(hasRef.GetRef(), "refs/tags/")
		for _, tag := range wfCondition.Tags {
			if ref == tag {
				return false, nil, nil
			}
		}
		for _, tag := range wfCondition.CompiledTags {
			if tag.MatchString(ref) {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	return true, nil, nil
}

func (handler *Handler) matchPathsIgnore(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) {
	if len(wfCondition.PathsIgnore) == 0 {
		return true, nil, nil
	}
	for _, changedFile := range event.ChangedFiles {
		if !matchPath(changedFile, wfCondition.PathsIgnore, wfCondition.CompiledPaths) {
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

func (handler *Handler) matchIf(wfCondition *config.WorkflowCondition, payload interface{}, event *Event) (bool, *Response, error) {
	return true, nil, nil
}
