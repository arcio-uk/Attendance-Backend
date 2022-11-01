/*
 * user.go contains handlers for endpoints under `/user`.
 * Author: Isaac George, Danny, Piper
 */

package routes

import (
	"arcio/attendance-system/middleware"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func addUserRoutes(rg *gin.Engine) {
	userRoutes := rg.Group("/user")
	userRoutes.Use(middleware.CheckAuth(NonceManager))
	userRoutes.GET("/get", middleware.CheckPermissions(security.Global, DatabasePool, security.PERMS_CAN_READ), GetUsersHandler)
	userRoutes.GET("/get-modules", middleware.CheckPermissions(security.Global, DatabasePool, security.PERMS_CAN_READ), GetUserModulesHandler)
	userRoutes.GET("/module-percentage", middleware.CheckPermissions(security.Module, DatabasePool, security.PERMS_CAN_READ), middleware.CheckPermissions(security.Attendance, DatabasePool, security.PERMS_CAN_READ), GetUserAttendance)
}

/*
 * Get all users.
 * Method: GET
 * URL: `/user/get`
 */
func GetUsersHandler(c *gin.Context) {
	users, err := model.GetUsers(DatabasePool)
	if err != nil {
		c.Error(errors.New("issue getting users from database"))
		log.Println(err)
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

/*
 * Get modules for given user.
 * Method: GET
 * URL: `/user/get-modules`
 */
func GetUserModulesHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	modules, err := model.GetModulesForUser(claims.Uuid, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("cannot get modules for user"))
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, modules)
}

/*
 * Returns for each module the user is
 * Method: GET
 * URL: `/user/module-percentage`
 */
func GetUserAttendance(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	attendance, err := model.GetStudentAttendancePercentages(claims.Uuid, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("unable to get student attendance"))
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, attendance)
}
