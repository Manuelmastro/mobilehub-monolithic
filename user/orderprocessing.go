/*package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"math"
	"net/http"

















	"github.com/gin-gonic/gin"
)

func Order(c *gin.Context){
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

	var OrderAdd models.OrderAdd
	err := c.BindJSON(&OrderAdd)
	if err!= nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"message": "binding of data failed",
		})
	}

	if err := helper.Validate(OrderAdd); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var count int
	config.DB.Raw(`SELECT COUNT(*) FROM addresses where id = ? AND user_id = ? AND deleted_at IS NULL`,OrderAdd.AddressID, userID).Scan(&count)
	if count == 0{
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "address not found for user",
		})
		return
	}

	var count1 int
	config.DB.Raw(`SELECT COUNT(*) FROM cart_items WHERE user_id = ? and deleted_at IS NULL`,userID).Scan(&count1)
	fmt.Println("count:",count1)
	if count1 == 0{
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"message": "this cart is empty",
		})
		return
	}
	var totalamount float64
	config.DB.Raw(`SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL`, userID).Scan(&totalamount)
	var totalamount1 float64
	config.DB.Raw(`SELECT SUM(total_amount) from cart_items where user_id = ? and deleted_at IS NULL`, userID).Scan(&totalamount1)
	var FinalAmount float64
	var discountamount float64

	fmt.Println("coupon code:", OrderAdd.CouponCode)
	if OrderAdd.CouponCode != ""{
		var count2 int
		config.DB.Raw(`SELECT COUNT(*) FROM coupons where code = ? and deleted_at IS NULL`, OrderAdd.CouponCode).Scan(&count2)
		if count2 == 0{
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "this coupon not found",
			})
			return
		}
		var minpurchase float64
		config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("min_purchase",&minpurchase)
		if totalamount1 > minpurchase{
			config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("discount", &discountamount)
		}else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "minimum purchase amount required to apply this coupon",
			})
			return
		}

	}
	var OfferApplied float64
	config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&OfferApplied)
	FinalAmount = totalamount - discountamount
	order := models.Order{
		UserID: userID,
		AddressID: OrderAdd.AddressID,
		TotalAmount: totalamount,
		OfferApplied: OfferApplied,
		CouponCode: OrderAdd.CouponCode,
		DiscountAmount: discountamount,
		FinalAmount: FinalAmount,
	}
	config.DB.Create(&order)
	var CartItems []models.CartItems
	config.DB.Where("user_id = ?", userID).Find(&CartItems)
	var ID uint
	config.DB.Raw(`SELECT id FROM orders where user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&ID)
	for _,v := range CartItems{
		if v.Qty == 0{
			continue
		}

		for i:= 0; i<int(v.Qty); i++{
			var price float64
			config.DB.Model(&models.CartItems{}).Where("product_id = ?", v.ProductID).Pluck("price", &price)
			var offerdiscount float64
			var coupondiscount float64
			var hasoffer bool
			config.DB.Model(models.Product{}).Where("id = ?", v.ProductID).Pluck("has_offer", &hasoffer)
			if hasoffer {
				var discountpercentage uint
				config.DB.Model(&models.Offer{}).Where("product_id = ?", v.ProductID).Pluck("discount_percentage",&discountpercentage)
				offerdiscount = price * float64(discountpercentage) / 100
			}
			var totalamount float64
			config.DB.Model(&models.Order{}).Where("id = ?", ID).Pluck("total_amount", &totalamount)
			coupondiscount = math.Round(coupondiscount*100)/ 100
			totaldiscount := offerdiscount + coupondiscount
			orderItems := models.OrderItems{
				OrderID: ID,
				ProductID: v.ProductID,
				Price: price,
				OfferDiscount: offerdiscount,
				CouponDiscount: coupondiscount,
				TotalDiscount: totaldiscount,
			}



		}
	}
	Payment := models.Payments{
		UserID: userID,
		OrderID: order.ID,
		TotalAmount: FinalAmount,
	}

	config.DB.Create(&Payment)
	config.DB.Where("user_id = ?", userID).Delete(&models.CartItems{})
	var order1 responsemodels.Order
	var address responsemodels.Address
	var orderitems1 []responsemodels.OrderItems
	config.DB.Raw(`SELECT orders.id,orders.created_at,orders.updated_at,orders.deleted_at,orders.user_id,orders.address_id,orders.total_amount,orders.payment_method,orders.order_status,orders.offer_applied,orders.coupon_code,orders.discount_amount,orders.final_amount FROM orders join addresses on orders.address_id=addresses.id WHERE orders.user_id = ? ORDER BY orders.created_at desc LIMIT 1`, userID).Scan(&order1)
	var orderid uint
	config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&orderid)
	var addressid uint
	config.DB.Raw(`SELECT address_id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&addressid)
	config.DB.Raw(`SELECT order_items.id,order_items.created_at,order_items.updated_at,order_items.deleted_at,order_items.order_id,order_items.product_id,products.product_name,order_items.price,order_items.order_status,order_items.payment_method,order_items.coupon_discount,order_items.offer_discount, order_items.total_discount,order_items.paid_amount FROM order_items join products on order_items.product_id=products.id WHERE order_items.order_id =? ORDER BY order_items.id`, orderid).Scan(&orderitems1)
	c.JSON(http.StatusOK, gin.H{
		"message": "order added successfully",
		"order": order1,
		"order_items": orderitems1,

	})




}*/

