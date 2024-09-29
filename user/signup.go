package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SignupUser(c *gin.Context) {
	var usersignup models.UserSignUp
	err := c.BindJSON(&usersignup)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  true,
			"message": "failed to bind request",
		})
		return
	}

	if err := helper.Validate(usersignup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var count1 int
	config.DB.Raw(`SELECT COUNT(*) FROM users where phone = ?`, usersignup.Phone).Scan(&count1)
	if count1 != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "mobile number alredy registered",
		})
		return
	}
	if usersignup.Password != usersignup.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "passwords does not match",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var count2 int
	err = config.DB.Raw(`SELECT COUNT(*) FROM user_login_methods WHERE user_login_method_email=?`, usersignup.Email).Scan(&count2).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to fetch data from database",
			"data":    gin.H{},
		})
		return

	}
	if count2 != 0 {
		var loginmethod string
		config.DB.Model(&models.UserLoginMethod{}).Where("user_login_method_email = ?", usersignup.Email).Pluck("login_method", &loginmethod)
		if loginmethod == "Google Authentication" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "login through google authentication",
				"data":    gin.H{},
			})
			return
		}
	}
	var count3 int
	err = config.DB.Raw(`SELECT COUNT(*) FROM users WHERE email=?`, usersignup.Email).Scan(&count3).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "failed to fetch data from database",
			"data":    gin.H{},
		})
		return
	}
	if count3 == 0 {
		Otp1 := helper.GenerateOTP()
		err := helper.SendOTPEmail(usersignup.Email, Otp1)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "failed to send otp"})
			return
		} else {
			Time := time.Now().Add(2 * time.Minute)
			var count4 int
			config.DB.Raw(`SELECT COUNT(*) FROM otps WHERE email=?`, usersignup.Email).Scan(&count4)
			if count4 > 0 {
				config.DB.Model(&models.OTP{}).Where("email = ?", usersignup.Email).Updates(models.OTP{OTP: Otp1, OtpExpiry: Time})
			} else {
				otp := models.OTP{
					Email:     usersignup.Email,
					OTP:       Otp1,
					OtpExpiry: Time,
				}
				config.DB.Create(&otp)
			}
			var count5 int
			config.DB.Raw(`SELECT COUNT(*) FROM temp_users WHERE email=?`, usersignup.Email).Scan(&count5)
			if count5 > 0 {
				config.DB.Model(&models.TempUser{}).Where("email = ?", usersignup.Email).Updates(models.TempUser{FirstName: usersignup.FirstName, LastName: usersignup.LastName, Password: usersignup.Password, Phone: usersignup.Phone})

			} else {
				User := models.TempUser{
					FirstName: usersignup.FirstName,
					LastName:  usersignup.LastName,
					Email:     usersignup.Email,
					Password:  usersignup.Password,
					Phone:     usersignup.Phone,
				}
				config.DB.Create(&User)
			}
			c.JSON(http.StatusOK, gin.H{"message": "otp generated successfully"})
			return
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "user alredy exists",
			"data":    gin.H{},
		})
		return
	}
}
