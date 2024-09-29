package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/responsemodels"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Listuser(c *gin.Context) {
	var userlist []responsemodels.User
	qry := `SELECT * FROM users`
	qr := config.DB.Raw(qry).Scan(&userlist)
	if qr.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "false",
			"message": "failed to fetch data from database, or data does not exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully fetched user details",
		"data": gin.H{
			"users": userlist,
		},
	})

}
