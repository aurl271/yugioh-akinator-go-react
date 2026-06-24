package api

import (
	"net/http"
)

// HealthHandler はバックエンドの起動確認用エンドポイント。
// デプロイ先やローカル疎通確認からGET /api/healthで呼ばれる。
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
