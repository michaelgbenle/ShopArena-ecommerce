package database

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/decadevs/shoparena/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"strconv"
	"time"
)

//PostgresDb implements the DB interface
type PostgresDb struct {
	DB *gorm.DB
}

// Init sets up the mongodb instance
func (pdb *PostgresDb) Init(host, user, password, dbName, port string) error {
	fmt.Println("connecting to Database.....")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Lagos", host, user, password, dbName, port)
	var err error
	if os.Getenv("DATABASE_URL") != "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	if db == nil {
		return fmt.Errorf("database was not initialized")
	} else {
		fmt.Println("Connected to Database")
	}

	pdb.DB = db
	err = pdb.PrePopulateTables()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil

}

func (pdb *PostgresDb) PrePopulateTables() error {
	err := pdb.DB.AutoMigrate(&models.Category{}, &models.Seller{}, &models.Product{}, &models.Image{},
		&models.Buyer{}, &models.Cart{}, &models.CartProduct{}, &models.Order{}, &models.Blacklist{})
	if err != nil {
		return fmt.Errorf("migration error: %v", err)
	}
	categories := []models.Category{{Name: "fashion"}, {Name: "electronics"}, {Name: "health & beauty"}, {Name: "baby products"}, {Name: "phones & tablets"}, {Name: "food drinks"}, {Name: "computing"}, {Name: "sporting goods"}, {Name: "others"}}
	result := pdb.DB.Find(&models.Category{})
	if result.RowsAffected < 1 {
		pdb.DB.Create(&categories)
	}

	user := models.User{
		Model:           gorm.Model{},
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "jdoe@gmail.com",
		Username:        "JD Baba",
		Password:        "12345678",
		ConfirmPassword: "12345678",
		PasswordHash:    "$2a$12$T2wSf1qgpTyhLOons3u4JOCqCwKDDL4J3UhGdOTEBL/CmAS/RNCPm",
		Address:         "aso rock",
		PhoneNumber:     "09091919292",
		Image:           "https://i.ibb.co/5jwDfyF/Photo-on-24-11-2021-at-20-45.jpg",
		IsActive:        true,
		Token:           "",
	}
	buyer := models.Buyer{
		Model:  gorm.Model{},
		User:   user,
		Orders: nil,
	}
	result = pdb.DB.Where("buyer = ?", "John").Find(&buyer)

	if result.RowsAffected < 1 {
		pdb.DB.Create(&buyer)
	}

	seller := models.Seller{
		Model:   gorm.Model{},
		User:    user,
		Product: nil,
		Orders:  nil,
		Rating:  5,
	}
	result = pdb.DB.Where("seller = ?", "John").Find(&seller)

	if result.RowsAffected < 1 {
		pdb.DB.Create(&seller)
	}
	product := models.Product{
		Model:       gorm.Model{},
		SellerId:    1,
		CategoryId:  1,
		Category:    models.Category{},
		Title:       "shoes",
		Description: "loafers",
		Price:       30,
		Images:      nil,
		Rating:      4,
		Quantity:    3,
	}
	Product1 := models.Product{
		Model:       gorm.Model{},
		SellerId:    1,
		CategoryId:  2,
		Category:    models.Category{},
		Title:       "toaster",
		Description: "sony press on toaster",
		Price:       420,
		Images:      nil,
		Rating:      4,
		Quantity:    3,
	}
	product2 := models.Product{
		Model:       gorm.Model{},
		SellerId:    1,
		CategoryId:  3,
		Category:    models.Category{},
		Title:       "lip gloss",
		Description: "fenty beauty shimmer gloss",
		Price:       76,
		Images:      nil,
		Rating:      4,
		Quantity:    3,
	}
	Product3 := models.Product{
		Model:       gorm.Model{},
		SellerId:    1,
		CategoryId:  4,
		Category:    models.Category{},
		Title:       "pampers",
		Description: "7 in 1 pamper pack",
		Price:       100,
		Images:      nil,
		Rating:      4,
		Quantity:    3,
	}
	result = pdb.DB.Where("title = ?", "shoes").Find(&product)
	if result.RowsAffected < 1 {
		pdb.DB.Create(&product)
		pdb.DB.Create(&Product1)
		pdb.DB.Create(&product2)
		pdb.DB.Create(&Product3)
	}
	return nil

}

