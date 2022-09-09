package route

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func matchTags(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil
	}
	if event.Payload.Ref != "" {
		ref := strings.TrimPrefix(event.Payload.Ref, "refs/tags/")
		for _, tag := range matchConfig.Tags {
			f, err := tag.Match(ref)
			if err != nil {
				return false, err
			}
			if f {
				return true, nil
			}
		}
		return false, nil
	}
	return false, nil
}

func matchTagsIgnore(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, error) {
	if len(matchConfig.TagsIgnore) == 0 {
		return true, nil
	}
	if event.Payload.Ref != "" {
		ref := strings.TrimPrefix(event.Payload.Ref, "refs/tags/")
		for _, tag := range matchConfig.TagsIgnore {
			f, err := tag.Match(ref)
			if err != nil {
				return false, err
			}
			if f {
				return false, nil
			}
		}
		return true, nil
	}
	return true, nil
}
