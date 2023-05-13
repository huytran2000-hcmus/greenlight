package main

import (
	"fmt"
	"huytran2000-hcmus/greenlight/internal/data"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Create a move\n")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "war", "romance"},
		Version:   1,
		CreatedAt: time.Time{},
	}

	err = app.writeJSON(w, http.StatusOK, nil, envelope{"movie": movie})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
