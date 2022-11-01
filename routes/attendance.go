/*
 * login_screen.go handler attendance for student and lecturers.
 * Author: John Costa
 */

package routes

import (
	"arcio/attendance-system/middleware"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func addAttendanceRoutes(rg *gin.Engine) {
	attendanceRoutes := rg.Group("/attendance")
	attendanceRoutes.Use(middleware.CheckAuth(NonceManager))

	attendanceRoutes.POST("/mark", middleware.CheckPermissions(security.Attendance, DatabasePool, security.PERMS_CAN_READ, security.PERMS_CAN_CREATE), PostMarkAttendance)
	attendanceRoutes.POST("/lecturer/mark", middleware.CheckPermissions(security.Attendance, DatabasePool, security.PERMS_CAN_READ_ALL, security.PERMS_CAN_CREATE), PostLecturerMarkAttendance)

}

type PostMarkAttendanceBody struct {
	LessonId string `json:"lesson-id"`
}

type LecturerMarkAttendance struct {
	LessonId string `json:"lesson-id"`
	UserId   string `json:"user-id"`
}

/*
 * Mark the attendance for a user, for a lesson.
 * Method: POST
 * URL: `/attendance/mark`
 */
func PostMarkAttendance(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var body PostMarkAttendanceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err := model.RegisterAttendance(claims.Uuid, body.LessonId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	utils.UpdateLogs(claims, c.Request, DatabasePool, GlobalConfig)
	c.Status(http.StatusOK)
}

/*
 * Mark the attendance for a user, for a lesson, by a lecturer.
 * Method: POST
 * URL: `/attendance/lecturer/mark`
 * Body Params: user-id, lesson-id
 */
func PostLecturerMarkAttendance(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var body LecturerMarkAttendance
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err := model.RegisterAttendance(body.UserId, body.LessonId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	utils.UpdateLogs(claims, c.Request, DatabasePool, GlobalConfig)
	c.Status(http.StatusCreated)
}
