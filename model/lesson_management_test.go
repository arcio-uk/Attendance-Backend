package model

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/utils"
	"fmt"
	"github.com/google/uuid"
	"log"
	"testing"
	"time"
)

var conf_lm config.Config
var pool_lm *utils.DatabasePool

func TestInitLmTests(t *testing.T) {
	conf_lm, err := config.LoadConfig()
	if err != nil {
		t.Log("failed to init database pool")
		t.Fail()
	}
	db, err := utils.InitDatabasePool(conf_lm)
	if err != nil {
		t.Log("failed to init database pool")
		t.Fail()
	}

	pool_lm = db
}

func TestCreateLessonBadGId(t *testing.T) {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
	newActualLesson := ActualLesson{GroupLessonId: uuid.New().String(),
		StartTime: time.Now(),
		EndTime:   time.Unix(time.Now().Unix()+4000, 0)}

	err := CreateActualLesson(newActualLesson, pool_lm)
	if err == nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetLessons(t *testing.T) {
	lessons, err := GetLessons(uuid.New().String(), pool_lm)
	if err != nil {
		t.Log("A random uuid should return no lessons and no err")
		t.Log(err)
		t.Fail()
	}

	if lessons == nil {
		t.Log("Lessons are nil")
		t.Fail()
	}
}

func TestGetLessonsWithRet(t *testing.T) {
	users, err := GetUsers(pool_lm)
	if err != nil {
		t.Log("users are nil")
		t.Fail()
	}

	count := 0
	for i := 0; i < len(users); i++ {
		if i > 20 {
			break
		}
		t.Log(fmt.Sprintf("%d/%d done...", i, len(users)))
		lessons, err := GetLessons(users[i].InternalId, pool_lm)
		t.Log(fmt.Sprintf("%s has %d lessons", users[i].Fname, len(lessons)))

		if err != nil {
			t.Log(fmt.Sprintf("%s has error getting lesson", users[i].Fname))
			t.Log(err)
			t.Fail()
		}

		if lessons == nil {
			t.Log(fmt.Sprintf("%s has error getting lesson - nil lessons", users[i].Fname))
			t.Fail()
		} else {
			count += len(lessons)
			for i := 0; i < len(lessons); i++ {
				if lessons[i].StartTime.After(lessons[i].EndTime) {
					t.Error("start time must be before end time")
					t.Error(lessons[i])
					t.Fail()
				}
			}

			// Test GetLessonsDetails
			deets, err := GetLessonsDetails(lessons, pool_lm)
			if err != nil {
				t.Log(err)
				t.Log("Cannot get lesson details")
				t.Fail()
			}

			if len(lessons) == 0 {
				continue
			}

			if len(deets.Modules) == 0 {
				t.Log("deets modules length == 0")
				t.Fail()
			}

			if len(deets.Lessons) != len(lessons) {
				t.Log("Deets are shittty")
				t.Fail()
			}

			if len(deets.GroupLessons) == 0 {
				t.Log("Deets group lessons have length 0")
				t.Fail()
			}
		}
	}

	if count == 0 {
		t.Log("NO LESSONS FOR THE TEST USERS")
		t.Fail()
	}
}
