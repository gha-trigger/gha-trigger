package github

import (
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
)

type Client struct {
	pr     PullRequestsService
	action ActionsService
}

func New(gh *V3Client) *Client {
	return &Client{
		pr:     gh.PullRequests,
		action: gh.Actions,
	}
}

type HasInstallation interface {
	GetInstallation() *Installation
}

type ParamNewApp struct {
	AppID          int64
	KeyFile        string
	InstallationID int64
}

func NewApp(param *ParamNewApp) (*Client, error) {
	itr, err := ghinstallation.New(http.DefaultTransport, param.AppID, param.InstallationID, []byte(param.KeyFile))
	if err != nil {
		return nil, fmt.Errorf("create a transport with private key: %w", err)
	}
	gh := NewV3Client(&http.Client{Transport: itr})
	return New(gh), nil
}
