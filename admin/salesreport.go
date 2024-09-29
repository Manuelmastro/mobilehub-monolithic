package admin

import (
	"MOBILEHUB/config"
	"MOBILEHUB/models"
	"MOBILEHUB/responsemodels"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/xuri/excelize/v2"
)

func GenerateSalesReport(c *gin.Context) {
	config.DB.Exec(`TRUNCATE sales_report_items`)
	var OrderItems []models.OrderItems
	config.DB.Raw(`SELECT * FROM order_items WHERE order_status='delivered' order by product_id,order_id`).Scan(&OrderItems)
	var orderid uint
	var productid string
	for _, v := range OrderItems {
		if orderid == v.OrderID && productid == v.ProductID {
			var qty uint
			config.DB.Model(&models.SalesReportItem{}).Where("order_id = ? AND product_id = ?", v.OrderID, v.ProductID).Pluck("qty", &qty)
			qty = qty + 1
			var cupondiscount float64
			config.DB.Model(&models.SalesReportItem{}).Where("order_id = ? AND product_id = ?", v.OrderID, v.ProductID).Pluck("coupon_discount", &cupondiscount)
			cupondiscount = cupondiscount + v.CouponDiscount
			cupondiscount = math.Round(cupondiscount*100) / 100
			var offerdiscount float64
			config.DB.Model(&models.SalesReportItem{}).Where("order_id = ? AND product_id = ?", v.OrderID, v.ProductID).Pluck("offer_discount", &offerdiscount)
			offerdiscount = offerdiscount + v.OfferDiscount
			var totaldiscount float64
			config.DB.Model(&models.SalesReportItem{}).Where("order_id = ? AND product_id = ?", v.OrderID, v.ProductID).Pluck("total_discount", &totaldiscount)
			totaldiscount = totaldiscount + v.TotalDiscount
			var paidamount float64
			config.DB.Model(&models.SalesReportItem{}).Where("order_id = ? AND product_id = ?", v.OrderID, v.ProductID).Pluck("paid_amount", &paidamount)
			paidamount = paidamount + v.PaidAmount
			salesreport := models.SalesReportItem{
				Qty:            qty,
				CouponDiscount: cupondiscount,
				OfferDiscount:  offerdiscount,
				TotalDiscount:  totaldiscount,
				PaidAmount:     paidamount,
			}
			config.DB.Model(&models.SalesReportItem{}).Where("order_id = ? AND product_id = ?", orderid, productid).Updates(&salesreport)
			continue
		}
		orderid = v.OrderID
		productid = v.ProductID
		var productname string
		config.DB.Raw(`SELECT product_name from products where id = ?`, v.ProductID).Scan(&productname)
		CuponDiscount1 := math.Round(v.CouponDiscount*100) / 100
		salesreport := models.SalesReportItem{
			OrderID:        orderid,
			ProductID:      productid,
			ProductName:    productname,
			Qty:            1,
			Price:          v.Price,
			OrderStatus:    v.OrderStatus,
			PaymentMethod:  v.PaymentMethod,
			CouponDiscount: CuponDiscount1,
			OfferDiscount:  v.OfferDiscount,
			TotalDiscount:  v.TotalDiscount,
			PaidAmount:     v.PaidAmount,
			OrderDate:      v.CreatedAt,
			DeliveredDate:  v.DeliveredDate,
		}
		config.DB.Create(&salesreport)
	}
	var salesreportitem []responsemodels.SalesReportItem
	config.DB.Find(&salesreportitem)
	c.JSON(http.StatusOK, gin.H{
		"sales report": salesreportitem,
		"message":      "sales report generated successfully",
	})
}

