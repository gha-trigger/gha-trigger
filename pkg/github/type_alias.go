package github

import (
	"github.com/google/go-github/v45/github"
)

type (
	CommitFile                         = github.CommitFile
	CreateWorkflowDispatchEventRequest = github.CreateWorkflowDispatchEventRequest
	ListOptions                        = github.ListOptions
	PullRequest                        = github.PullRequest
	Repository                         = github.Repository
	Response                           = github.Response
)

func ValidateSignature(signature string, payload, secretToken []byte) error {
	return github.ValidateSignature(signature, payload, secretToken)
}

func ParseWebHook(messageType string, payload []byte) (interface{}, error) {
	return github.ParseWebHook(messageType, payload)
}
