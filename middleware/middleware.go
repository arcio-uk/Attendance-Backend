/*
 * preflight.go Contains middleware that checks requests
 * meet requirements before they can run.
 * Author: John Costa, Isaac George
 */

package middleware

import (
	"arcio/attendance-system/config"
	"arcio/attendance-system/model"
	"arcio/attendance-system/security"
	"arcio/attendance-system/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

const (
	NONCE_HEADER     = "Nonce"
	JWT_HEADER       = "Authorization"
	CLAIMS_KEY       = "claims"
	AUTH_HEADER      = "Authorization"
	NOT_AUTHORISED   = "Not authorised"
	ICAL_AUTH_HEADER = "ical-auth"
)

// Standard error messages
const (
	INTERNAL_SERVER_ERROR_MSG = "Internal server error"
	UNAUTHORISED_ERROR_MSG    = "Unauthorised"
)

var GlobalConfig *config.Config

/*
 * Validates requests have appropriate headers.
 */
func Validate(c *gin.Context) error {
	//If nonces are enabled - Get the nonce header
	if GlobalConfig.NonceToggle && c.GetHeader(NONCE_HEADER) == "" {
		return errors.New("`Nonce` was missing from the request's header")
	} else if c.GetHeader(JWT_HEADER) == "" {
		return errors.New("`Authorization` was missing from the request's header")
	}
	return nil
}

/*
 * Ensures request isn't OPTIONS, then sets appropriate headers.
 */
func CheckPreflight() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
			return
		}

		c.Next()
	}
}

/*
 * Checks that requests has valid Nonce and JWT Token.
 */
func CheckAuth(NonceManager *security.NonceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := Validate(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if GlobalConfig.NonceToggle {
			nonceHeader, err := strconv.ParseInt(c.GetHeader(NONCE_HEADER), 10, 64)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Nonce must be a number.",
				})
				return
			}

			err = NonceManager.UseNonce(nonceHeader)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Nonce is invalid.",
				})
				return
			}
		}

		token := c.GetHeader(JWT_HEADER)
		claims, err := security.CheckHeaderJwt(token, *GlobalConfig)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid JWT",
			})
			return
		}
		c.Set(CLAIMS_KEY, claims)
		c.Next()
	}
}

func CheckPermissions(layer security.Layer, database *utils.DatabasePool, required ...security.Overrides) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get target ids
		moduleid := ""
		groupid := ""

		if c.Request.Method == "GET" {
			// Read parameters if GET
			id, exists := c.GetQuery("moduleGroupId")
			if exists {
				groupid = id
			}

			id, exists = c.GetQuery("moduleId")
			if exists {
				moduleid = id
			}
		} else {
			// Otherwise read body
			var ids struct {
				ModuleId      string `json:"module-id,omitempty"`
				ModuleGroupId string `json:"modu:le-group-id,omitempty"`
			}

			err := c.ShouldBindJSON(&ids)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Bad request govna",
				})
			}

			moduleid = ids.ModuleId
			groupid = ids.ModuleGroupId
		}

		// Get user id
		claims_tmp, found := c.Get(CLAIMS_KEY)
		if !found {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": INTERNAL_SERVER_ERROR_MSG,
			}) // They should be here as this middleware is chained init bruv
			return
		}
		claims := claims_tmp.(security.Claims)
		id := claims.Uuid

		// Get user perms
		perms, err := model.GetPermissions(id, moduleid, groupid, layer, database)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": INTERNAL_SERVER_ERROR_MSG,
			})
			return
		}

		// Test permissions
		valid := security.CheckPerms(perms, required...)
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": UNAUTHORISED_ERROR_MSG,
			})
			return
		}

		c.Next()
	}
}

func CheckIcalAuth(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var claims security.Claims
		var err error = nil

		token := c.GetHeader(AUTH_HEADER)
		if token != "" {
			claims, err = security.CheckHeaderJwt(token, *config)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		} else {
			token, present := c.GetQuery(ICAL_AUTH_HEADER)
			if !present {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "missing auth header",
				})
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			claims, err = security.CheckIcalJwt(token, *config)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		c.Writer.Header().Set("Content-Type", "application/Calender")
		c.Writer.Header().Set("Content-Disposition", "attatchment; filename=\"timetable.ics\"")
		c.Set("claims", claims)
		c.Next()
	}
}
