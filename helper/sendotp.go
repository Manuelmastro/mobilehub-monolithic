package helper

import (
	"fmt"
	"net/smtp"
	"strings"
)

func SendOTPEmail(email, otp string) error {
	fmt.Println(email, otp)

	from := "mobilehub.ecommerce@gmail.com"
	//password := os.Getenv("password")
	password := "gjbq brdn ybrw igwb"
	to := []string{email}
	subject := "OTP for Signup(mobilehub)"
	body := "Your OTP is:" + otp

	msg := "From: " + from + "\n" +
		"To: " + strings.Join(to, ",") + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	smtpServer := "smtp.gmail.com"
	auth := smtp.PlainAuth("", from, password, smtpServer)
	err := smtp.SendMail(smtpServer+":587", auth, from, to, []byte(msg))

	if err != nil {
		fmt.Println("error in email sending", err.Error())
		return err
	}
	return nil

}
