package api

import (
	"yugioh-akinator-backend/internal/game"
)

type GameHandler struct {
	service *game.Service
}

func NewGameHandler(service *game.Service) *GameHandler {
	return &GameHandler{
		service: service,
	}
}
