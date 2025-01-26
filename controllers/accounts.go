package controllers

import (
	"fmt"
	"net/http"
	"time"

	"ecom/backend/constants"
	"ecom/backend/database"
	"ecom/backend/errResponse"
	"ecom/backend/models"
	"ecom/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

func OnBoardingUser(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phone_number" validate:"required"`
		FirstName   string `json:"first_name" validate:"required"`
		LastName    string `json:"last_name" validate:"required"`
		Email       string `json:"email"`
		Password    string `json:"password" validate:"required"`
	}

	// Bind JSON request body to the struct
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResponse.Generate(constants.ErrorInvalidRequestPayload,
			constants.ErrorText(constants.ErrorInvalidRequestPayload), nil))
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		var errorMessages []string
		for _, fieldErr := range validationErrors {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is required", fieldErr.Field()))
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"message": errorMessages,
		})
		return
	}

	isValidPhoneNumber := utils.IsValidPhoneNumber(req.PhoneNumber)
	if !isValidPhoneNumber {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please enter a valid phone number",
		})
		return
	}

	var existingAccount models.Account
	if err := database.DB.Where("phone_number = ?", req.PhoneNumber).First(&existingAccount).Error; err == nil {
		c.JSON(http.StatusBadRequest, errResponse.Generate(constants.ErrorUserAlreadyExists,
			constants.ErrorText(constants.ErrorUserAlreadyExists), nil))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorHashingFailed,
			constants.ErrorText(constants.ErrorHashingFailed), nil))
		return
	}

	newAccount := models.Account{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: &req.PhoneNumber,
		Emails: []*models.Email{
			{
				Email:      req.Email,
				IsVerified: utils.BoolPtr(true),
			},
		},
		Credentials: []models.Credential{
			{
				Password: string(hashedPassword),
				Type:     models.CredentialsTypePassword,
			},
		},
		Role: models.Customer,
	}

	if err := database.DB.Create(&newAccount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorDatabaseCreateFailed,
			constants.ErrorText(constants.ErrorDatabaseCreateFailed), nil))
		return
	}

	newRole := models.UserRole{
		Role:     models.Customer,
		RoleName: "Customer",
	}

	if err := database.DB.Create(&newRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorDatabaseCreateFailed,
			constants.ErrorText(constants.ErrorDatabaseCreateFailed), nil))
		return
	}

	token, err := utils.NewTokenWithClaims(constants.JWT_SECRET, utils.CustomClaims{
		Role:        newAccount.Role,
		IsPartial:   false,
		PhoneNumber: *newAccount.PhoneNumber,
		Email:       newAccount.PrimaryEmail.Email,
		FirstName:   newAccount.FirstName,
		LastName:    newAccount.LastName,
	}, time.Now().Add(5*time.Minute))

	if err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorTokenGenerationFailed,
			constants.ErrorText(constants.ErrorTokenGenerationFailed), nil))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "User registered successfully",
		"user_id":      newAccount.UUID,
		"access_token": token,
	})
}
