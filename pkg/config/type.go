package config

import "regexp"

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
	AWS       *AWS       `yaml:"aws"`
	GitHubApp *GitHubApp `yaml:"github_app"`
	Events    []*EventConfig
}

type GitHubApp struct {
	AppID int64 `yaml:"app_id"`
}

type AWS struct {
	Region         string
	SecretsManager *SecretsManager `yaml:"secretsmanager"`
}

type SecretsManager struct {
	Region    string
	SecretID  string `yaml:"secret_id"`
	VersionID string `yaml:"version_id"`
}

type Secret struct {
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
	RepoOwner              string
	RepoName               string
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
	RepoOwner        string
	RepoName         string
	WorkflowFileName string `yaml:"workflow_file_name"`
	Ref              string
	Inputs           map[string]interface{}
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
