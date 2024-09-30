package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleResponse struct {
	//ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

var googleOauthConfig *oauth2.Config
var oauthStateString = "random" // Replace with a random string

var (
	clientid     = os.Getenv("GOOGLE_CLIENT_ID")
	clientsecret = os.Getenv("GOOGLE_CLIENT_SECRET")
)

func HandleGoogleLogin(c *gin.Context) {

	fmt.Println("clientid:", clientid)
	fmt.Println("clientsecret:", clientsecret)
	googleOauthConfig = &oauth2.Config{
		RedirectURL: "https://themobilehubs.xyz/auth/google/callback",
		//ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		//ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		ClientID:     clientid,
		ClientSecret: clientsecret,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)

}

func HandleGoogleCallback(c *gin.Context) {

	state := c.Query("state")
	if state != oauthStateString {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid oauth state"})
		return
	}

	code := c.Query("code")
	token, err := googleOauthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	// Use token.AccessToken to get user information from Google Identity Toolkit API
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"msg": err.Error()})
	}

	userinfo, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"msg": err.Error()})
	}

	//check

	var User GoogleResponse
	err = json.Unmarshal(userinfo, &User)
	if err != nil {
		fmt.Println("", err)
	}
	fmt.Println("----", User.Email, User.Name)
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM user_login_methods WHERE user_login_method_email=?`, User.Email).Scan(&count)
	if count != 0 {
		var loginmethod string
		config.DB.Model(&models.UserLoginMethod{}).Where("user_login_method_email = ?", User.Email).Pluck("login_method", &loginmethod)
		if loginmethod == "Manual" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "please log in through Manual Log in",
				"data":    gin.H{},
			})
			return
		}
	}

	if count == 0 {
		UserLoginMethod := models.UserLoginMethod{
			UserLoginMethodEmail: User.Email,
			LoginMethod:          "Google Authentication",
		}
		//var user models.User
		user := models.User{
			FirstName: User.Name,
			Email:     User.Email,
		}
		config.DB.Create(&UserLoginMethod)
		config.DB.Create(&user)
	}

	//generate JWT
	var id uint
	config.DB.Model(&models.User{}).Where("email = ?", User.Email).Pluck("id", &id)
	jwttoken, err := midleware.GenerateJWT("user", User.Email, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	fmt.Println("", jwttoken)
	// Set token as cookie
	c.SetCookie("jwt_token", jwttoken, 3600, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{"message": "User Login successful", "token": jwttoken})

}
