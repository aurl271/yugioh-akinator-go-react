package api

import (
	"yugioh-akinator-backend/internal/game"
)

// GameHandler はHTTP層の構造体。
// リクエスト/レスポンスのJSON処理だけを担当し、ゲーム処理はServiceへ渡す。
type GameHandler struct {
	service *game.Service
}

// NewGameHandler はServiceを持つHTTPハンドラを作る。
func NewGameHandler(service *game.Service) *GameHandler {
	return &GameHandler{
		service: service,
	}
}
