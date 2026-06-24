package api

import (
	"encoding/json"
	"net/http"
)

// WriteJSON はAPIレスポンスをJSONとして返す共通関数。
// 各HandlerでContent-TypeやEncode処理を重複させないために使う。
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
