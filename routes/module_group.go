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

func addModuleGroupRoutes(r *gin.Engine) {
	moduleGroupRoutes := r.Group("/module/group")
	moduleGroupRoutes.Use(middleware.CheckAuth(NonceManager))
	moduleGroupRoutes.POST("/add", middleware.CheckPermissions(security.Module, DatabasePool, security.PERMS_CAN_CREATE), CreateModuleGroupHandler)
	moduleGroupRoutes.POST("/add-user", middleware.CheckPermissions(security.ModuleGroup, DatabasePool, security.PERMS_CAN_UPDATE), AddUserToModuleGroupHandler)
	moduleGroupRoutes.POST("/add-users", middleware.CheckPermissions(security.ModuleGroup, DatabasePool, security.PERMS_CAN_UPDATE), AddUsersToModuleGroupHandler)
	moduleGroupRoutes.GET("/get", middleware.CheckPermissions(security.Module, DatabasePool, security.PERMS_CAN_READ), GetGroupsForModuleHandler)
	moduleGroupRoutes.DELETE("/rm-user", middleware.CheckPermissions(security.ModuleGroup, DatabasePool, security.PERMS_CAN_UPDATE), RemoveFromModuleGroupHandler)
	moduleGroupRoutes.GET("/users", middleware.CheckPermissions(security.ModuleGroup, DatabasePool, security.PERMS_CAN_READ_ALL), GetModuleGroupUsersHandler)
}

/*
 * Create new module group.
 * Method: POST
 * URL: `/module/group/add`
 * Body Params: name, module-id
 */
func CreateModuleGroupHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var newModuleGroup model.ModuleGroup
	err := c.ShouldBindJSON(&newModuleGroup)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err = model.CreateModuleGroup(&newModuleGroup, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue creating module group"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	utils.UpdateLogs(claims, c.Request, DatabasePool, GlobalConfig)

	c.JSON(http.StatusCreated, newModuleGroup)
}

/*
 * Add user to module group.
 * Method: POST
 * URL: `module/group/add-user`
 * Body Params: module-group-id, user-id
 */
func AddUserToModuleGroupHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var body model.ModuleGroupEdit
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err = model.AddUserToModuleGroup(body.UserId, body.ModuleGroupId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("failed to add user to module group"))
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
 * Add multiple users to module group.
 * Method: GET
 * URL: `/module/group/add-users`
 * Body params: module-group-id, user-ids
 */
func AddUsersToModuleGroupHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var body model.ModuleGroupBulkEdit
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
	}

	err = model.AddUsersToModuleGroup(body.UserIds, body.ModuleGroupId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue adding users to module group"))
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
 * Get module groups for given module.
 * Method: GET
 * URL: `/module/group/get`
 * Query params: moduleId
 */
func GetGroupsForModuleHandler(c *gin.Context) {
	moduleId, exists := c.GetQuery("moduleId")
	if !exists {
		c.Error(errors.New("missing query parameter moduleId"))
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	moduleGroups, err := model.GetGroupsForModule(moduleId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue getting module groups"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, moduleGroups)
}

/*
 * Remove user from module group.
 * Method: Delete
 * URL: `/module/group/rm-user`
 * Body Params: module-group-id, user-id
 */
func RemoveFromModuleGroupHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	var body model.ModuleGroupEdit
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	err = model.RemoveUserFromModuleGroup(body.UserId, body.ModuleGroupId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("Issue removing user from module"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
	}

	utils.UpdateLogs(claims, c.Request, DatabasePool, GlobalConfig)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

/*
 * Get all users in given module group.
 * Method: GET
 * URL: `module/group/users`
 * Params: moduleGroupId
 */
func GetModuleGroupUsersHandler(c *gin.Context) {
	moduleGroupId, exists := c.GetQuery("moduleGroupId")
	if !exists {
		c.Error(errors.New("missing required query parameter moduleGroupId"))
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	users, err := model.GetModuleGroupStudents(moduleGroupId, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("Cannot get module groups"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, users)
}
