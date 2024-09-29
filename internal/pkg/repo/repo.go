package repo

import "github.com/jackc/pgtype/pgxtype"

type Repo struct {
	db pgxtype.Querier
}

func NewRepo(db pgxtype.Querier) *Repo {
	return &Repo{
		db: db,
	}
}
