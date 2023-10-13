package github

import (
	"net/http"

	"github.com/google/go-github/v56/github"
)

const (
	SHA1SignatureHeader = github.SHA1SignatureHeader
	EventTypeHeader     = github.EventTypeHeader
)

type (
	AcceptedError                      = github.AcceptedError
	CommitFile                         = github.CommitFile
	CreateWorkflowDispatchEventRequest = github.CreateWorkflowDispatchEventRequest
	Deployment                         = github.Deployment
	HeadCommit                         = github.HeadCommit
	Installation                       = github.Installation
	Issue                              = github.Issue
	IssueComment                       = github.IssueComment
	IssueCommentEvent                  = github.IssueCommentEvent
	ListOptions                        = github.ListOptions
	PullRequest                        = github.PullRequest
	PullRequestBranch                  = github.PullRequestBranch
	PullRequestTargetEvent             = github.PullRequestTargetEvent
	PullRequestEvent                   = github.PullRequestEvent
	PushEvent                          = github.PushEvent
	ReleaseEvent                       = github.ReleaseEvent
	Repository                         = github.Repository
	RepositoryCommit                   = github.RepositoryCommit
	Response                           = github.Response
	StatusEvent                        = github.StatusEvent
	User                               = github.User
	V3Client                           = github.Client
)

func ValidateSignature(signature string, payload, secretToken []byte) error {
	return github.ValidateSignature(signature, payload, secretToken)
}

func ParseWebHook(messageType string, payload []byte) (interface{}, error) {
	return github.ParseWebHook(messageType, payload)
}

func NewV3Client(client *http.Client) *V3Client {
	return github.NewClient(client)
}
