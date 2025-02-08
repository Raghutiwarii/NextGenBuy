package controllers

import (
	"ecom/backend/constants"
	"ecom/backend/database"
	"ecom/backend/errResponse"
	"ecom/backend/models"
	"ecom/backend/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OnBoadingMerchantRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Email       string `json:"email"`
	Password    string `json:"password" validate:"required"`
}

func OnBoardingMerchant(c *gin.Context) {
	var (
		req          = OnBoadingMerchantRequest{}
		AccountRepo  = models.InitAccountRepo(database.DB)
		merchantRepo = models.InitMerchantRepo(database.DB)
	)

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

	existingAccount, err := AccountRepo.Get(&models.Account{
		PhoneNumber: &req.PhoneNumber})

	if err == nil || existingAccount != nil {
		c.JSON(http.StatusBadRequest, errResponse.Generate(constants.ErrorUserAlreadyExists,
			constants.ErrorText(constants.ErrorUserAlreadyExists), nil))
		return
	}

	hashedPassword, err := utils.HashPasswordWithSecret(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorHashingFailed,
			constants.ErrorText(constants.ErrorHashingFailed), nil))
		return
	}

	newAccount := models.Account{
		PhoneNumber: &req.PhoneNumber,
		Emails: []*models.Email{
			{
				Email:      req.Email,
				IsVerified: utils.BoolPtr(true),
			},
		},
		Password: string(hashedPassword),
		RoleID:   models.MerchantRole,
	}

	err = AccountRepo.Create(&newAccount)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorDatabaseCreateFailed,
			constants.ErrorText(constants.ErrorDatabaseCreateFailed), nil))
		return
	}

	newRole := models.UserRole{
		Role:     int(newAccount.RoleID),
		RoleName: models.GetRoleName(models.MerchantRole),
	}

	if err := database.DB.Create(&newRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorDatabaseCreateFailed,
			constants.ErrorText(constants.ErrorDatabaseCreateFailed), nil))
		return
	}

	newMerchant := models.Merchant{
		AccountUUID:              newAccount.AccountId,
		ApplicationCurrentStatus: models.MerchantOnboardingStateVerifyAccount,
	}

	if err := merchantRepo.Create(&newMerchant); err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorDatabaseCreateFailed,
			constants.ErrorText(constants.ErrorDatabaseCreateFailed), nil))
		return
	}

	token, err := utils.NewTokenWithClaims(constants.JWT_SECRET, utils.CustomClaims{
		Role:        models.GetRoleName(newAccount.RoleID),
		IsPartial:   true,
		AccountUUID: newAccount.AccountId,
	}, time.Now().Add(60*60*time.Minute))

	if err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorTokenGenerationFailed,
			constants.ErrorText(constants.ErrorTokenGenerationFailed), nil))
		return
	}

	otpCode := utils.GenerateOTP()

	// Create OTP instance
	otp := models.OTP{
		AccountUUID: newAccount.AccountId,
		Code:        otpCode,
		ExpiresAt:   time.Now().Add(time.Second * 50),
	}

	utils.Info("Otp sent to phone successfully. OTP is ", otpCode)

	// Save OTP to the database
	if err := database.DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create OTP",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"goto":         "VERIFY_OTP",
		"account_id":   newAccount.AccountId,
		"access_token": token,
	})
}
