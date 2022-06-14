package handlers

import (
	"github.com/decadevs/shoparena/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

func (h *Handler) BuyerForgotPasswordEMailHandler(c *gin.Context) {
	// crete a password reset struct and initialize it
	var forgotPassword models.ResetPasswordRequest

	err := c.BindJSON(&forgotPassword)
	if err != nil {
		c.JSON(400, gin.H{"message": "please fill all fields"})
		c.Abort()
		return
	}
	buyer, err := h.DB.FindBuyerByEmail(forgotPassword.Email)
	// return 404 status if there is an error getting the buyer
	if err != nil {
		c.JSON(404, gin.H{"message": "buyer not found"})
		c.Abort()
		return
	}

	// generate token that'll be used to reset the password
	secretString := os.Getenv("JWTSECRET")
	resetToken, _ := h.Mail.GenerateNonAuthToken(buyer.Email, secretString)
	// the link to be clicked in order to perform password reset
	//link := "http://localhost:8085/api/v1/buyerresetpassword?reset_token=" + *resetToken
	link := "https://shoparena-frontend.vercel.app/buyer/forgot/" + *resetToken
	// define the body of the email
	body := "Here is your reset <a href='" + link + "'>link</a>"
	html := "<strong>" + body + "</strong>"

	//intialize the email sendout
	privateAPIKey := os.Getenv("MAILGUN_API_KEY")
	yourDomain := os.Getenv("DOMAIN_STRING")
	err = h.Mail.SendMail("forgot Password", html, buyer.Email, privateAPIKey, yourDomain)

	//if email was sent return 200 status code
	if err == nil {
		c.JSON(200, gin.H{"message": "please check your email for password reset link"})
		c.Abort()
		return
	} else {
		log.Println(err)
		c.JSON(500, gin.H{"message": "something went wrong while trying to send you a mail, please try again"})
		c.Abort()
		return
	}

}

func (h *Handler) BuyerForgotPasswordResetHandler(c *gin.Context) {
	// instance of the forgot password request
	var password models.ForgotPassword
	// requires all fields and returns an error if some fields are missing
	err := c.BindJSON(&password)
	if err != nil {
		// Return response when we some fields are missing
		log.Println(err)
		c.JSON(406, gin.H{"message": "please provide all fields"})
		c.Abort()
		return
	}

	// verifies if the new password string and the string confirmation string match
	if password.NewPassword != password.ConfirmNewPassword {
		// Return response when passwords mismatch
		c.JSON(400, gin.H{"message": "password mismatch"})
		c.Abort()
		return
	}

	// gets the reset token from the dynamic route
	resetToken, _ := c.GetQuery("reset_token")

	// decodes token to get user email
	secretString := os.Getenv("JWTSECRET")
	userEmail, err := h.Mail.DecodeToken(resetToken, secretString)
	if err != nil {
		// Return response when we get an error while decoding the token
		log.Println(err)
		c.JSON(500, gin.H{"message": "Something wrong happened, Please try again later"})
		c.Abort()
		return
	}
	// check the db for a buyer with the email decoded from the token
	buyer, err := h.DB.FindBuyerByEmail(userEmail)
	if err != nil {
		// Return response when we get an error while fetching buyer
		log.Println(err)
		c.JSON(500, gin.H{"message": "buyer not found"})
		c.Abort()
		return
	}

	// generates a new password hash from the password string in the password reset request
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(password.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		// Return response when we get an error while hatching the password
		log.Println(err)
		c.JSON(500, gin.H{"message": "internal server error"})
		c.Abort()
		return
	}

	// resets the buyer password hash by email in the database
	_, err = h.DB.BuyerResetPassword(buyer.Email, string(newPasswordHash))
	if err != nil {
		// Return response if we are not able to update user password
		c.JSON(500, gin.H{"message": "Something happened while updating your password try again"})
		c.Abort()
		return
	}

	// returns a response after successfully resetting the password
	c.JSON(200, gin.H{"message": "password reset successful"})
	c.Abort()
	return
}

func (h *Handler) SellerForgotPasswordEMailHandler(c *gin.Context) {
	// crete a password reset struct and initialize it
	var forgotPassword models.ResetPasswordRequest

	err := c.BindJSON(&forgotPassword)
	if err != nil {
		c.JSON(400, gin.H{"message": "please fill all fields"})
		c.Abort()
		return
	}
	seller, err := h.DB.FindSellerByEmail(forgotPassword.Email)
	// return 404 status if there is an error getting the buyer
	if err != nil {
		c.JSON(404, gin.H{"message": "seller not found"})
		c.Abort()
		return
	}

	// generate token that'll be used to reset the password
	secretString := os.Getenv("JWTSECRET")
	resetToken, _ := h.Mail.GenerateNonAuthToken(seller.Email, secretString)
	// the link to be clicked in order to perform password reset
	//link := "http://localhost:8085/api/v1/sellerresetpassword?reset_token=" + *resetToken
	link := "https://shoparena-frontend.vercel.app/seller/forgot/" + *resetToken
	// define the body of the email
	body := "Here is your reset <a href='" + link + "'>link</a>"
	html := "<strong>" + body + "</strong>"

	//intialize the email sendout
	privateAPIKey := os.Getenv("MAILGUN_API_KEY")
	yourDomain := os.Getenv("DOMAIN_STRING")
	err = h.Mail.SendMail("forgot Password", html, seller.Email, privateAPIKey, yourDomain)

	//if email was sent return 200 status code
	if err == nil {
		c.JSON(200, gin.H{"message": "please check your email for password reset link"})
		c.Abort()
		return
	} else {
		log.Println(err)
		c.JSON(500, gin.H{"message": "something went wrong while trying to send you a mail, please try again"})
		c.Abort()
		return
	}

}

func (h *Handler) SellerForgotPasswordResetHandler(c *gin.Context) {
	// instance of the forgot password request
	var password models.ForgotPassword
	// requires all fields and returns an error if some fields are missing
	err := c.BindJSON(&password)
	if err != nil {
		// Return response when we some fields are missing
		log.Println(err)
		c.JSON(406, gin.H{"message": "please provide all fields"})
		c.Abort()
		return
	}

	// verifies if the new password string and the string confirmation string match
	if password.NewPassword != password.ConfirmNewPassword {
		// Return response when passwords mismatch
		c.JSON(400, gin.H{"message": "password mismatch"})
		c.Abort()
		return
	}

	// gets the reset token from the dynamic route
	resetToken, _ := c.GetQuery("reset_token")

	// decodes token to get user email
	secretString := os.Getenv("JWTSECRET")
	userEmail, err := h.Mail.DecodeToken(resetToken, secretString)
	if err != nil {
		// Return response when we get an error while decoding the token
		log.Println(err)
		c.JSON(500, gin.H{"message": "Something wrong happened, Please try again later"})
		c.Abort()
		return
	}
	// check the db for a buyer with the email decoded from the token
	seller, err := h.DB.FindSellerByEmail(userEmail)
	if err != nil {
		// Return response when we get an error while fetching buyer
		log.Println(err)
		c.JSON(500, gin.H{"message": "buyer not found"})
		c.Abort()
		return
	}

	// generates a new password hash from the password string in the password reset request
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(password.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		// Return response when we get an error while hatching the password
		log.Println(err)
		c.JSON(500, gin.H{"message": "internal server error"})
		c.Abort()
		return
	}

	// resets the buyer password hash by email in the database
	_, err = h.DB.SellerResetPassword(seller.Email, string(newPasswordHash))
	if err != nil {
		// Return response if we are not able to update user password
		c.JSON(500, gin.H{"message": "Something happened while updating your password try again"})
		c.Abort()
		return
	}

	// returns a response after successfully resetting the password
	c.JSON(200, gin.H{"message": "password reset successful"})
	c.Abort()
	return
}
