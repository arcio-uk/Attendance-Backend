/*
 * router.go is used for gin router initialisation.
 * Author: John Costa, Isaac George
 */

package routes

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/middleware"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"github.com/easonlin404/limit"
	"github.com/gin-gonic/gin"
)

var DatabasePool *utils.DatabasePool
var GlobalConfig *config.Config
var NonceManager *security.NonceManager

/*
 * Initialises the router and its routes & groups, and the nonce
 */
func InitRouter() *gin.Engine {
	router := gin.New()

	//Setups recovery to catch `fatals`, and initialises logger.
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.Use(middleware.CheckPreflight())
	router.Use(limit.Limit(1000))

	router.GET("/get-nonce", GetNonceHandler)
	router.GET("/status", StatusHandler)
	router.GET("/check-auth", middleware.CheckAuth(NonceManager), CheckAuthHandler)
	router.GET("/login-screen", middleware.CheckAuth(NonceManager), LoginScreenHandler)

	addUserRoutes(router)
	addModuleRoutes(router)
	addModuleGroupRoutes(router)
	addTimetableRoutes(router)
	addAttendanceRoutes(router)
	addLessonRoutes(router)

	return router
}
