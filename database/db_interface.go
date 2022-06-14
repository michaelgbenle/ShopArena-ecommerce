package database

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/decadevs/shoparena/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

// DB provides access to the different db

type DB interface {
	CreateBuyer(user *models.Buyer) (*models.Buyer, error)
	CreateSeller(user *models.Seller) (*models.Seller, error)
	FindAllSellersExcept(except string) ([]models.Seller, error)
	FindBuyerByEmail(email string) (*models.Buyer, error)
	FindBuyerByPhone(phone string) (*models.Buyer, error)
	FindBuyerByUsername(username string) (*models.Buyer, error)
	FindSellerByEmail(email string) (*models.Seller, error)
	FindSellerByPhone(phone string) (*models.Seller, error)
	UpdateBuyerImageURL(username, url string, buyerID uint) error
	UpdateSellerImageURL(username, url string, sellerID uint) error
	FindSellerByUsername(username string) (*models.Seller, error)
	FindSellerById(Id uint) (*models.Seller, error)
	FindProductById(Id uint) (*models.Product, error)
	SearchProduct(lowerPrice, upperPrice, category, name string) ([]models.Product, error)
	TokenInBlacklist(token *string) bool
	UpdateBuyerProfile(id uint, update *models.UpdateUser) error
	UpdateSellerProfile(id uint, update *models.UpdateUser) error
	UpdateSellerRating(id uint, update *models.UpdateRating) error
	UpdateProductRating(id uint, update *models.UpdateRating) error
	UploadFileToS3(h *session.Session, file multipart.File, fileName string, size int64) (string, error)
	CreateProduct(product models.Product) error
	GetCategory(category string) (*models.Category, error)
	DeleteProduct(productID, sellerID uint) error
	BuyerUpdatePassword(password, newPassword string) (*models.Buyer, error)
	SellerUpdatePassword(password, newPassword string) (*models.Seller, error)
	BuyerResetPassword(email, newPassword string) (*models.Buyer, error)
	SellerResetPassword(email, newPassword string) (*models.Seller, error)
	CreateBuyerCart(cart *models.Cart) (*models.Cart, error)
	FindIndividualSellerShop(sellerID uint) (*models.Seller, error)
	GetAllProducts() []models.Product
	UpdateProductByID(Id uint, prod models.Product) error
	GetAllSellers() ([]models.Seller, error)
	GetProductByID(id uint) (*models.Product, error)
	FindSellerProduct(sellerID uint) ([]models.Product, error)
	GetAllBuyerOrder(buyerId uint) ([]models.Order, error)
	GetAllSellerOrder(sellerId uint) ([]models.Order, error)
	GetAllSellerOrderCount(sellerId uint) (int, error)
	FindPaidProduct(sellerID uint) ([]models.CartProduct, error)
	AddToCart(product models.Product, buyer *models.Buyer) error
	GetCartProducts(buyer *models.Buyer) ([]models.CartProduct, error)
	ViewCartProducts(addedProducts []models.CartProduct) ([]models.ProductDetails, error)
	DeletePaidFromCart(cartID uint) error
	GetSellersProducts(sellerID uint) ([]models.Product, error)
	FindSellerIndividualProduct(sellerID uint) (*models.Product, error)
	FindCartProductSeller(sellerID, productID uint) (*models.CartProduct, error)
	DeleteCartProduct(buyerID, cartProductID uint) error
	DeleteAllFromCart(buyerID uint) error
	AddTokenToBlacklist(email string, token string) error
	DeleteAllSellerProducts(sellerID uint) error
	GetAllSellerOrders(sellerId uint) ([]models.OrderProducts, error)
	GetAllBuyerOrders(buyerId uint) ([]models.OrderProducts, error)
	ValidateBuyer(email string) error
}

// Mailer interface to implement mailing service
type Mailer interface {
	SendMail(subject, body, to, Private, Domain string) error
	GenerateNonAuthToken(UserEmail string, secret string) (*string, error)
	DecodeToken(token, secret string) (string, error)
}

//Paystack interface
type Paystack interface {
	InitializePayment(info []byte) (string, error)
	Callback(reference string) (*http.Response, error)
	PayStackDecodeToken(token, secret string) (jwt.MapClaims, error)
}

// ValidationError defines error that occur due to validation
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

type DBParams struct {
	Host     string
	User     string
	Password string
	DbName   string
	Port     string
}

func InitDBParams() DBParams {
	ginMode := os.Getenv("GIN_MODE")
	log.Println(ginMode)
	if ginMode != "release" {
		errEnv := godotenv.Load()
		if errEnv != nil {
			log.Fatal("Error loading .env file")
		}
	}

	host := os.Getenv("PDB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("PDB_PORT")

	return DBParams{
		Host:     host,
		User:     user,
		Password: password,
		DbName:   dbName,
		Port:     port,
	}
}
