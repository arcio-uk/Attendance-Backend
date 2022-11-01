package routes

import (
	"arcio/attendance-system/middleware"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

const (
	TIME_FORMAT = "15:04"      // HH:MM
	DATE_FORMAT = "2006-01-02" // YYYY-MM-DD
)

func addLessonRoutes(r *gin.Engine) {
	lessonRoutes := r.Group("/lesson")
	lessonRoutes.Use(middleware.CheckAuth(NonceManager))
	lessonRoutes.POST("/create-one-off", middleware.CheckPermissions(security.ModuleGroup, DatabasePool, security.PERMS_CAN_CREATE), CreateIndividualLessonHandler)
	lessonRoutes.POST("/create-repeating", middleware.CheckPermissions(security.ModuleGroup, DatabasePool, security.PERMS_CAN_CREATE), CreateRepeatingLessonHandler)
}

/*
 * Create single lesson.
 * Method: POST
 * URL: `/lesson/create-one-off`
 * Body Params: group-id, start-time, end-time
 */
func CreateIndividualLessonHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)
	requiredParams := []string{"group-lesson-id", "start-time", "end-time"}

	var body map[string]string
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	// Check all required parameters are passed in body
	for _, value := range requiredParams {
		if _, present := body[value]; !present {
			c.Error(errors.New("missing body parameter: " + value))
		}
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	sTime, err := time.Parse(time.RFC3339, body["start-time"])
	if err != nil {
		c.Error(errors.New("start-time date format error"))
	}

	eTime, err := time.Parse(time.RFC3339, body["end-time"])
	if err != nil {
		c.Error(errors.New("end-time date format error"))
	}

	if sTime.After(eTime) {
		c.Error(errors.New("start-time must be earlier than end-time"))
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	lesson := model.ActualLesson{
		StartTime:     sTime,
		EndTime:       eTime,
		GroupLessonId: body["group-lesson-id"],
	}

	err = model.CreateActualLesson(lesson, DatabasePool)
	if err != nil {
		log.Println(err)
		c.Error(errors.New("issue creating lesson"))
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
 * Create repeating lesson.
 * Method: POST
 * URL: `/lesson/create-repeating`
 * Body Params: group-lesson-id, start-repeating, stop-repeating,
 *	start-time, end-time, repeat-every
 */
func CreateRepeatingLessonHandler(c *gin.Context) {
	claims := c.MustGet("claims").(security.Claims)
	requiredParams := []string{"group-lesson-id", "start-repeating", "stop-repeating",
		"start-time", "end-time", "repeat-every"}

	// Interface used instead of corresponding struct to
	// give more control over how values are provided in json
	var body map[string]interface{}
	err := c.ShouldBindJSON(&body)
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	// Check all required parameters are passed in body
	for _, value := range requiredParams {
		if _, present := body[value]; !present {
			c.Error(errors.New("missing body parameter: " + value))
		}
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	// Check dates and time meet required format
	sRepeating, err := time.Parse(DATE_FORMAT, body["start-repeating"].(string))
	if err != nil {
		c.Error(errors.New("start-repeating time format error"))
	}

	eRepeating, err := time.Parse(DATE_FORMAT, body["stop-repeating"].(string))
	if err != nil {
		c.Error(errors.New("stop-repeating time format error"))
	}

	sTime, err := time.Parse(TIME_FORMAT, body["start-time"].(string))
	if err != nil {
		c.Error(errors.New("start-time date format error"))
	}

	eTime, err := time.Parse(TIME_FORMAT, body["end-time"].(string))
	if err != nil {
		c.Error(errors.New("end-time date format error"))
	}

	// Parse from seconds
	repeatInterval, err := time.ParseDuration(fmt.Sprintf("%ds",
		int(body["repeat-every"].(float64))))
	if err != nil {
		c.Error(errors.New("repeat-every format error"))
	}

	if sRepeating.After(eRepeating) {
		c.Error(errors.New("start-repeating must be earlier than stop-repeating"))
	}

	if sTime.After(eTime) {
		c.Error(errors.New("start-time must be earlier than end-time"))
	}

	if len(c.Errors) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": c.Errors,
		})
		return
	}

	lesson := model.RepeatingLesson{
		Id:             uuid.NewString(),
		GroupLessonId:  body["group-lesson-id"].(string),
		StartRepeating: sRepeating,
		StopRepeating:  eRepeating,
		StartTime:      sTime,
		EndTime:        eTime,
		CreationTime:   time.Now(),
		EditTime:       time.Now(),
		RepeatEvery:    repeatInterval,
	}

	if err = model.CreateRepeatingLesson(lesson, DatabasePool); err != nil {
		c.Error(err)
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
