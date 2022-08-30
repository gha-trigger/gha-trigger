package config

import (
	"regexp"

	"github.com/suzuki-shunsuke/gha-trigger/pkg/github"
)

/*
e.g.

events:
- matches:
  - repo_owner: suzuki-shunsuke
    repo_name: suzuki-shunsuke
    events: ["pull_request"]
    branches: ["main"]
  workflows:
  - repo_owner: suzuki-shunsuke
    repo_name: tfcmt-ci
    workflow_file_name: test.yaml
    ref: main
    inputs:
      event: foo
*/

// Repos -> Workflows -> Conditions
type Config struct {
	AWS        *AWS         `yaml:"aws"`
	GitHubApps []*GitHubApp `yaml:"github_apps"`
	Events     []*EventConfig
}

type AWS struct {
	Region string
}

type SecretsManager struct {
	SecretID  string
	VersionID string
}

type GlobalSecret struct {
	WebhookSecret string `json:"webhook_secret"`
}

type WebhookSecretConfig struct {
	SourceType string `yaml:"source_type"`
	Region     string
	SecretID   string `yaml:"secret_id"`
}

type GitHubApp struct {
	Name           string
	Org            string
	User           string
	AppID          int64 `yaml:"app_id"`
	InstallationID int64 `json:"installation_id"`
	Secret         *GitHubAppSecretConfig
}

type GitHubAppSecretConfig struct {
	Type     string
	Region   string
	SecretID string `yaml:"secret_id"`
}

type GitHubAppSecret struct {
	AppID               int64  `json:"app_id"`
	InstallationID      int64  `json:"installation_id"`
	WebhookSecret       string `json:"webhook_secret"`
	GitHubAppPrivateKey string `json:"github_app_private_key"`
}

type EventConfig struct {
	// OR Condition
	Matches   []*MatchConfig
	Workflows []*WorkflowConfig
}

type MatchConfig struct {
	// And Condition
	RepoOwner              string `yaml:"repo_owner"`
	RepoName               string `yaml:"repo_name"`
	Events                 []*Event
	Branches               []string
	Tags                   []string
	Paths                  []string
	BranchesIgnore         []string `yaml:"branches-ignore"`
	TagsIgnore             []string `yaml:"tags-ignore"`
	PathsIgnore            []string `yaml:"paths-ignore"`
	If                     string
	CompiledBranches       []*regexp.Regexp `yaml:"-"`
	CompiledTags           []*regexp.Regexp `yaml:"-"`
	CompiledPaths          []*regexp.Regexp `yaml:"-"`
	CompiledBranchesIgnore []*regexp.Regexp `yaml:"-"`
	CompiledTagsIgnore     []*regexp.Regexp `yaml:"-"`
	CompiledPathsIgnore    []*regexp.Regexp `yaml:"-"`
	CompiledIf             string           `yaml:"-"`
}

type WorkflowConfig struct {
	RepoOwner        string `yaml:"repo_owner"`
	RepoName         string `yaml:"repo_name"`
	WorkflowFileName string `yaml:"workflow_file_name"`
	Ref              string
	Inputs           map[string]interface{}
	GitHubAppName    string         `yaml:"github_app_name"`
	GitHub           *github.Client `yaml:"-"`
}

func compileStringsByRegexp(arr []string) []*regexp.Regexp {
	ret := make([]*regexp.Regexp, 0, len(arr))
	for _, s := range arr {
		p, err := regexp.Compile(s)
		if err != nil {
			continue
		}
		ret = append(ret, p)
	}
	return ret
}

func (mc *MatchConfig) Compile() error {
	mc.CompiledBranches = compileStringsByRegexp(mc.Branches)
	mc.CompiledTags = compileStringsByRegexp(mc.Tags)
	mc.CompiledPaths = compileStringsByRegexp(mc.Paths)
	mc.CompiledBranchesIgnore = compileStringsByRegexp(mc.BranchesIgnore)
	mc.CompiledTagsIgnore = compileStringsByRegexp(mc.TagsIgnore)
	mc.CompiledPathsIgnore = compileStringsByRegexp(mc.PathsIgnore)
	return nil
}

type Event struct {
	Name  string
	Types []string
}
