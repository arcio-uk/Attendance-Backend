package middleware

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var router *gin.Engine
var permsRouter *gin.Engine
var testNonceManager security.NonceManager

func init() {
	testNonceManager.InitNonceManager()

	router = gin.New()
	router.Use(CheckPreflight())
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/check-auth", CheckAuth(&testNonceManager), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Init the perms middleware
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
	conf_at, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to init database pool")
	}

	db, err := utils.InitDatabasePool(conf_at)
	if err != nil {
		log.Fatal("Failed to init database pool")
	}

	permsRouter = gin.New()
	permsRouter.Use(CheckPermissions(security.Global, db, security.PERMS_NONE))
	permsRouter.GET("/get", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	permsRouter.POST("/post", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}

//Checking if preflight OPTIONS requests are returned correctly.
func TestCheckPreflightOptions(t *testing.T) {
	conf, err := config.LoadConfig()
	if err != nil {
		t.Fail()
		return
	}

	GlobalConfig = &conf

	req, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("Expected Status Code 200 (http.StatusOK) but got", w.Result().StatusCode)
	}
}

//Check if GET request goes through middleware.
func TestCheckPreflightGet(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("Expected Status Code 200 (http.StatusOK) but got", w.Result().StatusCode)
	}
}

// Empty request should return validation errors
func TestCheckAuthNoHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "/check-auth", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Error("Expected Status Code 400 for request with no headers, but got", w.Result().StatusCode)
	}
}

// Request with string as a nonce should return error
func TestCheckAuthWrongTypeNonce(t *testing.T) {

	// Check if non-number nonce gives bad request.
	req, _ := http.NewRequest("GET", "/check-auth", nil)
	req.Header.Add("Nonce", "ndhjsnajdsa")
	req.Header.Add("Authorization", "insert jwt here at some point")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Error("Expected Status Code 401 for request with bad nonce, but got", w.Result().StatusCode)
	}

}

// Request with an invalid numeric nonce should return unauthorized.
func TestCheckInvalidNonce(t *testing.T) {
	req, _ := http.NewRequest("GET", "/check-auth", nil)
	req.Header.Add("Nonce", "321321")
	req.Header.Add("Authorization", "insert jwt here at some point")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Error("Expected Status Unauthorized, but got", w.Result().StatusCode)
	}
}

func TestCorrectNonce(t *testing.T) {
	req, _ := http.NewRequest("GET", "/check-auth", nil)
	nonce, err := testNonceManager.GetNonce()
	if err != nil {
		t.Fatal("Error getting nonce", err)
	}

	req.Header.Add("Nonce", strconv.FormatInt(nonce, 10))
	req.Header.Add("Authorization", "insert jwt here at some point")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Error("Expected Status UNAUTHORISED, but got", w.Result().StatusCode)
	}

	// Check that the error is "invalid JWT"
	body_bytes, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Error("Canny read the thing")
	}

	body := string(body_bytes)
	if body != "{\"error\":\"Invalid JWT\"}" {
		t.Error("Expecting JWT error not this error", body)
	}
}

func TestPermsMiddleWarePollutedJsonPOST(t *testing.T) {
	req, _ := http.NewRequest("POST", "/post", strings.NewReader(`{
    "pollution": 123
}`))
	w := httptest.NewRecorder()

	permsRouter.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Error("Expected Status 500, but got", w.Result().StatusCode)
	}
}