/*package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
)*/

/*func Order(c *gin.Context) {
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
	var OrderAdd models.OrderAdd
	err := c.BindJSON(&OrderAdd)
	response := gin.H{
		"status":  false,
		"message": "failed to bind request",
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, response)
		return
	}
	if err := helper.Validate(OrderAdd); err != nil {
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
			"status":  "false",
			"message": "cart is empty",
		})
		return
	}
	var totalquandity uint
	config.DB.Raw(`SELECT SUM(qty) FROM cart_items WHERE user_id=? and deleted_at IS NULL`, userID).Scan(&totalquandity)
	if totalquandity == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "cart is empty",
		})
	}
	var totalamount float64
	config.DB.Raw("SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount)
	var totalamount1 float64
	config.DB.Raw("SELECT SUM(total_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalamount1)
	if totalamount > 100000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "orders above 100000Rs cannot be done through cash on delivery",
		})
		return
	}
	var FinalAmount float64
	var discountamount float64
	if OrderAdd.CouponCode != "" {
		var count2 int64
		config.DB.Raw(`SELECT COUNT(*) FROM coupons where code = ? and deleted_at IS NULL`, OrderAdd.CouponCode).Scan(&count2)
		if count2 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "this coupon does not exist",
			})
			return
		}
		var minpurchase float64
		config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("min_purchase", &minpurchase)
		if totalamount1 > minpurchase {
			config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("discount", &discountamount)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "minimum purchase amount is required to apply this coupon",
			})
			return
		}
	}
	var offerapplied float64
	config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&offerapplied)
	FinalAmount = totalamount - discountamount
	order := models.Order{
		UserID:         userID,
		AddressID:      OrderAdd.AddressID,
		TotalAmount:    totalamount,
		OfferApplied:   offerapplied,
		CouponCode:     OrderAdd.CouponCode,
		DiscountAmount: discountamount,
		FinalAmount:    FinalAmount,
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
			orderItem := models.OrderItems{
				OrderID:        ID,
				ProductID:      v.ProductID,
				Price:          price,
				OfferDiscount:  offerdiscount,
				CouponDiscount: coupondiscount,
				TotalDiscount:  totaldiscount,
			}
			fmt.Println("order id:", orderItem.OrderID)
			fmt.Println("order item created")
			config.DB.Create(&orderItem)
		}
	}

	Payment := models.Payments{
		UserID:      userID,
		OrderID:     order.ID,
		TotalAmount: FinalAmount,
	}
	config.DB.Create(&Payment)
	config.DB.Where("user_id = ?", userID).Delete(&models.CartItems{})
	var order1 responsemodels.Order
	var address responsemodels.Address
	var orderitems1 []responsemodels.OrderItems
	config.DB.Raw(`SELECT orders.id,orders.created_at,orders.updated_at,orders.deleted_at,orders.user_id,orders.address_id,orders.total_amount,orders.payment_method,orders.order_status,orders.offer_applied,orders.coupon_code,orders.discount_amount,orders.final_amount FROM orders join addresses on orders.address_id=addresses.id WHERE orders.user_id = ? ORDER BY orders.created_at desc LIMIT 1`, userID).Scan(&order1)
	var orderid uint
	config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&orderid)
	var addressid uint
	config.DB.Raw(`SELECT address_id FROM orders WHERE user_id = ? ORDER BY created_at desc limit 1`, userID).Scan(&addressid)
	config.DB.Raw(`SELECT * FROM addresses WHERE id = ?`, addressid).Scan(&address)
	order1.Address = address
	config.DB.Raw(`SELECT order_items.id,order_items.created_at,order_items.updated_at,order_items.deleted_at,order_items.order_id,order_items.product_id,products.product_name,order_items.price,order_items.order_status,order_items.payment_method,order_items.coupon_discount,order_items.offer_discount,order_items.total_discount,order_items.paid_amount FROM order_items join products on order_items.product_id=products.id WHERE order_items.order_id = ? ORDER BY order_items.id`, orderid).Scan(&orderitems1)
	c.JSON(http.StatusOK, gin.H{"message": "Order added successfully",
		"order":       order1,
		"order_items": orderitems1})

}*/

