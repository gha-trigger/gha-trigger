package route

import (
	"context"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func matchEventType(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, error) {
	if len(matchConfig.Events) == 0 {
		return true, nil
	}
	for _, ev := range matchConfig.Events {
		if ev.Name != event.Type {
			continue
		}
		if ev.Name == "push" && event.Payload.Deleted {
			// https://github.com/gha-trigger/gha-trigger/issues/107
			// Ignore a push event when a branch or tag is deleted.
			return false, nil
		}
		if len(ev.Types) == 0 {
			return true, nil
		}
		for _, typ := range ev.Types {
			if typ == event.Payload.Action {
				return true, nil
			}
		}
	}
	return false, nil
}
