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

func init() {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
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

func TearDownModuleGroupTest() {

	queries := []string{
		"DELETE FROM module_user_groups WHERE module_group_id in (SELECT id FROM module_groups WHERE name = 'test module group');",
		"DELETE FROM module_users WHERE module_id in (SELECT id FROM modules WHERE external_id = '1234567890' AND name = 'test module');",
		"DELETE FROM module_groups WHERE module_id in (SELECT id FROM modules WHERE external_id = '1234567890' AND name = 'test module');",
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

func TestCreateModuleGroup(t *testing.T) {
	t.Cleanup(TearDownModuleGroupTest)

	// Create module
	newModule := model.Module{
		ExternalId: "1234567890",
		Name:       "test module",
	}
	err := model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create test module")
	}

	// Create module group

	testJson := []byte(fmt.Sprintf(`{
		"name":    "test module group",
		"module-id":  "%s"
	}`, newModule.Id))

	req, _ := http.NewRequest(http.MethodPost, "/module/group/add", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/module/group/add", CreateModuleGroupHandler)
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

	// Check can parse to module group
	var body model.ModuleGroup
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	// Check no fields are empty
	var moduleGroupMap map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &moduleGroupMap)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range moduleGroupMap {
		if v == "" {
			t.Error("Field", k, "is empty in module group")
		}
	}
}

func TestAddUserToModuleGroup(t *testing.T) {
	t.Cleanup(TearDownModuleGroupTest)

	// Create module
	newModule := model.Module{
		ExternalId: "1234567890",
		Name:       "test module",
	}
	err := model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create test module")
	}

	// Create Module Group
	newGroup := model.ModuleGroup{
		Name:     "test module group",
		ModuleId: newModule.Id,
	}
	err = model.CreateModuleGroup(&newGroup, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create module group")
	}

	// Add user to module
	err = model.AddUserToModule(user.InternalId, newModule.Id, DatabasePool)
	if err != nil {
		t.Fatal("Failed to add user to module")
	}

	// Add user to module group
	testJson := []byte(fmt.Sprintf(`{
		"module-group-id":    "%s",
		"user-id":            "%s"
	}`, newGroup.Id, user.InternalId))

	req, _ := http.NewRequest(http.MethodPost, "/module/group/add-user", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/module/group/add-user", AddUserToModuleGroupHandler)
	router.ServeHTTP(w, req)

	var body map[string]bool
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["success"] != true {
		t.Error("Expected status to be true but got", body["success"])
	}
}

func TestGetGroupsForModule(t *testing.T) {
	t.Cleanup(TearDownModuleGroupTest)

	// Create module
	newModule := model.Module{
		ExternalId: "1234567890",
		Name:       "test module",
	}
	err := model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create test module")
	}

	// Create Module Group
	newGroup := model.ModuleGroup{
		Name:     "test module group",
		ModuleId: newModule.Id,
	}
	err = model.CreateModuleGroup(&newGroup, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create module group")
	}
	req, _ := http.NewRequest(http.MethodGet, "/module/group/get?moduleId="+newModule.Id, nil)
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/module/group/get", GetGroupsForModuleHandler)
	router.ServeHTTP(w, req)

	// Check can parse to array of modules groups
	var body []model.ModuleGroup
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	// Check no fields are empty
	var groupMap []map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &groupMap)
	if err != nil {
		t.Fatal(err)
	}

	for _, group := range groupMap {
		for k, v := range group {
			if v == "" {
				t.Error("Field", k, "empty in module group.")
			}
		}
	}
}

