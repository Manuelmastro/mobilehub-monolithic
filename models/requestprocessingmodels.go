package models

type AdminLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type BlockUser struct {
	UserID uint `validate:"required" json:"user_id"`
}

type CategoryEdit struct {
	//ID           uint   ` json:"id"`
	CategoryName string `gorm:"unique" validate:"required,no_leading_trailing_spaces,no_repeating_spaces" json:"category_name"`
	Description  string `validate:"required,no_leading_trailing_spaces,no_repeating_spaces" json:"category_description"`
	ImgUrl       string `validate:"required" json:"category_imgurl"`
}

type ProductAdd struct {
	//ID         uint ` json:"id"`
	//CategoryID   uint   `json:"category_id" validate:"required"`
	CategoryName string `json:"category_name" validate:"required"`
	//Category    Category `gorm:"foriegnkey:CategoryID;references:ID"`
	ProductName string  `json:"product_name" validate:"required,no_leading_trailing_spaces,no_repeating_spaces,max=50"`
	Description string  `json:"product_description" validate:"required,no_leading_trailing_spaces,no_repeating_spaces,max=100"`
	ImageUrl    string  `json:"product_imageUrl" validate:"required,excludesall= "`
	Price       float64 ` json:"price" validate:"required,gt=0"`
	Stock       uint    `json:"stock" validate:"required"`
	//Popular     bool    `json:"popular" validate:"required"`
	//Size        string  ` json:"size" validate:"required"`
}

type ProductEdit struct {
	//ID         uint ` json:"id"`
	//CategoryID   uint   `json:"category_id" validate:"required"`
	CategoryName string `json:"category_name" validate:"required"`
	//Category    Category `gorm:"foriegnkey:CategoryID;references:ID"`
	ProductName string  `json:"product_name" validate:"required,no_leading_trailing_spaces,no_repeating_spaces,max=50"`
	Description string  `json:"product_description" validate:"required,no_leading_trailing_spaces,no_repeating_spaces,max=100"`
	ImageUrl    string  `json:"product_imageUrl" validate:"required,excludesall= "`
	Price       float64 ` json:"price" validate:"required,gt=0"`
	Stock       uint    `json:"stock" validate:"required"`
	//Popular     bool    `json:"popular" validate:"required"`
	//Size        string  ` json:"size" validate:"required"`
}

type UserSignUp struct {
	FirstName       string `validate:"required,excludesall= " json:"name"`
	LastName        string `validate:"required,nameOrInitials" json:"last_name"`
	Email           string `gorm:"unique" validate:"required,email" json:"email"`
	Password        string `validate:"required,min=8,password" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
	Phone           string `json:"phone" validate:"required,numeric,len=10"`
}

type VerifyOTP struct {
	Otp string `json:"otp"`
}

type UserLogin struct {
	Email    string `gorm:"unique" validate:"required,email" json:"email"`
	Password string `validate:"required" json:"password"`
}

type ProfileEdit struct {
	FirstName string `validate:"required,excludesall= " json:"name"`
	LastName  string `validate:"required,nameOrInitials" json:"last_name"`
	//Email           string `gorm:"unique" validate:"required,email" json:"email"`
	//Password        string `validate:"required" json:"password"`
	//ConfirmPassword string `validate:"required" json:"confirmpassword"`
	Phone string `json:"phone" validate:"required,numeric,len=10"`
}

type PasswordChange struct {
	Password        string `validate:"required,min=8,password" json:"password"`
	ConfirmPassword string `validate:"required" json:"confirmpassword"`
}

type AddressAdd struct {
	//UserID     uint   `validate:"required"`
	//User       User   `gorm:"foriegnkey:UserID;references:ID"`
	Country    string `json:"country" validate:"required,no_leading_trailing_spaces,no_repeating_spaces,max=50,alpha"`
	State      string `json:"state" validate:"required,no_leading_trailing_spaces,no_repeating_spaces,max=50,alpha"`
	District   string `json:"district" validate:"required,no_leading_trailing_spaces,no_repeating_spaces,max=50,alpha"`
	StreetName string `json:"street_name" validate:"required,no_leading_trailing_spaces,no_repeating_spaces,max=50,alpha"`
	PinCode    string `json:"pin_code" validate:"required,numeric"`
	Phone      string `json:"phone" validate:"required,numeric,len=10"`
	Default    bool   `json:"Default" `
}

type CartAdd struct {
	ProductID string `json:"product_id" validate:"required,numeric"`
}

type CouponCheckout struct {
	CouponCode string `json:"coupon_code"`
}

type OrderAdd struct {
	AddressID  uint   `json:"address_id" validate:"required"`
	CouponCode string `json:"coupon_code"`
}

type CancelOrder struct {
	OrderStatus string `json:"order_status" validate:"required"`
}
type WishlistAdd struct {
	ProductID uint `json:"product_id" validate:"required"`
}

type CouponAdd struct {
	Code        string  `validate:"required" json:"code"`
	Discount    float64 `validate:"required" json:"discount"`
	MinPurchase float64 `validate:"required" json:"min_purchase"`
}

type OfferAdd struct {
	ProductID          uint `validate:"required" json:"product_id"`
	DiscountPercentage uint `validate:"required" json:"discount_percentage"`
}
