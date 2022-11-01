/*
 * misc.go stores miscellanious endpoints.
 * Author: John Costa, Isaac George, Danny Piper
 */

package routes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

/*
 * Method to get nonces from manager and return them to user.
 * Method: GET
 * URL: `/get-nonce`
 */
func GetNonceHandler(c *gin.Context) {
	nonce, err := NonceManager.GetNonce()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Server internal error getting nonce",
		})
	}

	//nonce is converted to float64, as json only supports that numeric type
	c.JSON(http.StatusOK, gin.H{
		"nonce": strconv.FormatInt(nonce, 10),
	})
}

/*
 * Returns status of server running.
 * Method: GET
 * URL: `/status`
 */
func StatusHandler(c *gin.Context) {
	c.String(http.StatusOK, "online")
}

/*
 * Returns if requesting client is authenticated.
 * Method: GET
 * URL: `/check-auth`
 */
func CheckAuthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
