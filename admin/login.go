package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {

	var Adminlogin models.AdminLogin
	err := c.BindJSON(&Adminlogin)
	response := gin.H{
		"status":  false,
		"message": "failed to bind request",
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	err = helper.Validate(Adminlogin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": err.Error(),
			"data":    gin.H{},
		})
		return
	}

	var Admin models.Admin
	qr := config.DB.Where("email =?", Adminlogin.Email).First(&Admin)
	if qr.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid credentials",
			"data":    gin.H{},
		})
		return
	}
	var password string
	config.DB.Model(&models.Admin{}).Where("email = ? AND deleted_at IS NULL", Adminlogin.Email).Pluck("password", &password)
	if password != Adminlogin.Password {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "invalid email or password",
			"data":    gin.H{},
		})
		return
	}
	var id uint
	config.DB.Model(&models.Admin{}).Where("email = ?", Adminlogin.Email).Pluck("id", &id)
	token, err := midleware.GenerateJWT("admin", Adminlogin.Email, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	fmt.Println("", token)

	c.Header("Authorization", "Bearer "+token)

	c.JSON(http.StatusOK, gin.H{"message": "Admin successfully logged in", "token": token})

}
