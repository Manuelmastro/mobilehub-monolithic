package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Product(c *gin.Context) {
	var product []responsemodels.Product
	qry := `SELECT * FROM categories join products on categories.id=products.category_id and products.deleted_at IS NULL AND categories.deleted_at IS NULL ORDER BY products.id`
	qr := config.DB.Raw(qry).Scan(&product)
	if qr.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "false",
			"message": "failed to fetch data from database, or data does not exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "successfully fetched category details",
		"data": gin.H{
			"products": product,
		},
	})
}

func AddProduct(c *gin.Context) {

	var Product models.ProductAdd

	err := c.BindJSON(&Product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "request failed to bind",
		})
		return
	}
	if err := helper.Validate(Product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	if Product.Price != float64(int(Product.Price)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "price should not contain decimal places",
		})
		return
	}

	var productcount int
	err = config.DB.Raw(`SELECT COUNT(*) FROM categories WHERE categories.category_name=? and categories.deleted_at is NULL`, Product.CategoryName).Scan(&productcount).Error
	if err != nil {
		fmt.Println("failed to execute query", err)
	}
	if productcount != 0 {
		var categoryid uint
		config.DB.Raw(`SELECT id from categories where category_name = ?`, Product.CategoryName).Scan(&categoryid)
		product := models.Product{
			CategoryID:  categoryid,
			ProductName: Product.ProductName,
			Description: Product.Description,
			ImageUrl:    Product.ImageUrl,
			Price:       Product.Price,
			Stock:       Product.Stock,
		}
		config.DB.Create(&product)

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category does not exist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Product Added Successfully"})

}

func EditProduct(c *gin.Context) {
	productID := c.Param("id")
	var productcount int
	config.DB.Raw(`SELECT COUNT(*) FROM products WHERE id = ? AND deleted_at IS NULL`, productID).Scan(&productcount)
	if productcount == 0 {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "product does not exist",
		})
		return
	}

	var Product models.ProductEdit
	err := c.BindJSON(&Product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "failed to bind request",
		})
	}
	if err := helper.Validate(Product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	var productcount1 int
	err = config.DB.Raw(`SELECT COUNT(*) FROM categories join products on categories.id=products.category_id WHERE categories.category_name=? and categories.deleted_at is NULL`, Product.CategoryName).Scan(&productcount1).Error
	if err != nil {
		fmt.Println("failed to execute query", err)
	}
	if productcount1 != 0 {
		var categoryid uint
		config.DB.Raw(`SELECT id from categories where category_name = ?`, Product.CategoryName).Scan(&categoryid)
		product := models.Product{
			CategoryID:  categoryid,
			ProductName: Product.ProductName,
			Description: Product.Description,
			ImageUrl:    Product.ImageUrl,
			Price:       Product.Price,
			Stock:       Product.Stock,
		}
		config.DB.Model(&models.Product{}).Where("id = ?", productID).Updates(&product)

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category does not exist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Product updated Successfully"})
}

func DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	var productcount int
	config.DB.Raw(`SELECT COUNT(*) FROM products WHERE id = ? AND deleted_at IS NULL`, productID).Scan(&productcount)
	if productcount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "product does not exist",
		})
		return
	}

	config.DB.Where("id = ?", productID).Delete(&models.Product{})
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "product deleted",
	})

}
