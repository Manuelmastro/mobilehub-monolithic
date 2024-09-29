package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user model
/*type User struct {
	gorm.Model
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string `json:"name" gorm:"type:varchar(15);not null"`
	Phone    string `json:"phone" gorm:"type:varchar(15);not null;unique"`
	Email    string `json:"email" gorm:"type:varchar(100);unique;not null"`
	Password string `json:"password" gorm:"type:varchar(255);not null"`
	Status   string `json:"status" gorm:"type:enum('Active','Blocked','Deleted');default:'Active'"`
}*/

type User struct {
	gorm.Model
	//ID        uint   `gorm:"primary key" json:"id"`
	FirstName string `validate:"required"`
	LastName  string `validate:"required"`
	Email     string `gorm:"unique" validate:"required"`
	Password  string `validate:"required"`
	Phone     string `json:"phone" validate:"required,numeric,len=10"`
	Status    string `gorm:"type:varchar(10); check(status IN ('Active', 'Blocked', 'Deleted')) ;default:'Active'" json:"status" validate:"required"`
}

// Admin represents the admin model
type Admin struct {
	gorm.Model
	//ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string `json:"name" gorm:"type:varchar(15);not null"`
	Email    string `json:"email" gorm:"type:varchar(100);unique;not null"`
	Password string `json:"password" gorm:"type:varchar(255);not null"`
}

type Category struct {
	gorm.Model
	//ID           uint   `gorm:"primary key" json:"id"`
	CategoryName string `json:"category_name" validate:"required"`
	Description  string `json:"category_description" validate:"required"`
	ImageUrl     string `json:"category_imageUrl" validate:"required"`
}

type Product struct {
	gorm.Model
	//ID          uint     `gorm:"primary key" json:"id"`
	CategoryID  uint     `json:"category_id" validate:"required"`
	Category    Category `gorm:"foriegnkey:CategoryID;references:ID" json:"category,omitempty"`
	ProductName string   `json:"product_name" validate:"required"`
	Description string   `json:"product_description" validate:"required"`
	ImageUrl    string   `json:"product_imageUrl" validate:"required"`
	Price       float64  `gorm:"type:decimal(10,2)" json:"price" validate:"required"`
	Stock       uint     `json:"stock"`
	//Popular              bool     `gorm:"type:boolean;default:false" json:"popular" validate:"required"`
	//Size                 string   `gorm:"type:varchar(10); check:size IN ('Medium', 'Small', 'Large')" json:"size" validate:"required,oneof=Medium Small Large"`
	HasOffer             bool `gorm:"default:false"`
	OfferDiscountPercent uint `gorm:"default:0"`
}

type UserLoginMethod struct {
	UserLoginMethodEmail string
	LoginMethod          string
}

type OTP struct {
	Email     string `gorm:"primary key" json:"email"`
	OTP       string
	OtpExpiry time.Time
}

type TempUser struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	Phone     string
}

/*type Addresse struct {
	gorm.Model
	UserID     uint   `validate:"required"`
	User       User   `gorm:"foriegnkey:UserID;references:ID"`
	Country    string `validate:"required"`
	State      string `validate:"required"`
	District   string `validate:"required"`
	StreetName string `validate:"required"`
	PinCode    string `validate:"required,numeric"`
	Phone      string `validate:"required,numeric,len=10"`
	Default    bool   `gorm:"default:false" validate:"required"`
}*/

type Addresse struct {
	gorm.Model
	UserID     uint   `json:"user_id" validate:"required"`
	User       User   `gorm:"foreignkey:UserID;references:ID"`
	Country    string `json:"country" validate:"required"`
	State      string `json:"state" validate:"required"`
	District   string `json:"district" validate:"required"`
	StreetName string `json:"street_name" validate:"required"`
	PinCode    string `json:"pin_code" validate:"required,numeric"`
	Phone      string `json:"phone" validate:"required,numeric,len=10"`
	Default    bool   `json:"default" gorm:"default:false" validate:"required"`
}

type CartItems struct {
	gorm.Model
	UserID uint `validate:"required"`
	User   User `gorm:"foriegnkey:UserID;references:ID"`
	// CartID      string  `validate:"required,numeric"`
	// Cart        Cart    `gorm:"foriegnkey:CartID;references:ID"`
	ProductID   string  `validate:"required,numeric"`
	Product     Product `gorm:"foriegnkey:ProductID;references:ID"`
	TotalAmount float64 `gorm:"type:decimal(10,2);default:0.00"  `
	Qty         uint    `gorm:"default:0"`
	Price       float64 `gorm:"type:decimal(10,2)" `
	Discount    float64 `gorm:"default:0.00"`
	FinalAmount float64
}

