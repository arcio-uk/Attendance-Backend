package model

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/utils"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Global test vars
var confLdTest config.Config
var poolLdTest *utils.DatabasePool

func TestInitRepeatingLesson(t *testing.T) {
	confLdTest, err := config.LoadConfig()
	if err != nil {
		t.Log("Failed to init the pool")
		t.Log(err)
		t.Fail()
	}

	pooltmp, err := utils.InitDatabasePool(confLdTest)
	if err != nil {
		t.Log("Failed to init the pool")
		t.Log(err)
		t.Fail()
	}

	poolLdTest = pooltmp
}

func TestGetDaemonRepeatingLessons(t *testing.T) {
	repeatingLessons, err := GetDaemonRepeatingLessons(poolLdTest)

	if err != nil {
		t.Log("Cannot get repeating lessons")
		t.Log(err)
		t.Fail()
	}

	if repeatingLessons == nil {
		t.Log("Repeating Lessons are nil :(")
		t.Log(err)
		t.Fail()
	}
}

func TestCheckLessonNotHappening(t *testing.T) {
	//Started repeating in 1970, seems legit.
	var startRepeating int64 = 0
	var stopRepeating int64 = 60 * 60 * 24 * 7

	var startTime int64 = 7 * 60 * 60
	var endTime int64 = 19 * 60 * 60

	var interval int64 = 60 * 60 * 24

	happeningNow := RepeatingLesson{Id: uuid.New().String(),
		GroupLessonId:   uuid.New().String(),
		StartRepeating:  time.Unix(startRepeating, 0),
		StopRepeating:   time.Unix(stopRepeating, 0),
		StartTime:       time.Unix(startTime, 0),
		EndTime:         time.Unix(endTime, 0),
		CreationTime:    time.Now(),
		EditTime:        time.Now(),
		LastSpawnedTime: time.Now(),
		RepeatEvery:     time.Duration(interval)}

	_, err := CheckLessonHappening(happeningNow)

	if err == nil {
		t.Log("The lesson which is not happening is apparently happening - blame danny")
		t.Fail()
	}
}

func TestCheckLessonHappening(t *testing.T) {
	//Started repeating in 1970, seems legit.
	var startRepeating int64 = 0
	var stopRepeating int64 = 60 * 60 * 24 * 7

	var startTime int64 = 7 * 60 * 60
	var endTime int64 = 19 * 60 * 60

	var interval int64 = 60 * 60 * 24

	happeningNow := RepeatingLesson{Id: uuid.New().String(),
		GroupLessonId:   uuid.New().String(),
		StartRepeating:  time.Unix(startRepeating, 0),
		StopRepeating:   time.Unix(stopRepeating, 0),
		StartTime:       time.Unix(startTime, 0),
		EndTime:         time.Unix(endTime, 0),
		CreationTime:    time.Now(),
		EditTime:        time.Now(),
		LastSpawnedTime: time.Now(),
		RepeatEvery:     time.Duration(interval)}

	simulatedCurrentTime := time.Unix(startTime, 0)

	_, err := __CheckLessonHappening(simulatedCurrentTime, happeningNow)

	if err != nil {
		t.Log("The lesson which is happening is apparently not happening - blame danny")
		t.Fail()
	}
}

func TestCheckLessonNotHappeningWithinRepeating(t *testing.T) {
	//Started repeating in 1970, seems legit.
	var startRepeating int64 = 0
	var stopRepeating int64 = 60 * 60 * 24 * 7

	var startTime int64 = 7 * 60 * 60
	var endTime int64 = 19 * 60 * 60

	var interval int64 = 60 * 60 * 24

	happeningNow := RepeatingLesson{Id: uuid.New().String(),
		GroupLessonId:   uuid.New().String(),
		StartRepeating:  time.Unix(startRepeating, 0),
		StopRepeating:   time.Unix(stopRepeating, 0),
		StartTime:       time.Unix(startTime, 0),
		EndTime:         time.Unix(endTime, 0),
		CreationTime:    time.Now(),
		EditTime:        time.Now(),
		LastSpawnedTime: time.Now(),
		RepeatEvery:     time.Duration(interval)}

	simulatedCurrentTime := time.Unix(0, 0)

	_, err := __CheckLessonHappening(simulatedCurrentTime, happeningNow)

	if err == nil {
		t.Log("The lesson which is not happening is apparently happening - blame danny")
		t.Fail()
	}
}
