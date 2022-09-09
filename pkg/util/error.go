package util

import "errors"

type WarnError struct {
	err error
}

func (err *WarnError) Error() string {
	return err.err.Error()
}

func IsWarn(err error) bool {
	var w *WarnError
	return errors.As(err, &w)
}

func WithWarn(err error) error {
	return &WarnError{err: err}
}