type Order struct {
	gorm.Model
	UserID uint `validate:"required"`
	//OrderDate   time.Time
	AddressID   uint
	Address     Addresse `gorm:"foriegnkey:AddressID;references:ID"`
	TotalAmount float64
	// OrderStatus string `gorm:"type:varchar(10); check(order_status IN ('pending', 'delivered', 'cancelled')) ;default:'pending'" json:"order_status" validate:"required"`

	PaymentMethod string `gorm:"type:varchar(10); check(order_status IN ('COD', 'RazorPay','Wallet')) ;default:'COD'" json:"payment_method" validate:"required"`
	//PaymentStatus  string  `gorm:"type:varchar(10);check:payment_status IN ('pending', 'completed', 'failed');default:'pending'" json:"payment_status" validate:"required,oneof=pending completed failed"`
	OrderStatus    string  `gorm:"type:varchar(10);check:order_status IN ('pending','shipped', 'delivered', 'cancelled','failed');default:'pending'" json:"order_status" validate:"required,oneof=pending delivered shipped cancelled failed"`
	OfferApplied   float64 `gorm:"default:0.00"`
	CouponCode     string
	DiscountAmount float64 `gorm:"type:decimal(10,2);default:0.00"`
	FinalAmount    float64 `gorm:"type:decimal(10,2);not null"`
}

type OrderItems struct {
	gorm.Model
	OrderID   uint    `validate:"required"`
	Order     Order   `gorm:"foriegnkey:OrderID;references:ID"`
	ProductID string  `validate:"required,numeric"`
	Product   Product `gorm:"foriegnkey:ProductID;references:ID"`
	Qty       uint
	Price     float64
	//TotalAmount float64
	OrderStatus    string  `gorm:"type:varchar(10);check:order_status IN ('pending','shipped', 'delivered', 'cancelled','failed','return');default:'pending'" json:"order_status" validate:"required,oneof=pending delivered shipped cancelled failed return"`
	PaymentMethod  string  `gorm:"type:varchar(10); check(order_status IN ('COD', 'RazorPay','Wallet')) ;default:'COD'" json:"payment_method" validate:"required"`
	CouponDiscount float64 `gorm:"default:0.00"`
	OfferDiscount  float64 `gorm:"default:0.00"`
	TotalDiscount  float64 `gorm:"default:0.00"`
	PaidAmount     float64 `gorm:"default:0.00"`
	DeliveredDate  string
}

type Offer struct {
	gorm.Model
	ProductID          uint `gorm:"not null"`
	DiscountPercentage uint `gorm:"not null"`
}

type Coupon struct {
	gorm.Model
	Code        string  `gorm:"not null" json:"code"`
	Discount    float64 `gorm:"type:decimal(5,2);not null" json:"discount"`
	MinPurchase float64 `gorm:"type:decimal(10,2)" json:"min_purchase"`
}

type Payments struct {
	gorm.Model
	UserID        uint `validate:"required"`
	OrderID       uint `validate:"required"`
	OrderItemID   uint
	TotalAmount   float64 `validate:"required,numeric"`
	TransactionID string
	PaymentDate   string
	PaymentType   string `gorm:"type:varchar(10); check(status IN ('COD', 'RazorPay')) ;default:'COD'" json:"payment_type" validate:"required"`
	PaymentStatus string `gorm:"type:varchar(10); check(status IN ('pending', 'paid', 'refund')) ;default:'pending'" json:"payment_status" validate:"required"`
	Description   string
}

type Wishlist struct {
	gorm.Model
	UserID      uint `gorm:"not null"`
	ProductID   uint `gorm:"not null"`
	ProductName string
}

type WalletTransaction struct {
	gorm.Model
	UserID          uint    `gorm:"not null"`
	Amount          float64 `gorm:"not null"`
	TransactionType string  `gorm:"size:50;not null"`
	Description     string  `gorm:"size:255"`
}

type Wallet struct {
	gorm.Model
	UserID  uint    `gorm:"not null"`
	Balance float64 `gorm:"type:decimal(10,2);default:0.00"`
}

type SalesReportItem struct {
	OrderID        uint
	ProductID      string
	ProductName    string
	Qty            uint
	Price          float64
	OrderStatus    string
	PaymentMethod  string
	CouponDiscount float64
	OfferDiscount  float64
	TotalDiscount  float64
	PaidAmount     float64
	OrderDate      time.Time
	DeliveredDate  string
}
type TempAddress struct {
	AddressID  uint   `json:"address_id" validate:"required"`
	CouponCode string `json:"coupon_code"`
}

type Invoice struct {
	No             int
	ProductID      string
	ProductName    string
	Quantity       uint
	MRP            float64
	CouponDiscount float64
	OfferDiscount  float64
	TotalDiscount  float64
	FinalPrice     float64
}
