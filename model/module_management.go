package model

import (
	"arcio/attendance-system/utils"
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

func CreateModule(module *Module, pool *utils.DatabasePool) error {
	if module.Name == "" {
		return errors.New("The module name cannot be empty")
	}

	if module.ExternalId == "" {
		return errors.New("The external id (tag) cannot be empty")
	}

	module.CreationTime = time.Now()
	module.EditTime = module.CreationTime
	module.Id = uuid.New().String()
	stmt, err := pool.Database.Prepare("insert into modules (id, name, external_id, creation_time, edit_time) values ($1, $2, $3, $4, $5);")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(module.Id, module.Name, module.ExternalId, module.CreationTime, module.EditTime)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func CreateModuleGroup(moduleGroup *ModuleGroup, pool *utils.DatabasePool) error {
	if moduleGroup.Name == "" {
		return errors.New("The module group name cannot be empty")
	}

	moduleGroup.CreationTime = time.Now()
	moduleGroup.EditTime = moduleGroup.CreationTime
	moduleGroup.Id = uuid.New().String()

	createModuleGroup, err := pool.Database.Prepare("insert into module_groups (id, module_id, name, creation_time, edit_time) " +
		"values ($1, $2, $3, $4, $5);")
	if err != nil {
		log.Println(err)
		return err
	}
	defer createModuleGroup.Close()

	_, err = createModuleGroup.Exec(moduleGroup.Id, moduleGroup.ModuleId, moduleGroup.Name, moduleGroup.CreationTime, moduleGroup.EditTime)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetAllModules(DatabasePool *utils.DatabasePool) ([]Module, error) {
	var returnModules []Module

	getUsers, err := DatabasePool.Database.Prepare("SELECT id, external_id, name, creation_time, edit_time FROM modules;")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer getUsers.Close()

	rows, err := getUsers.Query()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var newModule Module
		rows.Scan(&newModule.Id, &newModule.ExternalId, &newModule.Name, &newModule.CreationTime,
			&newModule.EditTime)
		returnModules = append(returnModules, newModule)
	}

	return returnModules, nil
}

func GetModuleStudents(moduleId string, DatabasePool *utils.DatabasePool) ([]ModuleUserAttendanceRet, error) {

	//Gets the number of lessons for a given module.
	getModuleCount, err := DatabasePool.Database.Query(`
SELECT COUNT(actual_lessons.id), modules.name
FROM modules
LEFT JOIN module_groups ON module_groups.module_id = modules.id
LEFT JOIN group_lessons ON group_lessons.module_group_id = module_groups.id
LEFT JOIN actual_lessons ON actual_lessons.group_lesson_id = group_lessons.id
WHERE modules.id = $1
GROUP BY modules.name
  `, moduleId)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	moduleInfo := struct {
		Count int
		Name  string
	}{}

	if getModuleCount.Next() {
		err = getModuleCount.Scan(&moduleInfo.Count, &moduleInfo.Name)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	//Gets the users information with attendance for this module.
	getModuleUsers, err := DatabasePool.Database.Query(`SELECT COUNT(attendance.id), users.id, users.external_id, users.firstname, users.surname, users.email
FROM users
INNER JOIN module_users ON module_users.user_id = users.id
INNER JOIN modules ON module_users.module_id = modules.id
LEFT JOIN module_groups ON module_groups.module_id = modules.id
LEFT JOIN group_lessons ON group_lessons.module_group_id = module_groups.id
LEFT JOIN actual_lessons ON actual_lessons.group_lesson_id = group_lessons.id
LEFT JOIN attendance ON attendance.lesson_id = actual_lessons.id AND attendance.user_id = users.id
WHERE modules.id = $1
GROUP BY users.id, users.external_id, users.firstname, users.surname, users.email
	`, moduleId)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	var users []ModuleUserAttendanceRet

	for getModuleUsers.Next() {
		var user ModuleUserAttendanceRet
		var attendance AttendanceRecord
		attendance.TotalSessions = moduleInfo.Count

		getModuleUsers.Scan(&attendance.MarkedSessions, &user.InternalId, &user.ExternalId, &user.Fname,
			&user.Sname, &user.Email)

		user.Attendance = attendance
		users = append(users, user)
	}

	return users, nil
}

func AddUserToModule(uid string, mid string, pool *utils.DatabasePool) error {
	id := uuid.New().String()
	creationTime := time.Now()
	editTime := creationTime

	stmt, err := pool.Database.Prepare("insert into module_users (id, user_id, module_id, creation_time, edit_time) values ($1, $2, $3, $4, $5);")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, uid, mid, creationTime, editTime)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func AddUsersToModule(uids []string, mid string, pool *utils.DatabasePool) error {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := pool.Database.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if success {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	creationTime := time.Now()
	editTime := creationTime

	stmt, err := pool.Database.PrepareContext(ctx, "insert into module_users (id, user_id, module_id, creation_time, edit_time) values ($1, $2, $3, $4, $5);")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	for i := 0; i < len(uids); i++ {
		id := uuid.New().String()
		_, err = stmt.Exec(id, uids[i], mid, creationTime, editTime)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	success = true
	return nil
}

func RemoveUserFromModule(uid string, mid string, pool *utils.DatabasePool) error {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := pool.Database.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if success {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// Get module user id
	stmt, err := pool.Database.PrepareContext(ctx, "select id from module_users where user_id = $1 and module_id = $2;")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(uid, mid)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Println("Cannot find user")
		return errors.New("cannot find user")
	}
	var moduleUserId string
	rows.Scan(&moduleUserId)

	// Delete module user from groups
	stmt, err = pool.Database.PrepareContext(ctx, "delete from module_user_groups where module_user_id = $1;")

	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(moduleUserId)
	if err != nil {
		log.Println(err)
		return err
	}

	// Remove module user
	stmt, err = pool.Database.PrepareContext(ctx, "delete from module_users where id = $1;")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(moduleUserId)
	if err != nil {
		log.Println(err)
		return err
	}

	success = true
	return nil
}

func RemoveUserFromModuleGroup(uid string, mid string, pool *utils.DatabasePool) error {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := pool.Database.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if success {
			tx.Commit()
		} else {
			tx.Rollback() // Always roll back as funky stuff is bad innit
		}
	}()

	// Get module user id
	stmt, err := pool.Database.PrepareContext(ctx, "select id from module_users where user_id = $1;")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(uid)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Println("Cannot find user")
		return errors.New("Cannot find user")
	}

	var muid string
	rows.Scan(&muid)

	// Get groups
	stmt, err = pool.Database.PrepareContext(ctx, "select id from module_groups where module_id = $1;")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	rows, err = stmt.Query(mid)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var mgid string
		rows.Scan(&mgid)
		stmt, err = pool.Database.PrepareContext(ctx, "delete from module_user_groups "+
			"where module_user_groups.module_user_id = $1 and module_user_groups.module_group_id = $2;")
		if err != nil {
			log.Println(err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(muid, mgid)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	success = true
	return nil
}

func GetModulesForUser(userid string, DatabasePool *utils.DatabasePool) ([]Module, error) {
	var returnModules []Module

	getUserModules, err := DatabasePool.Database.Prepare("SELECT modules.id, modules.external_id, modules.name, " +
		"modules.creation_time, modules.edit_time FROM modules " +
		"INNER JOIN module_users ON module_users.module_id = modules.id " +
		"INNER JOIN users ON users.id = module_users.user_id " +
		"WHERE users.id=$1")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer getUserModules.Close()

	rows, err := getUserModules.Query(userid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var newModule Module
		rows.Scan(&newModule.Id, &newModule.ExternalId, &newModule.Name,
			&newModule.CreationTime, &newModule.EditTime)
		returnModules = append(returnModules, newModule)
	}

	return returnModules, nil
}

func AddUserToModuleGroup(userid string, modulegroupid string, DatabasePool *utils.DatabasePool) error {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := DatabasePool.Database.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if success {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// Get the module user id, to check if the user is in this module.
	stmt, err := DatabasePool.Database.PrepareContext(ctx, "SELECT module_users.id "+
		"FROM module_users "+
		"INNER JOIN modules ON modules.id = module_users.module_id "+
		"INNER JOIN module_groups ON module_groups.module_id = modules.id "+
		"WHERE module_groups.id=$1 "+
		"AND module_users.user_id=$2;")

	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(modulegroupid, userid)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		err = errors.New("User not found in module")
		log.Println(err)
		return err
	}

	var muid string
	rows.Scan(&muid)

	stmt, err = DatabasePool.Database.PrepareContext(ctx, "insert into module_user_groups (id, module_user_id, module_group_id) values ($1, $2, $3);")
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = stmt.Exec(uuid.New().String(), muid, modulegroupid)
	if err != nil {
		log.Println(err)
		return err
	}
	success = true

	return nil
}

func GetModuleGroupStudents(moduleGroupId string, DatabasePool *utils.DatabasePool) (ModuleGroupRet, error) {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := DatabasePool.Database.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return ModuleGroupRet{}, err
	}
	defer func() {
		if success {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	var moduleGroup ModuleGroupRet

	//Get the module name
	nameQuery, err := DatabasePool.Database.QueryContext(ctx,
		`SELECT name FROM module_groups WHERE id=$1`, moduleGroupId)

	if err != nil {
		log.Println(err)
		return ModuleGroupRet{}, err
	}
	defer nameQuery.Close()

	if !nameQuery.Next() {
		log.Println("Error getting module group name")
		return ModuleGroupRet{}, errors.New("error getting module group lesson count")
	}
	nameQuery.Scan(&moduleGroup.ModuleName)

	// Module inf
	groupLessons, err := DatabasePool.Database.QueryContext(ctx, `SELECT COUNT(actual_lessons.id) 
FROM module_groups
LEFT JOIN group_lessons ON group_lessons.module_group_id = module_groups.id
LEFT JOIN actual_lessons ON actual_lessons.group_lesson_id = group_lessons.id
WHERE module_groups.id = $1
AND actual_lessons.end_time <= current_timestamp;`, moduleGroupId)
	if err != nil {
		log.Println(err)
		return ModuleGroupRet{}, err
	}
	defer groupLessons.Close()

	if !groupLessons.Next() {
		log.Println("Error getting module group info")
		return ModuleGroupRet{}, errors.New("error getting module group lesson count")
	}

	var groupLessonCount int
	err = groupLessons.Scan(&groupLessonCount)
	if err != nil {
		log.Println(err)
		return ModuleGroupRet{}, err
	}

	moduleGroup.ModuleGroupId = moduleGroupId

	// Get attendance and, users
	userAttendance, err := DatabasePool.Database.Query("SELECT COUNT(attendance.id), users.id, users.external_id, users.firstname, users.surname, users.email "+
		"FROM users "+
		"INNER JOIN module_users ON module_users.user_id = users.id "+
		"INNER JOIN module_user_groups ON module_user_groups.module_user_id = module_users.id "+
		"INNER JOIN module_groups ON module_groups.id = module_user_groups.module_group_id "+
		"LEFT JOIN group_lessons ON group_lessons.module_group_id = module_groups.id "+
		"LEFT JOIN actual_lessons ON actual_lessons.group_lesson_id = group_lessons.id "+
		"LEFT JOIN attendance ON attendance.lesson_id = actual_lessons.id AND attendance.user_id = users.id "+
		"WHERE module_user_groups.module_group_id  = $1 "+
		"GROUP BY users.id, users.firstname, users.surname, users.email, users.external_id;", moduleGroupId)

	if err != nil {
		log.Println(err)
		return ModuleGroupRet{}, err
	}
	defer userAttendance.Close()

	users_arr := make([]ModuleUserAttendanceRet, 0)
	for userAttendance.Next() {
		var userAttRet ModuleUserAttendanceRet
		err = userAttendance.Scan(&userAttRet.Attendance.MarkedSessions, &userAttRet.InternalId, &userAttRet.ExternalId, &userAttRet.Fname, &userAttRet.Sname, &userAttRet.Email)
		if err != nil {
			log.Println(err)
			return ModuleGroupRet{}, err
		}

		userAttRet.Attendance.TotalSessions = groupLessonCount
		users_arr = append(users_arr, userAttRet)
	}
	moduleGroup.UserAttendance = users_arr

	success = true
	return moduleGroup, nil
}

func AddUsersToModuleGroup(userid []string, modulegroupid string, DatabasePool *utils.DatabasePool) error {
	for i := 0; i < len(userid); i++ {
		err := AddUserToModuleGroup(userid[i], modulegroupid, DatabasePool)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
