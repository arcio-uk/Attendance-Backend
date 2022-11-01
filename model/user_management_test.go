package model

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/utils"
	"math/rand"
	"strings"
	"testing"
)

var conf_um config.Config
var pool_um *utils.DatabasePool

func TestInitUmTests(t *testing.T) {
	conf_um, err := config.LoadConfig()
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}

	db, err := utils.InitDatabasePool(conf_um)
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}

	pool_um = db
}

func TestAddUserToModuleGroup(t *testing.T) {
	users, err := GetUsers(pool_um)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	r := rand.Intn(len(users))
	uid := users[r].InternalId
	t.Log(users[r])
	modules, err := GetModulesForUser(uid, pool_um)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	mgroups, err := GetModuleGroupsForUserModule(uid, modules[0].Id, pool_um)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	if len(mgroups) == 0 {
		t.Log("No groups?")
		t.Fail()
		return
	}

	r = rand.Intn(len(mgroups))

	err = AddUserToModuleGroup(uid, mgroups[r].Id, pool_um)
	t.Log(err)

	if strings.Index(string(err.Error()), "duplicate key value violates") != -1 {
		t.Log("This should fail because the user is there")
	} else {
		t.Fail()
	}
}
