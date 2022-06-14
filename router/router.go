package router

import (
	"github.com/decadevs/shoparena/handlers"
	"github.com/decadevs/shoparena/server/middleware"
	"github.com/gin-contrib/cors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Router struct {
	ContentType string
	handlers    map[string]func(w http.ResponseWriter, r *http.Request)
}

func SetupRouter(h *handlers.Handler) (*gin.Engine, string) {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	apirouter := router.Group("/api/v1")

	apirouter.GET("/ping", handlers.PingHandler)
	apirouter.GET("/searchproducts", h.SearchProductHandler)
	apirouter.GET("/products", h.GetAllProducts)
	apirouter.GET("/sellers", h.GetSellers)
	apirouter.GET("/product/:id", h.GetProductById)
	apirouter.POST("/loginbuyer", h.LoginBuyerHandler)
	apirouter.POST("/loginseller", h.LoginSellerHandler)
	apirouter.POST("/buyersignup", h.BuyerSignUpHandler)
	apirouter.PUT("/buyer/success", h.BuyerSignUpActivation)
	apirouter.POST("/sellersignup", h.SellerSignUpHandler)
	apirouter.GET("/callback", h.Callback)

	apirouter.GET("/seller/shop/:id", h.HandleGetSellerShopByProfileAndProduct())

	apirouter.GET("/seller/totalorder/:id", h.SellerTotalOrders)
	apirouter.POST("buyer/forgotpassword", h.BuyerForgotPasswordEMailHandler)
	apirouter.POST("seller/forgotpassword", h.SellerForgotPasswordEMailHandler)
	apirouter.PUT("/sellerresetpassword", h.SellerForgotPasswordResetHandler)
	apirouter.PUT("/buyerresetpassword", h.BuyerForgotPasswordResetHandler)

	//All authorized routes here
	authorizedRoutesBuyer := apirouter.Group("/")
	authorizedRoutesBuyer.Use(middleware.AuthorizeBuyer(h.DB.FindBuyerByEmail, h.DB.TokenInBlacklist))
	{
		authorizedRoutesBuyer.PUT("/updatebuyerprofile", h.UpdateBuyerProfileHandler)
		authorizedRoutesBuyer.GET("/getbuyerprofile", h.GetBuyerProfileHandler)
		authorizedRoutesBuyer.POST("/addtocart", h.AddToCart)
		authorizedRoutesBuyer.GET("/viewcart", h.ViewCartProducts)
		authorizedRoutesBuyer.POST("/pay", h.Pay)
		authorizedRoutesBuyer.PUT("/buyer/updatepassword", h.BuyerUpdatePassword)
		authorizedRoutesBuyer.PUT("/uploadbuyerpic", h.UploadBuyerImageHandler)
		authorizedRoutesBuyer.DELETE("/deletefromcart/:id", h.DeleteFromCart)
		authorizedRoutesBuyer.DELETE("/deleteallcart", h.DeleteAllCartProducts)
		authorizedRoutesBuyer.POST("/buyer/logout", h.HandleLogoutBuyer)
		authorizedRoutesBuyer.GET("/buyerorders", h.AllBuyerOrders)
		authorizedRoutesBuyer.POST("/buyer/rateaseller", h.SellerRating)
		authorizedRoutesBuyer.POST("/buyer/rateaproduct", h.ProductRating)
	}
	authorizedRoutesSeller := apirouter.Group("/")
	authorizedRoutesSeller.Use(middleware.AuthorizeSeller(h.DB.FindSellerByEmail, h.DB.TokenInBlacklist))
	{

		authorizedRoutesSeller.PUT("/updatesellerprofile", h.UpdateSellerProfileHandler)
		authorizedRoutesSeller.GET("/sellerorders", h.AllSellerOrders)
		authorizedRoutesSeller.GET("/seller/totalorder/", h.SellerTotalOrders)
		authorizedRoutesSeller.GET("/getsellerprofile", h.GetSellerProfileHandler)
		authorizedRoutesSeller.GET("/seller/total/product/sold", h.GetTotalSoldProductCount)
		authorizedRoutesSeller.DELETE("/deleteproduct/:id", h.DeleteSellerProduct)
		authorizedRoutesSeller.POST("/createproduct", h.CreateProducts)
		authorizedRoutesSeller.PUT("/seller/updatepassword", h.SellerUpdatePassword)

		authorizedRoutesSeller.GET("/seller/total/product/count", h.GetTotalProductCountForSeller)
		authorizedRoutesSeller.GET("/seller/product", h.SellerIndividualProduct)
		authorizedRoutesSeller.PUT("/update/product/:id", h.UpdateProduct)
		authorizedRoutesSeller.GET("/seller/allproducts", h.SellerAllProducts)
		authorizedRoutesSeller.GET("/seller/remaining/product/count", h.GetRemainingProductsCountSellerCount)
		authorizedRoutesBuyer.PUT("/uploadsellerpic", h.UploadSellerImageHandler)
		authorizedRoutesSeller.POST("/seller/logout", h.HandleLogoutSeller)
		authorizedRoutesSeller.DELETE("/deleteallsellerproducts/:seller_id", h.DeleteAllSellerProducts)
	}

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8081"
	}
	return router, port
}
