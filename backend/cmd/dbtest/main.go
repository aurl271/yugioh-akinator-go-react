package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM cards").Scan(&count); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("cards: %d\n", count)

	var cardID int64
	var name string
	var reading sql.NullString
	if err := db.QueryRow(`
		SELECT card_id, name, reading
		FROM cards
		ORDER BY id
		LIMIT 1
	`).Scan(&cardID, &name, &reading); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("first card: %d %s", cardID, name)
	if reading.Valid {
		fmt.Printf(" (%s)", reading.String)
	}
	fmt.Println()
}
