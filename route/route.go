package route

import (
	"MOBILEHUB/admin"
	"MOBILEHUB/midleware"
	"MOBILEHUB/user"

	"github.com/gin-gonic/gin"
)

func UrlRouteSet(router *gin.Engine) {
	adminGroup := router.Group("/admin")
	//admin-users
	adminGroup.GET("/listusers", midleware.AuthMiddleware("admin"), admin.Listuser)
	adminGroup.PUT("/listusers/blockuser", midleware.AuthMiddleware("admin"), admin.BloskUser)
	adminGroup.PUT("/listusers/unblockuser", midleware.AuthMiddleware("admin"), admin.Unblockuser)
	//admin-login
	adminGroup.POST("/", admin.Login)

	//admin-category
	adminGroup.GET("/category", midleware.AuthMiddleware("admin"), admin.Category)
	adminGroup.POST("/category", midleware.AuthMiddleware("admin"), admin.AddCategory)
	adminGroup.PUT("/category/:id", midleware.AuthMiddleware("admin"), admin.EditCategory)
	adminGroup.DELETE("/category/:id", midleware.AuthMiddleware("admin"), admin.DeleteCategory)

	//admin-product
	adminGroup.GET("/product", midleware.AuthMiddleware("admin"), admin.Product)
	adminGroup.POST("/product", midleware.AuthMiddleware("admin"), admin.AddProduct)
	adminGroup.PUT("/product/:id", midleware.AuthMiddleware("admin"), admin.EditProduct)
	adminGroup.DELETE("/product/:id", midleware.AuthMiddleware("admin"), admin.DeleteProduct)

	//admin-order
	adminGroup.GET("/orderlist", midleware.AuthMiddleware("admin"), admin.OrderList)
	adminGroup.GET("/orderlist/items:order_id", midleware.AuthMiddleware("admin"), admin.OrderItemsList)
	adminGroup.PUT("/order/changestatus/:id", midleware.AuthMiddleware("admin"), admin.ChangeOrderStatus)

	//admin-coupon
	adminGroup.GET("/coupon", midleware.AuthMiddleware("admin"), admin.CouponList)
	adminGroup.POST("/coupon", midleware.AuthMiddleware("admin"), admin.CouponAdd)
	adminGroup.DELETE("/coupon/:id", midleware.AuthMiddleware("admin"), admin.CouponRemove)

	//admin-productoffer
	adminGroup.GET("/offer", midleware.AuthMiddleware("admin"), admin.OfferList)
	adminGroup.POST("/offer", midleware.AuthMiddleware("admin"), admin.OfferAdd)
	adminGroup.DELETE("/offer/:id", midleware.AuthMiddleware("admin"), admin.OfferRemove)
	//admin salesreport------------!!!!!!!!
	adminGroup.POST("/salesreport", midleware.AuthMiddleware("admin"), admin.GenerateSalesReport)
	adminGroup.GET("/salesreport", midleware.AuthMiddleware("admin"), admin.FilterSalesReport)
	adminGroup.GET("/salesreportdownload", midleware.AuthMiddleware("admin"), admin.FilterSalesReportPdfExcel)

	//admin - bestselling and invoice
	adminGroup.GET("/bestselling", midleware.AuthMiddleware("admin"), admin.Bestselling)
	adminGroup.GET("invoice/:order_id", midleware.AuthMiddleware("admin"), admin.GenerateInvoice)

	//user
	router.POST("/signup/", user.SignupUser)
	router.POST("/login/", user.UserLogin)
	router.POST("/signup/verifyotp/:email", user.VerifyotpWindow)
	router.POST("/signup/resendotp/:email", user.ResendOtp)
	router.GET("/", midleware.AuthMiddleware("user"), user.ListProducts)

	router.GET("/auth/google/login", user.HandleGoogleLogin)
	router.GET("/auth/google/callback", user.HandleGoogleCallback)

	router.GET("/searchproduct", midleware.AuthMiddleware("user"), user.SearchProduct)

	//user-profile

	router.GET("/profile", midleware.AuthMiddleware("user"), user.Profile)
	router.PUT("/profile", midleware.AuthMiddleware("user"), user.EditProfile)

	//user profile order
	router.GET("/profile/userorders", midleware.AuthMiddleware("user"), user.OrderList)
	router.GET("/profile/userorders/items/:order_id", midleware.AuthMiddleware("user"), user.OrderItemsListt)
	router.PUT("/profile/userorders/cancelorder/:order_id", midleware.AuthMiddleware("user"), user.CancelOrder)
	router.PUT("/profile/userorders/cancelsingleorderitem/:orderitem_id", midleware.AuthMiddleware("user"), user.CancelSingleOrderItem)
	router.PUT("/profile/userorders/returnsingleorderitem/:orderitem_id", midleware.AuthMiddleware("user"), user.ReturnSingleOrderItem)

	//user profile changepassword
	router.PUT("/profile/changepassword", midleware.AuthMiddleware("user"), user.ChangePassword)

	//user profile address
	router.GET("/profile/useraddress", midleware.AuthMiddleware("user"), user.ListAddress)
	router.POST("/profile/useraddress", midleware.AuthMiddleware("user"), user.AddAddress)
	router.PUT("/profile/useraddress/:address_id", midleware.AuthMiddleware("user"), user.EditAddress)
	router.DELETE("/profile/useraddress/:address_id", midleware.AuthMiddleware("user"), user.DeleteAddress)

	//user cart

	router.GET("/cart", midleware.AuthMiddleware("user"), user.Cart)
	router.POST("/cart", midleware.AuthMiddleware("user"), user.AddCart)
	router.DELETE("/cart", midleware.AuthMiddleware("user"), user.RemoveCart)

	//user checkout

	router.GET("/checkout", midleware.AuthMiddleware("user"), user.Checkout)
	router.PUT("/checkout/address/:address_id", midleware.AuthMiddleware("user"), user.EditCheckOutAddress)
	router.POST("/checkout/order", midleware.AuthMiddleware("user"), user.Order)
	router.POST("/checkout/razorpay", midleware.AuthMiddleware("user"), user.CreateOrder)
	router.POST("/checkout/razorpay/paymentverification", midleware.AuthMiddleware("user"), user.PaymentWebhook)

	router.POST("/checkout/wallet", midleware.AuthMiddleware("user"), user.WalletOrder)

	//user wallet
	router.GET("/proflie/wallet", midleware.AuthMiddleware("user"), user.WalletListing)
	router.GET("/proflie/wallet/transactionlist", midleware.AuthMiddleware("user"), user.WalletTransactionListing)
	//router.POST("/checkout/wallet", midleware.AuthMiddleware("user"), user.WalletOrder)

	//user wishlist
	router.GET("/profile/wishlist", midleware.AuthMiddleware("user"), user.Wishlist)
	router.POST("/profile/wishlist", midleware.AuthMiddleware("user"), user.AddWishlist)
	router.DELETE("/profile/wishlist", midleware.AuthMiddleware("user"), user.RemoveWishlist)

	//invoice
	router.GET("/profile/invoice/:order_id", midleware.AuthMiddleware("user"), user.GenerateInvoice)

}
