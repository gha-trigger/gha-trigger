package route

import (
	"context"
	"strings"

	"github.com/gha-trigger/gha-trigger/pkg/config"
	"github.com/gha-trigger/gha-trigger/pkg/domain"
)

func matchTags(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil, nil
	}
	if event.Payload.Ref != "" {
		ref := strings.TrimPrefix(event.Payload.Ref, "refs/tags/")
		for _, tag := range matchConfig.Tags {
			f, err := tag.Match(ref)
			if err != nil {
				return false, nil, err
			}
			if f {
				return true, nil, nil
			}
		}
		return false, nil, nil
	}
	return false, nil, nil
}

func matchTagsIgnore(ctx context.Context, matchConfig *config.Match, event *domain.Event) (bool, *domain.Response, error) {
	if len(matchConfig.Tags) == 0 {
		return true, nil, nil
	}
	if event.Payload.Ref != "" {
		ref := strings.TrimPrefix(event.Payload.Ref, "refs/tags/")
		for _, tag := range matchConfig.TagsIgnore {
			f, err := tag.Match(ref)
			if err != nil {
				return false, nil, err
			}
			if f {
				return false, nil, nil
			}
		}
		return true, nil, nil
	}
	return true, nil, nil
}
