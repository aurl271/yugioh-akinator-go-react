package api

import (
	"encoding/json"
	"log"
	"net/http"
	"yugioh-akinator-backend/internal/model"
)

// AnswerQuestionHandler は質問への回答API。
// frontendが保持する回答履歴を受け取り、次の質問または回答候補を返す。
func (h *GameHandler) AnswerQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("answer question: invalid method=%s", r.Method)
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	var req model.GameRequest
	// 回答履歴と推理設定をリクエストbodyから復元する。
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("answer question: invalid request body: %v", err)
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}
	// ログでは履歴数と最後の質問IDを出し、ユーザー操作の流れを追いやすくする。
	lastQuestionID := 0
	if len(req.AnsweredQuestions) > 0 {
		lastQuestionID = req.AnsweredQuestions[len(req.AnsweredQuestions)-1].Question.ID
	}
	log.Printf(
		"answer question: history=%d lastQuestionID=%d beta=%.3f answerThreshold=%.3f topCandidatesCount=%d",
		len(req.AnsweredQuestions),
		lastQuestionID,
		req.Meta.Beta,
		req.Meta.AnswerThreshold,
		req.Meta.TopCandidatesCount,
	)

	// Serviceは履歴からEngineを再構築して次状態を計算する。
	res, err := h.service.AnswerQuestion(req)
	if err != nil {
		log.Printf("answer question: service error: %v", err)
		WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	// 回答提示前はAnswerCardがnilなので、ログ用の値を分けて扱う。
	answerCardID := int64(0)
	if res.AnswerCard != nil {
		answerCardID = res.AnswerCard.CardID
	}
	log.Printf(
		"answer question: ok isAnswer=%t nextQuestionNil=%t answerCardID=%d candidates=%d",
		res.IsAnswer,
		res.Question == nil,
		answerCardID,
		len(res.Candidates),
	)

	WriteJSON(w, http.StatusOK, res)
}
