package config

import "regexp"

/*
e.g.

repos:
- event_repo: suzuki-shunsuke/tfcmt
  workflow_repo: suzuki-shunsuke/tfcmt-ci
  workflows:
    - name: test
      repo: suzuki-shunsuke/tfcmt-ci
      conditions:
      - events: ["pull_request"]
        branches: ["main"]
*/

// Repos -> Workflows -> Conditions
type Config struct {
	Repos []*RepoConfig
}

type RepoConfig struct {
	Name      string
	Workflows []*WorkflowConfig
	Repo      string
}

type WorkflowConfig struct {
	FileName string `yaml:"file_name"`
	// OR Condition
	Conditions []*WorkflowCondition
}

type WorkflowCondition struct {
	// And Condition
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

func (wfc *WorkflowCondition) Compile() error {
	wfc.CompiledBranches = compileStringsByRegexp(wfc.Branches)
	wfc.CompiledTags = compileStringsByRegexp(wfc.Tags)
	wfc.CompiledPaths = compileStringsByRegexp(wfc.Paths)
	wfc.CompiledBranchesIgnore = compileStringsByRegexp(wfc.BranchesIgnore)
	wfc.CompiledTagsIgnore = compileStringsByRegexp(wfc.TagsIgnore)
	wfc.CompiledPathsIgnore = compileStringsByRegexp(wfc.PathsIgnore)
	return nil
}

type Event struct {
	Name  string
	Types []string
}
