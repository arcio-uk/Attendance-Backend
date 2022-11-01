package routes

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testModuleId string

func init() {
	userConfig, err := config.LoadConfig()
	if err != nil {
		log.Println("Failed to load config from .env")
		log.Println(err)
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

	stmt, err = pooltmp.Database.Prepare("SELECT id FROM modules LIMIT 1;")
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()

	moduleRes, err := stmt.Query()
	if err != nil {
		log.Println(err)
	}
	defer moduleRes.Close()
	if moduleRes.Next() {
		moduleRes.Scan(&testModuleId)
	}
}

func TestCreateModule(t *testing.T) {
	t.Cleanup(TearDownModuleTest)
	jsonData := []byte(`{
		"name":        "test module",
		"external-id": "1234567890"
	}`)

	req, _ := http.NewRequest(http.MethodPost, "/module/add", bytes.NewBuffer(jsonData))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/module/add", CreateModuleHandler)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Error("Expected status code", http.StatusCreated, "but got", w.Result().StatusCode)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		if err != nil {
			t.Fatal(err)
		}
		if val, ok := body["errors"]; ok {
			t.Fatal("Endpoint returned errors", val)
		}
	}

	// Check can be parsed to module
	var body model.Module
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	// Check no fields are empty
	var moduleMap map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &moduleMap)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range moduleMap {
		if v == "" || v == nil {
			t.Error("Field", k, "in returned module is empty.")
		}
	}
}

func TearDownModuleTest() {

	queries := []string{
		"delete from module_users where module_id in (select id from modules where external_id = '1234567890' and name = 'test module');",
		"DELETE FROM modules WHERE external_id = '1234567890' and name = 'test module';",
	}

	for _, query := range queries {
		stmt, err := DatabasePool.Database.Prepare(query)
		if err != nil {
			log.Println(err)
		}
		defer stmt.Close()

		_, err = stmt.Exec()
		if err != nil {
			log.Println(err)
		}
	}
}

func TestGetModule(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/module/get", nil)
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/module/get", GetModulesHandler)
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

	// Check if can be parsed to array of modules
	var body []model.Module
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	// Check no fields are empty
	var moduleMap []map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &moduleMap)
	if err != nil {
		t.Fatal(err)
	}
	for _, module := range moduleMap {
		for k, v := range module {
			if v == "" {
				t.Error("Field", k, "in module", module, "is empty")
			}
		}
	}
}

func TestGetModuleUsers(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/module/get-users?moduleId="+testModuleId, nil)
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/module/get-users", GetModuleUsersHandler)
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

	// Check can parse to UserAttendanceRet
	var body []model.UserAttendanceRet
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	// Check no fields are empty
	var userMap []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &userMap)
	if err != nil {
		t.Fatal(err)
	}

	for _, cUser := range userMap {
		for k, v := range cUser {
			// Handle embedded structs module-group and attendance
			switch v := v.(type) {
			case string:
				if v == "" {
					t.Error("Field", k, "in user is empty")
				}
			case map[string]interface{}:
				for field, value := range v {
					if value == nil || value == "" {
						t.Error("field", field, "is empty in object field", k)
					}
				}
			}
		}
	}
}

func TestAddUserToModule(t *testing.T) {
	t.Cleanup(TearDownModuleTest)
	// Create module to add user too
	setupData := []byte(`{
		"name":        "test module",
		"external-id": "1234567890"
	}`)

	setupReq, _ := http.NewRequest(http.MethodPost, "/module/add", bytes.NewBuffer(setupData))
	setupWriter := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/module/add", CreateModuleHandler)
	router.POST("/module/add-user", AddUserToModuleHandler)
	router.ServeHTTP(setupWriter, setupReq)
	if setupWriter.Result().StatusCode != http.StatusCreated {
		t.Fatal("failed to setup test")
	}
	var setupRes map[string]string
	err := json.Unmarshal(setupWriter.Body.Bytes(), &setupRes)
	if err != nil {
		t.Fatal(err, "issue parsing setup response")
	}

	testJson := []byte(fmt.Sprintf(`{
		"user-id":   "%s",
		"module-id": "%s"
		}`, user.InternalId, setupRes["id"]))
	req, _ := http.NewRequest(http.MethodPost, "/module/add-user", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Error("Expected status code", http.StatusCreated, "but got", w.Result().StatusCode)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		if err != nil {
			t.Fatal(err)
		}
		if val, ok := body["errors"]; ok {
			t.Fatal("Endpoint returned errors", val)
		}
	}

	var body map[string]bool
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["success"] != true {
		t.Error("Expected status true but got", body["success"])
	}
}

func TestAddUsersToModule(t *testing.T) {
	t.Cleanup(TearDownModuleTest)

	// Create new module
	newModule := model.Module{
		ExternalId: "1234567890",
		Name:       "test module",
	}
	err := model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		t.Fatal(err)
	}

	testJson := []byte(fmt.Sprintf(`{
		"user-ids":   ["%s"],
		"module-id": "%s"
		}`, user.InternalId, newModule.Id))
	req, _ := http.NewRequest(http.MethodPost, "/module/add-users", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/module/add-users", AddUsersToModuleHandler)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Error("Expected status code", http.StatusCreated, "but got", w.Result().StatusCode)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		if err != nil {
			t.Fatal(err)
		}
		if val, ok := body["errors"]; ok {
			t.Fatal("Endpoint returned errors", val)
		}
	}

	var body map[string]bool
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["success"] != true {
		t.Error("Expected stats to be true but got", body["success"])
	}
}

func TestRemoveUserFromModule(t *testing.T) {
	t.Cleanup(TearDownModuleTest)

	// Create new module
	newModule := model.Module{
		ExternalId: "1234567890",
		Name:       "test module",
	}
	err := model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		t.Fatal(err)
	}

	// Add user to module
	err = model.AddUserToModule(user.InternalId, newModule.Id, DatabasePool)
	if err != nil {
		t.Fatal(err)
	}

	testJson := []byte(fmt.Sprintf(`{
		"user-id":    "%s",
		"module-id":  "%s"
	}`, user.InternalId, newModule.Id))

	req, _ := http.NewRequest(http.MethodDelete, "/module/rm-user", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.DELETE("/module/rm-user", RemoveUserFromModuleHandler)
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

	var body map[string]bool
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["success"] != true {
		t.Error("Expected status to be true but got", body["success"])
	}
}
