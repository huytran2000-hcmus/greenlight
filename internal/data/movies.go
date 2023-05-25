package data

import (
	"database/sql"
	"errors"
	"fmt"
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

	err := m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
	if err != nil {
		return fmt.Errorf("data: insert a movie: %s", err)
	}

	return nil
}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	if id <= 0 {
		return nil, ErrRecordNotFound
	}

	query := `
    SELECT id, title, year, runtime, genres, created_at, version
    FROM movies
    WHERE id = $1
    `

	var movie Movie
	err := m.DB.QueryRow(query, id).Scan(&movie.ID, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.CreatedAt, &movie.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, fmt.Errorf("data: query a movie: %s", err)
	}

	return &movie, nil
}

func (m *MovieModel) Update(movie *Movie) error {
	query := `
    UPDATE movies
    SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
    WHERE id = $5 and version = $6
    returning version
    `
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}

	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		}

		return fmt.Errorf("data: update a movie: %s", err)
	}

	return nil
}

func (m *MovieModel) Delete(id int64) error {
	stmt := `
    DELETE FROM movies
    WHERE id = $1
    `

	result, err := m.DB.Exec(stmt, id)
	if err != nil {
		return fmt.Errorf("data: delete a movie: %s", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
