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

func init() {

	userConfig, err := config.LoadConfig()

	if err != nil {
		log.Println("Failed to load config from env")
		log.Println(err)
		return
	}

	pooltmp, err := utils.InitDatabasePool(userConfig)

	if err != nil {
		log.Println("Failed to init database pool")
		log.Println(err)
		return
	}

	DatabasePool = pooltmp

	stmt, err := pooltmp.Database.Prepare("SELECT id, external_id, firstname, " +
		"surname, email, creation_time, edit_time FROM users LIMIT 1;")
	if err != nil {
		log.Println(err)
		return
	}
	defer stmt.Close()

	userRes, err := stmt.Query()
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

func TestGetUsers(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/user/get", nil)
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/user/get", GetUsersHandler)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("Expected status code", http.StatusOK, "but got", w.Result().StatusCode)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		if err != nil {
			t.Fatal(err)
		}
		if val, ok := body["errors"]; ok {
			t.Fatal("Endpoint returned errors", val)
		}
	}

	var body []model.User
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	// Check no fields are empty in users
	var userMap []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &userMap)
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range userMap {
		for k, v := range user {
			if v == nil || v == "" {
				t.Error("Field", k, "in user is empty")
			}
		}
	}
}

func TestGetUserModules(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/user/get-modules", nil)
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.GET("/user/get-modules", GetUserModulesHandler)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("Expected status code", http.StatusOK, "but got", w.Result().StatusCode)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		if err != nil {
			t.Fatal(err)
		}
		if val, ok := body["errors"]; ok {
			t.Fatal("Endpoint returned errors", val)
		}
	}

	// Check can parse to module
	var body []model.Module
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	// Check no fields are empty
	var moduleMap []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &moduleMap)
	if err != nil {
		t.Fatal(err)
	}

	for _, module := range moduleMap {
		for k, v := range module {
			if v == nil || v == "" {
				t.Error("Field", k, "is empty in module", module)
			}
		}
	}
}

func TestGetUserAttendance(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/user/module-percentage", nil)
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.GET("/user/module-percentage", GetUserAttendance)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("Expected status code", http.StatusOK, "but got", w.Result().StatusCode)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		if err != nil {
			t.Fatal(err)
		}
		if val, ok := body["errors"]; ok {
			t.Fatal("Endpoint returned errors", val)
		}
	}

	// Check can parse to AttendanceRet
	var body model.AttendanceRet
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	// Check no fields are empty
	var attendanceMap map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &attendanceMap)
	if err != nil {
		t.Fatal(err)
	}

	// Force go to treat as map
	moduleAttendance := attendanceMap["module-attendance"].(map[string]interface{})

	for _, att := range moduleAttendance {
		// Force got to treat as map
		newAtt := att.(map[string]interface{})
		for k, v := range newAtt {
			if v == nil || v == "" {
				t.Error("Field", k, "is empty in attendance", att)
			}
		}
	}
}