func TestAddUsersToModuleGroup(t *testing.T) {
	t.Cleanup(TearDownModuleGroupTest)

	// Create module
	newModule := model.Module{
		ExternalId: "1234567890",
		Name:       "test module",
	}
	err := model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create test module")
	}

	// Create Module Group
	newGroup := model.ModuleGroup{
		Name:     "test module group",
		ModuleId: newModule.Id,
	}
	err = model.CreateModuleGroup(&newGroup, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create module group")
	}

	// Add user to module
	err = model.AddUserToModule(user.InternalId, newModule.Id, DatabasePool)
	if err != nil {
		t.Fatal("Failed to add user to module")
	}

	// Add user to module group
	testJson := []byte(fmt.Sprintf(`{
		"module-group-id":     "%s",
		"user-ids":            ["%s"]
	}`, newGroup.Id, user.InternalId))

	req, _ := http.NewRequest(http.MethodPost, "/module/group/add-users", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/module/group/add-users", AddUsersToModuleGroupHandler)
	router.ServeHTTP(w, req)

	var body map[string]bool
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["success"] != true {
		t.Error("Expected status to be true but got", body["success"])
	}
}

func TestRemoveFromModuleGroup(t *testing.T) {
	t.Cleanup(TearDownModuleGroupTest)

	// Create module
	newModule := model.Module{
		ExternalId: "1234567890",
		Name:       "test module",
	}
	err := model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create test module")
	}

	// Create Module Group
	newGroup := model.ModuleGroup{
		Name:     "test module group",
		ModuleId: newModule.Id,
	}
	err = model.CreateModuleGroup(&newGroup, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create module group")
	}

	// Add user to module
	err = model.AddUserToModule(user.InternalId, newModule.Id, DatabasePool)
	if err != nil {
		t.Fatal("Failed to add user to module")
	}

	// Add user to module group
	err = model.AddUserToModuleGroup(user.InternalId, newGroup.Id, DatabasePool)
	if err != nil {
		t.Fatal("Failed to add user to module group")
	}

	// Remove user from module group
	testJson := []byte(fmt.Sprintf(`{
			"module-group-id":    "%s",
			"user-id":            "%s"
		}`, newGroup.Id, user.InternalId))

	req, _ := http.NewRequest(http.MethodDelete, "/module/group/rm-user", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.DELETE("/module/group/rm-user", RemoveFromModuleGroupHandler)
	router.ServeHTTP(w, req)

	var body map[string]bool
	err = json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["success"] != true {
		t.Error("Expected status to be true but got", body["success"])
	}
}

func TestGetModuleGroupStudents(t *testing.T) {
	t.Cleanup(TearDownModuleGroupTest)

	// Create module
	newModule := model.Module{
		ExternalId: "1234567890",
		Name:       "test module",
	}
	err := model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create test module")
	}

	// Create Module Group
	newGroup := model.ModuleGroup{
		Name:     "test module group",
		ModuleId: newModule.Id,
	}
	err = model.CreateModuleGroup(&newGroup, DatabasePool)
	if err != nil {
		t.Fatal("Failed to create module group")
	}

	// Add user to module
	err = model.AddUserToModule(user.InternalId, newModule.Id, DatabasePool)
	if err != nil {
		t.Fatal("Failed to add user to module")
	}

	// Add user to module group
	err = model.AddUserToModuleGroup(user.InternalId, newGroup.Id, DatabasePool)
	if err != nil {
		t.Fatal("Failed to add user to module group")
	}

	req, _ := http.NewRequest(http.MethodGet, "/module/group/users?moduleGroupId="+newGroup.Id, nil)
	w := httptest.NewRecorder()
	router := gin.New()
	router.GET("/module/group/users", GetModuleGroupUsersHandler)
	router.ServeHTTP(w, req)

	// Check can parse to user
	var body model.ModuleGroupRet
	log.Println("Testinggg")
	err = json.Unmarshal(w.Body.Bytes(), &body)
	log.Println(w.Body.String())
	log.Println(body)
	if err != nil {
		t.Fatal(err)
	}

	if len(body.UserAttendance) == 0 {
		t.Error("No users")
	}

	if body.ModuleName == "" {
		t.Error("No module name")
	}

	if body.ModuleGroupId == "" {
		t.Error("No id")
	}
}
