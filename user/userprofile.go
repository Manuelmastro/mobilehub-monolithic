package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/midleware"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Profile(c *gin.Context) {
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
	var User responsemodels.User
	config.DB.Where("id = ?", userID).First(&User)
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully fetched user informations",
		"data": gin.H{
			"User": User,
		},
	})
}

func EditProfile(c *gin.Context) {
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

	var Profile models.ProfileEdit
	err := c.BindJSON(&Profile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind data",
		})
		return
	}
	//validation of data fetched from input
	if err := helper.Validate(Profile); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var phonecount int
	config.DB.Raw(`SELECT COUNT(*) FROM users where phone = ?`, Profile.Phone).Scan(&phonecount)
	if phonecount != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "this number alredy exists"})
		return
	}
	user := models.User{
		FirstName: Profile.FirstName,
		LastName:  Profile.LastName,
		Phone:     Profile.Phone,
	}
	config.DB.Model(&models.User{}).Where("id = ?", userID).Updates(&user)
	c.JSON(http.StatusOK, gin.H{"message": "User profile updated"})

}

func ChangePassword(c *gin.Context) {
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

	var PasswordChange models.PasswordChange

	err := c.BindJSON(&PasswordChange)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind data",
		})
		return
	}

	if err := helper.Validate(PasswordChange); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if PasswordChange.Password != PasswordChange.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    "passwords should match",
			"error_code": http.StatusBadRequest,
		})
		return
	}

	Passwordchange := models.User{
		Password: PasswordChange.Password,
	}

	config.DB.Model(&models.User{}).Where("id = ?", userID).Updates(&Passwordchange)
	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})

}

func OrderList(c *gin.Context) {
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
	var orders []responsemodels.Order
	var address responsemodels.Address
	qry := `SELECT orders.id,orders.created_at,orders.updated_at,orders.deleted_at,orders.user_id,orders.address_id,orders.total_amount,orders.offer_applied,orders.order_status,orders.coupon_code,orders.discount_amount,orders.final_amount,orders.payment_method,addresses.user_id,addresses.country,addresses.state,addresses.street_name,addresses.district,addresses.pin_code,addresses.phone,addresses.default
	FROM orders
	JOIN addresses ON orders.address_id = addresses.id where orders.user_id = ?`
	config.DB.Raw(qry, userID).Scan(&orders)

	for i, v := range orders {
		config.DB.Raw(`SELECT *
	        FROM orders
	        JOIN addresses ON orders.address_id = addresses.id
	        WHERE orders.user_id = ? AND orders.id = ?`, userID, v.ID).Scan(&address)
		orders[i].Address = address
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "order details retrived",
		"data": gin.H{
			"Order": orders,
		},
	})
}

func OrderItemsListt(c *gin.Context) {
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
	orderId := c.Param("order_id")
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM orders where id = ? AND user_id = ?`, orderId, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order id does not exist",
		})
		return
	}
	var orderitems []responsemodels.OrderItems
	qry := `SELECT order_items.id,order_items.created_at,order_items.updated_at,order_items.deleted_at,order_items.order_id,order_items.product_id,products.product_name,order_items.qty,order_items.price,order_items.order_status,order_items.payment_method,order_items.coupon_discount,order_items.offer_discount,order_items.total_discount,order_items.paid_amount FROM order_items join products on order_items.product_id=products.id WHERE order_items.order_id = ?`
	config.DB.Raw(qry, orderId).Scan(&orderitems)
	c.JSON(http.StatusOK, gin.H{
		"order items": orderitems,
	})
}

/*func CancelOrder(c *gin.Context) {
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
	orderID := c.Param("order_id")
	var orderid uint
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("id", &orderid)
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM orders where id = ? AND user_id = ?`, orderID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order id not found for this user",
		})
		return
	}
	var orderstatus string
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("order_status", &orderstatus)
	if orderstatus == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this item is alredy cancelled",
		})
		return
	}
	if orderstatus == "shipped" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this item alredy shipped",
		})
		return
	}
	if orderstatus == "delivered" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "ths item alredy deliveres ",
		})
		return
	}
	if orderstatus == "failed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order is alredy failed due to some reasons",
		})
		return
	}
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("order_status", "cancelled")
	config.DB.Model(&models.OrderItems{}).Where("order_id = ?", orderID).Update("order_status", "cancelled")
	var paymentmethod string
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("payment_method", &paymentmethod)
	var totalamount float64
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("final_amount", &totalamount)
	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})

}*/

