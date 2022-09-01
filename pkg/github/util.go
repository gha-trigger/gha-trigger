package github

import (
	"errors"
)

func ignoreAcceptedError(err error) error {
	var acceptedError *AcceptedError
	if errors.As(err, &acceptedError) {
		return nil
	}
	return err
}
