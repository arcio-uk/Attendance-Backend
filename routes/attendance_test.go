package routes

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var lesson model.ActualLesson

//1. Looks for a user in the database.
//2 .Gets a group lesson the user is in.
//3. Creates an actual lesson for right now
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

	userRes, err := pooltmp.Database.Query("SELECT id, external_id, firstname, " +
		"surname, email, creation_time, edit_time FROM users LIMIT 1;")
	if err != nil {
		log.Println(err)
		return
	}
	defer userRes.Close()

	if userRes.Next() {
		err = userRes.Scan(&user.InternalId, &user.ExternalId, &user.Fname, &user.Sname,
			&user.Email, &user.CreationTime, &user.EditTime)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func CreateLesson() {
	lessonStmt, err := DatabasePool.Database.Prepare("select group_lessons.id from group_lessons " +
		"inner join module_groups on group_lessons.module_group_id = module_groups.id " +
		"inner join modules on modules.id = module_groups.module_id " +
		"inner join module_users on module_users.module_id = modules.id " +
		"inner join users on users.id = module_users.user_id " +
		"where user_id=$1;")
	if err != nil {
		log.Println(err)
		return
	}
	defer lessonStmt.Close()

	lessonRes, err := lessonStmt.Query(user.InternalId)
	if err != nil {
		log.Println(err)
		return
	}
	defer lessonRes.Close()

	var groupLesson model.GroupLesson
	if lessonRes.Next() {
		lessonRes.Scan(&groupLesson.Id)
	} else {
		log.Println("Error, cannot find a group lesson for user")
	}

	createActualStmt, err := DatabasePool.Database.Prepare("INSERT INTO actual_lessons " +
		"(id, group_lesson_id, start_time, end_time, description, " +
		" location, summary) VALUES ($1, $2, $3, $4, $5, $6, $7);")
	if err != nil {
		log.Println(err)
	}
	defer createActualStmt.Close()

	lesson.Id = uuid.New().String()
	lesson.GroupLessonId = groupLesson.Id
	lesson.StartTime = time.Now().Add(time.Hour * -5)
	lesson.EndTime = time.Now().Add(time.Hour * 5)
	lesson.CreationTime = time.Now().Add(time.Hour * -5)
	lesson.EditTime = time.Now()
	lesson.Location = "Your mum's house"
	lesson.Description = "Me teaching your mum"
	lesson.Summary = "This is a summary"

	_, err = createActualStmt.Exec(lesson.Id, lesson.GroupLessonId,
		lesson.StartTime,
		lesson.EndTime,
		lesson.Description,
		lesson.Location,
		lesson.Summary)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println(lesson)
}

func TestMarkAttendance(t *testing.T) {
	CreateLesson()
	jsonData := []byte(`{
    "lesson-id": "` + lesson.Id + `"
  }`)

	req, _ := http.NewRequest(http.MethodPost, "/attendance/mark", bytes.NewBuffer(jsonData))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/attendance/mark", PostMarkAttendance)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("Expected Status Code OK, but got", w.Result().StatusCode)
	}
}

func TestLecturerMarkAttendance(t *testing.T) {
	CreateLesson()
	testJson := []byte(fmt.Sprintf(`{
		"user-id":   "%s",
		"lesson-id": "%s"
		}`, user.InternalId, lesson.Id))

	req, _ := http.NewRequest(http.MethodPost, "/attendance/lecturer/mark", bytes.NewBuffer(testJson))
	w := httptest.NewRecorder()
	router := gin.New()

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claims", security.Claims{
				Uuid: user.InternalId,
			})
		}
	}())

	router.POST("/attendance/lecturer/mark", PostLecturerMarkAttendance)
	router.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Error("Expected Status Code Created, but got", w.Result().StatusCode)
	}
}
