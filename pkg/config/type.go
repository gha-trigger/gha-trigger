package config

import (
	"context"
	"errors"
	"path"
	"regexp"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/github"
)

type Config struct {
	AWS        *AWS         `yaml:"aws"`
	GitHubApps []*GitHubApp `yaml:"github_apps"`
	Repos      []*Repo
}

type Repo struct {
	RepoOwner             string `yaml:"repo_owner" validate:"required"`
	RepoName              string `yaml:"repo_name" validate:"required"`
	WorkflowGitHubAppName string `yaml:"workflow_github_app_name" validate:"required"`
	CIRepoName            string `yaml:"ci_repo_name" validate:"required"`
	Events                []*Event
	GitHub                *github.Client `yaml:"-"`
}

type AWS struct {
	Region string
}

func (a *AWS) GetRegion() string {
	if a == nil {
		return ""
	}
	return a.Region
}

type GitHubApp struct {
	Name           string
	Org            string
	User           string
	AppID          int64                  `yaml:"app_id"`
	InstallationID int64                  `json:"installation_id"`
	Secret         *GitHubAppSecretConfig `validate:"required"`
}

type GitHubAppSecretConfig struct {
	Type      string `validate:"required,oneof=aws_secretsmanager"`
	SecretID  string `yaml:"secret_id" validate:"required"`
	VersionID string `yaml:"version_id"`
}

type GitHubAppSecret struct {
	AppID               int64  `json:"app_id"`
	InstallationID      int64  `json:"installation_id"`
	WebhookSecret       string `json:"webhook_secret" validate:"required"`
	GitHubAppPrivateKey string `json:"github_app_private_key" validate:"required"`
}

type Event struct {
	// OR Condition
	Matches  []*Match
	Workflow *Workflow `validate:"required"`
}

type StringMatch struct {
	Type   string `validate:"required,oneof=equal contain regexp prefix suffix glob"`
	Value  string `validate:"required"`
	regexp *regexp.Regexp
}

var errInvalidStringType = errors.New("type is invalid")

func validateStringType(s string) error {
	switch s {
	case "equal", "contain", "regexp", "prefix", "suffix", "glob":
		return nil
	default:
		return errInvalidStringType
	}
}

func (sm *StringMatch) Match(s string) (bool, error) {
	switch sm.Type {
	case "equal":
		return sm.Value == s, nil
	case "contain":
		return strings.Contains(s, sm.Value), nil
	case "regexp":
		return sm.regexp.MatchString(s), nil
	case "prefix":
		return strings.HasPrefix(s, sm.Value), nil
	case "suffix":
		return strings.HasSuffix(s, sm.Value), nil
	case "glob":
		return path.Match(sm.Value, s)
	default:
		return false, errInvalidStringType
	}
}

func (sm *StringMatch) Compile() error {
	if sm.Type != "regexp" {
		return nil
	}
	p, err := regexp.Compile(sm.Value)
	if err != nil {
		return err
	}
	sm.regexp = p
	return nil
}

func (sm *StringMatch) Validate() error {
	return validateStringType(sm.Type)
}

type Match struct {
	// And Condition
	Events         []*EventType
	Branches       []*StringMatch
	Tags           []*StringMatch
	Paths          []*StringMatch
	BranchesIgnore []*StringMatch `yaml:"branches-ignore"`
	TagsIgnore     []*StringMatch `yaml:"tags-ignore"`
	PathsIgnore    []*StringMatch `yaml:"paths-ignore"`
	If             string
	CompiledIf     string `yaml:"-"`
}

type Workflow struct {
	WorkflowFileName string `yaml:"workflow_file_name" validate:"required"`
	Ref              string
	GitHub           GitHubWorkflowClient `yaml:"-"`
}

type GitHubWorkflowClient interface {
	RunWorkflow(ctx context.Context, owner, repo, workflowFileName string, event github.CreateWorkflowDispatchEventRequest) (*github.Response, error)
}

func compileStringsByRegexp(arr []*StringMatch) error {
	for _, s := range arr {
		if err := s.Compile(); err != nil {
			return err
		}
		if err := s.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (mc *Match) Compile() error {
	if err := compileStringsByRegexp(mc.Branches); err != nil {
		return err
	}
	if err := compileStringsByRegexp(mc.Tags); err != nil {
		return err
	}
	if err := compileStringsByRegexp(mc.Paths); err != nil {
		return err
	}
	if err := compileStringsByRegexp(mc.BranchesIgnore); err != nil {
		return err
	}
	if err := compileStringsByRegexp(mc.TagsIgnore); err != nil {
		return err
	}
	if err := compileStringsByRegexp(mc.PathsIgnore); err != nil {
		return err
	}
	return nil
}

type EventType struct {
	// https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows
	Name  string `validate:"required,oneof=branch_protection_rule check_run check_suite create delete deployment deployment_status discussion discussion_comment fork gollum issue_comment issues label merge_group milestone page_build project project_card project_column public pull_request pull_request_review pull_request_review_comment pull_request_target push registry_package release repository_dispatch schedule status watch workflow_call workflow_dispatch workflow_run"`
	Types []string
}
