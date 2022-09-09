package route

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func matchIf(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, error) {
	// TODO
	return true, nil
}
