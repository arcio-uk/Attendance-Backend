package routes

import (
	"arcio/attendance-system/middleware"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"time"
)

const ICAL_JWT_EXPIRY = 5 * 365 * 24 * 60 * 60 // Five years in seconds

func addTimetableRoutes(r *gin.Engine) {
	timetableRoutes := r.Group("/timetable")

	timetableRoutes.GET("/ical", middleware.CheckIcalAuth(GlobalConfig), CalenderExportHandler)
	timetableRoutes.GET("/get-timetable-jwt", middleware.CheckAuth(NonceManager), CalenderJwtHandler)
	timetableRoutes.GET("/upcoming-lessons", middleware.CheckAuth(NonceManager), UpcomingLessonsHandler)
	timetableRoutes.GET("/happening-now", middleware.CheckAuth(NonceManager), GetActiveLessonsHandler)
}

/*
 * Provide lesson that student can import to their calendar.
 * Method: GET
 * URL: `/timetable/ical`
 */
func CalenderExportHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	lessons, err := model.GetLessons(claims.Uuid, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue getting lessons"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	icalLessons := model.ExportLessonsAsIcal(lessons)
	c.String(http.StatusOK, icalLessons)
}

/*
 * Returns Jwt for accessing calendar
 * Method: GET
 * URL: `/timetable/get-timetable-jwt`
 */
func CalenderJwtHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	// Create a test JWT
	// HS512 is the standard we use.
	nclaims := security.Claims{
		Uuid:      claims.Uuid,
		Type:      "ICAL",
		Firstname: claims.Firstname,
		Surname:   claims.Surname,
		Iat:       time.Now().Unix(),
		Exp:       time.Now().Unix() + ICAL_JWT_EXPIRY,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, nclaims)
	signedToken, err := token.SignedString([]byte(GlobalConfig.JwtIcalSecret))
	if err != nil {
		log.Println(err)
		c.Error(errors.New("cannot get jwt token"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.String(http.StatusOK, signedToken)
}

/*
 * Returns the upcoming lessons for the user
 * Method: GET
 * URL: `/timetable/upcoming-lessons
 */
func UpcomingLessonsHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)

	lessons, err := model.GetLessons(claims.Uuid, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	lessonDetail, err := model.GetLessonsDetails(lessons, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, lessonDetail)
}

/*
 * Return currently active lessons.
 * Method: GET
 * URL: `/timetable/happening-now`
 */
func GetActiveLessonsHandler(c *gin.Context) {
	activeLessons, err := model.GetAllCurrentLessons(DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue getting active lessons"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"errors": c.Errors,
		})
	}

	c.JSON(http.StatusOK, activeLessons)
}
