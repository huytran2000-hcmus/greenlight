package data

import (
	"fmt"
	"strconv"
)

type RunTime int32

func (r RunTime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quoteJSONVaue := strconv.Quote(jsonValue)

	return []byte(quoteJSONVaue), nil
}
