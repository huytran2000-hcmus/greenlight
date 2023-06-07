package data

import (
	"database/sql"
)

type Models struct {
	Movie MovieModel
	User  UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movie: MovieModel{DB: db},
		User:  UserModel{DB: db},
	}
}
