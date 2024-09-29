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

func Wishlist(c *gin.Context) {
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
	var wishlist []responsemodels.Wishlist
	config.DB.Raw(`SELECT wishlists.id,wishlists.created_at,wishlists.updated_at,wishlists.deleted_at,wishlists.user_id,wishlists.product_id,products.product_name,categories.category_name,products.description,products.image_url,products.price,products.stock,products.has_offer,products.offer_discount_percent FROM wishlists JOIN products ON wishlists.product_id = products.id JOIN categories ON categories.id = products.category_id  WHERE wishlists.user_id = ? AND wishlists.deleted_at IS NULL`, userID).Scan(&wishlist)
	c.JSON(http.StatusOK, gin.H{
		"data":    wishlist,
		"message": "wishlist data retrived succesfully",
	})

}

func AddWishlist(c *gin.Context) {

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

	var wishlistadd models.WishlistAdd
	err := c.BindJSON(&wishlistadd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}
	if err := helper.Validate(wishlistadd); err != nil {
		fmt.Println("validation error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	var count int
	config.DB.Raw(`SELECT COUNT(*) FROM wishlists WHERE product_id = ? AND deleted_at IS NULL`, wishlistadd.ProductID).Scan(&count)
	if count != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "product already exists in wishlist",
		})
		return
	}
	wishlist := models.Wishlist{
		UserID:    userID,
		ProductID: wishlistadd.ProductID,
	}
	config.DB.Create(&wishlist)
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "added to wishlist",
	})

}

func RemoveWishlist(c *gin.Context) {
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

	var wishlistadd models.WishlistAdd
	err := c.BindJSON(&wishlistadd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "binding of data failed",
		})
		return
	}
	if err := helper.Validate(wishlistadd); err != nil {
		fmt.Println("validation error:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	var count int
	config.DB.Raw(`SELECT COUNT(*) FROM wishlists WHERE product_id = ? AND user_id = ? AND deleted_at IS NULL`, wishlistadd.ProductID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "product does not exist in wishlist",
		})
		return
	}
	config.DB.Where("product_id = ? AND user_id = ?", wishlistadd.ProductID, userID).Delete(&models.Wishlist{})
	c.JSON(http.StatusOK, gin.H{
		"message": "product removed from wishlist",
	})

}
