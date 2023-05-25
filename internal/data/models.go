package data

import (
	"database/sql"
)

type Models struct {
	Movie MovieModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movie: MovieModel{DB: db},
	}
}
