package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CouponList(c *gin.Context) {
	var coupon []responsemodels.Coupon
	config.DB.Raw(`SELECT * FROM coupons WHERE deleted_at IS NULL`).Scan(&coupon)
	c.JSON(http.StatusOK, gin.H{
		"data":    coupon,
		"message": "listed coupons cuccessfully",
	})
}

func CouponAdd(c *gin.Context) {
	var couponadd models.CouponAdd
	err := c.BindJSON(&couponadd)
	response := gin.H{
		"status":  false,
		"message": "binding of data failed",
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, response)
		return
	}
	if err := helper.Validate(couponadd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM coupons WHERE code = ? AND deleted_at IS NULL`, couponadd.Code).Scan(&count)
	if count != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this cupon already exists",
		})
		return
	}
	coupon := models.Coupon{
		Code:        couponadd.Code,
		Discount:    couponadd.Discount,
		MinPurchase: couponadd.MinPurchase,
	}
	config.DB.Create(&coupon)
	c.JSON(http.StatusOK, gin.H{
		"message": "coupon added successfully",
	})
}
func CouponRemove(c *gin.Context) {
	CouponID := c.Param("id")
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM coupons WHERE id = ? AND deleted_at IS NULL`, CouponID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this coupon not exists",
		})
		return
	}
	config.DB.Where("id = ?", CouponID).Delete(&models.Coupon{})
	c.JSON(http.StatusOK, gin.H{
		"message": "Coupon deleted successfully",
	})
}
