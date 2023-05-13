package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	evlp := envelope{
		"status": "available",
		"system_info": map[string]any{
			"environment": app.cfg.env,
			"version":     version,
		},
	}
	err := app.writeJSON(w, http.StatusOK, nil, evlp)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
