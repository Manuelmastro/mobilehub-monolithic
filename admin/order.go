package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/helper"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func OrderList(c *gin.Context) {
	//listorder:= c.Query("list_order")
	var orders []responsemodels.Order
	var address responsemodels.Address
	qry := `SELECT orders.id,orders.created_at,orders.updated_at,orders.deleted_at,orders.user_id,orders.address_id,orders.total_amount,orders.payment_method,orders.order_status,orders.offer_applied,orders.coupon_code,orders.discount_amount,orders.final_amount,addresses.created_at,addresses.updated_at,addresses.deleted_at,addresses.user_id,addresses.country,addresses.state,addresses.street_name,addresses.district,addresses.pin_code,addresses.phone,addresses.default
	FROM orders
	JOIN addresses ON orders.address_id = addresses.id`
	config.DB.Raw(qry).Scan(&orders)
	for i, v := range orders {
		config.DB.Raw(`SELECT *
	        FROM orders
	        JOIN addresses ON orders.address_id = addresses.id
	        WHERE orders.id = ?`, v.ID).Scan(&address)
		orders[i].Address = address
	}
	c.JSON(http.StatusOK, gin.H{
		"order": orders,
	})

}

func OrderItemsList(c *gin.Context) {
	orderId := c.Param("order_id")
	//listorder := c.Query("list_order")
	var count int
	config.DB.Raw(`SELECT COUNT(*) FROM orders where id = ?`, orderId).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order id does not exist",
		})
		return
	}
	var orderitems []responsemodels.OrderItems

	qry := `SELECT order_items.id,order_items.created_at,order_items.updated_at,order_items.deleted_at,order_items.order_id,order_items.product_id,products.product_name,order_items.price,order_items.order_status,order_items.payment_method,order_items.coupon_discount,order_items.offer_discount,order_items.total_discount,order_items.paid_amount FROM order_items join products on order_items.product_id=products.id WHERE order_items.order_id = ?`
	config.DB.Raw(qry, orderId).Scan(&orderitems)
	c.JSON(http.StatusOK, gin.H{
		"order_items": orderitems,
	})
}

func ChangeOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	var count int64
	config.DB.Raw(`SELECT COUNT(*) FROM orders where id = ?`, orderID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadGateway, gin.H{
			"message": "order id does not exist",
		})
		return
	}
	var Order models.CancelOrder
	err := c.BindJSON(&Order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "false",
			"message": "binding of data failed",
		})
	}

	if err := helper.Validate(Order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"message":    err.Error(),
			"error_code": http.StatusBadRequest,
		})
		return
	}

	p := Order.OrderStatus
	if p != "delivered" && p != "cancelled" && p != "pending" && p != "shipped" && p != "failed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid order status",
		})
		return
	}
	var orderstatus string
	config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("order_status", &orderstatus)
	if orderstatus == "shipped" {
		if Order.OrderStatus == "pending" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "cannot change order status since item already shipped",
			})
			return
		}
		if Order.OrderStatus == "cancelled" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "cannot cancel order  since item already shipped",
			})
			return
		}
	}
	if orderstatus == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order is alredy cancelled",
		})
		return
	}
	if orderstatus == "failed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order has already failed",
		})
		return
	}
	if Order.OrderStatus == "delivered" {
		now := time.Now()
		today := now.Format("2024-02-03")
		var paymentmethod string
		config.DB.Model(&models.Order{}).Where("id = ?", orderID).Pluck("payment_method", &paymentmethod)
		if paymentmethod == "COD" {
			payment := models.Payments{
				PaymentDate:   today,
				PaymentStatus: "paid",
			}
			config.DB.Model(&models.Payments{}).Where("order_id = ?", orderID).Updates(&payment)
		}
		var OrderItems []models.OrderItems
		config.DB.Where("order_id = ?", orderID).Find(&OrderItems)
		for _, v := range OrderItems {
			if v.OrderStatus != "cancelled" && v.OrderStatus != "return" {
				var stock uint
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Pluck("stock", &stock)
				stock = stock - 1
				config.DB.Model(&models.Product{}).Where("id = ?", v.ProductID).Update("stock", stock)
				paidamount := v.Price - v.TotalDiscount
				config.DB.Model(&models.OrderItems{}).Where("id = ?", v.ID).Update("paid_amount", paidamount)
				config.DB.Model(&models.OrderItems{}).Where("id = ?", v.ID).Update("delivered_date", today)

			}
		}
	}
	order := models.Order{
		OrderStatus: Order.OrderStatus,
	}
	config.DB.Model(&models.Order{}).Where("ID = ?", orderID).Updates(&order)
	config.DB.Model(&models.OrderItems{}).Where("order_id = ? and order_status != 'cancelled'", orderID).Update("order_status", Order.OrderStatus)
	c.JSON(http.StatusOK, gin.H{
		"message": "order status changed successfully",
	})
}
