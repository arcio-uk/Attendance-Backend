/*
 * module.go contains handlers for endpoints under `/module`.
 * `/module` handles interactions concerned with modules.
 * Author: Isaac George, Danny, Piper
 */
package routes

import (
	"arcio/attendance-system/middleware"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func addModuleRoutes(r *gin.Engine) {
	moduleRoutes := r.Group("/module")
	moduleRoutes.Use(middleware.CheckAuth(NonceManager))
	moduleRoutes.POST("/add", middleware.CheckPermissions(security.Global, DatabasePool, security.PERMS_CAN_CREATE), CreateModuleHandler)
	moduleRoutes.GET("/get", middleware.CheckPermissions(security.Global, DatabasePool, security.PERMS_CAN_READ), GetModulesHandler)
	moduleRoutes.GET("/get-users", middleware.CheckPermissions(security.Module, DatabasePool, security.PERMS_CAN_READ_ALL), GetModuleUsersHandler)
	moduleRoutes.POST("/add-user", middleware.CheckPermissions(security.Module, DatabasePool, security.PERMS_CAN_UPDATE), AddUserToModuleHandler)
	moduleRoutes.POST("/add-users", middleware.CheckPermissions(security.Module, DatabasePool, security.PERMS_CAN_UPDATE), AddUsersToModuleHandler)
	moduleRoutes.DELETE("/rm-user", middleware.CheckPermissions(security.Module, DatabasePool, security.PERMS_CAN_UPDATE), RemoveUserFromModuleHandler)
}

/*
 * Create new module
 * Method: POST
 * URL: `/module/add`
 * Body Params: name, external-id
 */
func CreateModuleHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var newModule model.Module
	err := c.ShouldBindJSON(&newModule)

	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err = model.CreateModule(&newModule, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	utils.UpdateLogs(claims, c.Request, DatabasePool, GlobalConfig)

	c.JSON(http.StatusCreated, newModule)
}

/*
 * Gets all modules.
 * Method: GET
 * URL: `/module/get`
 */
func GetModulesHandler(c *gin.Context) {
	modules, err := model.GetAllModules(DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(err)
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
 * Returns list of users in a given module.
 * Method: GET
 * URL: `/module/get-users`
 * Query Params: moduleId
 */
func GetModuleUsersHandler(c *gin.Context) {
	moduleId, exists := c.GetQuery("moduleId")
	if !exists {
		c.Error(errors.New("moduleId parameter is missing"))
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	moduleUsers, err := model.GetModuleStudents(moduleId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(err)
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, moduleUsers)
}

/*
 * Add single user to module.
 * Method: GET
 * PATH: `/module/add-user`
 * Body Params: user-id, module-id
 */
func AddUserToModuleHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)
	var input model.ModuleUserEdit
	err := c.ShouldBindJSON(&input)

	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err = model.AddUserToModule(input.UserId, input.ModuleId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue adding user to module"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	utils.UpdateLogs(claims, c.Request, DatabasePool, GlobalConfig)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
	})
}

/*
 * Add multiple users to module.
 * Method: GET
 * PATH: `/module/add-users`
 * Body Params: user-ids, module-id
 */
func AddUsersToModuleHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var body model.ModuleUserBulkEdit
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err = model.AddUsersToModule(body.Users, body.ModuleId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue adding users to module"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	utils.UpdateLogs(claims, c.Request, DatabasePool, GlobalConfig)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
	})
}

/*
 * Remove user from module.
 * Method: DELETE
 * URL: `/module/rm-user`
 * Body Params: user-id, module-id
 */
func RemoveUserFromModuleHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var body model.ModuleUserEdit
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err = model.RemoveUserFromModule(body.UserId, body.ModuleId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue adding user to module"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	utils.UpdateLogs(claims, c.Request, DatabasePool, GlobalConfig)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
