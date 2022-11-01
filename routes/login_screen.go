/*
 * login_screen.go Provides information about user when first logging in.
 * Author: John Costa, Isaac George, Danny Piper
 */

package routes

import (
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type LoginScreen struct {
	Courses         []model.Module       `json:"modules"`
	UpcomingLessons []model.ActualLesson `json:"upcoming-lessons"`
}

/*
 * Returns users courses and upcoming lessons
 * Method: GET
 * URL: `/login-screen`
 */
func LoginScreenHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	courses, err := model.GetUsersModules(claims.Uuid, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue getting courses"))
	}

	lessons, err := model.GetLessons(claims.Uuid, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issues getting lessons"))
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, LoginScreen{
		Courses:         courses,
		UpcomingLessons: lessons,
	})
}
