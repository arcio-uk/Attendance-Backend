package model

import (
	"github.com/arran4/golang-ical"
)

func ExportLessonsAsIcal(lessons []ActualLesson) string {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetName("arcio timetable, CHANGE ME")
	cal.SetXWRTimezone("Europe/London")
	cal.SetProductId("arcio")

	for i := 0; i < len(lessons); i++ {
		lesson := lessons[i]

		event := cal.AddEvent(lesson.Id)
		event.SetCreatedTime(lesson.CreationTime)
		event.SetDtStampTime(lesson.CreationTime)
		event.SetModifiedAt(lesson.EditTime)
		event.SetStartAt(lesson.StartTime)
		event.SetEndAt(lesson.EndTime)
		event.SetSummary(lesson.Summary)
		event.SetLocation(lesson.Location)
		event.SetDescription(lesson.Description)
	}

	return cal.Serialize()
}
