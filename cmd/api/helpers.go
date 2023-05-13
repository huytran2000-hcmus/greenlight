package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	rawID := params.ByName("id")
	id, err := strconv.ParseInt(rawID, 10, 64)

	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id %d", id)
	}

	return id, nil
}
