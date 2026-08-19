package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWriteJSONWritesStatusContentTypeAndBody はWriteJSONがstatus、Content-Type、JSON本文を設定することを確認する。
func TestWriteJSONWritesStatusContentTypeAndBody(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteJSON(recorder, http.StatusCreated, map[string]string{"status": "created"})

	assertJSONResponse(t, recorder, http.StatusCreated)
	var response map[string]string
	decodeJSONResponse(t, recorder, &response)
	if response["status"] != "created" {
		t.Errorf("response status = %q; want %q", response["status"], "created")
	}
}
