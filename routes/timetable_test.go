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
		log.Println("Failed to load config from .env")
		log.Println(err)
	}
	GlobalConfig = &userConfig

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

func TestCalenderExportHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/timetable/ical", nil)
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.GET("/timetable/ical", CalenderExportHandler)
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

	// Body is just a string so isn't being checked
}

func TestCalenderJwt(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/timetable/get-timetable-jwt", nil)
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid:      user.InternalId,
				Firstname: user.Fname,
				Surname:   user.Sname,
			})
		}
	}())

	router.GET("/timetable/get-timetable-jwt", CalenderJwtHandler)
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

	token := w.Body.String()
	_, err := security.CheckIcalJwt(token, *GlobalConfig)
	if err != nil {
		t.Error(err)
		t.Error("Returned invalid jwt token")
	}
}

func TestUpcomingLessons(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/timetable/upcoming-lessons", nil)
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid:      user.InternalId,
				Firstname: user.Fname,
				Surname:   user.Sname,
			})
		}
	}())

	router.GET("/timetable/upcoming-lessons", UpcomingLessonsHandler)
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
}

func TestGetActiveLessons(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/timetable/happening-now", nil)
	w := httptest.NewRecorder()
	router := gin.New()

	router.GET("/timetable/happening-now", GetActiveLessonsHandler)
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

	// Check can parse to lessons
	var body []model.ActualLesson
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		t.Fatal("failed to parse response to lesson")
		t.Fatal(err)
	}

	// Check no field are empty
	var lessonMap []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &lessonMap)
	if err != nil {
		t.Fatal(err)
	}

	for _, lesson := range lessonMap {
		for k, v := range lesson {
			if v == nil || v == "" {
				t.Error("Field", k, "is empty in lesson")
			}
		}
	}
}
