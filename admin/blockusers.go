package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func BloskUser(c *gin.Context) {
	var blockuser models.BlockUser
	err := c.BindJSON(&blockuser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "request failed to bind",
		})
		return
	}
	if err := helper.Validate(blockuser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	var usercount int
	//qry := `SELECT COUNT(*) FROM users WHERE id = ?`
	config.DB.Raw(`SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, blockuser.UserID).Scan(&usercount)
	if usercount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "user does not exist",
		})
		return
	}
	config.DB.Model(&models.User{}).Where("id = ?", blockuser.UserID).Update("Status", "Blocked")
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "blocked user"})

}
