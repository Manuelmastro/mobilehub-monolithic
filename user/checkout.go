package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Checkout(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "failed to get claims",
		})
		return
	}
	CustomClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := CustomClaims.ID
	fmt.Println("id from claims:", userID)

	var couponcheckout models.CouponCheckout
	err := c.BindJSON(&couponcheckout)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind request",
		})
		return
	}

	var discountamount float64
	if couponcheckout.CouponCode != "" {
		var count int
		config.DB.Raw(`select count (*) from coupons where code = ? and deleted_at is null`, couponcheckout.CouponCode).Scan(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "this coupon does not exist",
			})
			return
		}

		config.DB.Model(&models.Coupon{}).Where("code =?", couponcheckout.CouponCode).Pluck("discount", &discountamount)

	}

	type info struct {
		CartItem            []responsemodels.CartItems `json:"cart_items"`
		Totalamount         float64                    `json:"total_amount"`
		OfferApplied        float64                    `json:"offer_applied"`
		CouponDiscount      float64                    `json:"coupon_discount"`
		CouponAppliedAmount float64                    `json:"coupon_applied_amount"`
		Address             []responsemodels.Address   `json:"address"`
	}

	var Info info
	config.DB.Raw(`select cart_items.user_id,cart_items.product_id,products.product_name,cart_items.total_amount,cart_items.qty,cart_items.price,cart_items.discount,cart_items.final_amount from cart_items join products on cart_items.product_id = products.id where cart_items.user_id = ? and cart_items.qty != 0 and cart_items.deleted_at is null`, userID).Scan(&Info.CartItem)
	var count1 int
	config.DB.Raw("SELECT COUNT(*) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&count1)
	if count1 != 0 {
		err := config.DB.Raw("SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&Info.Totalamount).Error
		fmt.Println("", Info.Totalamount)
		if err != nil {
			fmt.Println("query execution failed", err)
		}
		err = config.DB.Raw("SELECT SUM(discount) from cart_items where user_id =? and deleted_at IS NULL", userID).Scan(&Info.OfferApplied).Error
		if err != nil {
			fmt.Println("query execution failed", err)
		}
	}
	var minpurchase float64
	config.DB.Raw(`select min_purchase from coupons where code = ? and deleted_at is null`, couponcheckout.CouponCode).Scan(&minpurchase)
	if Info.Totalamount+Info.OfferApplied < float64(minpurchase) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "not eligible for cupon apply",
		})
		return
	}
	Info.CouponDiscount = discountamount
	Info.CouponAppliedAmount = Info.Totalamount - discountamount
	config.DB.Where("user_id = ? and deleted_at is null", userID).Find(&Info.Address)
	lastMsg := helper.Responses("showing checkout page", Info, nil)
	c.JSON(http.StatusOK, lastMsg)
}

func EditCheckOutAddress(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "failed to get claims",
		})
		return
	}
	CustomClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := CustomClaims.ID
	fmt.Println("id from claims:", userID)
	addressID := c.Param("address_id")
	var count int
	config.DB.Raw(`SELECT COUNT(*) FROM addresses where id = ? AND user_id = ? and deleted_at IS NULL`, addressID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "address does not found",
		})
		return
	}
	var Address models.Addresse
	err := c.BindJSON(&Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "binding of data failed",
		})
	}

	/*if err := helper.Validate(Address); err != nil {
		fmt.Println("", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}*/
	address := models.Addresse{
		Country:    Address.Country,
		State:      Address.State,
		District:   Address.State,
		StreetName: Address.StreetName,
		PinCode:    Address.PinCode,
		Phone:      Address.Phone,
		Default:    Address.Default,
	}
	config.DB.Model(&models.Addresse{}).Where("id = ? and user_id = ?", addressID, userID).Updates(&address)
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Adress updated successfully",
	})
}
