package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

/*
type Course struct {
	Id           string
	Name         string
	Teacher      string
	CreationTime time.Time
	EditTime     time.Time
}

// Group
type CourseGroup struct {
	Id           string
	CourseId     string
	CreationTime time.Time
	EditTime     time.Time
}
*/

func TestUserAsJson(t *testing.T) {
	user := User{InternalId: uuid.New().String(),
		ExternalId:   uuid.New().String(),
		Fname:        "John",
		Sname:        "Acosta",
		Email:        "test@gmail.com",
		CreationTime: time.Now(),
		EditTime:     time.Now()}

	jsonout, err := json.Marshal(user)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Logf("User:\t%s\n", string(jsonout))
}

func TestGroupLessonAsJson(t *testing.T) {
	lessons := GroupLesson{Id: uuid.New().String(),
		ModuleGroupId: uuid.New().String(),
		CreationTime:  time.Now(),
		EditTime:      time.Now()}

	jsonout, err := json.Marshal(lessons)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Logf("Lessons:\t%s\n", string(jsonout))
}
