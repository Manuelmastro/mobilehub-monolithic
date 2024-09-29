package user

import (
	"MOBILEHUB/config"
	"MOBILEHUB/responsemodels"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SearchProduct(c *gin.Context) {
	query := c.Query("search")
	sortName := c.Query("name_sort")
	sortPrice := c.Query("price_sort")
	latestModels := c.Query("latest_models")
	category := c.Query("category")
	var products []responsemodels.Product

	qry := `SELECT products.*, categories.category_name AS category_name FROM products
            JOIN categories ON products.category_id = categories.id`

	if query != "" {
		qry += ` WHERE (products.product_name ILIKE '%` + query + `%')`
	}

	var count int
	config.DB.Raw(`SELECT COUNT(*) FROM categories WHERE category_name = ? AND deleted_at IS NULL`, category).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "this category does not exist",
		})
		return
	}

	fmt.Println("category: ", category)
	if category != "" {
		if query != "" {
			qry += ` AND categories.category_name = '` + category + `'`
		} else {
			qry += ` WHERE categories.category_name = '` + category + `'`
		}
	}
	if sortName != "aA-zZ" && sortName != "zZ-aA" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Give correct name sort filter",
		})
		return
	}

	if sortName == "aA-zZ" {
		qry += ` ORDER BY products.product_name ASC`
	} else if sortName == "zZ-aA" {
		qry += ` ORDER BY products.product_name DESC`
	}

	if sortPrice != "low-high" && sortPrice != "high-low" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Apply correct filter for price sorting",
		})
		return
	}
	if sortPrice == "low-high" {
		if sortName != "" {
			qry += `, products.price ASC`
		} else {
			qry += ` ORDER BY products.price ASC`
		}
	} else if sortPrice == "high-low" {
		if sortName != "" {
			qry += `, products.price DESC`
		} else {
			qry += ` ORDER BY products.price DESC`
		}
	}

	if latestModels != "true" && latestModels != "false" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "latest models can either be true or false",
		})
		return
	}
	if latestModels == "true" {
		if sortName != "" || sortPrice != "" {
			qry += `, products.created_at DESC`
		} else {
			qry += ` ORDER BY products.created_at DESC`
		}
	}
	config.DB.Raw(qry).Scan(&products)
	c.JSON(http.StatusOK, products)
}
