package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type RunTime int32

func (r RunTime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quoteJSONVaue := strconv.Quote(jsonValue)

	return []byte(quoteJSONVaue), nil
}

func (r *RunTime) UnmarshalJSON(b []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(b))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	parts := strings.Fields(unquotedJSONValue)
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i64, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = RunTime(i64)
	return nil
}
