package model

import (
	"time"
)

// Users
type User struct {
	InternalId   string    `json:"id"`
	ExternalId   string    `json:"external-id"`
	Fname        string    `json:"firstname"`
	Sname        string    `json:"surname"`
	Email        string    `json:"email"`
	CreationTime time.Time `json:"creation-time,omitempty"`
	EditTime     time.Time `json:"edit-time,omitempty"`
}

// Courses
type Module struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	ExternalId   string    `json:"external-id"`
	CreationTime time.Time `json:"creation-time,omitempty"`
	EditTime     time.Time `json:"edit-time,omitempty"`
}

type ModuleUser struct {
	Id           string    `json:"id"`
	UserId       string    `json:"user-id"`
	ModuleId     string    `json:"module-id"`
	RoleId       string    `json:"role-id"`
	CreationTime time.Time `json:"creation-time,omitempty"`
	EditTime     time.Time `json:"edit-time,omitempty"`
}

// Group
type ModuleGroup struct {
	Id           string    `json:"id"`
	ModuleId     string    `json:"module-id,omitempty"`
	Name         string    `json:"name"`
	CreationTime time.Time `json:"creation-time,omitempty"`
	EditTime     time.Time `json:"edit-time,omitempty"`
}

// Group Lessons
type GroupLesson struct {
	Id                 string    `json:"id"`
	ModuleGroupId      string    `json:"module-group-id"`
	AttendanceRequired bool      `json:"attendance-required"`
	CreationTime       time.Time `json:"creation-time,omitempty"`
	EditTime           time.Time `json:"edit-time,omitempty"`
	Name               string    `json:"name"`
	Summary            string    `json:"summary"`
	Description        string    `json:"description"`
	Location           string    `json:"location"`
}

// Lessons
type ActualLesson struct {
	Id            string    `json:"id"`
	GroupLessonId string    `json:"group-lesson-id"`
	StartTime     time.Time `json:"start-time"`
	EndTime       time.Time `json:"end-time"`
	CreationTime  time.Time `json:"creation-time"`
	EditTime      time.Time `json:"edit-time"`
	Summary       string    `json:"summary"`
	Description   string    `json:"description"`
	Location      string    `json:"location"`
	IsAbstract    bool      `json:"is-abstract,omitempty"`
}

type RepeatingLesson struct {
	Id              string        `json:"id,omitempty"`
	GroupLessonId   string        `json:"group-lesson-id"`
	StartRepeating  time.Time     `json:"start-repeating"`             // Date
	StopRepeating   time.Time     `json:"stop-repeating"`              // Date
	StartTime       time.Time     `json:"start-time"`                  // Time
	EndTime         time.Time     `json:"end-time"`                    // Time
	CreationTime    time.Time     `json:"creation-time,omitempty"`     // Datetime
	EditTime        time.Time     `json:"edit-time"`                   // Datetime
	LastSpawnedTime time.Time     `json:"last-spawned-time,omitempty"` // Datetime
	RepeatEvery     time.Duration `json:"repeat-every"`
}

// Attendance
type LessonAttendance struct {
	LessonId     string
	UserId       string
	RegisterTime time.Time
}

// Ret Structures
type LessonsRet struct {
	Lessons      []ActualLesson `json:"lessons"`
	GroupLessons []GroupLesson  `json:"group-lessons"`
	Modules      []Module       `json:"modules"`
}

type AttendanceRecord struct {
	MarkedSessions int `json:"marked-sessions"`
	TotalSessions  int `json:"total-sessions"`
}

type AttendanceRet struct {
	ModuleAttendance map[string]AttendanceRecord `json:"module-attendance"`
}

type ModuleUsersRet struct {
	ModuleUsers     []User                      `json:"module-users"`
	UsersAttendance map[string]AttendanceRecord `json:"users-attendance"`
}

type UserAttendanceRet struct {
	UserId     string           `json:"id"`
	Attendance AttendanceRecord `json:"attendance"`
}

type ModuleUserAttendanceRet struct {
	InternalId   string           `json:"id"`
	ExternalId   string           `json:"external-id"`
	Fname        string           `json:"firstname"`
	Sname        string           `json:"surname"`
	Email        string           `json:"email"`
	CreationTime time.Time        `json:"creation-time,omitempty"`
	EditTime     time.Time        `json:"edit-time,omitempty"`
	Attendance   AttendanceRecord `json:"attendance"`
}

type ModuleUserEdit struct {
	UserId   string `json:"user-id"`
	ModuleId string `json:"module-id"`
}

type ModuleUserBulkEdit struct {
	ModuleId string   `json:"module-id"`
	Users    []string `json:"user-ids"`
}

type ModuleGroupEdit struct {
	ModuleGroupId string `json:"module-group-id"`
	UserId        string `json:"user-id"`
}

type ModuleGroupBulkEdit struct {
	ModuleGroupId string   `json:"module-group-id"`
	UserIds       []string `json:"user-ids"`
}

type CalenderJwtRet struct {
	IcalJwt string `json:"ical-jwt"`
}

type MarkAttendanceJson struct {
	LessonId string `json:"lesson-id"`
}

type ModuleGroupRet struct {
	UserAttendance []ModuleUserAttendanceRet `json:"users"`
	ModuleName     string                    `json:"module-group-name"`
	ModuleGroupId  string                    `json:"module-group-id"`
}
