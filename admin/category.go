package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Category(c *gin.Context) {
	var category []responsemodels.Category
	qry := `SELECT * FROM categories WHERE deleted_at IS NULL`
	qr := config.DB.Raw(qry).Scan(&category)
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
			"categories": category,
		},
	})

}

func AddCategory(c *gin.Context) {
	var Category models.CategoryEdit
	err := c.BindJSON(&Category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "request failed to bind",
		})
		return
	}
	if err := helper.Validate(Category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	var categorycount int
	config.DB.Raw(`SELECT COUNT(*) FROM categories WHERE category_name = ? AND deleted_at IS NULL `, Category.CategoryName).Scan(&categorycount)
	if categorycount != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "category name alredy exist",
		})
		return
	}
	category := models.Category{
		CategoryName: Category.CategoryName,
		Description:  Category.Description,
		ImageUrl:     Category.ImgUrl,
	}

	config.DB.Create(&category)
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "new category added"})

}

func EditCategory(c *gin.Context) {
	CategoryID := c.Param("id")
	fmt.Println("id is", CategoryID)
	//var category models.Category
	var categorycount int
	config.DB.Raw(`SELECT COUNT(*) FROM categories WHERE id = ? AND deleted_at IS NULL `, CategoryID).Scan(&categorycount)
	if categorycount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "category does not exist",
		})
		return
	}

	var Category models.CategoryEdit
	err := c.BindJSON(&Category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "request failed to bind",
		})
		return
	}
	if err := helper.Validate(Category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}
	category := models.Category{
		CategoryName: Category.CategoryName,
		Description:  Category.Description,
		ImageUrl:     Category.ImgUrl,
	}
	config.DB.Model(&models.Category{}).Where("id = ?", CategoryID).Updates(&category)
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "category updated"})

}
func DeleteCategory(c *gin.Context) {
	CategoryID := c.Param("id")
	var categorycount int
	config.DB.Raw(`SELECT COUNT(*) FROM categories WHERE id = ? AND deleted_at IS NULL `, CategoryID).Scan(&categorycount)
	if categorycount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "category does not exist",
		})
		return
	}
	config.DB.Where("id = ?", CategoryID).Delete(&models.Category{})
	config.DB.Model(&models.Product{}).Where("category_id = ?", CategoryID).Update("deleted_at", gorm.DeletedAt{Time: time.Now(), Valid: true})
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "category deleted"})

}
