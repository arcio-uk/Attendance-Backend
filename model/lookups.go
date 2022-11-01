package model

import (
	"arcio/attendance-system/utils"
	"log"
)

// Funcs
func GetUsersForModule(ModuleId string, Pool *utils.DatabasePool) ([]User, error) {
	stmt, err := Pool.Database.Prepare("select (users.id, users.external_id, firstname, surname, email, users.creation_time, users.edit_time) " +
		"from users, module_users " +
		"where user_id = users.id and module_id = $1;")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(ModuleId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	ret := make([]User, 0)
	for rows.Next() {
		var user User
		rows.Scan(&user.InternalId, &user.ExternalId, &user.Fname, &user.Sname, &user.Email, &user.CreationTime, &user.EditTime)

		ret = append(ret, user)
	}

	return ret, nil
}

func GetGroupsForModule(ModuleId string, Pool *utils.DatabasePool) ([]ModuleGroup, error) {
	stmt, err := Pool.Database.Prepare("select id, module_id, name, creation_time, edit_time from module_groups where module_id = $1;")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(ModuleId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	ret := make([]ModuleGroup, 0)
	for rows.Next() {
		var moduleGroup ModuleGroup
		rows.Scan(&moduleGroup.Id, &moduleGroup.ModuleId, &moduleGroup.Name, &moduleGroup.CreationTime, &moduleGroup.EditTime)
		ret = append(ret, moduleGroup)
	}

	return ret, nil
}

const GET_USER_COURSES = "select modules.id, modules.external_id, modules.name, modules.creation_time, modules.edit_time " +
	"from modules, module_users " +
	"where modules.id = module_users.module_id and module_users.user_id = $1;"

func GetUsersModules(UserId string, Pool *utils.DatabasePool) ([]Module, error) {
	stmt, err := Pool.Database.Prepare(GET_USER_COURSES)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(UserId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	ret := make([]Module, 0)
	for rows.Next() {
		var module Module
		rows.Scan(&module.Id, &module.ExternalId, &module.Name, &module.CreationTime, &module.EditTime)

		ret = append(ret, module)
	}

	return ret, nil
}

/*
* Returns all the module groups the user is in by inner joining users, module_users, module_users_groups.
 */
func GetUsersModuleGroups(UserId string, pool *utils.DatabasePool) ([]ModuleGroup, error) {

	userModuleGroups := make([]ModuleGroup, 0)

	getModuleGroupsId, err := pool.Database.Prepare("select module_groups.id, module_groups.module_id, module_groups.name, module_groups.creation_time, module_groups.edit_time " +
		"from module_groups " +
		"INNER JOIN module_user_groups ON module_user_groups.module_group_id = module_groups.id " +
		"INNER JOIN module_users ON module_users.id = module_user_groups.module_user_id " +
		"INNER JOIN users ON users.id = module_users.user_id " +
		"WHERE users.id=$1;")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	rows, err := getModuleGroupsId.Query(UserId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		var singleModuleGroup ModuleGroup
		rows.Scan(&singleModuleGroup.Id, &singleModuleGroup.ModuleId, &singleModuleGroup.Name, &singleModuleGroup.CreationTime, &singleModuleGroup.EditTime)
		userModuleGroups = append(userModuleGroups, singleModuleGroup)
	}

	return userModuleGroups, nil
}
