/*package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Bestselling(c *gin.Context) {
	day := c.Query("day")
	month := c.Query("month")
	year := c.Query("year")

	if day != "" {
		_, err := time.Parse("2006-01-02", day)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid date format",
			})
			return
		}
	}
	if month != "" {
		int1, err := strconv.Atoi(month)
		if err != nil {
			fmt.Println("error  in string to int conversion:", err)
		}
		if int1 >= 1 && int1 <= 12 {
			fmt.Println("ok")
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid month format",
			})
			return
		}
	}
	if year != "" {
		yearint, err := strconv.Atoi(year)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid year format",
			})
			return
		}
		currentYear := time.Now().Year()
		if yearint < 2000 || yearint > currentYear {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "year must be between 2000 and curent year",
			})
			return
		}
	}
	var orderitems []models.OrderItems
	var orderitems1 []responsemodels.BestSelling

	sql := `select product_id,count(*) AS count from order_items where order_status='delivered'`
	sqla := ` group by product_id order by count DESC limit 10`
	sql1 := `select count(*),categories.category_name from order_items join products on products.id = order_items.product_id join categories on categories.id = products.category_id where order_items.order_status='delivered'`
	sql1a := ` group by categories.category_name order by count desc`

	if day != "" {
		sql = sql + ` AND DATE(created_at) ='` + day + `'` + sqla
		config.DB.Raw(sql).Scan(&orderitems)
		sql1 = sql1 + ` AND DATE(order_items.created_at) ='` + day + `'` + sql1a
		config.DB.Raw(sql1).Scan(&orderitems1)
	} else if month != "" {
		sql = sql + ` AND EXTRACT(MONTH FROM created_at)` + month + `'` + sqla
		config.DB.Raw(sql).Scan(&orderitems)
		sql1 = sql1 + ` AND EXTRACT(MONTH FROM order_items.created_at) ='` + month + `'` + sql1a
		config.DB.Raw(sql1).Scan(&orderitems1)

	} else if year != "" {
		sql = sql + ` AND EXTRACT(YEAR FROM created_at)` + year + `'` + sqla
		config.DB.Raw(sql).Scan(&orderitems)
		sql1 = sql1 + ` AND EXTRACT(YEAR FROM order_items.created_at) ='` + year + `'` + sql1a
		config.DB.Raw(sql1).Scan(&orderitems1)
	} else {
		config.DB.Raw(sql).Scan(&orderitems)
		config.DB.Raw(sql1).Scan(&orderitems1)
	}
	var product1 responsemodels.Product
	var products []responsemodels.Product
	for _, v := range orderitems {
		fmt.Println("product id count:", v.ProductID)
		config.DB.Raw("SELECT * FROM products join categories on products.category_id=categories.id WHERE products.id = ?", v.ProductID).Scan(&product1)
		products = append(products, product1)
	}
	var category []string
	for _, v := range orderitems1 {
		fmt.Println("category-count:", v.Count)
		category = append(category, v.CategoryName)
	}
	c.JSON(http.StatusOK, gin.H{
		"best_selling_product":  products,
		"best_selling_category": category,
		"message":               "retrieved best selling details",
	})

}*/

