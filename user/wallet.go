package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func WalletListing(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Claims not found",
		})
		return
	}
	customClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid claims",
		})
		return
	}
	userID := customClaims.ID
	fmt.Println("user id:", userID)
	var wallet responsemodels.Wallet
	config.DB.Raw(`SELECT * from wallets where user_id = ?`, userID).Scan(&wallet)
	c.JSON(http.StatusOK, gin.H{
		"data":    wallet,
		"message": "wallete data retrived successfully",
	})
}

func WalletOrder(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Claims not found",
		})
		return
	}
	customClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid claims",
		})
		return
	}
	userID := customClaims.ID
	fmt.Println("user id:", userID)

	var OrderAdd models.OrderAdd
	err := c.BindJSON(&OrderAdd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}
	if err := helper.Validate(OrderAdd); err != nil {
		fmt.Println("validation error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	var count1 int64
	config.DB.Raw(`SELECT COUNT(*) FROM addresses where id = ? AND user_id = ? AND deleted_at IS NULL`, OrderAdd.AddressID, userID).Scan(&count1)
	if count1 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "address_id not exist for this user",
		})
		return
	}
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM cart_items WHERE user_id=? and deleted_at IS NULL`, userID).Scan(&count)
	fmt.Println("count:", count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "cart is empty",
		})
		return
	}
	var totalquandity uint
	config.DB.Raw(`SELECT SUM(qty) FROM cart_items WHERE user_id=? and deleted_at IS NULL`, userID).Scan(&totalquandity)
	fmt.Println("total quandity:", totalquandity)
	if totalquandity == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "cart is empty",
		})
		return
	}
	var totalamount float64
	config.DB.Raw("SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount)
	var totalamount1 float64
	config.DB.Raw("SELECT SUM(total_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount1)
	var Finalamount float64
	var discountamount float64
	fmt.Println("cupon code:", OrderAdd.CouponCode)
	if OrderAdd.CouponCode != "" {
		fmt.Println("cupon processing")
		var count2 int64
		config.DB.Raw(`SELECT COUNT(*) FROM coupons where code = ? AND deleted_at IS NULL`, OrderAdd.CouponCode).Scan(&count2)
		if count2 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "this cupon does not exist",
			})
			return
		}
		var minpurchase float64
		config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("min_purchase", &minpurchase)
		if totalamount1 > minpurchase {
			config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("discount", &discountamount)

		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "not eligible to apply this cupon",
			})
			return
		}
	}
	var OfferApplied float64
	config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&OfferApplied)
	Finalamount = totalamount - discountamount
	var balance float64
	config.DB.Raw(`SELECT balance FROM wallets WHERE user_id = ?`, userID).Scan(&balance)
	fmt.Println("wallet balance:", balance)
	if balance-Finalamount < 0.00 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "insufficient balance in wallet",
		})
		return
	}
	order := models.Order{
		UserID:         userID,
		AddressID:      OrderAdd.AddressID,
		TotalAmount:    totalamount,
		OfferApplied:   OfferApplied,
		PaymentMethod:  "Wallet",
		CouponCode:     OrderAdd.CouponCode,
		DiscountAmount: discountamount,
		FinalAmount:    Finalamount,
	}
	config.DB.Create(&order)
	balance = balance - Finalamount
	fmt.Println("wallet balance after finalamount:", balance)
	result := config.DB.Exec(`UPDATE wallets SET balance = ? WHERE user_id = ?`, balance, userID)
	if result.Error != nil {
		fmt.Println("", result.Error)
	}
	transaction := models.WalletTransaction{
		UserID:          userID,
		Amount:          Finalamount,
		TransactionType: "Debit",
		Description:     "Purchase thrugh Wallet",
	}
	config.DB.Create(&transaction)
	var CartItems []models.CartItems
	config.DB.Where("user_id = ?", userID).Find(&CartItems)
	var orderID uint
	config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&orderID)
	for _, item := range CartItems {
		if item.Qty == 0 {
			continue
		}
		var price float64
		config.DB.Model(&models.CartItems{}).Where("product_id = ?", item.ProductID).Pluck("price", &price)
		var offerDiscount float64
		var couponDiscount float64
		var hasOffer bool
		config.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).Pluck("has_offer", &hasOffer)
		if hasOffer {
			var discountPercentage uint
			config.DB.Model(&models.Offer{}).Where("product_id = ?", item.ProductID).Pluck("discount_percentage", &discountPercentage)
			offerDiscount = price * float64(discountPercentage) / 100
		}
		couponDiscount = (price / totalamount) * discountamount
		couponDiscount = math.Round(couponDiscount*100) / 100
		totalDiscount := offerDiscount + couponDiscount
		paidamount := price - totalDiscount

		var stock uint
		config.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).Pluck("stock", &stock)

		if stock < item.Qty {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": "Insufficient stock for product ID: " + string(item.ProductID),
			})
			return
		}

		config.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).Update("stock", stock-item.Qty)
		orderItem := models.OrderItems{
			OrderID:        orderID,
			ProductID:      item.ProductID,
			Price:          price,
			PaidAmount:     paidamount,
			Qty:            item.Qty,
			OfferDiscount:  offerDiscount,
			CouponDiscount: couponDiscount,
			TotalDiscount:  totalDiscount,
			PaymentMethod:  "Wallet",
		}
		config.DB.Create(&orderItem)

	}
	now := time.Now()
	today := now.Format("2024-01-02")
	Payment := models.Payments{
		UserID:        userID,
		OrderID:       order.ID,
		TotalAmount:   Finalamount,
		PaymentDate:   today,
		PaymentType:   "wallet",
		PaymentStatus: "paid",
	}
	config.DB.Create(&Payment)
	config.DB.Where("user_id = ?", userID).Delete(&models.CartItems{})
	var order1 responsemodels.Order1
	//var address responsemodels.Address
	var orderitems1 []responsemodels.OrderItems
	config.DB.Raw(`SELECT orders.id,orders.created_at,orders.updated_at,orders.deleted_at,orders.user_id,orders.address_id,orders.total_amount,orders.offer_applied,orders.payment_method,orders.order_status,orders.coupon_code,orders.discount_amount,orders.final_amount FROM orders join addresses on orders.address_id=addresses.id WHERE orders.user_id = ? ORDER BY orders.created_at desc LIMIT 1`, userID).Scan(&order1)
	var orderid uint
	config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&orderid)
	var addressid uint
	config.DB.Raw(`SELECT address_id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&addressid)
	//order1.Address = address
	config.DB.Raw(`SELECT order_items.id,order_items.created_at,order_items.updated_at,order_items.deleted_at,order_items.order_id,order_items.product_id,products.product_name,order_items.qty,order_items.price,order_items.order_status,order_items.payment_method,order_items.coupon_discount,order_items.offer_discount,order_items.total_discount,order_items.paid_amount FROM order_items join products on order_items.product_id=products.id WHERE order_items.order_id = ? ORDER BY order_items.id`, orderid).Scan(&orderitems1)
	c.JSON(http.StatusOK, gin.H{"message": "Order added successfully",
		"order":       order1,
		"order_items": orderitems1})

}

func WalletTransactionListing(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Claims not found",
		})
		return
	}
	customClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid claims",
		})
		return
	}
	userID := customClaims.ID
	fmt.Println("user id:", userID)
	var wallettransaction []responsemodels.WalletTransaction
	config.DB.Raw(`SELECT * from wallet_transactions where user_id = ?`, userID).Scan(&wallettransaction)
	c.JSON(http.StatusOK, gin.H{
		"data":    wallettransaction,
		"message": "wallet transaction data succesfully retrieve",
	})

}
