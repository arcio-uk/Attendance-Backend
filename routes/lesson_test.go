package routes

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

func TearDownLessonTest() {

	queries := []string{
		"DELETE FROM repeating_lessons WHERE group_lesson_id IN (SELECT id FROM group_lessons WHERE name = 'test lesson');",
		"DELETE FROM actual_lessons WHERE group_lesson_id IN (SELECT id from group_lessons WHERE name= 'test lesson');",
		"DELETE FROM group_lessons WHERE name = 'test lesson';",
		"DELETE FROM module_groups WHERE name = 'test module group';",
		"DELETE FROM modules WHERE name = 'test module';",
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

func TestCreateRepeatingLesson(t *testing.T) {
	t.Cleanup(TearDownLessonTest)

	// DB Setup
	module := model.Module{
		Id:         uuid.NewString(),
		Name:       "test module",
		ExternalId: "test module",
	}
	if err := model.CreateModule(&module, DatabasePool); err != nil {
		t.Fatal("failed to create test module")
	}

	moduleGroup := model.ModuleGroup{
		Id:       uuid.NewString(),
		ModuleId: module.Id,
		Name:     "test module group",
	}
	if err := model.CreateModuleGroup(&moduleGroup, DatabasePool); err != nil {
		t.Fatal("failed to create test module group")
	}

	groupLesson := model.GroupLesson{
		Id:                 uuid.NewString(),
		ModuleGroupId:      moduleGroup.Id,
		AttendanceRequired: false,
		Name:               "test lesson",
		Summary:            "test lesson",
		Description:        "test lesson",
		Location:           "test lesson",
	}
	if err := model.CreateGroupLesson(&groupLesson, DatabasePool); err != nil {
		t.Fatal("failed to create test group lesson")
	}

	// Main testing
	testJson := []byte(fmt.Sprintf(`{
		"group-lesson-id": "%s",
		"start-repeating": "2022-07-05",
		"stop-repeating": "2022-07-30",
		"start-time": "11:00",
		"end-time": "15:00",
		"repeat-every": 604800
	}`, groupLesson.Id))

	req, _ := http.NewRequest(http.MethodPost, "/lesson/create-repeating", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/lesson/create-repeating", CreateRepeatingLessonHandler)
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
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal("failed to bind body to map[string]bool")
	}

	if body["success"] != true {
		t.Error("Excepted status to be true but got", body["success"])
	}
}

func TestCreateIndividualLesson(t *testing.T) {
	t.Cleanup(TearDownLessonTest)

	// DB Setup
	module := model.Module{
		Id:         uuid.NewString(),
		Name:       "test module",
		ExternalId: "test module",
	}
	if err := model.CreateModule(&module, DatabasePool); err != nil {
		t.Fatal("failed to create test module")
	}

	moduleGroup := model.ModuleGroup{
		Id:       uuid.NewString(),
		ModuleId: module.Id,
		Name:     "test module group",
	}
	if err := model.CreateModuleGroup(&moduleGroup, DatabasePool); err != nil {
		t.Fatal("failed to create test module group")
	}

	groupLesson := model.GroupLesson{
		Id:                 uuid.NewString(),
		ModuleGroupId:      moduleGroup.Id,
		AttendanceRequired: false,
		Name:               "test lesson",
		Summary:            "test lesson",
		Description:        "test lesson",
		Location:           "test lesson",
	}
	if err := model.CreateGroupLesson(&groupLesson, DatabasePool); err != nil {
		t.Fatal("failed to create test group lesson")
	}

	// Main testing
	testJson := []byte(fmt.Sprintf(`{
		"group-lesson-id": "%s",
		"start-time": "2019-10-12T07:19:50.52Z",
		"end-time": "2019-10-12T07:20:50.52Z"
	}`, groupLesson.Id))

	req, _ := http.NewRequest(http.MethodPost, "/lesson/create-one-off", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/lesson/create-one-off", CreateIndividualLessonHandler)
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
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal("failed to bind body to map[string]bool")
	}

	if body["success"] != true {
		t.Error("Excepted status to be true but got", body["success"])
	}
}
