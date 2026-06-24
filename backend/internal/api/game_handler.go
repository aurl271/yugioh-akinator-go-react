package api

import (
	"yugioh-akinator-backend/internal/game"
	"yugioh-akinator-backend/internal/repository"
)

// GameHandler はHTTP層の構造体。
// リクエスト/レスポンスのJSON処理だけを担当し、ゲーム処理はServiceへ渡す。
type GameHandler struct {
	service              *game.Service
	gameResultRepository *repository.GameResultRepository
}

// NewGameHandler はServiceを持つHTTPハンドラを作る。
func NewGameHandler(service *game.Service, gameResultRepository *repository.GameResultRepository) *GameHandler {
	return &GameHandler{
		service:              service,
		gameResultRepository: gameResultRepository,
	}
}
