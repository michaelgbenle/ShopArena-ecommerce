package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/decadevs/shoparena/models"
	"github.com/gin-gonic/gin"
)

func (h *Handler) BuyerSignUpHandler(c *gin.Context) {
	buyer := &models.Buyer{}
	err := c.ShouldBindJSON(buyer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to bind JSON",
		})
		return
	}
	if buyer.Username == "" || buyer.FirstName == "" || buyer.LastName == "" || buyer.Password == "" || buyer.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Enter all fields",
		})
		return
	}
	validEmail := buyer.ValidMailAddress()
	if validEmail == false {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "enter valid email",
		})
		return
	}

	_, err = h.DB.FindBuyerByUsername(buyer.Username)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username exists",
		})
		return
	}
	_, err = h.DB.FindBuyerByEmail(buyer.Email)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email exists",
		})
		return
	}

	_, err = h.DB.FindBuyerByPhone(buyer.PhoneNumber)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "phone number exists",
		})
		return
	}

	if err = buyer.HashPassword(); err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	_, err = h.DB.CreateBuyer(buyer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not create buyer",
		})
		return
	}
	cart := &models.Cart{BuyerID: buyer.ID}
	_, err = h.DB.CreateBuyerCart(cart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": error.Error(err),
		})
		return
	}
	h.BuyerSignUpValidation(buyer.Email, c)

}

func (h *Handler) BuyerSignUpValidation(email string, c *gin.Context) {
	// generate token that'll be used to activate the account
	secretString := os.Getenv("JWT_SECRET")
	verify_Token, _ := h.Mail.GenerateNonAuthToken(email, secretString)
	// the link to be clicked in order to perform password reset
	//link := "https://shoparena-frontend.vercel.app/buyer/success" + *verifyToken
	link := "localhost:9094/api/v1/buyer/success?verify_token=" + *verify_Token
	// define the body of the email
	body := " <a href='" + link + "'>Click here to activate your account</a>"
	//html := "<strong>" + body + "</strong>"

	//initialize the email send out
	privateAPIKey := os.Getenv("MAILGUN_API_KEY")
	yourDomain := os.Getenv("DOMAIN_STRING")
	err := h.Mail.SendMail("Email Activation", body, email, privateAPIKey, yourDomain)

	//if email was sent return 200 status code
	if err == nil {
		c.JSON(200, gin.H{"message": "please check your email for activation link"})
		c.Abort()
		return
	} else {
		log.Println(err)
		c.JSON(500, gin.H{"error": "please try again"})
		c.Abort()
		return
	}
}

func (h *Handler) BuyerSignUpActivation(c *gin.Context) {
	token, _ := c.GetQuery("verify_token")
	secret_key := os.Getenv("JWT_SECRET")
	email, err := h.Mail.DecodeToken(token, secret_key)
	if err != nil {
		c.IndentedJSON(400, gin.H{
			"error": "internal server error",
		})
		return
	}
	err = h.DB.ValidateBuyer(email)
	if err != nil {
		c.IndentedJSON(400, gin.H{
			"error": "internal server error",
		})
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"message": "congratulations, your account is now activated",
	})
}

func (h *Handler) SellerSignUpHandler(c *gin.Context) {

	seller := &models.Seller{}
	err := c.ShouldBindJSON(seller)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to bind json",
		})
		return
	}

	if seller.Username == "" || seller.FirstName == "" || seller.LastName == "" || seller.Password == "" || seller.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Enter all fields",
		})
		return
	}
	validEmail := seller.ValidMailAddress()
	if validEmail == false {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "enter valid email",
		})
		return
	}

	_, err = h.DB.FindSellerByUsername(seller.Username)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username exists",
		})
		return
	}
	_, err = h.DB.FindSellerByEmail(seller.Email)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email exists",
		})
		return
	}

	_, err = h.DB.FindSellerByPhone(seller.PhoneNumber)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "phone number exists",
		})
		return

	}
	if err := seller.HashPassword(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}
	_, err = h.DB.CreateSeller(seller)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not create seller",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Sign Up Successful",
	})
}
