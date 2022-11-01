package model

import (
	"arcio/attendance-system/utils"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

/**
 * Registers a user as attending a lesson if they are part of it.
 *
 * @param UserId the user id
 * @param Pool   the database pool
 * @return error the error message to send to the user,
 *               could be the user is not in the lesson
 */
func RegisterAttendance(UserId string, LessonId string, Pool *utils.DatabasePool) error {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := Pool.Database.BeginTx(ctx, nil)
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
	// 1. Check the lesson is happening now
	// 2. Check that the user is in the group which is in the group_lesson which is in the actual lesson

	// Insert with checks
	RecordTime := time.Now()

	id := uuid.New().String()
	stmt, err := Pool.Database.PrepareContext(ctx, "select 1 "+
		"from actual_lessons, group_lessons, module_groups, module_user_groups, "+
		"module_users, users "+
		"where actual_lessons.group_lesson_id = group_lessons.id and "+
		"group_lessons.module_group_id = module_groups.id and "+
		"module_user_groups.module_group_id = module_groups.id and "+
		"module_user_groups.module_user_id = module_users.id and "+
		"module_users.user_id = users.id and "+
		"users.id = $1 and "+
		"actual_lessons.end_time >= CURRENT_TIMESTAMP and "+
		"actual_lessons.start_time <= CURRENT_TIMESTAMP and "+
		"actual_lessons.id = $2;")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(UserId, LessonId)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Println("Lesson not found")
		return errors.New("Lesson not found")
	}

	stmt, err = Pool.Database.PrepareContext(ctx, "insert into attendance (id, lesson_id, user_id, register_time) values ($1, $2, $3, $4);")
	_, err = stmt.Exec(id, LessonId, UserId, RecordTime)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Registered attendance for user %s lesson %s successfully\n", UserId, LessonId)

	success = true
	return nil
}

func GetStudentAttendancePercentages(UserId string, Pool *utils.DatabasePool) (AttendanceRet, error) {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := Pool.Database.BeginTx(ctx, nil)
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

	// Get lesson count
	stmt, err := Pool.Database.PrepareContext(ctx, "select count(actual_lessons.id), module_groups.module_id "+
		"from actual_lessons, group_lessons, module_groups, module_user_groups, module_users "+
		"where actual_lessons.group_lesson_id = group_lessons.id and "+
		"actual_lessons.end_time <= CURRENT_TIMESTAMP and "+
		"group_lessons.module_group_id = module_groups.id and "+
		"module_user_groups.module_group_id = module_groups.id and "+
		"module_user_groups.module_user_id = module_users.id and "+
		"module_users.user_id = $1 "+
		"group by module_groups.module_id;")

	if err != nil {
		log.Println(err)
		return AttendanceRet{}, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(UserId)
	if err != nil {
		log.Println(err)
		return AttendanceRet{}, err
	}
	defer rows.Close()

	ret := make(map[string]AttendanceRecord)
	for rows.Next() {
		var rcrd AttendanceRecord
		var module_id string
		err = rows.Scan(&rcrd.TotalSessions, &module_id)
		if err != nil {
			log.Println(err)
			return AttendanceRet{}, err
		}

		ret[module_id] = rcrd
	}

	// Get attendance mark count
	stmt, err = Pool.Database.PrepareContext(ctx, "select count(attendance.id), module_groups.module_id "+
		"from attendance, actual_lessons, group_lessons, module_groups, module_user_groups, module_users "+
		"where attendance.lesson_id = actual_lessons.id and "+
		"attendance.user_id = $1 and "+
		"actual_lessons.end_time <= CURRENT_TIMESTAMP and "+
		"actual_lessons.group_lesson_id = group_lessons.id and "+
		"group_lessons.module_group_id = module_groups.id and "+
		"module_user_groups.module_group_id = module_groups.id and "+
		"module_user_groups.module_user_id = module_users.id and "+
		"module_users.user_id = $1 "+
		"group by module_groups.module_id;")
	if err != nil {
		log.Println(err)
		return AttendanceRet{}, err
	}
	defer stmt.Close()

	rows, err = stmt.Query(UserId)
	if err != nil {
		log.Println(err)
		return AttendanceRet{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var count int
		var module_id string
		err = rows.Scan(&count, &module_id)

		if err != nil {
			log.Println(err)
			return AttendanceRet{}, err
		}

		rcrd, found := ret[module_id]
		if !found {
			return AttendanceRet{}, errors.New(fmt.Sprintf("Cannot find module %s for user %s", module_id, UserId))
		} else {
			rcrd.MarkedSessions = count
			ret[module_id] = rcrd
		}
	}

	success = true
	return AttendanceRet{ModuleAttendance: ret}, nil
}

func GetStudentModuleAttendancePercentages(UserId string, ModuleId string, Pool *utils.DatabasePool) (AttendanceRecord, error) {
	// Create transaction
	success := false
	ctx := context.Background()
	tx, err := Pool.Database.BeginTx(ctx, nil)
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

	var ret AttendanceRecord

	// Get lesson count
	stmt, err := Pool.Database.PrepareContext(ctx, "select count(actual_lessons.id) "+
		"from actual_lessons, group_lessons, module_groups, module_user_groups, module_users "+
		"where actual_lessons.group_lesson_id = group_lessons.id and "+
		"actual_lessons.end_time <= CURRENT_TIMESTAMP and "+
		"group_lessons.module_group_id = module_groups.id and "+
		"module_user_groups.module_group_id = module_groups.id and "+
		"module_user_groups.module_user_id = module_users.id and "+
		"module_users.user_id = $1 and "+
		"module_groups.module_id = $2;")

	if err != nil {
		log.Println(err)
		return AttendanceRecord{}, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(UserId, ModuleId)
	if err != nil {
		log.Println(err)
		return AttendanceRecord{}, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&ret.TotalSessions)
		if err != nil {
			log.Println(err)
			return AttendanceRecord{}, err
		}
	} else {
		log.Println("No marks found")
		return AttendanceRecord{}, errors.New("No marks found")
	}

	// Get attendance mark count
	stmt, err = Pool.Database.PrepareContext(ctx, "select count(attendance.id) "+
		"from attendance, actual_lessons, group_lessons, module_groups, module_user_groups, module_users "+
		"where attendance.lesson_id = actual_lessons.id and "+
		"attendance.user_id = $1 and "+
		"actual_lessons.end_time <= CURRENT_TIMESTAMP and "+
		"actual_lessons.group_lesson_id = group_lessons.id and "+
		"group_lessons.module_group_id = module_groups.id and "+
		"module_user_groups.module_group_id = module_groups.id and "+
		"module_user_groups.module_user_id = module_users.id and "+
		"module_users.user_id = $1 and "+
		"module_groups.module_id = $2;")
	if err != nil {
		log.Println(err)
		return AttendanceRecord{}, err
	}
	defer stmt.Close()

	rows, err = stmt.Query(UserId, ModuleId)
	if err != nil {
		log.Println(err)
		return AttendanceRecord{}, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&ret.MarkedSessions)

		if err != nil {
			log.Println(err)
			return AttendanceRecord{}, err
		}

	} else {
		log.Println("No marks found")
		return AttendanceRecord{}, errors.New("No marks found")
	}

	success = true
	return ret, nil
}
