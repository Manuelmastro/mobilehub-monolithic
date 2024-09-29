/*package user



import (
	"MOBILEHUB/config"
	"MOBILEHUB/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)*/

/*func VerifyotpWindow(c *gin.Context) {
	fmt.Println("Hi")
	Email := c.Param("email")
	var VerifyOTP models.VerifyOTP
	if err := c.BindJSON(&VerifyOTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("", VerifyOTP.Otp)
	fmt.Println("hello")
	var otp string
	config.DB.Model(&models.OTP{}).Where("email = ?", Email).Pluck("otp", &otp)
	var otptime time.Time
	config.DB.Model(&models.OTP{}).Where("email = ?", Email).Pluck("otp_expiry", &otptime)
	if VerifyOTP.Otp != otp || time.Now().After(otptime) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	} else {
		var tempuser models.TempUser
		if err := config.DB.Model(&models.TempUser{}).Select("first_name, last_name, email, password, phone").Where("email = ?", Email).Scan(&tempuser).Error; err != nil {
			panic(err)
		}
		newUser := models.User{
			FirstName: tempuser.FirstName,
			LastName:  tempuser.LastName,
			Email:     tempuser.Email,
			Password:  tempuser.Password,
			Phone:     tempuser.Phone,
		}
		config.DB.Create(&newUser)
		UserLoginMethod := models.UserLoginMethod{
			UserLoginMethodEmail: Email,
			LoginMethod:          "Manual",
		}

		var count6 int
		err := config.DB.Raw(`SELECT COUNT(*) FROM user_login_methods WHERE user_login_method_email=? group by user_login_method_email,login_method`, Email).Scan(&count6).Error
		if err != nil {
			fmt.Println("failed to execute query", err)
		}
		if count6 == 0 {
			config.DB.Create(&UserLoginMethod)
		} else {
			fmt.Println("user alredy exists")
		}

		config.DB.Where("email = ?", Email).Delete(&models.OTP{})
		config.DB.Where("email = ?", Email).Delete(&models.TempUser{})
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "new user created",
			"data":    gin.H{},
		})

	}

}*/

package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func VerifyotpWindow(c *gin.Context) {
	fmt.Println("Hi")

	// Retrieve the email from URL parameters
	rawEmail := c.Param("email")
	fmt.Println("Raw Email:", rawEmail)

	// Remove any leading or trailing colons
	sanitizedEmail := strings.Trim(rawEmail, ":")
	fmt.Println("Sanitized Email:", sanitizedEmail)

	// Retrieve OTP from request body
	var VerifyOTP models.VerifyOTP
	if err := c.BindJSON(&VerifyOTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Received OTP from request:", VerifyOTP.Otp)

	// Retrieve stored OTP and expiry time from the database
	var otp string
	if err := config.DB.Model(&models.OTP{}).Where("email = ?", sanitizedEmail).Pluck("otp", &otp).Error; err != nil {
		fmt.Println("Error retrieving OTP:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve OTP"})
		return
	}
	if otp == "" {
		fmt.Println("No OTP found for this email:", sanitizedEmail)
		c.JSON(http.StatusNotFound, gin.H{"error": "No OTP found for this email"})
		return
	}
	fmt.Println("Stored OTP:", otp)

	var otptime time.Time
	if err := config.DB.Model(&models.OTP{}).Where("email = ?", sanitizedEmail).Pluck("otp_expiry", &otptime).Error; err != nil {
		fmt.Println("Error retrieving OTP expiry time:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve OTP expiry time"})
		return
	}
	if otptime.IsZero() {
		fmt.Println("No OTP expiry time found for this email:", sanitizedEmail)
		c.JSON(http.StatusNotFound, gin.H{"error": "No OTP expiry time found for this email"})
		return
	}
	fmt.Println("Stored OTP Expiry Time:", otptime)

	// Validate OTP and expiry time
	if VerifyOTP.Otp != otp {
		fmt.Println("OTP does not match")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}
	if time.Now().After(otptime) {
		fmt.Println("OTP has expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP has expired"})
		return
	}

	// Retrieve temporary user data
	var tempuser models.TempUser
	if err := config.DB.Model(&models.TempUser{}).Select("first_name, last_name, email, password, phone").Where("email = ?", sanitizedEmail).Scan(&tempuser).Error; err != nil {
		fmt.Println("Error retrieving temporary user data:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
		return
	}

	// Create new user
	newUser := models.User{
		FirstName: tempuser.FirstName,
		LastName:  tempuser.LastName,
		Email:     tempuser.Email,
		Password:  tempuser.Password,
		Phone:     tempuser.Phone,
	}
	if err := config.DB.Create(&newUser).Error; err != nil {
		fmt.Println("Error creating new user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new user"})
		return
	}

	// Create or update user login method
	UserLoginMethod := models.UserLoginMethod{
		UserLoginMethodEmail: sanitizedEmail,
		LoginMethod:          "Manual",
	}
	var count6 int
	if err := config.DB.Raw(`SELECT COUNT(*) FROM user_login_methods WHERE user_login_method_email=? AND login_method=?`, sanitizedEmail, "Manual").Scan(&count6).Error; err != nil {
		fmt.Println("Error checking existing login methods:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing login methods"})
		return
	}
	if count6 == 0 {
		if err := config.DB.Create(&UserLoginMethod).Error; err != nil {
			fmt.Println("Error creating user login method:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user login method"})
			return
		}
	}

	// Clean up OTP and temporary user data
	if err := config.DB.Where("email = ?", sanitizedEmail).Delete(&models.OTP{}).Error; err != nil {
		fmt.Println("Error deleting OTP:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete OTP"})
		return
	}
	if err := config.DB.Where("email = ?", sanitizedEmail).Delete(&models.TempUser{}).Error; err != nil {
		fmt.Println("Error deleting temporary user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete temporary user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "New user created",
		"data":    gin.H{},
	})
}