//GET ALL PRODUCTS FROM DB
func (pdb *PostgresDb) GetAllProducts() []models.Product {
	var products []models.Product
	if err := pdb.DB.Preload("Images").Find(&products).Error; err != nil {
		log.Println("Could not find product", err)
	}

	return products
}

//UPDATE PRODUCT BY ID
func (pdb *PostgresDb) UpdateProductByID(Id uint, prod models.Product) error {
	products := models.Product{}

	err := pdb.DB.Model(&products).Where("id = ?", Id).Update("title", prod.Title).
		Update("description", prod.Description).Update("price", prod.Price).
		Update("rating", prod.Rating).Update("quantity", prod.Quantity).Error
	if err != nil {
		fmt.Println("error in updating in postgres db")
		return err
	}
	return nil
}

// SearchProduct Searches all products from DB
func (pdb *PostgresDb) SearchProduct(lowerPrice, upperPrice, categoryName, name string) ([]models.Product, error) {
	var products []models.Product

	var LPInt int
	var UPInt int
	var categoryInt int

	LPInt, _ = strconv.Atoi(lowerPrice)
	UPInt, _ = strconv.Atoi(upperPrice)
	categoryInt, _ = strconv.Atoi(categoryName)

	if categoryInt == 0 {
		if LPInt == 0 && UPInt == 0 && name == "" {
			err := pdb.DB.Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			return products, nil
		} else if LPInt == 0 && UPInt != 0 && name == "" {
			err := pdb.DB.Where("price <= ?", uint(UPInt)).Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt != 0 && UPInt == 0 && name == "" {
			err := pdb.DB.Where("price >= ?", uint(LPInt)).Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt != 0 && UPInt != 0 && name == "" {
			err := pdb.DB.Where("price >= ?", uint(LPInt)).
				Where("price <= ?", uint(UPInt)).Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt == 0 && UPInt == 0 && name != "" {
			err := pdb.DB.Where("title LIKE ?", "%"+name+"%").Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt == 0 && name != "" {
			err := pdb.DB.Where("price <= ?", uint(UPInt)).
				Where("title LIKE ?", "%"+name+"%").Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if UPInt == 0 && name != "" {
			err := pdb.DB.Where("price >= ?", uint(LPInt)).
				Where("title LIKE ?", "%"+name+"%").Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt != 0 && UPInt != 0 && name != "" {
			err := pdb.DB.Where("price >= ?", uint(LPInt)).Where("price <= ?", uint(UPInt)).
				Where("title LIKE ?", "%"+name+"%").Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		}
	} else if categoryInt != 0 {
		if LPInt == 0 && UPInt == 0 && name == "" {
			err := pdb.DB.Where("category_id = ?", uint(categoryInt)).Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt == 0 && name == "" {
			err := pdb.DB.Where("category_id = ?", uint(categoryInt)).
				Where("price <= ?", uint(UPInt)).Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if UPInt == 0 && name == "" {
			err := pdb.DB.Where("category_id = ?", uint(categoryInt)).
				Where("price >= ?", uint(LPInt)).Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt != 0 && UPInt != 0 && name == "" {
			err := pdb.DB.Where("category_id = ?", uint(categoryInt)).Where("price >= ?", uint(LPInt)).
				Where("price <= ?", uint(UPInt)).Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt == 0 && UPInt == 0 && name != "" {
			err := pdb.DB.Where("category_id = ?", uint(categoryInt)).
				Where("title LIKE ?", "%"+name+"%").Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if LPInt == 0 && name != "" {
			err := pdb.DB.Where("category_id = ?", uint(categoryInt)).
				Where("price <= ?", uint(UPInt)).
				Where("title LIKE ?", "%"+name+"%").Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else if UPInt == 0 && name != "" {
			err := pdb.DB.Where("category_id = ?", uint(categoryInt)).
				Where("price >= ?", uint(LPInt)).
				Where("title LIKE ?", "%"+name+"%").Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		} else {
			err := pdb.DB.Where("category_id = ?", uint(categoryInt)).Where("price >= ?", uint(LPInt)).
				Where("price <= ?", uint(UPInt)).
				Where("title LIKE ?", "%"+name+"%").Preload("Images").Find(&products).Error
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		}
	}
	return products, nil
}

// CreateSeller creates a new Seller in the DB
func (pdb *PostgresDb) CreateSeller(user *models.Seller) (*models.Seller, error) {
	var err error
	user.CreatedAt = time.Now()
	user.IsActive = true
	err = pdb.DB.Create(user).Error
	return user, err
}

// CreateBuyer creates a new Buyer in the DB
func (pdb *PostgresDb) CreateBuyer(user *models.Buyer) (*models.Buyer, error) {
	var err error
	user.CreatedAt = time.Now()
	user.IsActive = true
	err = pdb.DB.Create(user).Error
	return user, err
}

// ValidateBuyer finds updates buyer status to true
func (pdb *PostgresDb) ValidateBuyer(email string) error {
	buyer := &models.Buyer{}
	if err := pdb.DB.Model(buyer).Where("email =?", email).Update("is_active", true).Error; err != nil {
		return err
	}
	return nil
}

//CreateBuyerCart creates a new cart for the buyer
func (pdb *PostgresDb) CreateBuyerCart(cart *models.Cart) (*models.Cart, error) {
	var err error
	cart.CreatedAt = time.Now()
	err = pdb.DB.Create(cart).Error
	return cart, err
}

// FindSellerByUsername finds a user by the username
func (pdb *PostgresDb) FindSellerByUsername(username string) (*models.Seller, error) {
	user := &models.Seller{}

	if err := pdb.DB.Where("username = ?", username).First(user).Error; err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, errors.New("user inactive")
	}
	return user, nil
}

// FindBuyerByUsername finds a user by the username
func (pdb *PostgresDb) FindBuyerByUsername(username string) (*models.Buyer, error) {
	buyer := &models.Buyer{}

	if err := pdb.DB.Where("username = ?", username).First(buyer).Error; err != nil {
		return nil, err
	}
	if !buyer.IsActive {
		return nil, errors.New("user inactive")
	}
	return buyer, nil
}

// FindSellerByEmail finds a user by email
func (pdb *PostgresDb) FindSellerByEmail(email string) (*models.Seller, error) {
	seller := &models.Seller{}
	if err := pdb.DB.Where("email = ?", email).First(seller).Error; err != nil {
		return nil, errors.New(email + " does not exist" + " seller not found")
	}

	return seller, nil
}

// FindSellerByID finds a user by Id
func (pdb *PostgresDb) FindSellerById(Id uint) (*models.Seller, error) {
	seller := &models.Seller{}
	if err := pdb.DB.Where("seller_id = ?", Id).First(seller).Error; err != nil {
		return nil, errors.New("Id does not, exist seller not found")
	}

	return seller, nil
}

// FindProductByID finds a product by Id
func (pdb *PostgresDb) FindProductById(Id uint) (*models.Product, error) {
	product := &models.Product{}
	if err := pdb.DB.Where("product_id = ?", Id).First(product).Error; err != nil {
		return nil, errors.New("Id does not exist, seller not found")
	}

	return product, nil
}

// FindBuyerByEmail finds a user by email
func (pdb *PostgresDb) FindBuyerByEmail(email string) (*models.Buyer, error) {
	buyer := &models.Buyer{}
	if err := pdb.DB.Where("email = ?", email).First(buyer).Error; err != nil {
		return nil, errors.New(email + " does not exist" + " buyer not found")
	}

	return buyer, nil
}

// FindSellerByPhone finds a user by the phone
func (pdb PostgresDb) FindSellerByPhone(phone string) (*models.Seller, error) {
	user := &models.Seller{}
	if err := pdb.DB.Where("phone_number =?", phone).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// FindBuyerByPhone finds a user by the phone
func (pdb PostgresDb) FindBuyerByPhone(phone string) (*models.Buyer, error) {
	buyer := &models.Buyer{}
	if err := pdb.DB.Where("phone_number =?", phone).First(buyer).Error; err != nil {
		return nil, err
	}
	return buyer, nil
}

// TokenInBlacklist checks if token is already in the blacklist collection
func (pdb *PostgresDb) TokenInBlacklist(token *string) bool {
	tok := &models.Blacklist{}
	if err := pdb.DB.Where("token = ?", token).First(&tok).Error; err != nil {
		return false
	}

	return true
}

// FindAllUsersExcept returns all the users expcept the one specified in the except parameter
func (pdb *PostgresDb) FindAllSellersExcept(except string) ([]models.Seller, error) {
	sellers := []models.Seller{}
	if err := pdb.DB.Not("username = ?", except).Find(sellers).Error; err != nil {

		return nil, err
	}
	return sellers, nil
}

func (pdb *PostgresDb) UpdateBuyerProfile(id uint, update *models.UpdateUser) error {
	result :=
		pdb.DB.Model(models.Buyer{}).
			Where("id = ?", id).
			Updates(
				models.User{
					FirstName:   update.FirstName,
					LastName:    update.LastName,
					PhoneNumber: update.PhoneNumber,
					Address:     update.Address,
					Email:       update.Email,
				},
			)
	return result.Error
}

func (pdb *PostgresDb) UpdateSellerProfile(id uint, update *models.UpdateUser) error {
	result :=
		pdb.DB.Model(models.Seller{}).
			Where("id = ?", id).
			Updates(
				models.User{
					FirstName:   update.FirstName,
					LastName:    update.LastName,
					PhoneNumber: update.PhoneNumber,
					Address:     update.Address,
					Email:       update.Email,
				},
			)
	return result.Error
}

// UpdateSellerRating updates a sellers' rating by Id
func (pdb *PostgresDb) UpdateSellerRating(id uint, update *models.UpdateRating) error {
	result :=
		pdb.DB.Model(models.Seller{}).
			Where("id = ?", id).
			Updates(
				models.Seller{
					Rating:                  update.Rating,
					TotalRatings:            update.TotalRatings,
					NumberOfRatingsReceived: update.NumberOfRatingsReceived,
				},
			)
	return result.Error
}

// UpdateProductRating updates a products' rating by Id
func (pdb *PostgresDb) UpdateProductRating(id uint, update *models.UpdateRating) error {
	result :=
		pdb.DB.Model(models.Product{}).
			Where("id = ?", id).
			Updates(
				models.Product{
					Rating:                  update.Rating,
					TotalRatings:            update.TotalRatings,
					NumberOfRatingsReceived: update.NumberOfRatingsReceived,
				},
			)
	return result.Error
}

// UploadFileToS3 send FileToS3 saves a file to aws bucket and returns the url to the file and an error if there's any
func (pdb *PostgresDb) UploadFileToS3(h *session.Session, file multipart.File, fileName string, size int64) (string, error) {
	// get the file size and read the file content into a buffer
	buffer := make([]byte, size)
	_, err2 := file.Read(buffer)
	if err2 != nil {
		return "", err2
	}
	// config settings: this is where you choose the bucket,
	// filename, content-type and storage class of the file you're uploading
	url := "https://s3-eu-west-3.amazonaws.com/arp-rental/" + fileName
	_, err := s3.New(h).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(os.Getenv("S3_BUCKET_NAME")),
		Key:                  aws.String(fileName),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(int64(size)),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})
	return url, err
}

func (pdb *PostgresDb) UpdateBuyerImageURL(username, url string, buyerID uint) error {
	buyer := models.Buyer{}
	buyer.Image = url
	result :=
		pdb.DB.Model(models.Buyer{}).
			Where("username = ?", username).
			Updates(buyer)
	return result.Error
}
func (pdb *PostgresDb) UpdateSellerImageURL(username, url string, sellerID uint) error {
	seller := models.Seller{}
	seller.Image = url
	result :=
		pdb.DB.Model(models.Seller{}).
			Where("username = ?", username).
			Updates(seller)
	return result.Error
}
func (pdb *PostgresDb) BuyerUpdatePassword(password, newPassword string) (*models.Buyer, error) {
	buyer := &models.Buyer{}
	if err := pdb.DB.Model(buyer).Where("password_hash =?", password).Update("password_hash", newPassword).Error; err != nil {
		return nil, err
	}
	return buyer, nil
}
func (pdb *PostgresDb) SellerUpdatePassword(password, newPassword string) (*models.Seller, error) {
	seller := &models.Seller{}
	if err := pdb.DB.Model(seller).Where("password_hash =?", password).Update("password_hash", newPassword).Error; err != nil {
		return nil, err
	}
	return seller, nil
}
func (pdb *PostgresDb) BuyerResetPassword(email, newPassword string) (*models.Buyer, error) {
	buyer := &models.Buyer{}
	if err := pdb.DB.Model(buyer).Where("email =?", email).Update("password_hash", newPassword).Error; err != nil {
		return nil, err
	}
	return buyer, nil
}
func (pdb *PostgresDb) SellerResetPassword(email, newPassword string) (*models.Seller, error) {
	seller := &models.Seller{}
	if err := pdb.DB.Model(seller).Where("email =?", email).Update("password_hash", newPassword).Error; err != nil {
		return nil, err
	}
	return seller, nil
}

//FindIndividualSellerShop return the individual seller and its respective shop gotten by its unique ID
func (pdb *PostgresDb) FindIndividualSellerShop(sellerID uint) (*models.Seller, error) {
	//create instance of a seller and its respective product, and unmarshal data into them
	seller := &models.Seller{}

	if err := pdb.DB.Preload("Product").Where("id = ?", sellerID).Find(&seller).Error; err != nil {
		log.Println("Error in finding", err)
		return nil, err
	}

	return seller, nil
}

//GetAllBuyerOrder fetches all buyer orders
func (pdb *PostgresDb) GetAllBuyerOrder(buyerId uint) ([]models.Order, error) {
	var buyerOrder []models.Order
	if err := pdb.DB.Where("buyer_id =?", buyerId).
		Preload("Seller").
		Preload("Buyer").
		Preload("Product").
		Preload("Product.Category").
		Find(&buyerOrder).
		Error; err != nil {
		log.Println("could not find order", err)
		return nil, err
	}

	return buyerOrder, nil
}

//GetAllBuyerOrders fetches all buyer orders
func (pdb *PostgresDb) GetAllBuyerOrders(buyerId uint) ([]models.OrderProducts, error) {
	var buyerOrder []models.Order

	var result []models.OrderProducts

	if err := pdb.DB.Where("buyer_id =?", buyerId).
		Preload("Seller").
		Preload("Buyer").
		Preload("Product").
		Preload("Product.Category").
		Find(&buyerOrder).
		Error; err != nil {
		log.Println("could not find order", err)
		return nil, err
	}
	for i := 0; i < len(buyerOrder); i++ {
		re := models.OrderProducts{
			Fname:        buyerOrder[i].Seller.FirstName,
			Lname:        buyerOrder[i].Seller.LastName,
			CategoryName: buyerOrder[i].Product.Category.Name,
			Title:        buyerOrder[i].Product.Title,
			Price:        buyerOrder[i].Product.Price,
			Quantity:     buyerOrder[i].Product.Quantity,
		}
		result = append(result, re)
	}
	return result, nil
}

// GetAllSellerOrder fetches all seller orders
func (pdb *PostgresDb) GetAllSellerOrder(sellerId uint) ([]models.Order, error) {
	var sellerOrder []models.Order
	if err := pdb.DB.Where("seller_id= ?", sellerId).Preload("Seller").
		Preload("Buyer").
		Preload("Product").
		Preload("Product.Category").
		Find(&sellerOrder).
		Error; err != nil {
		return nil, err
	}
	return sellerOrder, nil
}

// GetAllSellerOrders fetches all seller orders
func (pdb *PostgresDb) GetAllSellerOrders(sellerId uint) ([]models.OrderProducts, error) {
	var sellerOrder []models.Order

	var result []models.OrderProducts
	if err := pdb.DB.Where("seller_id= ?", sellerId).Preload("Seller").
		Preload("Buyer").
		Preload("Product").
		Preload("Product.Category").
		Find(&sellerOrder).
		Error; err != nil {
		return nil, err
	}
	for i := 0; i < len(sellerOrder); i++ {
		re := models.OrderProducts{
			Fname:        sellerOrder[i].Buyer.FirstName,
			Lname:        sellerOrder[i].Buyer.LastName,
			CategoryName: sellerOrder[i].Product.Category.Name,
			Title:        sellerOrder[i].Product.Title,
			Price:        sellerOrder[i].Product.Price,
			Quantity:     sellerOrder[i].Product.Quantity,
		}
		result = append(result, re)
	}
	return result, nil
}

// GetAllSellerOrderCount fetches all buyer orders
func (pdb *PostgresDb) GetAllSellerOrderCount(sellerId uint) (int, error) {
	var sellerOrder []models.Order
	if err := pdb.DB.Where("seller_id= ?", sellerId).Preload("Seller").
		Preload("Buyer").
		Preload("Product").
		Preload("Product.Category").
		Find(&sellerOrder).
		Error; err != nil {
		return 0, err
	}
	count := len(sellerOrder)

	return count, nil
}

// GetAllSellers returns all the sellers in the updated database
func (pdb *PostgresDb) GetAllSellers() ([]models.Seller, error) {
	var seller []models.Seller
	err := pdb.DB.Model(&models.Seller{}).Find(&seller).Error
	if err != nil {
		return nil, err
	}
	return seller, nil
}

// GetProductByID returns a particular product by it's ID
func (pdb *PostgresDb) GetProductByID(id uint) (*models.Product, error) {
	product := &models.Product{}
	if err := pdb.DB.Where("ID=?", id).Preload("Images").First(product).Error; err != nil {
		return nil, err
	}
	return product, nil

}

//GET INDIVIDUAL SELLER PRODUCT
func (pdb *PostgresDb) FindSellerProduct(sellerID uint) ([]models.Product, error) {

	product := []models.Product{}

	if err := pdb.DB.Preload("Category").Where("seller_id = ?", sellerID).Find(&product).Error; err != nil {
		log.Println("Error finding seller product", err)
		return nil, err
	}
	return product, nil

}

//GET PAID PRODUCTS FROM DATABASE
func (pdb *PostgresDb) FindPaidProduct(sellerID uint) ([]models.CartProduct, error) {

	cartProduct := []models.CartProduct{}

	if err := pdb.DB.Where("order_status = ?", true).
		Where("seller_id = ?", sellerID).
		Find(&cartProduct).Error; err != nil {
		log.Println("Error finding products paid", err)
		return nil, err
	}

	return cartProduct, nil

}

func (pdb *PostgresDb) CreateProduct(product models.Product) error {

	err := pdb.DB.Create(&product).Error
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (pdb *PostgresDb) GetCategory(category string) (*models.Category, error) {
	categories := models.Category{}

	err := pdb.DB.Where("name = ?", category).First(&categories).Error
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &categories, nil
}

func (pdb *PostgresDb) DeleteProduct(productID, sellerID uint) error {
	product := models.Product{}

	err := pdb.DB.Where("id = ?", productID).Where("seller_id = ?", sellerID).Delete(&product).Error
	if err != nil {
		return err
	}

	return nil
}

func (pdb *PostgresDb) AddToCart(product models.Product, buyer *models.Buyer) error {
	var prod *models.Product
	var userBuyer *models.Buyer
	var cart *models.Cart

	err := pdb.DB.Where("id = ?", product.ID).First(&prod).Error
	if err != nil {
		return err
	}

	err = pdb.DB.Where("id = ?", buyer.ID).First(&userBuyer).Error
	if err != nil {
		return err
	}

	err = pdb.DB.Where("buyer_id = ?", buyer.ID).First(&cart).Error
	if err != nil {
		return err
	}

	cartProduct := models.CartProduct{
		CartID:        cart.ID,
		ProductID:     product.ID,
		TotalPrice:    prod.Price * product.Quantity,
		TotalQuantity: product.Quantity,
		OrderStatus:   false,
		BuyerId:       buyer.ID,
		SellerId:      product.SellerId,
	}

	cart.Product = append(cart.Product, cartProduct)

	err = pdb.DB.Where("id = ?", cart.ID).Save(&cart).Error
	if err != nil {
		return err
	}

	return nil

}

func (pdb *PostgresDb) GetCartProducts(buyer *models.Buyer) ([]models.CartProduct, error) {

	var cart *models.Cart
	var addedProducts []models.CartProduct

	err := pdb.DB.Where("buyer_id = ?", buyer.ID).First(&cart).Error
	if err != nil {
		return nil, err
	}

	err = pdb.DB.Where("cart_id = ?", cart.ID).Where("order_status = ?", false).
		Find(&addedProducts).Error
	if err != nil {
		return nil, err
	}

	return addedProducts, nil

}

func (pdb *PostgresDb) ViewCartProducts(addedProducts []models.CartProduct) ([]models.ProductDetails, error) {
	var details []models.ProductDetails

	for i := 0; i < len(addedProducts); i++ {
		var product *models.Product
		err := pdb.DB.Where("id = ?", addedProducts[i].ProductID).Preload("Images").First(&product).Error
		if err != nil {
			return nil, err
		}
		prodDetail := models.ProductDetails{
			Name:          product.Title,
			Price:         addedProducts[i].TotalPrice,
			Quantity:      addedProducts[i].TotalQuantity,
			Images:        product.Images,
			CartProductID: addedProducts[i].ID,
		}

		details = append(details, prodDetail)
	}

	return details, nil
}

func (pdb *PostgresDb) DeletePaidFromCart(cartID uint) error {
	var cartProducts []models.CartProduct
	var cart models.Cart

	log.Println("buyer_1d coming in", cartID)

	err := pdb.DB.Where("buyer_id = ?", cartID).First(&cart).Error
	if err != nil {
		return err
	}

	log.Println("cartid coming in", cart.ID)

	err = pdb.DB.Where("cart_id = ?", cart.ID).Where("order_status = ?", false).
		Find(&cartProducts).Error
	if err != nil {
		return err
	}

	log.Println("hello world", cartProducts)

	var totalOrders []models.Order
	for i := 0; i < len(cartProducts); i++ {

		orders := models.Order{
			SellerId:  cartProducts[i].SellerId,
			BuyerId:   cartProducts[i].BuyerId,
			ProductId: cartProducts[i].ProductID,
		}
		totalOrders = append(totalOrders, orders)

	}

	err = pdb.DB.Create(&totalOrders).Error
	if err != nil {
		return err
	}

	err = pdb.DB.Where("cart_id = ?", cart.ID).Delete(&cartProducts).Error
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDb) GetSellersProducts(sellerID uint) ([]models.Product, error) {
	var products []models.Product

	err := pdb.DB.Where("seller_id = ?", sellerID).Preload("Images").Find(&products).Error
	if err != nil {
		log.Println("Error from GetSellersProduct in DB")
		return nil, err
	}
	return products, nil
}

//GET INDIVIDUAL SELLER PRODUCT
func (pdb *PostgresDb) FindSellerIndividualProduct(sellerID uint) (*models.Product, error) {

	product := &models.Product{}

	if err := pdb.DB.Preload("Category").Where("seller_id = ?", sellerID).Find(&product).Error; err != nil {
		log.Println("Error finding seller product", err)
		return nil, err
	}
	return product, nil

}

func (pdb *PostgresDb) AddTokenToBlacklist(email string, token string) error {
	blacklisted := models.Blacklist{}
	blacklisted.Token = token
	blacklisted.Email = email
	blacklisted.CreatedAt = time.Now()

	err := pdb.DB.Create(&blacklisted).Error
	if err != nil {
		log.Println("error in ad token to blacklist")
		return err
	}
	log.Println("token added to blacklist")
	return nil

}

func (pdb *PostgresDb) FindCartProductSeller(sellerID, productID uint) (*models.CartProduct, error) {

	//initiating an instance of a cart product
	cartProduct := &models.CartProduct{}

	if err := pdb.DB.Where("seller_id = ?", sellerID).Where("product_id = ?", productID).Find(cartProduct).Error; err != nil {
		log.Println("Error In find Cart", err)
		return nil, err
	}

	return cartProduct, nil
}

func (pdb *PostgresDb) DeleteAllSellerProducts(sellerID uint) error {
	product := models.Product{}
	err := pdb.DB.Where("seller_id = ?", sellerID).Delete(&product).Error
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresDb) DeleteCartProduct(buyerID, cartProductID uint) error {
	cart := &models.Cart{}
	cartProduct := &models.CartProduct{}

	err := pdb.DB.Where("buyer_id = ?", buyerID).First(cart).Error
	if err != nil {
		return err
	}

	err = pdb.DB.Where("cart_id = ?", cart.ID).Where("id = ?", cartProductID).Delete(cartProduct).Error
	if err != nil {
		return err
	}

	return nil
}

func (pdb *PostgresDb) DeleteAllFromCart(buyerID uint) error {
	cart := &models.Cart{}
	cartProduct := &models.CartProduct{}

	err := pdb.DB.Where("buyer_id = ?", buyerID).First(cart).Error
	if err != nil {
		return err
	}

	err = pdb.DB.Where("cart_id = ?", cart.ID).Delete(cartProduct).Error
	if err != nil {
		return err
	}
	return nil
}
