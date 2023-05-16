package main

import (
	"fmt"
	"huytran2000-hcmus/greenlight/internal/data"
	"huytran2000-hcmus/greenlight/internal/validator"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.RunTime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	data.ValidateMovie(v, movie)
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movie.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	header := http.Header{}
	header.Set("Location", fmt.Sprintf("/v1/movie/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, header, envelope{"movie": movie})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
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
		return
	}
}
