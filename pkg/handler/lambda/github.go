package lambda

import (
	"github.com/suzuki-shunsuke/gha-dispatcher/pkg/github"
)

func (handler *Handler) setGitHub(param *github.ParamNewApp) error {
	if handler.gh != nil {
		return nil
	}
	return handler.refreshGitHub(param)
}

func (handler *Handler) refreshGitHub(param *github.ParamNewApp) error {
	gh, err := github.NewApp(param)
	if err != nil {
		return err
	}
	handler.gh = gh
	return nil
}
