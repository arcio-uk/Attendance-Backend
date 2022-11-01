package routes

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var user model.User

func init() {
	var nonceManager security.NonceManager
	nonceManager.InitNonceManager()
	NonceManager = &nonceManager

	userConfig, _ := config.LoadConfig()
	pooltmp, err := utils.InitDatabasePool(userConfig)

	if err != nil {
		log.Println("Failed to init database pool")
		log.Println(err)
	}

	DatabasePool = pooltmp

	userRes, err := pooltmp.Database.Query("SELECT id, external_id, firstname, " +
		"surname, email, creation_time, edit_time FROM users LIMIT 1;")
	if err != nil {
		log.Println(err)
		return
	}
	defer userRes.Close()

	if userRes.Next() {
		userRes.Scan(&user.InternalId, &user.ExternalId, &user.Fname, &user.Sname,
			&user.Email, &user.CreationTime, &user.EditTime)
	}
}

func TestLoginScreenHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/login-screen", nil)

	w := httptest.NewRecorder()

	r := gin.New()

	r.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())
	r.GET("/login-screen", LoginScreenHandler)

	r.ServeHTTP(w, req)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	if val, ok := body["errors"]; ok {
		t.Error("Endpoint returned errors", val)
	}

	for _, module := range body["modules"].([]interface{}) {
		// force go to interpret interface as map
		newModule := module.(map[string]interface{})

		// For each field in module
		for k, v := range newModule {
			if v == nil || v == "" {
				t.Error("Field", k, "in module", newModule, "is empty")
			}
		}
	}

	for _, lesson := range body["upcoming-lessons"].([]interface{}) {
		newLesson := lesson.(map[string]interface{})

		for k, v := range newLesson {
			if v == nil || v == "" {
				t.Error("Field", k, "in lesson", newLesson, "is empty")
			}
		}
	}

}
