package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	rawID := params.ByName("id")
	id, err := strconv.ParseInt(rawID, 10, 64)

	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id %d", id)
	}

	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, header http.Header, data envelope) error {
	resBody, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resBody = append(resBody, '\n')

	for key, val := range header {
		w.Header()[key] = val
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(resBody)
	return nil
}
