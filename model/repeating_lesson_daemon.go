package model

import (
	"arcio/attendance-system/utils"
	"errors"
	"log"
	"time"
)

/*
* This file creates an actual lesson when a repeating lesson is about to happen.
* It does this by polling the database for new repeating lessons.
 */

func GetDaemonRepeatingLessons(pool *utils.DatabasePool) ([]RepeatingLesson, error) {
	rows, err := pool.Database.Query("SELECT (id, group_lesson_id, " +
		"last_spawned_time, start_repeating, " +
		"stop_repeating, start_time, end_time, repeat_every) " +
		"FROM repeating_lessons " +
		"WHERE start_repeating <= CURRENT_TIMESTAMP " +
		"AND stop_repeating >= CURRENT_TIMESTAMP;")

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	repeatingLessons := make([]RepeatingLesson, 0)

	for rows.Next() {
		var lesson RepeatingLesson
		rows.Scan(&lesson.Id,
			&lesson.GroupLessonId,
			&lesson.LastSpawnedTime,
			&lesson.StartRepeating,
			&lesson.StopRepeating,
			&lesson.RepeatEvery,
			&lesson.StartTime,
			&lesson.EndTime)

		repeatingLessons = append(repeatingLessons, lesson)
	}

	return repeatingLessons, nil
}

/*
* Checks if any repeating lesson needs to be spawned.
* Returns an ActualLesson or an error which indicates that the
* lesson does not need to be spawned.
 */
func __CheckLessonHappening(currentTime time.Time, lesson RepeatingLesson) (ActualLesson, error) {
	if lesson.StopRepeating.Unix() < currentTime.Unix() {
		return ActualLesson{}, errors.New("This lesson has stopped repeating")
	}

	var nullLessonTime time.Time
	if lesson.LastSpawnedTime == nullLessonTime {
		return ActualLesson{}, errors.New("This lesson was already made")
	}

	startTime := lesson.StartTime.Unix()
	endTime := lesson.EndTime.Unix()
	startRepeating := lesson.StartRepeating.Unix()
	interval := int64(lesson.RepeatEvery)

	numberOfIntervals := (currentTime.Unix() - startTime - startRepeating) / interval

	//If the lesson start is bigger than current time and smaller than end time
	//Then we know that the lesson is happening right now.
	if startRepeating+(numberOfIntervals*interval)+startTime <= currentTime.Unix() &&
		startRepeating+(numberOfIntervals*interval)+endTime >= currentTime.Unix() {
		newActualLesson := ActualLesson{GroupLessonId: lesson.GroupLessonId,
			StartTime: lesson.StartTime,
			EndTime:   lesson.EndTime}

		return newActualLesson, nil
	}

	return ActualLesson{}, errors.New("Lesson is not occuring")
}

func CheckLessonHappening(lesson RepeatingLesson) (ActualLesson, error) {
	return __CheckLessonHappening(time.Now(), lesson)
}

// Poll every second
const REPEATING_LESSON_POLL_TIME = time.Millisecond * 1000

func StartLessonSpawnDaemon(pool *utils.DatabasePool) {
	defer log.Println("Lesson spawner daemon crashed")
	for true {
		lessons, err := GetDaemonRepeatingLessons(pool)
		if err != nil {
			log.Println(err)
		} else {
			for i := 0; i < len(lessons); i++ {
				spawn, err := CheckLessonHappening(lessons[i])
				if err == nil {
					log.Printf("Creating new lesson %s %s\n", spawn.Summary, spawn.Description)
					CreateActualLesson(spawn, pool)
				}
			}
		}

		time.Sleep(REPEATING_LESSON_POLL_TIME)
	}
}
