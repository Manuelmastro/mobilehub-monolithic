package config

import (
	"MOBILEHUB/models"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Initialize() {
	// var err error
	// //dsn := "postgres://mobilehubawsdb:mobilehubawsdb@mobilehubawsdb.crc2mmq6agnb.ap-south-1.rds.amazonaws.com:5432/mobilehub"
	// dsn := "postgres://postgres:mobilehubawsdb@mobilehubawsdb.crc2mmq6agnb.ap-south-1.rds.amazonaws.com:5432/mobilehub?sslmode=require"

	// DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	// 	fmt.Println("connection failed due to ", err)
	// }
	var err error

	// Fetch database configuration from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Construct the DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open a connection to the database
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Connection failed due to:", err)
		return
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
	DB.AutoMigrate(&models.Addresse{})
	DB.AutoMigrate(&models.CartItems{})
}