func CancelOrder(c *gin.Context) {
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
	orderID := c.Param("order_id")
	var orderid uint
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("id", &orderid)

	// Check if the order exists for the user
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM orders WHERE id = ? AND user_id = ?`, orderID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order id not found for this user",
		})
		return
	}

	// Fetch the current order status
	var orderStatus string
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("order_status", &orderStatus)
	switch orderStatus {
	case "cancelled":
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this item is already cancelled",
		})
		return
	case "shipped":
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this item already shipped",
		})
		return
	case "delivered":
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this item already delivered",
		})
		return
	case "failed":
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order is already failed due to some reasons",
		})
		return
	}

	// Update the order and order items status to "cancelled"
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("order_status", "cancelled")
	config.DB.Model(&models.OrderItems{}).Where("order_id = ?", orderID).Update("order_status", "cancelled")

	var paymentmethod string
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("payment_method", &paymentmethod)
	var totalamount float64
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("final_amount", &totalamount)
	if paymentmethod == "RazorPay" || paymentmethod == "Wallet" {
		var count int64
		config.DB.Raw(`SELECT COUNT(*) FROM wallets WHERE user_id = ?`, userID).Scan(&count)
		if count == 0 {
			wallet := models.Wallet{
				UserID:  userID,
				Balance: totalamount,
			}
			config.DB.Create(&wallet)
		} else {
			var balance float64
			config.DB.Model(&models.Wallet{}).Where("user_id = ?", userID).Pluck("balance", &balance)
			balance = balance + totalamount
			wallet := models.Wallet{
				UserID:  userID,
				Balance: balance,
			}
			config.DB.Model(&models.Wallet{}).Where("user_id = ?", userID).Updates(&wallet)
		}
		transaction := models.WalletTransaction{
			UserID:          userID,
			Amount:          totalamount,
			TransactionType: "Credit",
			Description:     "Refund for cancelling order",
		}
		config.DB.Create(&transaction)
		now := time.Now()
		today := now.Format("2024-01-02")
		payment := models.Payments{
			UserID:        userID,
			OrderID:       orderid,
			TotalAmount:   totalamount,
			PaymentDate:   today,
			PaymentType:   paymentmethod,
			PaymentStatus: "refund",
			Description:   "refund for cancelling order",
		}
		config.DB.Create(&payment)
	}

	// Fetch order items to update stock
	var orderItems []models.OrderItems
	config.DB.Where("order_id = ?", orderID).Find(&orderItems)

	// Update stock for each product based on the cancelled quantities
	for _, item := range orderItems {
		config.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).
			UpdateColumn("stock", gorm.Expr("stock + ?", item.Qty))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

/*--------------------------------------addition-------------------------------------*/
func CancelSingleOrderItem(c *gin.Context) {
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
	ItemID := c.Param("orderitem_id")
	var itemid uint
	config.DB.Model(&models.OrderItems{}).Where("id = ?", itemid).Pluck("id", &itemid)
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM order_items where id = ?`, ItemID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order item id not exist for this user",
		})
		return
	}
	var orderstatus string
	config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("order_status", &orderstatus)
	fmt.Println("order status:", orderstatus)
	if orderstatus == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order item is alredy cancelled",
		})
		return
	}
	if orderstatus == "return" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order item is alredy returnes",
		})
		return
	}
	if orderstatus == "shipped" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "item is alredy shipped",
		})
		return
	}
	if orderstatus == "delivered" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "item is already delivered",
		})
		return
	}
	if orderstatus == "failed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order is already failed due to some issues",
		})
		return
	}
	var price float64
	config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("price", &price)
	var orderid uint
	config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("order_id", &orderid)
	fmt.Println("price:", price)
	var offerdiscount float64
	config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("offer_discount", &offerdiscount)
	fmt.Println("offer discount", offerdiscount)
	offerdiscounted := price - offerdiscount
	var totalamount float64
	config.DB.Model(&models.Order{}).Where("id = ?", orderid).Pluck("total_amount", &totalamount)
	var offerapplied float64
	config.DB.Model(&models.Order{}).Where("id = ?", orderid).Pluck("offer_applied", &offerapplied)
	total := totalamount + offerapplied
	var cuponcode string
	config.DB.Raw(`select coupon_code from orders where id = ?`, orderid).Scan(&cuponcode)
	var minpurchase float64
	config.DB.Raw(`select min_purchase from coupons where code = ?`, cuponcode).Scan(&minpurchase)
	totalamount = totalamount - offerdiscounted
	config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("total_amount", totalamount)
	fmt.Println("total amount:", totalamount)
	var paidamount float64
	var paymentmethod string
	var cupondiscount float64
	if total-price < minpurchase {
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("final_amount", totalamount)
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("coupon_code", "")
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("discount_amount", 0.00)
		config.DB.Model(&models.OrderItems{}).Where("order_id = ?", orderid).Update("coupon_discount", 0.00)
		var OrderItems []models.OrderItems
		config.DB.Where("order_id = ?", orderid).Find(&OrderItems)

		for _, v := range OrderItems {
			config.DB.Model(&models.OrderItems{}).Where("id = ?", v.ID).Update("total_discount", v.OfferDiscount)
			if v.PaymentMethod == "RazorPay" || v.PaymentMethod == "Wallet" {
				var orderitem models.OrderItems
				config.DB.Raw(`UPDATE order_items SET paid_amount = ? WHERE id = ?`, v.Price-v.OfferDiscount, v.ID).Scan(&orderitem)

			}
		}

	} else {
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("paid_amount", &paidamount)
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("payment_method", &paymentmethod)
		var totaldiscount float64
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("total_discount", &totaldiscount)
		if paymentmethod == "COD" {
			paidamount = price - totaldiscount
		}
		var finalamount float64
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("final_amount", finalamount)
		finalamount = finalamount - paidamount
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("final_amount", finalamount)
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("cupon_discount", &cupondiscount)
		var discountamount float64
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Pluck("discount_amount", &discountamount)
		discountamount = discountamount - cupondiscount
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("discount_amount", discountamount)

	}
	var count1 int
	config.DB.Raw(`SELECT COUNT(*) FROM order_items WHERE order_id = ? AND order_status != 'cancelled' AND order_status != 'return'`, orderid).Scan(&count1)
	if count1 == 1 {
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("order_status", "cancelled")
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("final_amount", 0.00)
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("discount_amount", 0.00)
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("offer_applied", 0.00)
	}
	config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Update("order_status", "cancelled")
	if paymentmethod == "RazorPay" || paymentmethod == "Wallet" {
		if total-price < minpurchase {
			var discount float64
			config.DB.Raw(`SELECT discount from coupons where code = ? and deleted_at is null`, cuponcode).Scan((&discount))
			paidamount = offerdiscounted - discount
		}
		var count int64
		config.DB.Raw(`SELECT COUNT(*) FROM wallets WHERE user_id = ?`, userID).Scan(&count)
		if count == 0 {
			wallet := models.Wallet{
				UserID:  userID,
				Balance: paidamount,
			}
			config.DB.Create(&wallet)
		} else {
			var balance float64
			config.DB.Model(&models.Wallet{}).Where("user_id = ?", userID).Pluck("balance", &balance)
			balance = balance + paidamount
			wallet := models.Wallet{
				UserID:  userID,
				Balance: balance,
			}
			config.DB.Model(&models.Wallet{}).Where("user_id = ?", userID).Updates(&wallet)
		}
		transaction := models.WalletTransaction{
			UserID:          userID,
			Amount:          paidamount,
			TransactionType: "Credit",
			Description:     "Refund for cancelling single order item",
		}
		config.DB.Create(&transaction)
		now := time.Now()
		today := now.Format("2024-01-02")
		payment := models.Payments{
			UserID:        userID,
			OrderID:       orderid,
			OrderItemID:   itemid,
			TotalAmount:   paidamount,
			PaymentDate:   today,
			PaymentType:   paymentmethod,
			PaymentStatus: "refund",
			Description:   "refund for cancelling single order item",
		}
		config.DB.Create(&payment)
	}
	var productID uint
	var quantity uint
	config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("product_id", &productID)
	config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("qty", &quantity)
	var product models.Product
	config.DB.Model(&models.Product{}).Where("id = ?", productID).First(&product)
	newStock := product.Stock + quantity
	config.DB.Model(&models.Product{}).Where("id = ?", productID).Update("stock", newStock)

	config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("offer_applied", offerapplied-offerdiscount)
	c.JSON(http.StatusOK, gin.H{"message": "Order item cancelled successfully"})

}

