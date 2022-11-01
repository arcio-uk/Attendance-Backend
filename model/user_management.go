package model

import (
	"arcio/attendance-system/utils"
	"log"
)

func GetUsers(DatabasePool *utils.DatabasePool) ([]User, error) {
	var returnUsers []User

	getUsers, err := DatabasePool.Database.Prepare("SELECT id, external_id, firstname, surname, " +
		"email, creation_time, edit_time " +
		"FROM users;")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	rows, err := getUsers.Query()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		var newUser User
		err = rows.Scan(&newUser.InternalId, &newUser.ExternalId, &newUser.Fname, &newUser.Sname, &newUser.Email,
			&newUser.CreationTime, &newUser.EditTime)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		returnUsers = append(returnUsers, newUser)
	}

	return returnUsers, nil
}

func GetModuleGroupsForUserModule(moduleuserid string, moduleid string, DatabasePool *utils.DatabasePool) ([]ModuleGroup, error) {
	stmt, err := DatabasePool.Database.Prepare("select module_groups.id, module_groups.module_id, module_groups.name, module_groups.creation_time, module_groups.edit_time " +
		"from module_groups, module_user_groups, module_users " +
		"where module_user_groups.module_user_id = module_users.id " +
		"and module_users.user_id = $1 and module_groups.module_id = $2 " +
		"and module_user_groups.module_group_id = module_groups.id;")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(moduleuserid, moduleid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Next()

	var ret []ModuleGroup = make([]ModuleGroup, 0)
	for rows.Next() {
		var mg ModuleGroup
		rows.Scan(&mg.Id, &mg.ModuleId, &mg.Name, &mg.CreationTime, &mg.EditTime)

		ret = append(ret, mg)
	}

	return ret, nil
}

func GetUserModuleGroups(userid string, moduleid string, pool *utils.DatabasePool) ([]ModuleGroup, error) {
	var moduleGroups []ModuleGroup

	stmt, err := pool.Database.Prepare("SELECT " +
		"module_groups.id, module_groups.module_id, module_groups.name, module_groups.creation_time, module_groups.edit_time " +
		"FROM module_users " +
		"INNER JOIN module_user_groups ON module_user_groups.module_user_id = module_users.id " +
		"INNER JOIN module_groups ON module_user_groups.module_group_id = module_groups.id " +
		"WHERE module_users.user_id = $1 AND module_users.module_id = $2;")

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userid, moduleid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var moduleGroup ModuleGroup
		rows.Scan(&moduleGroup.Id, &moduleGroup.ModuleId,
			&moduleGroup.Name, &moduleGroup.CreationTime, moduleGroup.EditTime)

		moduleGroups = append(moduleGroups, moduleGroup)
	}

	return moduleGroups, nil
}
