package model

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/utils"
	"context"
	"github.com/google/uuid"
	"log"
	"testing"
)

var conf_cm config.Config
var pool_cm *utils.DatabasePool

func TestInitCmTests(t *testing.T) {
	conf_cm, err := config.LoadConfig()
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}
	db, err := utils.InitDatabasePool(conf_cm)
	if err != nil {
		t.Log("Failed to init database pool")
		t.Fail()
	}

	pool_cm = db
}

func TestAddBadCourse(t *testing.T) {
	var module Module
	err := CreateModule(&module, pool_cm)

	if err == nil {
		t.Log("Adding a bad course should fail")
		t.Fail()
	}

	t.Log(err)

	module.Name = "test"
	err = CreateModule(&module, pool_cm)

	if err == nil {
		t.Log("Adding a bad course should fail")
		t.Fail()
	}

	t.Log(err)
}

func TestAddGoodCourse(t *testing.T) {
	module := Module{Name: "Test", ExternalId: "CS1812"}
	err := CreateModule(&module, pool_cm)

	if err != nil {
		t.Log(err)
		t.Log("Cannot add a good course :(")
		t.Fail()
	}

	if module.Id == "" {
		t.Log("Course id was not changed")
		t.Fail()
	}

	if module.CreationTime != module.EditTime && module.CreationTime.Unix() == 0 {
		t.Log("Course times were not updated")
		t.Fail()
	}
}

func TestAddRandUserToRandModule(t *testing.T) {
	uid := uuid.New().String()
	mid := uuid.New().String()

	err := AddUserToModule(uid, mid, pool_cm)
	if err == nil {
		t.Log("Expected failure for insert of random module and user")
		t.Fail()
	}

	t.Log(err)
}

func TestRmUserFromModuleGroup(t *testing.T) {
	uid := uuid.New().String()
	mid := uuid.New().String()

	err := RemoveUserFromModuleGroup(uid, mid, pool_cm)
	if err == nil {
		t.Log("Expected failure for insert of random module and user")
		t.Fail()
	}

	t.Log(err)
}

func TestAddThenRmUserFromModuleGroup(t *testing.T) {
	// Create transaction
	ctx := context.Background()
	tx, err := pool.Database.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}
	defer func() {
		tx.Rollback() // Always roll back as funky stuff is bad innit
	}()
	id := uuid.New().String()

	// Create user
	stmt, err := pool_cm.Database.PrepareContext(ctx, "insert into users (id, external_id, firstname, surname, email, password, salt) values ($1, 'asdf', 'Bob', 'McTestingTon', $2, 'password', 'salt');")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, id+"@testmail.com")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	// Get random module
	rows, err := pool_cm.Database.QueryContext(ctx, "select id from modules limit 1;")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer rows.Close()

	if !rows.Next() {
		t.Log("no modules")
		t.Log("*megamind ascii art*")
		t.Fail()
	}
	var mid string
	rows.Scan(&mid)

	// Add to module
	muid := uuid.New().String()
	stmt, err = pool_cm.Database.PrepareContext(ctx, "insert into module_users (id, user_id, module_id) values ($1, $2, $3);")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(muid, id, mid)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	// Get random module group
	stmt, err = pool_cm.Database.PrepareContext(ctx, "select module_groups.id from module_groups "+
		"where module_groups.module_id = $1 limit 1;")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer stmt.Close()

	rows, err = stmt.Query(mid)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	if !rows.Next() {
		t.Log("no module groups")
		t.Log("*megamind ascii art*")
		t.Fail()
	}
	var mgid string
	rows.Scan(&mgid)

	// Add to module group
	stmt, err = pool_cm.Database.PrepareContext(ctx, "insert into module_user_groups (id, module_user_id, module_group_id) values ($1, $2, $3);")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(uuid.New().String(), muid, mgid)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	// Remove from module group
	err = RemoveUserFromModuleGroup(id, mgid, pool_cm)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	// Kill the user at the end init govna
	defer func() {
		recover()
		// Rm user
		stmt, err = pool_cm.Database.PrepareContext(ctx, "delete from module_user_groups where module_user_id = $1;")
		if err != nil {
			t.Log(err)
			t.Fail()
			return
		}

		_, err = stmt.Exec(muid)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		stmt, err = pool_cm.Database.PrepareContext(ctx, "delete from module_users where user_id = $1;")
		if err != nil {
			t.Log(err)
			t.Fail()
			return
		}

		_, err = stmt.Exec(id)
		if err != nil {
			t.Log(err)
			t.Fail()
		}

		stmt, err = pool_cm.Database.PrepareContext(ctx, "delete from users where id = $1;")
		if err != nil {
			t.Log(err)
			t.Fail()
			return
		}

		_, err = stmt.Exec(id)
		if err != nil {
			t.Log(err)
			t.Fail()
			return

		}
	}()
}

func TestRmRandomUserFromRandModule(t *testing.T) {
	uid := uuid.New().String()
	mid := uuid.New().String()

	err := RemoveUserFromModule(uid, mid, pool_cm)
	if err == nil {
		t.Log("Expected failure for delete of random module and user")
		t.Fail()
	}

	t.Log(err)
}

func TestGetAllModules(t *testing.T) {
	ret, err := GetAllModules(pool_cm)
	if err != nil {
		t.Log(err)
		t.Log("Cannot get modules")
		t.Fail()
	}

	if ret == nil {
		t.Log("Return was nil")
		t.Fail()
	}

	if len(ret) == 0 {
		t.Log("No modules were returned")
		t.Fail()
	}
}