/*--------------------------------------addition---------------------------------------------------------------------------*/

func ReturnSingleOrderItem(c *gin.Context) {
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
	fmt.Println("user id:", userID)
	ItemID := c.Param("orderitem_id")
	var itemid uint
	config.DB.Model(models.OrderItems{}).Where("id = ?", ItemID).Pluck("id", &itemid)
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM order_items where id = ?`, ItemID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "order item not exist for this user"})
		return
	}
	var orderstatus string
	config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("order_status", &orderstatus)
	fmt.Println("order status:", orderstatus)
	if orderstatus == "delivered" {
		var paidamount float64
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("paid_amount", &paidamount)
		var orderid uint
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("order_id", &orderid)
		var totalamount float64
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Pluck("total_amount", &totalamount)
		var price float64
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("price", &price)
		var offerdiscount float64
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("offer_discount", &offerdiscount)
		offerdiscounted := price - offerdiscount
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("price", &price)
		totalamount = totalamount - offerdiscounted
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("total_amount", totalamount)
		var finalamount float64
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Pluck("final_amount", &finalamount)
		finalamount = finalamount - paidamount
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("final_amount", finalamount)
		var cupondiscount float64
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("cupon_discount", &cupondiscount)
		var discountamount float64
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Pluck("discount_amount", &discountamount)
		discountamount = discountamount - cupondiscount
		config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("discount_amount", discountamount)
		var balance float64
		config.DB.Model(&models.Wallet{}).Where("user_id = ?", userID).Pluck("balance", &balance)
		balance = balance + paidamount
		config.DB.Model(&models.Wallet{}).Where("user_id = ?", userID).Update("balance", balance)
		var count1 int
		config.DB.Raw(`SELECT COUNT(*) FROM order_items WHERE order_id = ? AND order_status != 'cancelled' AND order_status != 'return'`, orderid).Scan(&count1)
		if count1 == 1 {
			config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("order_status", "cancelled")
			config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("final_amount", 0.00)
			config.DB.Model(&models.Order{}).Where("id = ?", orderid).Update("discount_amount", 0.00)
		}
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Update("order_status", "return")
		transaction := models.WalletTransaction{
			UserID:          userID,
			Amount:          paidamount,
			TransactionType: "credit",
			Description:     "Refund for return of single order item",
		}
		config.DB.Create(&transaction)
		now := time.Now()
		today := now.Format("2006-01-02")
		var paymentmethod string
		config.DB.Model(&models.OrderItems{}).Where("id = ?", ItemID).Pluck("payment_method", &paymentmethod)
		payment := models.Payments{
			UserID:        userID,
			OrderID:       orderid,
			OrderItemID:   itemid,
			TotalAmount:   paidamount,
			PaymentDate:   today,
			PaymentType:   paymentmethod,
			PaymentStatus: "refund",
			Description:   "refund for returning single order item",
		}
		config.DB.Create(&payment)
		c.JSON(http.StatusOK, gin.H{
			"message": "order returned successfully",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "cannot return an item until it is delivered",
		})
	}
}
