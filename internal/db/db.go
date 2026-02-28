package db

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dbURL string) *sql.DB {
	pgDB, err := sql.Open("pgx", dbURL)

	if err != nil {
		log.Fatalln("Connection to Postgres failed:", err)
	}

	pingErr := pgDB.PingContext(context.Background())

	if pingErr != nil {
		log.Fatalln("Ping to Postgres failed:", pingErr)
	}

	return pgDB
}
