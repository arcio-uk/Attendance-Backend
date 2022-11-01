package model

import (
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"context"
	"log"
)

func getAttendancePerms(userId string, pool *utils.DatabasePool, ctx context.Context) (security.Overrides, error) {
	stmt, err := pool.Database.PrepareContext(ctx, "select overrides from roles, attendance_user_roles "+
		"where attendance_user_roles.user_id = $1 and roles.id = attendance_user_roles.role_id;")
	if err != nil {
		return security.PERMS_NONE, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return security.PERMS_NONE, err
	}
	defer rows.Close()

	overrides := make([]security.Overrides, 0)
	for rows.Next() {
		var override security.Overrides

		err = rows.Scan(&override)
		if err != nil {
			return security.PERMS_NONE, err
		}

		overrides = append(overrides, override)
	}

	return security.CalculatePermissionsInner(overrides), nil
}

func getModuleGroupPerms(userId string, moduleId string, moduleGroupId string, pool *utils.DatabasePool, ctx context.Context) (security.Overrides, error) {
	stmt, err := pool.Database.PrepareContext(ctx, "select overrides from module_user_group_roles, "+
		"module_groups, module_user_groups, module_users, roles "+
		"where roles.id = module_user_group_roles.role_id and "+
		"module_user_group_roles.module_user_group_id = module_user_groups.id and "+
		"module_users.user_id = $1 and "+
		"module_user_groups.module_user_id = module_users.id and "+
		"module_users.module_id = $2 and "+
		"module_user_roles.role_id = roles.id and "+
		"module_group_roles.module_group_id = $3;")
	if err != nil {
		return security.PERMS_NONE, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId, moduleId, moduleGroupId)
	if err != nil {
		return security.PERMS_NONE, err
	}
	defer rows.Close()

	overrides := make([]security.Overrides, 0)
	for rows.Next() {
		var override security.Overrides

		err = rows.Scan(&override)
		if err != nil {
			return security.PERMS_NONE, err
		}

		overrides = append(overrides, override)
	}

	return security.CalculatePermissionsInner(overrides), nil
}

func getModulePerms(userId string, moduleId string, pool *utils.DatabasePool, ctx context.Context) (security.Overrides, error) {
	stmt, err := pool.Database.PrepareContext(ctx, "select overrides from roles, module_user_roles, module_users "+
		"where module_users.user_id = $1 and "+
		"module_users.module_id = $2 and "+
		"roles.id = module_user_roles.role_id and "+
		"module_user_roles.module_user_id = module_users.id;")
	if err != nil {
		return security.PERMS_NONE, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId, moduleId)
	if err != nil {
		return security.PERMS_NONE, err
	}
	defer rows.Close()

	overrides := make([]security.Overrides, 0)
	for rows.Next() {
		var override security.Overrides

		err = rows.Scan(&override)
		if err != nil {
			return security.PERMS_NONE, err
		}

		overrides = append(overrides, override)
	}

	return security.CalculatePermissionsInner(overrides), nil
}

func getGlobalPerms(userId string, pool *utils.DatabasePool, ctx context.Context) (security.Overrides, error) {
	stmt, err := pool.Database.PrepareContext(ctx, "select overrides from roles, role_users "+
		"where role_users.user_id = $1 and role_users.role_id = roles.id;")
	if err != nil {
		return security.PERMS_NONE, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		return security.PERMS_NONE, err
	}
	defer rows.Close()

	overrides := make([]security.Overrides, 0)
	for rows.Next() {
		var override security.Overrides

		err = rows.Scan(&override)
		if err != nil {
			return security.PERMS_NONE, err
		}

		overrides = append(overrides, override)
	}

	return security.CalculatePermissionsInner(overrides), nil
}

/*
 * Permissions getter
 * For moduleId and, moduleGroupId these are set to blank if the layer is above their respective location.
 * If a layer has a blank ID then it skips the layer continues to the parent layer
 */
func GetPermissions(userId string, moduleId string, moduleGroupId string, l security.Layer, pool *utils.DatabasePool) (security.Overrides, error) {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := pool.Database.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}
	defer func() {
		if success {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()
	// Init ret
	overrides := make([]security.Overrides, 0)

	// All cases are meant to flow
	switch l {
	case security.Attendance:
		t, err := getAttendancePerms(userId, pool, ctx)
		if err != nil {
			log.Println(err)
			return security.PERMS_NONE, err
		}

		overrides = append(overrides, t)
		fallthrough
	case security.ModuleGroup:
		if moduleGroupId != "" && moduleId != "" {
			t, err := getModuleGroupPerms(userId, moduleId, moduleGroupId, pool, ctx)
			if err != nil {
				log.Println(err)
				return security.PERMS_NONE, err
			}

			overrides = append(overrides, t)
		}
		fallthrough
	case security.Module:
		if moduleId != "" {
			t, err := getModulePerms(userId, moduleId, pool, ctx)
			if err != nil {
				log.Println(err)
				return security.PERMS_NONE, err
			}

			overrides = append(overrides, t)
		}
		fallthrough
	case security.Global:
		t, err := getGlobalPerms(userId, pool, ctx)
		if err != nil {
			log.Println(err)
			return security.PERMS_NONE, err
		}

		overrides = append(overrides, t)
	}

	if len(overrides) == 0 {
		return security.PERMS_NONE, nil
	}

	success = true
	return security.CalculatePermissions(overrides[0], overrides[1:]), nil
}