func FilterSalesReport(c *gin.Context) {
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
			fmt.Println("Error converting string to int:", err)
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
		if yearint < 1900 || yearint > currentYear {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "year out of range, must be between 1900 and current year",
			})
			return
		}
	}
	var salesreportitem []responsemodels.SalesReportItem
	type SalesReportSummary struct {
		TotalQuantity   uint    `json:"total_quantity"`
		TotalPaidAmount float64 `json:"total_paid_amount"`
		TotalDiscount   float64 `json:"total_discount"`
	}
	var summary SalesReportSummary
	qry := `SELECT * FROM sales_report_items `
	qry1 := `SELECT SUM(qty)AS total_quantity, SUM(paid_amount) AS total_paid_amount, SUM(total_discount) AS total_discount FROM sales_report_items `
	if day != "" {
		qry = qry + `WHERE DATE(order_date) ='` + day + `'`
		config.DB.Raw(qry).Scan(&salesreportitem)
		qry1 = qry1 + `WHERE DATE(order_date) ='` + day + `'`
		config.DB.Raw(qry1).Scan(&summary)
	} else if month != "" {
		qry = qry + `WHERE EXTRACT(MONTH FROM order_date) = '` + month + `'`
		config.DB.Raw(qry).Scan(&salesreportitem)
		qry1 = qry1 + `WHERE EXTRACT(MONTH FROM order_date) = '` + month + `'`
		config.DB.Raw(qry1).Scan(&summary)
	} else if year != "" {
		qry = qry + `WHERE EXTRACT(YEAR FROM order_date) = '` + year + `'`
		config.DB.Raw(qry).Scan(&salesreportitem)
		qry1 = qry1 + `WHERE EXTRACT(YEAR FROM order_date) = '` + year + `'`
		config.DB.Raw(qry1).Scan(&summary)
	} else {
		config.DB.Raw(qry).Scan(&salesreportitem)
		config.DB.Raw(qry1).Scan(&summary)
	}
	c.JSON(http.StatusOK, gin.H{
		"sales_report":         salesreportitem,
		"message":              "sales report generated successfully",
		"overall_sales_count":  summary.TotalQuantity,
		"overall_order_amount": summary.TotalPaidAmount,
		"overall_discount":     summary.TotalDiscount,
	})

}

type SalesReportSummary struct {
	TotalQuantity   uint    `json:"total_quantity"`
	TotalPaidAmount float64 `json:"total_paid_amount"`
	TotalDiscount   float64 `json:"total_discount"`
}

func FilterSalesReportPdfExcel(c *gin.Context) {
	day := c.Query("day")
	month := c.Query("month")
	year := c.Query("year")

	var salesreportitem []models.SalesReportItem
	var summary SalesReportSummary
	qry := `SELECT * FROM sales_report_items `
	qry1 := `SELECT SUM(qty)AS total_quantity, SUM(paid_amount) AS total_paid_amount, SUM(total_discount) AS total_discount FROM sales_report_items `
	if day != "" {
		qry = qry + `WHERE DATE(order_date) ='` + day + `'`
		qry1 = qry1 + `WHERE DATE(order_date) ='` + day + `'`
	} else if month != "" {
		qry = qry + `WHERE EXTRACT(MONTH FROM order_date) = '` + month + `'`
		qry1 = qry1 + `WHERE EXTRACT(MONTH FROM order_date) = '` + month + `'`

	} else if year != "" {
		qry = qry + `WHERE EXTRACT(YEAR FROM order_date) = '` + year + `'`
		qry1 = qry1 + `WHERE EXTRACT(YEAR FROM order_date) = '` + year + `'`

	}
	config.DB.Raw(qry).Scan(&salesreportitem)
	config.DB.Raw(qry1).Scan(&summary)

	format := c.Query("format")
	fileName := "sales_report." + format
	filePath := fileName
	if format == "xlsx" {
		err := GenerateExcelReport(salesreportitem, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate xcel report"})
			return
		}
	} else if format == "pdf" {
		err := GeneratePDFReport(salesreportitem, summary, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF report"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format"})
		return
	}

	c.File(filePath)
	c.JSON(http.StatusOK, gin.H{
		"sales_report":         salesreportitem,
		"message":              "sales report generated successfully",
		"overall_sales_count":  summary.TotalQuantity,
		"overall_order_amount": summary.TotalPaidAmount,
		"overall_discount":     summary.TotalDiscount,
	})

}

