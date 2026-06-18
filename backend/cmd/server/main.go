package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"yugioh-akinator-backend/internal/api"
	"yugioh-akinator-backend/internal/game"
	"yugioh-akinator-backend/internal/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gameData, err := repository.LoadGameData(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	gameService := game.NewService(gameData)
	gameHandler := api.NewGameHandler(gameService)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", api.HealthHandler)
	mux.HandleFunc("/api/game/start", gameHandler.StartGameHandler)
	mux.HandleFunc("/api/game/answer", gameHandler.AnswerQuestionHandler)
	mux.HandleFunc("/api/game/confirm", gameHandler.ConfirmAnswerHandler)

	addr := "0.0.0.0:" + port
	log.Println("server listening on", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
