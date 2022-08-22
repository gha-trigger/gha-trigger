package lambda

import (
	"strconv"
)

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64) //nolint:gomnd
}
