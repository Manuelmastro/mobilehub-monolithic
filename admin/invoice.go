package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/models"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf/v2"
)

func GenerateInvoice(c *gin.Context) {
	OrderID := c.Param("order_id")
	var count int64

	config.DB.Raw(`SELECT COUNT(*) FROM orders WHERE id = ? AND order_status!='cancelled'`, OrderID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order is cancelled or order not exists",
		})
		return
	}
	config.DB.Exec(`TRUNCATE invoices`)
	var userid uint
	config.DB.Model(&models.Order{}).Where("id = ?", OrderID).Pluck("user_id", &userid)

	var firstname string
	config.DB.Model(&models.User{}).Where("id = ?", userid).Pluck("first_name", &firstname)

	var lastname string
	config.DB.Model(&models.User{}).Where("id = ?", userid).Pluck("last_name", &lastname)

	fullname := firstname + " " + lastname

	var orderdate time.Time
	config.DB.Raw(`SELECT created_at from orders where id = ?`, OrderID).Scan(&orderdate)

	var OrderStatus string
	config.DB.Raw(`SELECT order_status from orders where id = ?`, OrderID).Scan(&OrderStatus)

	var paymentmethod string
	config.DB.Raw(`SELECT payment_method from orders where id = ?`, OrderID).Scan(&paymentmethod)

	var paymentstatus string

	if paymentmethod == "COD" {
		if OrderStatus == "delivered" {
			paymentstatus = "paid"
		} else {
			paymentstatus = "pending"
		}
	} else {
		paymentstatus = "paid"
	}
	var order models.Order
	config.DB.Preload("Address").First(&order, OrderID)

	Country := order.Address.Country
	State := order.Address.State
	Districr := order.Address.District
	Streetname := order.Address.StreetName
	PinCode := order.Address.PinCode
	Phone := order.Address.Phone

	var orderitems []models.OrderItems
	config.DB.Raw(`SELECT * FROM order_items WHERE order_status!='return' AND order_status!='cancelled' AND order_id = ? order by product_id`, OrderID).Scan(&orderitems)

	var productid string
	for i, v := range orderitems {
		if productid == v.ProductID {
			var qty uint
			config.DB.Model(&models.Invoice{}).Where("product_id = ?", v.ProductID).Pluck("quantity", &qty)
			qty = qty + 1
			var mrp float64
			config.DB.Model(&models.Invoice{}).Where("product_id = ?", v.ProductID).Pluck("mrp", &mrp)
			mrp = mrp + v.Price
			var cupondiscount float64
			config.DB.Model(&models.Invoice{}).Where("product_id = ?", v.ProductID).Pluck("coupon_discount", &cupondiscount)
			cupondiscount = cupondiscount + v.CouponDiscount
			cupondiscount = math.Round(cupondiscount*100) / 100
			var offerdiscount float64
			config.DB.Model(&models.Invoice{}).Where("product_id = ?", v.ProductID).Pluck("offer_discount", &offerdiscount)
			offerdiscount = offerdiscount + v.OfferDiscount
			var totaldiscount float64
			config.DB.Model(&models.Invoice{}).Where("product_id = ?", v.ProductID).Pluck("total_discount", &totaldiscount)
			totaldiscount = totaldiscount + v.TotalDiscount
			var finalprice float64
			config.DB.Model(&models.Invoice{}).Where("product_id = ?", v.ProductID).Pluck("final_price", &finalprice)
			finalprice = finalprice + finalprice
			invoice := models.Invoice{
				Quantity:       qty,
				MRP:            mrp,
				CouponDiscount: cupondiscount,
				OfferDiscount:  offerdiscount,
				TotalDiscount:  totaldiscount,
				FinalPrice:     finalprice,
			}
			config.DB.Model(&models.Invoice{}).Where("product_id = ?", productid).Updates(&invoice)
			continue
		}
		productid = v.ProductID
		CouponDiscount1 := math.Round(v.CouponDiscount*100) / 100
		var orderitem1 models.OrderItems
		config.DB.Preload("Product").First(&orderitem1, v.ID)
		invoice := models.Invoice{
			No:             i + 1,
			ProductID:      v.ProductID,
			ProductName:    orderitem1.Product.ProductName,
			Quantity:       1,
			MRP:            v.Price,
			CouponDiscount: CouponDiscount1,
			OfferDiscount:  v.OfferDiscount,
			TotalDiscount:  v.TotalDiscount,
			FinalPrice:     v.Price - v.TotalDiscount,
		}
		config.DB.Create(&invoice)
	}
	var invoiceitem []models.Invoice
	config.DB.Find(&invoiceitem)
	var grandtotal float64
	config.DB.Raw(`SELECT SUM(final_price) from invoices`).Scan(&grandtotal)

	pdf := gofpdf.New("P", "mm", "A2", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Tax Invoice")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Tax Invoice")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "MobileHub")
	pdf.Ln(12)

	orderDateString := orderdate.Format("2006-01-02 15:04:05")

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Order ID: %s", OrderID))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Order Date: %s", orderDateString))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Payment Status: %s", paymentstatus))
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Address")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Name: %s", fullname))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Address:%s,%s,%s,%s-%s", Streetname, Districr, State, Country, PinCode))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Phone: %s", Phone))
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	headers := []string{"No", "Product ID", "Product Name", "Quantity", "MRP", "Coupon Discount", "Offer Discount", "Total Discount", "Final Price"}
	widths := []float64{15, 20, 30, 20, 25, 35, 35, 35, 35}
	for i, header := range headers {
		pdf.CellFormat(widths[i], 10, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	fill := false
	for i, item := range invoiceitem {
		if fill {
			pdf.SetFillColor(230, 230, 230)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		fill = !fill
		fmt.Println("gsdfh", i)
		pdf.CellFormat(widths[0], 10, strconv.Itoa(int(item.No)), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[1], 10, item.ProductID, "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[2], 10, item.ProductName, "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[3], 10, strconv.Itoa(int(item.Quantity)), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[4], 10, fmt.Sprintf("%.2f", item.MRP), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[5], 10, fmt.Sprintf("%.2f", item.CouponDiscount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[6], 10, fmt.Sprintf("%.2f", item.OfferDiscount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[7], 10, fmt.Sprintf("%.2f", item.TotalDiscount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[8], 10, fmt.Sprintf("%.2f", item.FinalPrice), "1", 0, "C", true, 0, "")
		pdf.Ln(-1)

	}
	grandtotalString := strconv.FormatFloat(grandtotal, 'f', 2, 64)
	pdf.Cell(0, 10, fmt.Sprintf("Grand Total: %s", grandtotalString))
	pdf.Ln(8)

	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Generated on %s", time.Now().Format("2006-01-02")))
	pdf.OutputFileAndClose("invoice.pdf")
	c.File("invoice.pdf")

}