func GenerateExcelReport(salesReportItems []models.SalesReportItem, filePath string) error {
	f := excelize.NewFile()
	index, _ := f.NewSheet("SalesReport")
	headers := []string{"Order ID", "Product ID", "Product Name", "Quantity", "Price", "Order Status", "Payment Method", "Coupon Discount", "Offer Discount", "Total Discount", "Paid Amount", "Order Date"}
	for i, header := range headers {
		col := string('A' + i)
		f.SetCellValue("SalesReport", col+"1", header)
	}

	for i, item := range salesReportItems {
		row := strconv.Itoa(i + 2)
		f.SetCellValue("SalesReport", "A"+row, item.OrderID)
		f.SetCellValue("SalesReport", "B"+row, item.ProductID)
		f.SetCellValue("SalesReport", "C"+row, item.ProductName)
		f.SetCellValue("SalesReport", "D"+row, item.Qty)
		f.SetCellValue("SalesReport", "E"+row, item.Price)
		f.SetCellValue("SalesReport", "F"+row, item.OrderStatus)
		f.SetCellValue("SalesReport", "G"+row, item.PaymentMethod)
		f.SetCellValue("SalesReport", "H"+row, item.CouponDiscount)
		f.SetCellValue("SalesReport", "I"+row, item.OfferDiscount)
		f.SetCellValue("SalesReport", "J"+row, item.TotalDiscount)
		f.SetCellValue("SalesReport", "K"+row, item.PaidAmount)
		f.SetCellValue("SalesReport", "L"+row, item.OrderDate.Format("2006-01-02"))
		f.SetCellValue("SalesReport", "M"+row, item.DeliveredDate)

	}
	f.SetActiveSheet(index)
	if err := f.SaveAs(filePath); err != nil {
		fmt.Println("Error saving file:", err)
		return err
	}
	return nil
}

func GeneratePDFReport(salesReportItems []models.SalesReportItem, summary SalesReportSummary, filePath string) error {
	pdf := gofpdf.New("P", "mm", "A2", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Sales Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Overall Sales Count: %d", summary.TotalQuantity))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Overall Order Amount: Rs.%.2f", summary.TotalPaidAmount))
	pdf.Ln(8)
	pdf.Cell(0, 10, fmt.Sprintf("Overall Discount: Rs.%.2f", summary.TotalDiscount))
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	headers := []string{"Order ID", "Product ID", "Product Name", "Quantity", "Price", "Order Status", "Payment Method", "Coupon Discount", "Offer Discount", "Total Discount", "Paid Amount", "Order Date", "Delivered Date"}
	widths := []float64{15, 20, 30, 20, 25, 35, 35, 35, 35, 35, 30, 25, 35}
	for i, header := range headers {
		pdf.CellFormat(widths[i], 10, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	fill := false
	for _, item := range salesReportItems {
		if fill {
			pdf.SetFillColor(230, 230, 230)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		fill = !fill
		pdf.CellFormat(widths[0], 10, strconv.Itoa(int(item.OrderID)), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[1], 10, item.ProductID, "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[2], 10, item.ProductName, "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[3], 10, strconv.Itoa(int(item.Qty)), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[4], 10, fmt.Sprintf("%.2f", item.Price), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[5], 10, item.OrderStatus, "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[6], 10, item.PaymentMethod, "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[7], 10, fmt.Sprintf("%.2f", item.CouponDiscount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[8], 10, fmt.Sprintf("%.2f", item.OfferDiscount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[9], 10, fmt.Sprintf("%.2f", item.TotalDiscount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[10], 10, fmt.Sprintf("%.2f", item.PaidAmount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[11], 10, item.OrderDate.Format("2006-01-02"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[12], 10, item.DeliveredDate, "1", 0, "C", true, 0, "")
		pdf.Ln(-1)

	}
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Generated on %s", time.Now().Format("2006-01-02")))
	return pdf.OutputFileAndClose(filePath)
}
