package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/responsemodels"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListProducts(c *gin.Context) {
	var Products []responsemodels.Product
	qr := config.DB.Raw(`SELECT * FROM categories JOIN products ON products.category_id=categories.id WHERE products.deleted_at IS NULL AND categories.deleted_at IS NULL ORDER BY products.id`).Scan(&Products)
	if qr.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":     false,
			"message":    "failed to fetch data from database or product does not exist",
			"error_code": http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "product details fetched",
		"data": gin.H{
			"products": Products,
		},
	})
}
