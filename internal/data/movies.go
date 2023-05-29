package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"huytran2000-hcmus/greenlight/internal/validator"
)

var standardTimeout = 3 * time.Second

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
	v.CheckError(
		validator.LengthLessOrEqual(m.Title, 500),
		"title",
		"must not be greater than 500 characters",
	)

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
	ctx, cancel := context.WithTimeout(context.Background(), standardTimeout)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).
		Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
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
	ctx, cancel := context.WithTimeout(context.Background(), standardTimeout)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.CreatedAt,
		&movie.Version)
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
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), standardTimeout)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
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

	ctx, cancel := context.WithTimeout(context.Background(), standardTimeout)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, stmt, id)
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

func (m *MovieModel) GetAll(title string, genres []string, filter Filters) ([]Movie, Metadata, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) OVER(), id, title, year, runtime, genres, version
    FROM movies
    WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
    AND (genres @> $2 OR $2 = '{}')
    ORDER BY %s %s, id ASC
    LIMIT $3 OFFSET $4`, filter.sortColumn(), filter.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), standardTimeout)
	defer cancel()
	row, err := m.DB.QueryContext(
		ctx,
		query,
		title,
		pq.Array(genres),
		filter.limit(),
		filter.offset(),
	)
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("query all movie: %s", err)
	}
	defer row.Close()

	totalRecords := 0
	movies := []Movie{}
	for row.Next() {
		var mv Movie
		err = row.Scan(
			&totalRecords,
			&mv.ID,
			&mv.Title,
			&mv.Year,
			&mv.Runtime,
			pq.Array(&mv.Genres),
			&mv.Version,
		)
		if err != nil {
			return nil, Metadata{}, fmt.Errorf("scan a movie: %s", err)
		}

		movies = append(movies, mv)

	}

	err = row.Err()
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("iterate all movie: %s", err)
	}

	return movies, makeMetadata(totalRecords, filter.Page, filter.PageSize), nil
}