/*func Order(c *gin.Context) {
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

	var OrderAdd models.OrderAdd
	err := c.BindJSON(&OrderAdd)
	response := gin.H{
		"status":  false,
		"message": "failed to bind request",
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, response)
		return
	}
	if err := helper.Validate(OrderAdd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	// Check if address exists for user
	var count1 int64
	config.DB.Raw(`SELECT COUNT(*) FROM addresses WHERE id = ? AND user_id = ? AND deleted_at IS NULL`, OrderAdd.AddressID, userID).Scan(&count1)
	if count1 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "address not found",
		})
		return
	}

	// Check if cart is empty
	var totalquantity uint
	config.DB.Raw(`SELECT SUM(qty) FROM cart_items WHERE user_id = ? AND deleted_at IS NULL`, userID).Scan(&totalquantity)
	if totalquantity == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "cart is empty",
		})
		return
	}

	// Calculate total amounts
	var totalamount float64
	config.DB.Raw("SELECT SUM(final_amount) FROM cart_items WHERE user_id = ? AND deleted_at IS NULL", userID).Scan(&totalamount)
	var totalamount1 float64
	config.DB.Raw("SELECT SUM(total_amount) FROM cart_items WHERE user_id = ? AND deleted_at IS NULL", userID).Scan(&totalamount1)

	// Check cash on delivery condition
	if totalamount > 100000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "orders above 100000Rs cannot be done through cash on delivery",
		})
		return
	}

	var FinalAmount float64
	var discountamount float64

	// Coupon code processing
	if OrderAdd.CouponCode != "" {
		var count2 int64
		config.DB.Raw(`SELECT COUNT(*) FROM coupons WHERE code = ? AND deleted_at IS NULL`, OrderAdd.CouponCode).Scan(&count2)
		if count2 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "this coupon does not exist",
			})
			return
		}

		var minpurchase float64
		config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("min_purchase", &minpurchase)
		if totalamount1 >= minpurchase {
			config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("discount", &discountamount)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "minimum purchase amount is required to apply this coupon",
			})
			return
		}
	}

	var offerapplied float64
	config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&offerapplied)

	FinalAmount = totalamount - discountamount

	// Create order
	order := models.Order{
		UserID:         userID,
		AddressID:      OrderAdd.AddressID,
		TotalAmount:    totalamount,
		OfferApplied:   offerapplied,
		CouponCode:     OrderAdd.CouponCode,
		DiscountAmount: discountamount,
		FinalAmount:    FinalAmount,
	}
	config.DB.Create(&order)

	// Fetch cart items for processing order items
	var CartItems []models.CartItems
	config.DB.Where("user_id = ?", userID).Find(&CartItems)

	var ID uint
	config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&ID)

	for _, v := range CartItems {
		if v.Qty == 0 {
			continue
		}

		// Offer and coupon discount calculation
		var price float64
		config.DB.Model(&models.CartItems{}).Where("product_id = ?", v.ProductID).Pluck("price", &price)

		var offerdiscount float64
		var coupondiscount float64
		var hasoffer bool
		config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Pluck("has_offer", &hasoffer)

		if hasoffer {
			var discountpercentage uint
			config.DB.Model(&models.Offer{}).Where("product_id = ?", v.ProductID).Pluck("discount_percentage", &discountpercentage)
			offerdiscount = price * float64(discountpercentage) / 100
		}

		coupondiscount = (price / totalamount1) * discountamount
		coupondiscount = math.Round(coupondiscount*100) / 100

		totaldiscount := offerdiscount + coupondiscount

		// Create the order item based on total quantity without looping over each unit
		orderItem := models.OrderItems{
			OrderID:        ID,
			ProductID:      v.ProductID,
			Price:          price,
			OfferDiscount:  offerdiscount,
			CouponDiscount: coupondiscount,
			TotalDiscount:  totaldiscount,
		}
		config.DB.Create(&orderItem)

		// Update stock based on total quantity ordered
		var currentStock int64
		config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Pluck("stock", &currentStock)
		newStock := currentStock - int64(v.Qty) // Correct stock update
		config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Update("stock", newStock)
	}

	// Create payment record
	Payment := models.Payments{
		UserID:      userID,
		OrderID:     order.ID,
		TotalAmount: FinalAmount,
	}
	config.DB.Create(&Payment)

	// Clear cart items for the user
	config.DB.Where("user_id = ?", userID).Delete(&models.CartItems{})

	// Fetch order details for response
	var order1 responsemodels.Order
	var address responsemodels.Address
	var orderitems1 []responsemodels.OrderItems

	config.DB.Raw(`SELECT orders.id, orders.created_at, orders.updated_at, orders.deleted_at, orders.user_id, orders.address_id, orders.total_amount, orders.payment_method, orders.order_status, orders.offer_applied, orders.coupon_code, orders.discount_amount, orders.final_amount FROM orders JOIN addresses ON orders.address_id = addresses.id WHERE orders.user_id = ? ORDER BY orders.created_at DESC LIMIT 1`, userID).Scan(&order1)

	var orderid uint
	config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&orderid)
	var addressid uint
	config.DB.Raw(`SELECT address_id FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&addressid)

	config.DB.Raw(`SELECT * FROM addresses WHERE id = ?`, addressid).Scan(&address)
	order1.Address = address

	config.DB.Raw(`SELECT order_items.id, order_items.created_at, order_items.updated_at, order_items.deleted_at, order_items.order_id, order_items.product_id, products.product_name, order_items.price, order_items.order_status, order_items.payment_method, order_items.coupon_discount, order_items.offer_discount, order_items.total_discount, order_items.paid_amount FROM order_items JOIN products ON order_items.product_id = products.id WHERE order_items.order_id = ? ORDER BY order_items.id`, orderid).Scan(&orderitems1)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Order added successfully",
		"order":       order1,
		"order_items": orderitems1,
	})
}*/