package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Bestselling(c *gin.Context) {
	day := c.Query("day")
	month := c.Query("month")
	year := c.Query("year")

	// Validate day format
	if day != "" {
		_, err := time.Parse("2006-01-02", day)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid date format",
			})
			return
		}
	}

	// Validate month format
	if month != "" {
		int1, err := strconv.Atoi(month)
		if err != nil {
			fmt.Println("Error in string to int conversion:", err)
		}
		if int1 >= 1 && int1 <= 12 {
			fmt.Println("ok")
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid month format",
			})
			return
		}
	}

	// Validate year format
	if year != "" {
		yearint, err := strconv.Atoi(year)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid year format",
			})
			return
		}
		currentYear := time.Now().Year()
		if yearint < 2000 || yearint > currentYear {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Year must be between 2000 and the current year",
			})
			return
		}
	}

	// Define variables for database queries
	var orderitems []models.OrderItems
	var orderitems1 []responsemodels.BestSelling

	// Base SQL queries for products and categories
	sql := `SELECT product_id, COUNT(*) AS count FROM order_items WHERE order_status='delivered'`
	sqla := ` GROUP BY product_id ORDER BY count DESC LIMIT 10`
	sql1 := `SELECT COUNT(*), categories.category_name FROM order_items JOIN products ON products.id = order_items.product_id JOIN categories ON categories.id = products.category_id WHERE order_items.order_status='delivered'`
	sql1a := ` GROUP BY categories.category_name ORDER BY COUNT(*) DESC`

	// Build query based on day, month, and year filters
	if day != "" {
		sql = sql + ` AND DATE(created_at) ='` + day + `'` + sqla
		config.DB.Raw(sql).Scan(&orderitems)
		sql1 = sql1 + ` AND DATE(order_items.created_at) ='` + day + `'` + sql1a
		config.DB.Raw(sql1).Scan(&orderitems1)
	} else if month != "" {
		sql = sql + ` AND EXTRACT(MONTH FROM created_at) =` + month + sqla
		config.DB.Raw(sql).Scan(&orderitems)
		sql1 = sql1 + ` AND EXTRACT(MONTH FROM order_items.created_at) =` + month + sql1a
		config.DB.Raw(sql1).Scan(&orderitems1)
	} else if year != "" {
		sql = sql + ` AND EXTRACT(YEAR FROM created_at) =` + year + sqla
		config.DB.Raw(sql).Scan(&orderitems)
		sql1 = sql1 + ` AND EXTRACT(YEAR FROM order_items.created_at) =` + year + sql1a
		config.DB.Raw(sql1).Scan(&orderitems1)
	} else {
		config.DB.Raw(sql + sqla).Scan(&orderitems)
		config.DB.Raw(sql1 + sql1a).Scan(&orderitems1)
	}

	// Fetch product and category details
	var product1 responsemodels.Product
	var products []responsemodels.Product
	for _, v := range orderitems {
		fmt.Println("product id count:", v.ProductID)
		config.DB.Raw("SELECT * FROM products JOIN categories ON products.category_id=categories.id WHERE products.id = ?", v.ProductID).Scan(&product1)
		products = append(products, product1)
	}

	var category []string
	for _, v := range orderitems1 {
		fmt.Println("category-count:", v.Count)
		category = append(category, v.CategoryName)
	}

	type SalesReportSummary struct {
		TotalQuandity         uint    `json:"total_quandity"`
		TotalPaidAmount       float64 `json:"total_paid_amount"`
		TotalDiscount         float64 `json:"total_discount"`
		TottalPendingOrder    int64   `json:"total_pendingorders"`
		TottalCanceledOrder   int64   `json:"total_cancelledorders"`
		Tottaldeliveredorders int64   `json:"total_deliveredorders"`
		TottalCouponsRedeemed int64   `json:"total_coupons_redeemed"`
	}
	var summary SalesReportSummary
	qry1 := `SELECT SUM(qty)AS total_quantity, SUM(paid_amount) AS total_paid_amount, SUM(total_discount) AS total_discount FROM sales_report_items `
	if day != "" {
		qry1 = qry1 + `WHERE DATE(order_date) ='` + day + `'`
		config.DB.Raw(qry1).Scan(&summary)
	} else if month != "" {
		qry1 = qry1 + `WHERE EXTRACT(MONTH FROM order_date) = '` + month + `'`
		config.DB.Raw(qry1).Scan(&summary)
	} else if year != "" {
		qry1 = qry1 + `WHERE EXTRACT(YEAR FROM order_date) = '` + year + `'`
		config.DB.Raw(qry1).Scan(&summary)
	} else {
		config.DB.Raw(qry1).Scan(&summary)
	}

	// qry2 := `SELECT COUNT(*) AS total_pendingorders FROM orders WHERE order_status = 'pending' `
	// qry3 := `SELECT COUNT(*) AS total_cancelledorders FROM orders WHERE order_status = 'cancelled' `
	// qry4 := `SELECT COUNT(*) AS total_coupons_redeemed FROM orders WHERE coupon_code != '' `
	// if day!= ""{
	// 	qry2 = qry2 + `AND DATE(created_at) = '` + day + `'`
	// 	qry3 = qry3 + `AND DATE(created_at) = '` + day + `'`
	// 	qry4 = qry4 + `AND DATE(created_at) = '` + day + `'`
	// 	config.DB.Raw(qry2).Scan(&summary)
	// 	config.DB.Raw(qry3).Scan(&summary)
	// 	config.DB.Raw(qry4).Scan(&summary)

	// } else if month != ""{
	// 	qry2 = qry2 + `AND EXTRACT(MONTH FROM order_date) = '` + month + `'`
	// 	qry3 = qry3 + `AND EXTRACT(MONTH FROM order_date) = '` + month + `'`
	// 	qry4 = qry4 + `AND EXTRACT(MONTH FROM order_date) = '` + month + `'`
	// 	config.DB.Raw(qry2).Scan(&summary)
	// 	config.DB.Raw(qry3).Scan(&summary)
	// 	config.DB.Raw(qry4).Scan(&summary)

	// } else if year != ""{
	// 	qry2 = qry2 + `AND EXTRACT(YEAR FROM order_date) = '` + year + `'`
	// 	qry3 = qry3 + `AND EXTRACT(YEAR FROM order_date) = '` + year + `'`
	// 	qry4 = qry4 + `AND EXTRACT(YEAR FROM order_date) = '` + year + `'`
	// 	config.DB.Raw(qry2).Scan(&summary)
	// 	config.DB.Raw(qry3).Scan(&summary)
	// 	config.DB.Raw(qry4).Scan(&summary)
	// }else{
	// 	config.DB.Raw(qry2).Scan(&summary)
	// 	config.DB.Raw(qry3).Scan(&summary)
	// 	config.DB.Raw(qry4).Scan(&summary)

	// }

	qry2 := `SELECT COUNT(*) AS total_pending_orders FROM orders WHERE order_status = 'pending'`
	qry3 := `SELECT COUNT(*) AS total_cancelled_orders FROM orders WHERE order_status = 'cancelled'`
	qry5 := `SELECT COUNT(*) AS total_delivered_orders FROM orders WHERE order_status = 'delivered'`
	qry4 := `SELECT COUNT(*) AS total_coupons_redeemed FROM orders WHERE coupon_code != ''`

	if day != "" {
		qry2 = qry2 + ` AND DATE(created_at) = '` + day + `'`
		qry3 = qry3 + ` AND DATE(created_at) = '` + day + `'`
		qry4 = qry4 + ` AND DATE(created_at) = '` + day + `'`
		qry5 = qry5 + ` AND DATE(created_at) = '` + day + `'`

		config.DB.Raw(qry2).Scan(&summary.TottalPendingOrder)
		config.DB.Raw(qry3).Scan(&summary.TottalCanceledOrder)
		config.DB.Raw(qry4).Scan(&summary.TottalCouponsRedeemed)
		config.DB.Raw(qry5).Scan(&summary.Tottaldeliveredorders)

	} else if month != "" {
		qry2 = qry2 + ` AND EXTRACT(MONTH FROM created_at) = '` + month + `'`
		qry3 = qry3 + ` AND EXTRACT(MONTH FROM created_at) = '` + month + `'`
		qry4 = qry4 + ` AND EXTRACT(MONTH FROM created_at) = '` + month + `'`
		qry5 = qry5 + ` AND EXTRACT(MONTH FROM created_at) = '` + month + `'`

		config.DB.Raw(qry2).Scan(&summary.TottalPendingOrder)
		config.DB.Raw(qry3).Scan(&summary.TottalCanceledOrder)
		config.DB.Raw(qry4).Scan(&summary.TottalCouponsRedeemed)
		config.DB.Raw(qry5).Scan(&summary.Tottaldeliveredorders)

	} else if year != "" {
		qry2 = qry2 + ` AND EXTRACT(YEAR FROM created_at) = '` + year + `'`
		qry3 = qry3 + ` AND EXTRACT(YEAR FROM created_at) = '` + year + `'`
		qry4 = qry4 + ` AND EXTRACT(YEAR FROM created_at) = '` + year + `'`
		qry5 = qry5 + ` AND EXTRACT(YEAR FROM created_at) = '` + year + `'`

		config.DB.Raw(qry2).Scan(&summary.TottalPendingOrder)
		config.DB.Raw(qry3).Scan(&summary.TottalCanceledOrder)
		config.DB.Raw(qry4).Scan(&summary.TottalCouponsRedeemed)
		config.DB.Raw(qry5).Scan(&summary.Tottaldeliveredorders)

	} else {
		config.DB.Raw(qry2).Scan(&summary.TottalPendingOrder)
		config.DB.Raw(qry3).Scan(&summary.TottalCanceledOrder)
		config.DB.Raw(qry4).Scan(&summary.TottalCouponsRedeemed)
		config.DB.Raw(qry5).Scan(&summary.Tottaldeliveredorders)
	}

	// Return the response with best selling products and categories
	c.JSON(http.StatusOK, gin.H{
		"best_selling_product":  products,
		"best_selling_category": category,
		//"overall_sales_count":           summary.TotalQuandity,
		//"overall_order_amount":          summary.TotalPaidAmount,
		//"overall_discount":              summary.TotalDiscount,
		"tottal_pending_orders":         summary.TottalPendingOrder,
		"tottal_cancelled_orders":       summary.TottalCanceledOrder,
		"total_delivered_orders":        summary.Tottaldeliveredorders,
		"tottal_times_coupons_redeemed": summary.TottalCouponsRedeemed,
		"message":                       "Retrieved best selling details",
	})

}
