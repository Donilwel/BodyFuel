package postgres

import "github.com/jmoiron/sqlx"

type UserParamsRepo struct {
	getter dbClientGetter
}

func NewUserParamsRepository(db *sqlx.DB) *UserParamsRepo {
	return &UserParamsRepo{getter: dbClientGetter{db: db}}
}

//TODO: доделать полностью
