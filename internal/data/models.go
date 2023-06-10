package data

import (
	"database/sql"
)

type Models struct {
	Movie MovieModel
	User  UserModel
	Token TokenModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movie: MovieModel{DB: db},
		User:  UserModel{DB: db},
		Token: TokenModel{DB: db},
	}
}
