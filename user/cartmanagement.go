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

func Cart(c *gin.Context) {
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

	var cart []responsemodels.CartItems
	//qr := config.DB.Raw(`SELECT cart_items.user_id,cart_items.product_id,product.product_name,cart_items.total_amount,cart_items.qty,cart_items.price,cart_items.discount,cart_items.final_amount FROM cart_items join products on cart_items.product_id=products.id where user_id = ? AND cart_items.deleted_at IS NULL AND cart_items.qty != 0`, userID).Scan(&cart)
	qr := config.DB.Raw(`SELECT cart_items.user_id, cart_items.product_id, products.product_name, cart_items.total_amount, cart_items.qty, cart_items.price, cart_items.discount, cart_items.final_amount 
		FROM cart_items 
		JOIN products ON cart_items.product_id = products.id 
		WHERE cart_items.user_id = ? 
		AND cart_items.deleted_at IS NULL 
		AND cart_items.qty != 0`, userID).Scan(&cart)

	if qr.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "failed to retrive data from database or data does not exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully retrieved cart data",
		"data": gin.H{
			"cart_items": cart,
		},
	})
}

/*func AddCart(c *gin.Context) {
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

	var Cart models.CartAdd
	err := c.BindJSON(&Cart)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}

	if err := helper.Validate(Cart); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var cartcount1 int
	config.DB.Raw("SELECT COUNT(*) FROM products where id =? AND deleted_at IS NULL", Cart.ProductID).Scan(&cartcount1)
	if cartcount1 == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "product id does not exist",
		})
		return
	}

	var cartcount2 int
	config.DB.Raw(`SELECT COUNT(*) FROM cart_items WHERE user_id =? and product_id=? and deleted_at IS NULL`, userID, Cart.ProductID).Scan(&cartcount2)
	if cartcount2 != 0 {
		var price float64
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("price", &price)
		var hasoffer bool
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("has_offer", &hasoffer)
		var finalamount float64
		finalamount = price
		var discount float64
		if hasoffer {
			var discountpercentage uint
			config.DB.Model(&models.Offer{}).Where("product_id = ?", Cart.ProductID).Pluck("discount_percentage", &discountpercentage)
			discount = price * float64(discountpercentage) / 100
			fmt.Println("discount-", discount)
			config.DB.Raw(`UPDATE cart_items SET discount = ? WHERE product_id = ?`, discount, Cart.ProductID)
			finalamount = price - discount
		}
		var totalamount float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? and product_id = ?", userID, Cart.ProductID).Pluck("total_amount", &totalamount)
		fmt.Println("total amount: ", totalamount)
		totalamount = totalamount + price
		var Finalamount1 float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? and product_id = ?", userID, Cart.ProductID).Pluck("final_amount", &Finalamount1)
		Finalamount1 = Finalamount1 + finalamount
		var Discount1 float64
		config.DB.Model(&models.CartItems{}).Where("user_id =? and product_id = ?", userID, Cart.ProductID).Pluck("discount", &Discount1)
		Discount1 = Discount1 + discount
		fmt.Println("total amount: ", totalamount)
		var quandity uint
		config.DB.Model(&models.CartItems{}).Where("user_id = ? and product_id = ?", userID, Cart.ProductID).Pluck("qty", &quandity)
		fmt.Println("quandity: ", quandity)
		var stock uint
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("stock", &stock)
		fmt.Println("stock:", stock)
		if quandity >= 6 {
			c.JSON(http.StatusOK, gin.H{"status": true, "message": "Exceeded maximum quandity for a product"})
			return
		}
		if quandity >= stock {
			c.JSON(http.StatusOK, gin.H{"status": true, "message": "product out of stock"})
			return
		}
		quandity = quandity + 1
		fmt.Println("quandity: ", quandity)
		cart := models.CartItems{
			TotalAmount: totalamount,
			Qty:         quandity,
			Price:       price,
			Discount:    Discount1,
			FinalAmount: Finalamount1,
		}
		config.DB.Model(&models.CartItems{}).Where("user_id = ? and product_id = ? ", userID, Cart.ProductID).Updates(&cart)
		c.JSON(http.StatusOK, gin.H{"status": true, "message": "product added to cart successfully"})
		return
	}
	var price float64
	config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("price", &price)
	var hasoffer bool
	config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("has_offer", &hasoffer)
	var discount float64
	var finalamount float64
	finalamount = price
	if hasoffer {
		fmt.Println("processing offer!!")
		var discountpercentage uint
		config.DB.Model(&models.Offer{}).Where("product_id = ?", Cart.ProductID).Pluck("discount_percentage", &discountpercentage)
		discount = (price * float64(discountpercentage) / 100)
		finalamount = price - (price * float64(discountpercentage) / 100)
		fmt.Println("tottal final price: ", finalamount)

	}
	fmt.Println("final amount(outside): ", finalamount)
	cart := models.CartItems{
		UserID:      userID,
		ProductID:   Cart.ProductID,
		TotalAmount: price,
		Qty:         1,
		Price:       price,
		Discount:    discount,
		FinalAmount: finalamount,
	}
	config.DB.Create(&cart)
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "product added to cart",
	})

}*/

