package config

import (
	"MOBILEHUB/models"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Initialize() {
	var err error
	dsn := "postgres://postgres:manu123@localhost:5432/mobilehub"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("connection failed due to ", err)
	}

}

func AutoMigrate() {
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Product{})
	DB.AutoMigrate(&models.Admin{})
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.OTP{})
	DB.AutoMigrate(&models.TempUser{})
	DB.AutoMigrate(&models.UserLoginMethod{})
	DB.AutoMigrate(&models.Payments{})
	DB.AutoMigrate(&models.TempAddress{})
	DB.AutoMigrate(&models.Wallet{})
	DB.AutoMigrate(&models.Wishlist{})
	DB.AutoMigrate(&models.Coupon{})
	DB.AutoMigrate(&models.Offer{})
	DB.AutoMigrate(&models.SalesReportItem{})
	DB.AutoMigrate(&models.WalletTransaction{})
	DB.AutoMigrate(&models.OrderItems{})
	DB.AutoMigrate(&models.Invoice{})
	DB.AutoMigrate(&models.Order{})

}
