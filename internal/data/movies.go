package data

import (
	"huytran2000-hcmus/greenlight/internal/validator"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   RunTime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
	CreatedAt time.Time `json:"-"`
}

func ValidateMovie(v *validator.Validator, m *Movie) {
	v.CheckError(validator.NotBlank(m.Title), "title", "must be provided")
	v.CheckError(validator.LengthLessOrEqual(m.Title, 500), "title", "must not be greater than 500 characters")

	v.CheckError(m.Year != 0, "year", "must be provided")
	v.CheckError(m.Year >= 1888, "year", "must be equal or greater than 1888")
	v.CheckError(m.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.CheckError(m.Runtime != 0, "runtime", "must be provided")
	v.CheckError(m.Runtime > 0, "runtime", "must be a positive integer")

	v.CheckError(m.Genres != nil, "genres", "must be provided")
	v.CheckError(len(m.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.CheckError(len(m.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.CheckError(validator.Unique(m.Genres), "genres", "must not contain duplicate values")
}
