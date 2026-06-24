package api

import (
	"encoding/json"
	"log"
	"net/http"
	"yugioh-akinator-backend/internal/model"
)

// StartGameHandler はゲーム開始API。
// 設定値を受け取り、Serviceから最初の質問を取得してJSONで返す。
func (h *GameHandler) StartGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("start game: invalid method=%s", r.Method)
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req model.SettingHyperParameter
	// リクエストbodyのmetaをGoの型へ変換する。
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("start game: invalid request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}
	log.Printf(
		"start game: beta=%.3f answerThreshold=%.3f topCandidatesCount=%d",
		req.Meta.Beta,
		req.Meta.AnswerThreshold,
		req.Meta.TopCandidatesCount,
	)

	// 実際のゲーム開始処理はServiceへ委譲し、HTTP層は変換とステータス管理に集中する。
	res, err := h.service.StartGame(req.Meta)
	if err != nil {
		log.Printf("start game: service error: %v", err)
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	// ログ用にnilの可能性がある質問を安全に取り出す。
	questionID := 0
	if res.Question != nil {
		questionID = res.Question.ID
	}
	log.Printf("start game: ok questionID=%d candidates=%d", questionID, len(res.Candidates))

	WriteJSON(w, http.StatusOK, res)
}
