package github

import (
	"context"
	"errors"
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
	Org            string
	User           string
}

func newTransport(ctx context.Context, param *ParamNewApp) (http.RoundTripper, error) {
	if param.InstallationID != 0 {
		return ghinstallation.New(http.DefaultTransport, param.AppID, param.InstallationID, []byte(param.KeyFile))
	}
	if param.Org != "" {
		atr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, param.AppID, []byte(param.KeyFile))
		if err != nil {
			return nil, err
		}
		aClient := NewV3Client(&http.Client{Transport: atr})
		inst, _, err := aClient.Apps.FindOrganizationInstallation(ctx, param.Org)
		if err != nil {
			return nil, err
		}
		return ghinstallation.NewFromAppsTransport(atr, inst.GetID()), nil
	}
	if param.User != "" {
		atr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, param.AppID, []byte(param.KeyFile))
		if err != nil {
			return nil, err
		}
		aClient := NewV3Client(&http.Client{Transport: atr})
		inst, _, err := aClient.Apps.FindUserInstallation(ctx, param.User)
		if err != nil {
			return nil, err
		}
		return ghinstallation.NewFromAppsTransport(atr, inst.GetID()), nil
	}
	return nil, errors.New("either installation id, org, or user is required")
}

func NewApp(ctx context.Context, param *ParamNewApp) (*Client, error) {
	itr, err := newTransport(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("create a transport with private key: %w", err)
	}
	gh := NewV3Client(&http.Client{Transport: itr})
	return New(gh), nil
}
