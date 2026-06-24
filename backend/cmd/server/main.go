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

	// DATABASE_URL はRender/Supabaseなどの本番環境でも同じ名前で渡す接続文字列。
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// pgxのdatabase/sqlドライバでPostgreSQLへ接続する。
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 起動時にDB接続を確認し、接続できない状態でサーバーを公開しない。
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	// PORT はRenderなどのホスティング環境から渡される。ローカルでは8080を使う。
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 推理に必要なカード・質問・回答データは起動時に一度だけ読み込む。
	gameData, err := repository.LoadGameData(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	// Handler -> Service -> Engine/Repository の順に責務を分ける。
	gameResultRepository := repository.NewGameResultRepository(db)
	gameService := game.NewService(gameData)
	gameHandler := api.NewGameHandler(gameService, gameResultRepository)

	// APIのURLはフロントエンド側のgameApi.tsと対応する。
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", api.HealthHandler)
	mux.HandleFunc("/api/game/start", gameHandler.StartGameHandler)
	mux.HandleFunc("/api/game/answer", gameHandler.AnswerQuestionHandler)
	mux.HandleFunc("/api/game/confirm", gameHandler.ConfirmAnswerHandler)

	addr := "0.0.0.0:" + port
	log.Println("server listening on", addr)

	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// withCORS はVite開発サーバーからバックエンドAPIを呼べるようにするミドルウェア。
func withCORS(next http.Handler) http.Handler {
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
