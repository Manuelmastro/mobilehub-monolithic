package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

var RazorpayClient *razorpay.Client

type orderID struct {
	//orderid string
	OrderID string
}

func CreateOrder(c *gin.Context) {
	RazorpayClient := razorpay.NewClient("rzp_test_0R0gzwKjIDaGFD", "Yix1qNSPiophDD7DTEHH5AaM")
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "claims not found"})
		return
	}

	customClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := customClaims.ID
	var tempaddress models.TempAddress
	err := c.BindJSON(&tempaddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}
	if err := helper.Validate(tempaddress); err != nil {
		fmt.Println("", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	var count1 int64
	config.DB.Raw(`SELECT COUNT(*) FROM addresses where id = ? AND user_id = ? AND deleted_at IS NULL`, tempaddress.AddressID, userID).Scan(&count1)
	if count1 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "address_id not exist for this user",
		})
		return
	}
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM cart_items WHERE user_id=? and deleted_at IS NULL`, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "cart is empty, cant place order",
		})
		return
	}
	var totalquantity uint
	config.DB.Raw(`SELECT SUM(qty) FROM cart_items WHERE user_id=? and deleted_at IS NULL`, userID).Scan(&totalquantity)
	if totalquantity == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "cart is empty, cant place order",
		})
		return
	}
	var totalamount float64
	config.DB.Raw("SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount)
	var totalamount1 float64
	config.DB.Raw("SELECT SUM(total_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount1)
	var Finalamount float64
	var discountamount float64
	if tempaddress.CouponCode != "" {
		var count2 int64
		config.DB.Raw(`SELECT COUNT(*) FROM coupons where code = ? AND deleted_at IS NULL`, tempaddress.CouponCode).Scan(&count2)
		if count2 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "this cupon not exist",
			})
			return
		}
		var minpurchase float64
		config.DB.Model(&models.Coupon{}).Where("code = ?", tempaddress.CouponCode).Pluck("min_purchase", &minpurchase)
		if totalamount1 > minpurchase {
			config.DB.Model(&models.Coupon{}).Where("code = ?", tempaddress.CouponCode).Pluck("discount", &discountamount)

		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "not eligible to apply cupon(totalamount < minpurchase)",
			})
			return
		}
	}
	Finalamount = totalamount - discountamount
	amountInPaise := int(Finalamount * 100)
	data := map[string]interface{}{
		"amount":   amountInPaise,
		"currency": "INR",
	}
	headers := map[string]string{}
	order, err := RazorpayClient.Order.Create(data, headers)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	if order["amount"] != nil {
		amountInPaise := order["amount"].(float64)
		amountInRupees := amountInPaise / 100
		order["amount"] = amountInRupees
		order["amount_due"] = amountInRupees
	}
	tempaddress1 := models.TempAddress{
		AddressID:  tempaddress.AddressID,
		CouponCode: tempaddress.CouponCode,
	}
	config.DB.Create(&tempaddress1)
	order_id := order["id"].(string)
	//added code for week 11
	//var offerApplied float64
	//config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&offerApplied)

	// order1 := models.Order{
	// 	UserID:         userID,
	// 	AddressID:      tempaddress.AddressID,
	// 	TotalAmount:    totalamount,
	// 	OfferApplied:   offerApplied,
	// 	CouponCode:     tempaddress.CouponCode,
	// 	DiscountAmount: discountamount,
	// 	FinalAmount:    Finalamount,
	// 	PaymentStatus:  "pending",
	// }
	// config.DB.Create(&order1)
	// var OorderID uint
	// config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&OorderID)

	c.HTML(http.StatusOK, "razorpay.html", orderID{
		OrderID: order_id, // Use the capitalized field name
	})

}

type Payload struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Signature string `json:"signature"`
}

func verifySignature(orderID string, paymentID string, razorpaySignaure string, secret string) bool {
	data := orderID + "|" + paymentID
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	generatedSignature := hex.EncodeToString(h.Sum(nil))
	return generatedSignature == razorpaySignaure
}

/*func PaymentWebhook(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "claims not found"})
		return
	}

	customClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := customClaims.ID
	var payload Payload
	if err := c.BindJSON(&payload); err != nil {
		log.Println("Error reading request body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ordderID := payload.OrderID
	paymentID := payload.PaymentID
	signature := payload.Signature
	secret := "Yix1qNSPiophDD7DTEHH5AaM"
	if verifySignature(ordderID, paymentID, signature, secret) {
		fmt.Println("payment verified:", payload)
		var totalamount float64
		config.DB.Raw("SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount)
		var totalamount1 float64
		config.DB.Raw("SELECT SUM(total_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount1)
		var addressid uint
		config.DB.Raw(`SELECT address_id from temp_addresses`).Scan(&addressid)
		var cuponcode string
		var count int64
		result := config.DB.Raw(`SELECT COUNT(*) FROM temp_addresses WHERE coupon_code != ''`).Scan(&count)
		if result.Error != nil {
			panic(result.Error)
		}
		if count != 0 {
			config.DB.Raw(`SELECT coupon_code from temp_addresses`).Scan(&cuponcode)
		}
		var Finalamount float64
		var discountamount float64
		if cuponcode != "" {
			var count2 int64
			config.DB.Raw(`SELECT COUNT(*) FROM coupons where code = ?`, cuponcode).Scan(&count2)
			if count2 == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "this cupon not exist",
				})
				return
			}
			var minpurchase float64
			config.DB.Model(&models.Coupon{}).Where("code = ?", cuponcode).Pluck("min_purchase", &minpurchase)
			if totalamount1 > minpurchase {
				config.DB.Model(&models.Coupon{}).Where("code = ?", cuponcode).Pluck("discount", &discountamount)
			}
		}
		var offerapplied float64
		config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&offerapplied)
		Finalamount = totalamount - discountamount
		order := models.Order{
			UserID:         userID,
			AddressID:      addressid,
			TotalAmount:    totalamount,
			PaymentMethod:  "RazorPay",
			OfferApplied:   offerapplied,
			CouponCode:     cuponcode,
			DiscountAmount: discountamount,
			FinalAmount:    Finalamount,
		}
		config.DB.Create(&order)
		var CartItems []models.CartItems
		config.DB.Where("user_id = ?", userID).Find(&CartItems)
		var ID uint
		config.DB.Raw(`SELECT id FROM orders where user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&ID)
		for _, v := range CartItems {
			if v.Qty == 0 {
				continue
			}
			for i := 0; i < int(v.Qty); i++ {
				var price float64
				config.DB.Model(&models.CartItems{}).Where("product_id = ?", v.ProductID).Pluck("price", &price)
				fmt.Println("id", ID)
				fmt.Println("price printing--", price)
				var offerdiscount float64
				var coupondiscount float64
				var hasoffer bool
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Pluck("has_offer", &hasoffer)
				if hasoffer {
					var discountpercentage uint
					config.DB.Model(&models.Offer{}).Where("product_id = ?", v.ProductID).Pluck("discount_percentage", &discountpercentage)
					offerdiscount = price * float64(discountpercentage) / 100
				}
				var totalamount float64
				config.DB.Model(&models.Order{}).Where("id = ?", ID).Pluck("total_amount", &totalamount)
				coupondiscount = (price / totalamount1) * discountamount
				coupondiscount = math.Round(coupondiscount*100) / 100
				totaldiscount := offerdiscount + coupondiscount
				paidamount := price - totaldiscount
				orderItem := models.OrderItems{
					OrderID:   ID,
					ProductID: v.ProductID,
					//Qty:         v.Qty,
					Price: price,
					//TotalAmount: float64(v.Qty) * price,
					PaymentMethod:  "RazorPay",
					OfferDiscount:  offerdiscount,
					CouponDiscount: coupondiscount,
					TotalDiscount:  totaldiscount,
					PaidAmount:     paidamount,
				}
				config.DB.Create(&orderItem)

			}
		}
		now := time.Now()
		today := now.Format("2006-01-02")
		Payment := models.Payments{
			UserID:        userID,
			OrderID:       order.ID,
			TotalAmount:   Finalamount,
			PaymentDate:   today,
			PaymentType:   "RazorPay",
			PaymentStatus: "paid",
		}
		config.DB.Create(&Payment)
		config.DB.Where("user_id = ?", userID).Delete(&models.CartItems{})
		var order1 responsemodels.Order
		var address responsemodels.Address
		var orderitems1 []responsemodels.OrderItems
		config.DB.Raw(`SELECT orders.id,orders.created_at,orders.updated_at,orders.deleted_at,orders.user_id,orders.address_id,orders.total_amount,orders.offer_applied,orders.coupon_code,orders.discount_amount,orders.final_amount,orders.order_status,orders.payment_method FROM orders join addresses on orders.address_id=addresses.id WHERE orders.user_id = ? ORDER BY orders.created_at desc LIMIT 1`, userID).Scan(&order1)
		fmt.Println("-----------------")
		fmt.Println("user id ", userID)
		var orderid uint
		config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&orderid)
		fmt.Println("order id ", orderid)
		//var addressid uint
		config.DB.Raw(`SELECT address_id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&addressid)
		fmt.Println("address id", addressid)
		config.DB.Raw(`SELECT * FROM addresses WHERE id = ?`, addressid).Scan(&address)
		order1.Address = address
		config.DB.Raw(`SELECT order_items.id,order_items.created_at,order_items.updated_at,order_items.deleted_at,order_items.order_id,order_items.product_id,products.product_name,order_items.price,order_items.order_status,order_items.payment_method,order_items.cupon_discount,order_items.offer_discount,order_items.total_discount,order_items.paid_amount FROM order_items join products on order_items.product_id=products.id WHERE order_items.order_id = ? ORDER BY order_items.id`, orderid).Scan(&orderitems1)
		result = config.DB.Exec("TRUNCATE temp_addresses")
		if result.Error != nil {
			panic(result.Error)
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order added successfully",
			"order":       order1,
			"order_items": orderitems1,
			"status":      "success"})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
	}

}*/

/*func PaymentWebhook(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "claims not found"})
		return
	}

	customClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := customClaims.ID
	var payload Payload
	if err := c.BindJSON(&payload); err != nil {
		log.Println("Error reading request body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ordderID := payload.OrderID
	paymentID := payload.PaymentID
	signature := payload.Signature
	secret := "Yix1qNSPiophDD7DTEHH5AaM"
	if verifySignature(ordderID, paymentID, signature, secret) {
		fmt.Println("payment verified:", payload)
		var totalamount float64
		config.DB.Raw("SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount)
		var totalamount1 float64
		config.DB.Raw("SELECT SUM(total_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount1)
		var addressid uint
		config.DB.Raw(`SELECT address_id from temp_addresses`).Scan(&addressid)
		var cuponcode string
		var count int64
		result := config.DB.Raw(`SELECT COUNT(*) FROM temp_addresses WHERE coupon_code != ''`).Scan(&count)
		if result.Error != nil {
			panic(result.Error)
		}
		if count != 0 {
			config.DB.Raw(`SELECT coupon_code from temp_addresses`).Scan(&cuponcode)
		}
		var Finalamount float64
		var discountamount float64
		if cuponcode != "" {
			var count2 int64
			config.DB.Raw(`SELECT COUNT(*) FROM coupons where code = ?`, cuponcode).Scan(&count2)
			if count2 == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "this cupon not exist",
				})
				return
			}
			var minpurchase float64
			config.DB.Model(&models.Coupon{}).Where("code = ?", cuponcode).Pluck("min_purchase", &minpurchase)
			if totalamount1 > minpurchase {
				config.DB.Model(&models.Coupon{}).Where("code = ?", cuponcode).Pluck("discount", &discountamount)
			}
		}
		var offerapplied float64
		config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&offerapplied)
		Finalamount = totalamount - discountamount
		order := models.Order{
			UserID:         userID,
			AddressID:      addressid,
			TotalAmount:    totalamount,
			PaymentMethod:  "RazorPay",
			OfferApplied:   offerapplied,
			CouponCode:     cuponcode,
			DiscountAmount: discountamount,
			FinalAmount:    Finalamount,
		}
		config.DB.Create(&order)
		var CartItems []models.CartItems
		config.DB.Where("user_id = ?", userID).Find(&CartItems)
		var ID uint
		config.DB.Raw(`SELECT id FROM orders where user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&ID)
		for _, v := range CartItems {
			if v.Qty == 0 {
				continue
			}
			for i := 0; i < int(v.Qty); i++ {
				var price float64
				config.DB.Model(&models.CartItems{}).Where("product_id = ?", v.ProductID).Pluck("price", &price)
				fmt.Println("id", ID)
				fmt.Println("price printing--", price)
				var offerdiscount float64
				var coupondiscount float64
				var hasoffer bool
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Pluck("has_offer", &hasoffer)
				if hasoffer {
					var discountpercentage uint
					config.DB.Model(&models.Offer{}).Where("product_id = ?", v.ProductID).Pluck("discount_percentage", &discountpercentage)
					offerdiscount = price * float64(discountpercentage) / 100
				}
				var totalamount float64
				config.DB.Model(&models.Order{}).Where("id = ?", ID).Pluck("total_amount", &totalamount)
				coupondiscount = (price / totalamount1) * discountamount
				coupondiscount = math.Round(coupondiscount*100) / 100
				totaldiscount := offerdiscount + coupondiscount
				paidamount := price - totaldiscount
				orderItem := models.OrderItems{
					OrderID:        ID,
					ProductID:      v.ProductID,
					Price:          price,
					PaymentMethod:  "RazorPay",
					OfferDiscount:  offerdiscount,
					CouponDiscount: coupondiscount,
					TotalDiscount:  totaldiscount,
					PaidAmount:     paidamount,
				}
				config.DB.Create(&orderItem)

				// Update product stock
				var currentStock int
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Pluck("stock", &currentStock)
				newStock := currentStock - 1
				if newStock < 0 {
					newStock = 0
				}
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Update("stock", newStock)
			}
		}
		now := time.Now()
		today := now.Format("2006-01-02")
		Payment := models.Payments{
			UserID:        userID,
			OrderID:       order.ID,
			TotalAmount:   Finalamount,
			PaymentDate:   today,
			PaymentType:   "RazorPay",
			PaymentStatus: "paid",
		}
		config.DB.Create(&Payment)
		config.DB.Where("user_id = ?", userID).Delete(&models.CartItems{})
		var order1 responsemodels.Order
		var address responsemodels.Address
		var orderitems1 []responsemodels.OrderItems
		config.DB.Raw(`SELECT orders.id,orders.created_at,orders.updated_at,orders.deleted_at,orders.user_id,orders.address_id,orders.total_amount,orders.offer_applied,orders.coupon_code,orders.discount_amount,orders.final_amount,orders.order_status,orders.payment_method FROM orders join addresses on orders.address_id=addresses.id WHERE orders.user_id = ? ORDER BY orders.created_at desc LIMIT 1`, userID).Scan(&order1)
		fmt.Println("-----------------")
		fmt.Println("user id ", userID)
		var orderid uint
		config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&orderid)
		fmt.Println("order id ", orderid)
		config.DB.Raw(`SELECT address_id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&addressid)
		fmt.Println("address id", addressid)
		config.DB.Raw(`SELECT * FROM addresses WHERE id = ?`, addressid).Scan(&address)
		order1.Address = address
		config.DB.Raw(`SELECT order_items.id,order_items.created_at,order_items.updated_at,order_items.deleted_at,order_items.order_id,order_items.product_id,products.product_name,order_items.price,order_items.order_status,order_items.payment_method,order_items.cupon_discount,order_items.offer_discount,order_items.total_discount,order_items.paid_amount FROM order_items join products on order_items.product_id=products.id WHERE order_items.order_id = ? ORDER BY order_items.id`, orderid).Scan(&orderitems1)
		result = config.DB.Exec("TRUNCATE temp_addresses")
		if result.Error != nil {
			panic(result.Error)
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order added successfully",
			"order":       order1,
			"order_items": orderitems1,
			"status":      "success"})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
	}
}*/

func PaymentWebhook(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "claims not found"})
		return
	}

	customClaims, ok := claims.(*midleware.CustomClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		return
	}
	userID := customClaims.ID
	var payload Payload
	if err := c.BindJSON(&payload); err != nil {
		log.Println("Error reading request body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ordderID := payload.OrderID
	paymentID := payload.PaymentID
	signature := payload.Signature
	secret := "Yix1qNSPiophDD7DTEHH5AaM"
	if verifySignature(ordderID, paymentID, signature, secret) {
		fmt.Println("payment verified:", payload)
		var totalamount float64
		config.DB.Raw("SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount)
		var totalamount1 float64
		config.DB.Raw("SELECT SUM(total_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount1)
		var addressid uint
		config.DB.Raw(`SELECT address_id from temp_addresses`).Scan(&addressid)
		var cuponcode string
		var count int64
		result := config.DB.Raw(`SELECT COUNT(*) FROM temp_addresses WHERE coupon_code != ''`).Scan(&count)
		if result.Error != nil {
			panic(result.Error)
		}
		if count != 0 {
			config.DB.Raw(`SELECT coupon_code from temp_addresses`).Scan(&cuponcode)
		}
		var Finalamount float64
		var discountamount float64
		if cuponcode != "" {
			var count2 int64
			config.DB.Raw(`SELECT COUNT(*) FROM coupons where code = ?`, cuponcode).Scan(&count2)
			if count2 == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "this cupon not exist",
				})
				return
			}
			var minpurchase float64
			config.DB.Model(&models.Coupon{}).Where("code = ?", cuponcode).Pluck("min_purchase", &minpurchase)
			if totalamount1 > minpurchase {
				config.DB.Model(&models.Coupon{}).Where("code = ?", cuponcode).Pluck("discount", &discountamount)
			}
		}
		var offerapplied float64
		config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&offerapplied)
		Finalamount = totalamount - discountamount
		order := models.Order{
			UserID:         userID,
			AddressID:      addressid,
			TotalAmount:    totalamount,
			PaymentMethod:  "RazorPay",
			OfferApplied:   offerapplied,
			CouponCode:     cuponcode,
			DiscountAmount: discountamount,
			FinalAmount:    Finalamount,
		}
		config.DB.Create(&order)
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
			var hasoffer bool
			config.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).Pluck("has_offer", &hasoffer)
			if hasoffer {
				var discountPercentage uint
				config.DB.Model(&models.Offer{}).Where("product_id = ?", item.ProductID).Pluck("discount_percentage", &discountPercentage)
				offerDiscount = price * float64(discountPercentage) / 100
			}

			couponDiscount = (price / totalamount) * discountamount
			couponDiscount = math.Round(couponDiscount*100) / 100
			totalDiscount := offerDiscount + couponDiscount
			paidamount := price - totalDiscount //new added

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
				Price:          price, //new added
				PaidAmount:     paidamount,
				Qty:            item.Qty,
				OfferDiscount:  offerDiscount,
				CouponDiscount: couponDiscount,
				TotalDiscount:  totalDiscount,
				PaymentMethod:  "RazorPay",
			}
			config.DB.Create(&orderItem)

		}
		/*var ID uint
		config.DB.Raw(`SELECT id FROM orders where user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&ID)
		for _, v := range CartItems {
			if v.Qty == 0 {
				continue
			}
			for i := 0; i < int(v.Qty); i++ {
				var price float64
				config.DB.Model(&models.CartItems{}).Where("product_id = ?", v.ProductID).Pluck("price", &price)
				fmt.Println("id", ID)
				fmt.Println("price printing--", price)
				var offerdiscount float64
				var coupondiscount float64
				var hasoffer bool
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Pluck("has_offer", &hasoffer)
				if hasoffer {
					var discountpercentage uint
					config.DB.Model(&models.Offer{}).Where("product_id = ?", v.ProductID).Pluck("discount_percentage", &discountpercentage)
					offerdiscount = price * float64(discountpercentage) / 100
				}
				var totalamount float64
				config.DB.Model(&models.Order{}).Where("id = ?", ID).Pluck("total_amount", &totalamount)
				coupondiscount = (price / totalamount1) * discountamount
				coupondiscount = math.Round(coupondiscount*100) / 100
				totaldiscount := offerdiscount + coupondiscount
				paidamount := price - totaldiscount
				orderItem := models.OrderItems{
					OrderID:        ID,
					ProductID:      v.ProductID,
					Price:          price,
					PaymentMethod:  "RazorPay",
					OfferDiscount:  offerdiscount,
					CouponDiscount: coupondiscount,
					TotalDiscount:  totaldiscount,
					PaidAmount:     paidamount,
				}
				config.DB.Create(&orderItem)

				// Update product stock
				var currentStock int
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Pluck("stock", &currentStock)
				newStock := currentStock - 1
				if newStock < 0 {
					newStock = 0
				}
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Update("stock", newStock)
			}
		}*/
		now := time.Now()
		today := now.Format("2006-01-02")
		Payment := models.Payments{
			UserID:        userID,
			OrderID:       order.ID,
			TotalAmount:   Finalamount,
			PaymentDate:   today,
			PaymentType:   "RazorPay",
			PaymentStatus: "paid",
		}
		config.DB.Create(&Payment)
		config.DB.Where("user_id = ?", userID).Delete(&models.CartItems{})
		var order1 responsemodels.Order
		var address responsemodels.Address
		var orderitems1 []responsemodels.OrderItems
		config.DB.Raw(`SELECT orders.id,orders.created_at,orders.updated_at,orders.deleted_at,orders.user_id,orders.address_id,orders.total_amount,orders.offer_applied,orders.coupon_code,orders.discount_amount,orders.final_amount,orders.order_status,orders.payment_method FROM orders join addresses on orders.address_id=addresses.id WHERE orders.user_id = ? ORDER BY orders.created_at desc LIMIT 1`, userID).Scan(&order1)
		fmt.Println("-----------------")
		fmt.Println("user id ", userID)
		var orderid uint
		config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&orderid)
		fmt.Println("order id ", orderid)
		config.DB.Raw(`SELECT address_id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&addressid)
		fmt.Println("address id", addressid)
		config.DB.Raw(`SELECT * FROM addresses WHERE id = ?`, addressid).Scan(&address)
		order1.Address = address
		config.DB.Raw(`SELECT order_items.id,order_items.created_at,order_items.updated_at,order_items.deleted_at,order_items.order_id,order_items.product_id,products.product_name,order_items.qty,order_items.price,order_items.order_status,order_items.payment_method,order_items.coupon_discount,order_items.offer_discount,order_items.total_discount,order_items.paid_amount FROM order_items join products on order_items.product_id=products.id WHERE order_items.order_id = ? ORDER BY order_items.id`, orderid).Scan(&orderitems1)
		result = config.DB.Exec("TRUNCATE temp_addresses")
		if result.Error != nil {
			panic(result.Error)
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order added successfully",
			"order":       order1,
			"order_items": orderitems1,
			"status":      "success"})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
	}
}
