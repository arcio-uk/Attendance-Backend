package routes

import (
	"arcio/attendance-system/security"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func init() {
	var nonceManager security.NonceManager
	nonceManager.InitNonceManager()
	NonceManager = &nonceManager
}

func TestGetNonce(t *testing.T) {
	req, _ := http.NewRequest("GET", "/get-nonce", nil)
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/get-nonce", GetNonceHandler)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("Expected Status Code 200 (http.StatusOK) but got", w.Result().StatusCode)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if _, ok := body["nonce"]; !ok {
		t.Error("Expected nonce value in response, but none was found")
	}

	switch v := body["nonce"].(type) {
	case string:
		_, err := strconv.ParseInt(body["nonce"].(string), 10, 64)
		if err != nil {
			t.Errorf("Expected nonce to be a number")
		}
	default:
		t.Errorf("Expected nonce to be returned as string, got %T\n", v)
	}

}

func TestStatus(t *testing.T) {
	req, _ := http.NewRequest("GET", "/status", nil)

	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/status", StatusHandler)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("Expected Status Code 200 (http.StatusOK) but got", w.Result().StatusCode)
	}

	if w.Body.String() != "online" {
		t.Error("Expected body to return `online` got", w.Body.String())
	}
}
