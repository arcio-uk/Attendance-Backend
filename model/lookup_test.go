package model

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/utils"
	"log"
	"testing"

	"github.com/google/uuid"
)

// Global test vars
var conf config.Config
var pool *utils.DatabasePool

func TestInit(t *testing.T) {
	conf, err := config.LoadConfig()
	if err != nil {
		t.Log("Failed to init the pool")
		t.Log(err)
		t.Fail()
	}
	pooltmp, err := utils.InitDatabasePool(conf)
	if err != nil {
		t.Log("Failed to init the pool")
		t.Log(err)
		t.Fail()
	}

	pool = pooltmp
}

func TestGetUsersForModule(t *testing.T) {
	empty_users, err := GetUsersForModule(uuid.New().String(), pool)
	if err != nil {
		t.Log("Expected success for random id")
		t.Log(err)
		t.Fail()
	}

	if len(empty_users) != 0 {
		t.Log("Expected an empty array for random id")
		t.Fail()
	}

	nil_users, err := GetUsersForModule("fail", pool)
	if err == nil {
		t.Log("Expected failure for bad course id")
		t.Fail()
	}

	if nil_users != nil {
		t.Log("Failure did not produce a nil array")
		t.Fail()
	}
}

func TestGetGroupsForCourse(t *testing.T) {
	groups, err := GetGroupsForModule(uuid.New().String(), pool)
	if err != nil {
		t.Log("Expected success for random id")
		t.Log(err)
		t.Fail()
	}

	if len(groups) != 0 {
		t.Log("Expected an empty array for random id")
		t.Fail()
	}

	nil_groups, err := GetGroupsForModule("fail", pool)
	if err == nil {
		t.Log("Expected failure for bad course id")
		t.Fail()
	}

	if nil_groups != nil {
		t.Log("Failure did not produce a nil array")
		t.Fail()
	}
}

func TestGetUserModules(t *testing.T) {
	groups, err := GetUsersModules("fail", pool)
	if err == nil {
		t.Log("Expected fail for bad id for get users courses")
		t.Log(err)
		t.Fail()
	}

	groups, err = GetUsersModules(uuid.New().String(), pool)
	if err != nil {
		t.Log("Expected success for getting user courses")
		t.Log(err)
		t.Fail()
	}

	if groups == nil {
		t.Log("Expected an array of courses")
		t.Fail()
	}
}

func TestGetUpcomingLessons(t *testing.T) {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
	lessons, err := GetLessons(uuid.New().String(), pool)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if lessons == nil {
		t.Log("Lessons cannot be nil")
		t.Fail()
	}
}