func AddCart(c *gin.Context) {
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

	var Cart models.CartAdd
	err := c.BindJSON(&Cart)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}

	if err := helper.Validate(Cart); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var cartcount1 int
	config.DB.Raw("SELECT COUNT(*) FROM products WHERE id = ? AND deleted_at IS NULL", Cart.ProductID).Scan(&cartcount1)
	if cartcount1 == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "product id does not exist",
		})
		return
	}

	var stock uint
	config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("stock", &stock)
	if stock == 0 {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": "product out of stock"})
		return
	}

	// Check if the product is already in the cart
	var cartcount2 int
	config.DB.Raw("SELECT COUNT(*) FROM cart_items WHERE user_id = ? AND product_id = ? AND deleted_at IS NULL", userID, Cart.ProductID).Scan(&cartcount2)
	if cartcount2 != 0 {
		// Product already in cart; update quantity and price
		var price float64
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("price", &price)
		var hasoffer bool
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("has_offer", &hasoffer)
		var discount float64
		var finalamount float64
		finalamount = price
		if hasoffer {
			var discountpercentage uint
			config.DB.Model(&models.Offer{}).Where("product_id = ?", Cart.ProductID).Pluck("discount_percentage", &discountpercentage)
			discount = price * float64(discountpercentage) / 100
			finalamount = price - discount
		}
		var totalamount float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("total_amount", &totalamount)
		totalamount += price
		var Finalamount1 float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("final_amount", &Finalamount1)
		Finalamount1 += finalamount
		var Discount1 float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("discount", &Discount1)
		Discount1 += discount
		var quantity uint
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("qty", &quantity)
		if quantity >= stock {
			c.JSON(http.StatusOK, gin.H{"status": false, "message": "product out of stock"})
			return
		}
		quantity++
		cart := models.CartItems{
			TotalAmount: totalamount,
			Qty:         quantity,
			Price:       price,
			Discount:    Discount1,
			FinalAmount: Finalamount1,
		}
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Updates(&cart)
	} else {
		// Product not in cart; add it
		var price float64
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("price", &price)
		var hasoffer bool
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("has_offer", &hasoffer)
		var discount float64
		var finalamount float64
		finalamount = price
		if hasoffer {
			var discountpercentage uint
			config.DB.Model(&models.Offer{}).Where("product_id = ?", Cart.ProductID).Pluck("discount_percentage", &discountpercentage)
			discount = price * float64(discountpercentage) / 100
			finalamount = price - discount
		}
		cart := models.CartItems{
			UserID:      userID,
			ProductID:   Cart.ProductID,
			TotalAmount: price,
			Qty:         1,
			Price:       price,
			Discount:    discount,
			FinalAmount: finalamount,
		}
		config.DB.Create(&cart)
	}

	// Decrease the stock of the product by 1
	/*if err := config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Update("stock", gorm.Expr("stock - ?", 1)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "failed to update stock"})
		return
	}*/

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "product added to cart successfully",
	})
}

