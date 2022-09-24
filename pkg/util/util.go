package util

import (
	"strconv"
)

func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64) //nolint:gomnd
}

func StrP(s string) *string {
	return &s
}

func BoolP(b bool) *bool {
	return &b
}
