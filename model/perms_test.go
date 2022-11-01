package model

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"log"
	"testing"
)

var conf_pt config.Config
var pool_pt *utils.DatabasePool

func TestInitPtTests(t *testing.T) {
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime | log.Lmicroseconds)
	conf_pt, err := config.LoadConfig()
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}
	db, err := utils.InitDatabasePool(conf_pt)

	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}

	pool_pt = db
}

func TestGetPermissionsGlobalLayer(t *testing.T) {
	rows, err := pool_pt.Database.Query("select id from users limit 1;")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer rows.Close()

	if !rows.Next() {
		t.Log("No users?")
		t.Fail()
		return
	}

	var id string
	err = rows.Scan(&id)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	_, err = GetPermissions(id, "", "", security.Global, pool_pt)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
}

func TestGetPermissionsAttendance1(t *testing.T) {
	rows, err := pool_pt.Database.Query("select id from users limit 1;")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer rows.Close()

	if !rows.Next() {
		t.Log("No users?")
		t.Fail()
		return
	}

	var id string
	err = rows.Scan(&id)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	_, err = GetPermissions(id, "", "", security.Attendance, pool_pt)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
}
