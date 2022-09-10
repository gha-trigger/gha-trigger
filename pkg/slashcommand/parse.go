package slashcommand

import (
	"fmt"

	"github.com/gha-trigger/gha-trigger/pkg/util"
)

func parseIDs(words []string) ([]int64, error) {
	ids := make([]int64, len(words))
	for i, idS := range words {
		id, err := util.ParseInt64(idS)
		if err != nil {
			return nil, fmt.Errorf("id must be int64: %w", err)
		}
		if id <= 0 {
			return nil, fmt.Errorf("id must be a positive number: %w", err)
		}
		ids[i] = id
	}
	return ids, nil
}
