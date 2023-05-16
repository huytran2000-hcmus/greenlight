package data

import (
	"database/sql"
	"huytran2000-hcmus/greenlight/internal/validator"
	"time"

	"github.com/lib/pq"
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

type MovieModel struct {
	DB *sql.DB
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

func (m *MovieModel) Insert(movie *Movie) error {
	query := `
    INSERT INTO movies (title, year, runtime, genres)
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at, version
    `

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

func (m *MovieModel) Update(movie *Movie) error {
	return nil
}

func (m *Movie) Delete(id int64) error {
	return nil
}
