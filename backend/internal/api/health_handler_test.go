package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHealthHandlerReturnsOK はGETリクエストに200とstatus okを返すことを確認する。
func TestHealthHandlerReturnsOK(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder := httptest.NewRecorder()

	HealthHandler(recorder, request)

	assertJSONResponse(t, recorder, http.StatusOK)
	var response map[string]string
	decodeJSONResponse(t, recorder, &response)
	if response["status"] != "ok" {
		t.Errorf("health status = %q; want %q", response["status"], "ok")
	}
}

// TestHealthHandlerRejectsNonGetMethod はGET以外のメソッドに405を返すことを確認する。
func TestHealthHandlerRejectsNonGetMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/health", nil)
	recorder := httptest.NewRecorder()

	HealthHandler(recorder, request)

	assertErrorResponse(t, recorder, http.StatusMethodNotAllowed, "method not allowed")
}
