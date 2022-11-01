package model

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/utils"
	"github.com/google/uuid"
	"log"
	"testing"
)

var conf_at config.Config
var pool_at *utils.DatabasePool

func TestInitAtTests(t *testing.T) {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
	conf_at, err := config.LoadConfig()
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}

	db, err := utils.InitDatabasePool(conf_at)
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}

	pool_at = db
}

func TestUserLessonDoesNotExist(t *testing.T) {
	userid := uuid.New().String()
	lessonid := uuid.New().String()

	err := RegisterAttendance(userid, lessonid, pool_at)
	if err == nil {
		t.Log("Expected error when marking attendance for fake users")
		t.Log(err)
		t.Fail()
	}
}

func TestAttendancePercentageUserNotThere(t *testing.T) {
	userid := uuid.New().String()

	_, err := GetStudentAttendancePercentages(userid, pool_at)
	if err != nil {
		t.Fail()
	}
}

func TestAttendancePercentageForStudents(t *testing.T) {
	rows, err := pool_at.Database.Query("select users.id from users order by creation_time asc limit 100;")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	defer rows.Close()
	net := 0
	att := 0

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			t.Log(err)
			t.Fail()
			return
		}

		ret, err := GetStudentAttendancePercentages(id, pool_at)
		if err != nil {
			t.Log(err)
			t.Log("FAILED to get attendance %")
			t.Fail()
		}

		for k, v := range ret.ModuleAttendance {
			if k == "" {
				t.Logf("No module id for user %s", id)
				t.Fail()
			}

			if v.MarkedSessions < 0 || v.MarkedSessions > v.TotalSessions {
				t.Logf("marked sesssions out of bound for user %s mareked: %d total: %d", id, v.MarkedSessions, v.TotalSessions)
				t.Fail()
			}

			if v.TotalSessions <= 0 {
				t.Logf("No lessons where found that attendnace could have been marked for %d", v.TotalSessions)
				t.Fail()
			}

			net += v.MarkedSessions
		}

		att += len(ret.ModuleAttendance)
	}

	if net <= 0 {
		t.Log("There were no marked sessions at all!!")
		t.Fail()
	}

	if att <= 0 {
		t.Fail()
	}
}

func TestAttendancePercentageForStudentModules(t *testing.T) {
	rows, err := pool_at.Database.Query("select users.id, module_users.module_id from users, module_users " +
		"where module_users.user_id = users.id " +
		"order by users.creation_time asc limit 100;")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	defer rows.Close()
	net := 0

	for rows.Next() {
		var id string
		var mid string
		err = rows.Scan(&id, &mid)
		if err != nil {
			t.Log(err)
			t.Fail()
			return
		}

		ret, err := GetStudentModuleAttendancePercentages(id, mid, pool_at)
		TotalSessions := ret.TotalSessions
		MarkedSessions := ret.MarkedSessions

		if err != nil {
			t.Log(err)
			t.Log("FAILED to get attendance %")
			t.Fail()
		}

		if MarkedSessions < 0 || MarkedSessions > TotalSessions {
			t.Logf("marked sesssions out of bound for user %s mareked: %d total: %d", id, MarkedSessions, TotalSessions)
			t.Fail()
		}

		net += MarkedSessions
	}
}
