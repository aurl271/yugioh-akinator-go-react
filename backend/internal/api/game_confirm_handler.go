package api

import (
	"encoding/json"
	"net/http"
	"yugioh-akinator-backend/internal/model"
)

func (h *GameHandler) ConfirmAnswerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req model.ConfirmAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	res, err := h.service.ConfirmAnswer(req)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	WriteJSON(w, http.StatusOK, res)
}
