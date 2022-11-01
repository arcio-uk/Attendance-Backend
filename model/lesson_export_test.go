package model

import (
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
)

const LESSON_EXPORT_TEST_COUNT = 100

func TestLessonExport(t *testing.T) {
	lessons := make([]ActualLesson, LESSON_EXPORT_TEST_COUNT)
	for i := 0; i < LESSON_EXPORT_TEST_COUNT; i++ {
		lessons[i] = ActualLesson{Id: uuid.New().String(),
			GroupLessonId: uuid.New().String(),
			StartTime:     time.Now(),
			EndTime:       time.Now(),
			CreationTime:  time.Now(),
			EditTime:      time.Now(),
			Summary:       fmt.Sprintf("Beans on toast %d", i),
			Description:   fmt.Sprintf("E: %d", i),
			Location:      fmt.Sprintf("Windsor Building")}
	}

	output := ExportLessonsAsIcal(lessons)

	if output == "" {
		t.Log("output was nil or \"\"")
		t.Fail()
	}
}

func BenchmarkAllUserLessonExports(t *testing.B) {
	users, err := GetUsers(pool_lm)
	if err != nil {
		t.Log("users are nil")
		t.Fail()
	}

	count := 0
	for i := 0; i < len(users); i++ {
		lessons, err := GetLessons(users[i].InternalId, pool_lm)

		if err != nil {
			t.Log(fmt.Sprintf("%s has error getting lesson", users[i].Fname))
			t.Log(err)
			t.Fail()
		}

		if lessons == nil {
			t.Log(fmt.Sprintf("%s has error getting lesson - nil lessons", users[i].Fname))
			t.Fail()
		} else if len(lessons) == 0 {
			t.Fail()
		} else {
			count += len(lessons)

			ExportLessonsAsIcal(lessons)
		}
	}

	if count == 0 {
		t.Log("NO LESSONS FOR THE TEST USERS")
		t.Fail()
	}
}
