package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserLogin(c *gin.Context) {
	var UserLogin models.UserLogin
	if err := c.BindJSON(&UserLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bund data",
			"data":    gin.H{},
		})
		return
	}

	err := helper.Validate(UserLogin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
			"data":    gin.H{},
		})
		return
	}
	var usercount int
	err = config.DB.Raw(`SELECT COUNT(*) FROM users WHERE email=?`, UserLogin.Email).Scan(&usercount).Error
	if err != nil {
		fmt.Println("failed to execute query")

	}
	if usercount == 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Invalid email or password"})
		return
	} else {
		fmt.Println("user exists")
	}

	var password, status string
	config.DB.Model(&models.User{}).Where("email = ?", UserLogin.Email).Pluck("password", &password)
	config.DB.Model(&models.User{}).Where("email = ?", UserLogin.Email).Pluck("status", &status)
	if password != UserLogin.Password {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid email or password",
			"data":    gin.H{},
		})
		return

	}
	if status == "Blocked" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "user is blocked by admin",
			"data":    gin.H{},
		})
		return

	}
	var id uint
	config.DB.Model(&models.User{}).Where("email = ?", UserLogin.Email).Pluck("id", &id)
	token, err := midleware.GenerateJWT("user", UserLogin.Email, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}
	fmt.Println(token)
	c.Header("Authorization", "Bearer "+token)
	c.JSON(http.StatusOK, gin.H{"message": "User Login successful", "token": token})
	//vallet part

}