/*func RemoveCart(c *gin.Context) {
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

	var Cart models.CartAdd
	err := c.BindJSON(&Cart)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}

	if err := helper.Validate(Cart); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var cartcount int
	config.DB.Raw(`SELECT COUNT(*) FROM cart_items WHERE user_id = ? AND product_id = ? and deleted_at IS NULL`, userID, Cart.ProductID).Scan(&cartcount)
	if cartcount != 0 {
		var quandity uint
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("qty", &quandity)
		if quandity == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "This item not found in cart",
			})
			return
		}
		fmt.Println("quandity:", quandity)
		quandity = quandity - 1
		fmt.Println("quandity:", quandity)
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Update("qty", quandity)
		var price float64
		config.DB.Model(&models.CartItems{}).Where("product_id = ?", Cart.ProductID).Pluck("price", &price)
		fmt.Println("price of product:", price)
		var totalamount float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("total_amount", &totalamount)
		fmt.Println("totalamount:", totalamount)
		totalamount = totalamount - price
		fmt.Println("reduced total amount:", totalamount)
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Order("total_amount DESC").Update("total_amount", totalamount)
		var hasoffer bool
		config.DB.Model(models.Product{}).Where("id = ?", Cart.ProductID).Pluck("has_offer", &hasoffer)
		fmt.Println("has offer:", hasoffer)
		var discount float64
		var finalamount float64
		if hasoffer {
			fmt.Println("processing offer")
			var discountpercentage uint
			config.DB.Model(&models.Offer{}).Where("product_id = ?", Cart.ProductID).Pluck("discount_percentage", &discountpercentage)
			discount = price * float64(discountpercentage) / 100
			finalamount = price - (price * float64(discountpercentage) / 100)
			fmt.Println("price:", finalamount)
		}
		var finalAmount1 float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? and product_id = ?", userID, Cart.ProductID).Pluck("final_amount", &finalAmount1)
		finalAmount1 = finalAmount1 - finalamount
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Update("final_amount", finalAmount1)
		var discount1 float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id =?", userID, Cart.ProductID).Pluck("discount", &discount1)
		discount1 = discount1 - discount
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Update("discount", discount1)
		c.JSON(http.StatusOK, gin.H{"status": true, "message": "product removed from cart"})
		return

	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "product does not exist in cart"})

}*/

func RemoveCart(c *gin.Context) {
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

	var Cart models.CartAdd
	err := c.BindJSON(&Cart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}

	if err := helper.Validate(Cart); err != nil {
		fmt.Println("error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	var cartcount int
	config.DB.Raw(`SELECT COUNT(*) FROM cart_items WHERE user_id = ? AND product_id = ? AND deleted_at IS NULL`, userID, Cart.ProductID).Scan(&cartcount)
	if cartcount != 0 {
		var quantity uint
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("qty", &quantity)
		if quantity == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "This item not found in cart",
			})
			return
		}
		fmt.Println("quantity:", quantity)
		quantity = quantity - 1
		fmt.Println("quantity:", quantity)
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Update("qty", quantity)

		var price float64
		config.DB.Model(&models.CartItems{}).Where("product_id = ?", Cart.ProductID).Pluck("price", &price)
		fmt.Println("price of product:", price)
		var totalamount float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("total_amount", &totalamount)
		fmt.Println("totalamount:", totalamount)
		totalamount = totalamount - price
		fmt.Println("reduced total amount:", totalamount)
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Update("total_amount", totalamount)

		var hasOffer bool
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("has_offer", &hasOffer)
		fmt.Println("has offer:", hasOffer)
		var discount float64
		var finalAmount float64
		if hasOffer {
			fmt.Println("processing offer")
			var discountPercentage uint
			config.DB.Model(&models.Offer{}).Where("product_id = ?", Cart.ProductID).Pluck("discount_percentage", &discountPercentage)
			discount = price * float64(discountPercentage) / 100
			finalAmount = price - (price * float64(discountPercentage) / 100)
			fmt.Println("price:", finalAmount)
		}
		var finalAmount1 float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("final_amount", &finalAmount1)
		finalAmount1 = finalAmount1 - finalAmount
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Update("final_amount", finalAmount1)

		var discount1 float64
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Pluck("discount", &discount1)
		discount1 = discount1 - discount
		config.DB.Model(&models.CartItems{}).Where("user_id = ? AND product_id = ?", userID, Cart.ProductID).Update("discount", discount1)

		// Increase stock by 1
		/*var stock uint
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Pluck("stock", &stock)
		config.DB.Model(&models.Product{}).Where("id = ?", Cart.ProductID).Update("stock", stock+1)*/

		c.JSON(http.StatusOK, gin.H{"status": true, "message": "product removed from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "product does not exist in cart"})
}
