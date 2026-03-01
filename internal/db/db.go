package db

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dbURL string) *sql.DB {
	log.Println("⏳Connecting to Postgres database...")
	pgDB, err := sql.Open("pgx", dbURL)

	if err != nil {
		log.Fatalln("❌Connection to Postgres failed:", err)
	}

	log.Println("⏳Pinging Postgres database...")
	pingErr := pgDB.PingContext(context.Background())

	if pingErr != nil {
		log.Fatalln("❌Ping to Postgres failed:", pingErr)
	}

	log.Println("✅Connection to Postgres database established")

	return pgDB
}