package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
)

func Order(c *gin.Context) {
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
	var OrderAdd models.OrderAdd
	err := c.BindJSON(&OrderAdd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind request",
		})
		return
	}
	if err := helper.Validate(OrderAdd); err != nil {
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
			"status":  false,
			"message": "address not found",
		})
		return
	}

	var totalQuantity uint
	config.DB.Raw(`SELECT SUM(qty) FROM cart_items WHERE user_id=? and deleted_at IS NULL`, userID).Scan(&totalQuantity)
	if totalQuantity == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "cart is empty",
		})
		return
	}

	var totalAmount float64
	config.DB.Raw("SELECT SUM(final_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalAmount)
	var totalAmount1 float64
	config.DB.Raw("SELECT SUM(total_amount) from cart_items where user_id = ? and deleted_at IS NULL", userID).Scan(&totalAmount1)

	if totalAmount > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "orders above 1000 Rs cannot be done through cash on delivery",
		})
		return
	}

	var finalAmount float64
	var discountAmount float64
	if OrderAdd.CouponCode != "" {
		var count2 int64
		config.DB.Raw(`SELECT COUNT(*) FROM coupons where code = ? and deleted_at IS NULL`, OrderAdd.CouponCode).Scan(&count2)
		if count2 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "this coupon does not exist",
			})
			return
		}

		var minPurchase float64
		config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("min_purchase", &minPurchase)
		if totalAmount1 > minPurchase {
			config.DB.Model(&models.Coupon{}).Where("code = ?", OrderAdd.CouponCode).Pluck("discount", &discountAmount)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "minimum purchase amount is required to apply this coupon",
			})
			return
		}
	}

	var offerApplied float64
	config.DB.Raw(`SELECT SUM(discount) FROM cart_items WHERE deleted_at IS NULL`).Scan(&offerApplied)
	finalAmount = totalAmount - discountAmount

	order := models.Order{
		UserID:         userID,
		AddressID:      OrderAdd.AddressID,
		TotalAmount:    totalAmount,
		OfferApplied:   offerApplied,
		CouponCode:     OrderAdd.CouponCode,
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
	}
	config.DB.Create(&order)

	var cartItems []models.CartItems
	config.DB.Where("user_id = ?", userID).Find(&cartItems)

	// Fetch the latest order ID
	var orderID uint
	config.DB.Raw(`SELECT id FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(&orderID)

	for _, item := range cartItems {
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

		couponDiscount = (price / totalAmount) * discountAmount
		couponDiscount = math.Round(couponDiscount*100) / 100
		totalDiscount := offerDiscount + couponDiscount

		// Update stock of the product
		var stock uint
		config.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).Pluck("stock", &stock)

		if stock < item.Qty {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": "Insufficient stock for product ID: " + string(item.ProductID),
			})
			return
		}

		// Decrement the stock
		config.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).Update("stock", stock-item.Qty)

		// Create an OrderItem with quantity
		orderItem := models.OrderItems{
			OrderID:        orderID,
			ProductID:      item.ProductID,
			Price:          price,
			Qty:            item.Qty, // Storing quantity here
			OfferDiscount:  offerDiscount,
			CouponDiscount: couponDiscount,
			TotalDiscount:  totalDiscount,
		}
		config.DB.Create(&orderItem)
	}

	payment := models.Payments{
		UserID:      userID,
		OrderID:     order.ID,
		TotalAmount: finalAmount,
	}
	config.DB.Create(&payment)

	// Clear user's cart after order creation
	config.DB.Where("user_id = ?", userID).Delete(&models.CartItems{})

	var orderResponse responsemodels.Order1
	var orderItemsResponse []responsemodels.OrderItems

	// Fetch the order response with latest order details
	config.DB.Raw(`SELECT orders.id, orders.created_at, orders.updated_at, orders.deleted_at, orders.user_id, orders.address_id, orders.total_amount, orders.payment_method, orders.order_status, orders.offer_applied, orders.coupon_code, orders.discount_amount, orders.final_amount FROM orders WHERE orders.user_id = ? ORDER BY orders.created_at DESC LIMIT 1`, userID).Scan(&orderResponse)

	// Fetch order items response
	config.DB.Raw(`SELECT order_items.id, order_items.created_at, order_items.updated_at, order_items.deleted_at, order_items.order_id, order_items.product_id, products.product_name, order_items.price, order_items.qty, order_items.order_status, order_items.payment_method, order_items.coupon_discount, order_items.offer_discount, order_items.total_discount FROM order_items JOIN products ON order_items.product_id = products.id WHERE order_items.order_id = ? ORDER BY order_items.id`, orderID).Scan(&orderItemsResponse)

	c.JSON(http.StatusOK, gin.H{
		"message":     "order added successfully",
		"order":       orderResponse,
		"order_items": orderItemsResponse,
	})
}
