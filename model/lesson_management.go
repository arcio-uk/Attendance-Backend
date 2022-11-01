package model

import (
	"arcio/attendance-system/utils"
	"errors"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

func CreateActualLesson(lesson ActualLesson, pool *utils.DatabasePool) error {
	lesson.CreationTime = time.Now()
	lesson.EditTime = lesson.CreationTime
	lesson.Id = uuid.New().String()

	stmt, err := pool.Database.Prepare("select (summary, description, location) from group_lessons where id = $1;")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(lesson.GroupLessonId)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return errors.New("Cannot find group lesson with matching id")
	}

	var summary string
	var description string
	var location string
	rows.Scan(&summary, &description, &location)

	stmt, err = pool.Database.Prepare("insert into actual_lessons (id, group_lesson_id, start_time, end_time, creation_time, edit_time, summary, description, location) values($1, $2, $3, $4, $5, $6, $7, $8, $9);")
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(lesson.Id, lesson.GroupLessonId, lesson.StartTime, lesson.EndTime, lesson.CreationTime, lesson.EditTime, summary, description, location)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

//
func GetRepeatingLessons(UserId string, pool *utils.DatabasePool) ([]RepeatingLesson, error) {
	repeatingLessonsRet := make([]RepeatingLesson, 0)

	getUpcomingRepeatingLessons, err := pool.Database.Prepare("SELECT repeating_lessons.id, repeating_lessons.start_repeating, repeating_lessons.stop_repeating, repeating_lessons.start_time, " +
		"repeating_lessons.end_time, repeating_lessons.group_lesson_id, " +
		"(EXTRACT(epoch FROM repeating_lessons.repeat_every) * 1000000000)::BIGINT, " +
		"repeating_lessons.creation_time, repeating_lessons.edit_time " +
		"FROM repeating_lessons " +
		"INNER JOIN group_lessons ON group_lessons.id = repeating_lessons.group_lesson_id " +
		"INNER JOIN module_groups ON module_groups.id = group_lessons.module_group_id " +
		"INNER JOIN module_user_groups ON module_user_groups.module_group_id = module_groups.id " +
		"INNER JOIN module_users ON module_users.id = module_user_groups.module_user_id " +
		"INNER JOIN users ON users.id = module_users.user_id " +
		"WHERE repeating_lessons.start_repeating <= CURRENT_DATE and repeating_lessons.stop_repeating >= CURRENT_DATE AND " +
		"users.id=$1")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	rows, err := getUpcomingRepeatingLessons.Query(UserId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		var singleRepeatingLesson RepeatingLesson
		err := rows.Scan(&singleRepeatingLesson.Id,
			&singleRepeatingLesson.StartRepeating,
			&singleRepeatingLesson.StopRepeating,
			&singleRepeatingLesson.StartTime,
			&singleRepeatingLesson.EndTime,
			&singleRepeatingLesson.GroupLessonId,
			&singleRepeatingLesson.RepeatEvery,
			&singleRepeatingLesson.CreationTime,
			&singleRepeatingLesson.EditTime)
		repeatingLessonsRet = append(repeatingLessonsRet, singleRepeatingLesson)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	return repeatingLessonsRet, nil
}

const LESSON_QUERY_LIMIT = 1000

func GetLessons(UserId string, Pool *utils.DatabasePool) ([]ActualLesson, error) {
	var wg sync.WaitGroup
	var lock sync.Mutex
	var reterr error = nil
	wg.Add(2)

	ret := make([]ActualLesson, 0)

	// Get one off lessons
	go func() {
		defer wg.Done()
		stmt, err := Pool.Database.Prepare("select actual_lessons.id, actual_lessons.group_lesson_id, " +
			"actual_lessons.start_time, actual_lessons.end_time, actual_lessons.creation_time, " +
			"actual_lessons.edit_time, actual_lessons.description, actual_lessons.summary, " +
			"actual_lessons.location " +
			"from group_lessons, actual_lessons, module_user_groups, module_users " +
			"where " +
			"group_lessons.module_group_id = module_user_groups.module_group_id and " +
			"module_user_groups.module_user_id = module_users.id and module_users.user_id = $1 and " +
			"actual_lessons.group_lesson_id = group_lessons.id and " +
			"end_time >= CURRENT_TIMESTAMP - interval '3 weeks'" +
			"order by start_time asc;")
		if err != nil {
			log.Println(err)
			reterr = err
			return
		}
		defer stmt.Close()

		rows, err := stmt.Query(UserId)
		if err != nil {
			log.Println(err)
			reterr = err
			return
		}
		defer rows.Close()

		for rows.Next() && len(ret) < LESSON_QUERY_LIMIT {
			var lesson ActualLesson
			err := rows.Scan(&lesson.Id, &lesson.GroupLessonId, &lesson.StartTime, &lesson.EndTime, &lesson.CreationTime, &lesson.EditTime, &lesson.Description, &lesson.Summary, &lesson.Location)

			if err != nil {
				reterr = err
				log.Println(err)
				return
			}

			lock.Lock()
			ret = append(ret, lesson)
			lock.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		// Get repeating lessons
		lessons, err := GetRepeatingLessons(UserId, Pool)
		if err != nil {
			log.Println(err)
			reterr = err
			return
		}

		var wg_inner sync.WaitGroup
		wg_inner.Add(len(lessons))

		for __i := 0; __i < len(lessons); __i++ {
			go func(i int) {
				defer wg_inner.Done()
				rlesson := lessons[i]

				stmt, err := Pool.Database.Prepare("select summary, description, location from group_lessons where id = $1;")
				if err != nil {
					reterr = err
					log.Println(err)
					return
				}
				defer stmt.Close()

				rows, err := stmt.Query(rlesson.GroupLessonId)
				if err != nil {
					reterr = err
					log.Println(err)
					return
				}
				defer rows.Close()

				if !rows.Next() {
					err := errors.New("Cannot find group lesson")
					reterr = err
					log.Println(err)
					return
				}

				var summary string
				var description string
				var location string
				err = rows.Scan(&summary, &description, &location)
				if err != nil {
					reterr = err
					log.Println(err)
					return
				}

				now := time.Now()
				var epoch time.Time
				// you need to add one because it was off by one for some reason - idk man
				lesson := ActualLesson{Id: rlesson.Id,
					GroupLessonId: rlesson.GroupLessonId,
					CreationTime:  rlesson.CreationTime,
					EditTime:      rlesson.EditTime,
					StartTime:     rlesson.StartRepeating.Add(rlesson.StartTime.Sub(epoch)),
					EndTime:       rlesson.StartRepeating.Add(rlesson.EndTime.Sub(epoch)),
					Summary:       summary,
					Description:   description,
					Location:      location,
					IsAbstract:    true}

				// Sanity check: I do not want to add too many lessons,
				// lets say they want lessons from now to 9999 then that will be silly
				for j := 0; j < LESSON_QUERY_LIMIT; j++ {
					// Check upper bound
					if lesson.StartTime.After(rlesson.StopRepeating) {
						break
					}

					// Spawned lessons are accounted in the actual lessons check
					if lesson.StartTime.After(rlesson.LastSpawnedTime) {

						// Only add lessons that after now()
						if lesson.StartTime.After(now) {
							lessonCopy := lesson
							lock.Lock()
							ret = append(ret, lessonCopy)
							lock.Unlock()
						}
					}

					lesson.StartTime = lesson.StartTime.Add(rlesson.RepeatEvery)
					lesson.EndTime = lesson.EndTime.Add(rlesson.RepeatEvery)
				}
			}(__i)
		}

		wg_inner.Wait()
	}()

	wg.Wait()
	if reterr != nil {
		return nil, reterr
	}

	// Sort and truncate
	if len(ret) > LESSON_QUERY_LIMIT {
		log.Println("Truncated lessons")
		ret = ret[0 : LESSON_QUERY_LIMIT+1] // Slice to LESSON_QUERY_LIMIT
	}
	sort.SliceStable(ret, func(i int, j int) bool {
		return ret[i].StartTime.Before(ret[j].StartTime)
	})

	return ret, nil
}

func GetAllCurrentLessons(DatabasePool *utils.DatabasePool) ([]ActualLesson, error) {
	var returnActualLessons []ActualLesson

	getActualLessons, err := DatabasePool.Database.Prepare("SELECT id, group_lesson_id, start_time, end_time, " +
		"creation_time, edit_time, summary, description, location " +
		"FROM actual_lessons " +
		"WHERE start_time <= CURRENT_TIMESTAMP AND end_time >= CURRENT_TIMESTAMP")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer getActualLessons.Close()

	rows, err := getActualLessons.Query()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var newActualLesson ActualLesson
		rows.Scan(&newActualLesson.Id, &newActualLesson.GroupLessonId, &newActualLesson.StartTime,
			&newActualLesson.EndTime, &newActualLesson.CreationTime, &newActualLesson.EditTime, &newActualLesson.Summary, &newActualLesson.Description, &newActualLesson.Location)
		returnActualLessons = append(returnActualLessons, newActualLesson)
	}

	return returnActualLessons, err
}

func GetLessonsDetails(lessons []ActualLesson, pool *utils.DatabasePool) (LessonsRet, error) {
	var ret LessonsRet = LessonsRet{Lessons: lessons,
		GroupLessons: make([]GroupLesson, 0),
		Modules:      make([]Module, 0)}

	// Get Group lessons
	var seenModules map[string]string = make(map[string]string)
	var seenGrouops map[string]string = make(map[string]string)

	for i := 0; i < len(lessons); i++ {
		_, seen := seenGrouops[lessons[i].GroupLessonId]
		if !seen {
			seenGrouops[lessons[i].GroupLessonId] = lessons[i].GroupLessonId
			stmt, err := pool.Database.Prepare("select id, module_group_id, attendance_required, creation_time, edit_time, summary, description, location from group_lessons where id = $1;")
			if err != nil {
				log.Println(err)
				return ret, err
			}
			defer stmt.Close()

			rows, err := stmt.Query(lessons[i].GroupLessonId)
			if err != nil {
				log.Println(err)
				return ret, err
			}
			defer rows.Close()

			for rows.Next() {
				var gl GroupLesson
				rows.Scan(&gl.Id, &gl.ModuleGroupId, &gl.AttendanceRequired, &gl.CreationTime, &gl.EditTime, &gl.Summary, &gl.Description, &gl.Location)

				ret.GroupLessons = append(ret.GroupLessons, gl)
				seenModules[gl.ModuleGroupId] = gl.ModuleGroupId
			}
		}
	}

	// Get Modules
	for _, key := range seenModules {
		stmt, err := pool.Database.Prepare("select modules.id, modules.name, modules.external_id, modules.creation_time, modules.edit_time from modules, module_groups where module_groups.id = $1 and modules.id = module_groups.module_id;")
		if err != nil {
			log.Println(err)
			return ret, err
		}

		rows, err := stmt.Query(key)
		if err != nil {
			log.Println(err)
			return ret, err
		}

		for rows.Next() {
			var module Module
			rows.Scan(&module.Id, &module.Name, &module.ExternalId, &module.CreationTime, &module.EditTime)

			ret.Modules = append(ret.Modules, module)
		}
	}

	return ret, nil
}

func CreateRepeatingLesson(lesson RepeatingLesson, pool *utils.DatabasePool) error {

	// Search for group lesson with matching id
	query, err := pool.Database.Query("SELECT * FROM group_lessons WHERE id = $1;", lesson.GroupLessonId)
	if err != nil || !query.Next() {
		log.Println(err)
		return errors.New("no group lesson found with matching id")
	}
	defer query.Close()

	// create repeating lesson
	_, err = pool.Database.Exec("INSERT INTO repeating_lessons "+
		"(id, group_lesson_id, start_repeating, stop_repeating,"+
		"start_time, end_time, repeat_every, creation_time, edit_time)"+
		"VALUES ($1,$2,$3,$4,$5,$6,to_timestamp($7),$8,$9);",
		lesson.Id, lesson.GroupLessonId, lesson.StartRepeating, lesson.StopRepeating,
		lesson.StartTime, lesson.EndTime, lesson.RepeatEvery.Seconds(), lesson.CreationTime, lesson.EditTime)

	if err != nil {
		log.Println(err)
		return errors.New("issuing creating new repeating lesson")
	}

	return nil
}

func CreateGroupLesson(group *GroupLesson, pool *utils.DatabasePool) error {

	group.CreationTime = time.Now()
	group.EditTime = time.Now()
	group.Id = uuid.NewString()

	// Check group lesson exists
	query, err := pool.Database.Query("SELECT * FROM module_groups WHERE id = $1;", group.ModuleGroupId)
	if err != nil || !query.Next() {
		return errors.New("ot module group with matching id found")
	}
	defer query.Close()

	// Create lesson group
	_, err = pool.Database.Exec("INSERT INTO public.group_lessons "+
		"(id, module_group_id, \"name\", creation_time, "+
		"edit_time, attendance_required, description, \"location\", summary) "+
		"VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);",
		group.Id, group.ModuleGroupId, group.Name, group.CreationTime, group.EditTime,
		group.AttendanceRequired, group.Description, group.Location, group.Summary)

	if err != nil {
		log.Println(err)
		return errors.New("issue creating lesson group")
	}

	return nil
}
