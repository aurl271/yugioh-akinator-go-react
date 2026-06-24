package api

import (
	"encoding/json"
	"log"
	"net/http"
	"yugioh-akinator-backend/internal/model"
)

// ConfirmAnswerHandler は最終回答が正しかったかを受け取るAPI。
// 今はDB保存せず、フィードバック画面に必要なレスポンスを返す。
func (h *GameHandler) ConfirmAnswerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("confirm answer: invalid method=%s", r.Method)
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req model.ConfirmAnswerRequest
	// answerCardとisCorrectを含む確認リクエストを読み込む。
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("confirm answer: invalid request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}
	log.Printf(
		"confirm answer: answerCardID=%d isCorrect=%t history=%d",
		req.AnswerCard.CardID,
		req.IsCorrect,
		len(req.AnsweredQuestions),
	)

	if h.gameResultRepository != nil {
		if err := h.gameResultRepository.SaveGameResult(r.Context(), req); err != nil {
			log.Printf("save game result failed: %v", err)
		}
	}

	res, err := h.service.ConfirmAnswer(req)
	if err != nil {
		log.Printf("confirm answer: service error: %v", err)
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	if res.IsCorrect == nil {
		log.Printf("confirm answer: ok isCorrect=<nil>")
	} else {
		log.Printf("confirm answer: ok isCorrect=%t", *res.IsCorrect)
	}

	WriteJSON(w, http.StatusOK, res)
}
