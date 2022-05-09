package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const stringConnection = "postgres://%s:%s@%s/%s?sslmode=disable"

type DBParams struct {
	DBName   string
	Host     string
	User     string
	Password string
}

func NewInstance(params DBParams) (*sql.DB, error) {
	dsn := fmt.Sprintf(stringConnection, params.User, params.Password, params.Host, params.DBName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
