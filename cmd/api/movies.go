package main

import (
	"errors"
	"fmt"
	"net/http"

	"huytran2000-hcmus/greenlight/internal/data"
	"huytran2000-hcmus/greenlight/internal/validator"
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
		app.serverErrorResponse(w, r, fmt.Errorf("create movie handler: %s", err))
		return
	}

	header := http.Header{}
	header.Set("Location", fmt.Sprintf("/v1/movie/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, header, envelope{"movie": movie})
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("create movie handler: %s", err))
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movie.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("show movie handler: %s", err))
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, nil, envelope{"movie": movie})
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("show movie handler: %s", err))
		return
	}
}

func (app *application) listAllMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
		Title  string
		Genres []string
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Sort = app.readString(qs, "sort", "id")
	input.SortWhiteList = []string{
		"id",
		"title",
		"year",
		"runtime",
		"-id",
		"-title",
		"-year",
		"-runtime",
	}

	data.ValidateFilter(v, input.Filters)
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	movies, err := app.models.Movie.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("list all movies handler: %s", err))
		return
	}

	err = app.writeJSON(w, http.StatusOK, nil, envelope{"movies": movies})
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("list all movies handler: %s", err))
		return
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movie.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("update movie handler: %s", err))
		}
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.RunTime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}
	v := validator.New()
	data.ValidateMovie(v, movie)
	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movie.Update(movie)
	if err != nil {
		if errors.Is(err, data.ErrEditConflict) {
			app.editConflictResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, fmt.Errorf("update movie handler: %s", err))
		return
	}

	err = app.writeJSON(w, http.StatusOK, nil, envelope{"movie": movie})
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("update movie handler: %s", err))
	}
}

func (app *application) deleteMovieHanlder(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movie.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, fmt.Errorf("delete movie handler: %s", err))
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, nil, envelope{"message": "movie succesfully deleted"})
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("delete movie handler: %s", err))
	}
}
