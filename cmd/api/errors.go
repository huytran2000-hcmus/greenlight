package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.Print(err)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	evlp := envelope{
		"error": message,
	}

	err := app.writeJSON(w, status, nil, evlp)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	app.errorResponse(w, r, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func (app *application) methodNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The method %s is not supported by this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}
