package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ResendOtp(c *gin.Context) {
	//Email := c.Param("email")

	// Retrieve the email from URL parameters
	rawEmail := c.Param("email")
	fmt.Println("Raw Email:", rawEmail)

	// Remove any leading or trailing colons
	sanitizedEmail := strings.Trim(rawEmail, ":")
	fmt.Println("Sanitized Email:", sanitizedEmail)

	Otp := helper.GenerateOTP()
	fmt.Println("", Otp)
	err := helper.SendOTPEmail(sanitizedEmail, Otp)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Failed to send otp"})
		return
	} else {
		Time := time.Now().Add(1 * time.Minute)
		otp := models.OTP{
			Email:     sanitizedEmail,
			OTP:       Otp,
			OtpExpiry: Time,
		}
		config.DB.Model(&models.OTP{}).Where("email = ?", sanitizedEmail).Updates(&otp)
		c.JSON(http.StatusOK, gin.H{"message": "Otp generated successfully"})
		return
	}

}
