package controllers

import (
	"fmt"
	"net/http"
	"time"

	"ecom/backend/constants"
	"ecom/backend/database"
	"ecom/backend/errResponse"
	"ecom/backend/middleware"
	"ecom/backend/models"
	"ecom/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type verifyRequest struct {
	OTP string `json:"otp"`
}

type OnBoadingCustomerRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Email       string `json:"email"`
	Password    string `json:"password" validate:"required"`
}

func OnBoardingCustomer(c *gin.Context) {
	var (
		req         = OnBoadingCustomerRequest{}
		AccountRepo = models.InitAccountRepo(database.DB)
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
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		PhoneNumber: &req.PhoneNumber,
		Emails: []*models.Email{
			{
				Email:      req.Email,
				IsVerified: utils.BoolPtr(true),
			},
		},
		Password: string(hashedPassword),
		RoleID:   models.CustomerRole,
	}

	err = AccountRepo.Create(&newAccount)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorDatabaseCreateFailed,
			constants.ErrorText(constants.ErrorDatabaseCreateFailed), nil))
		return
	}

	newRole := models.UserRole{
		Role:     int(newAccount.RoleID),
		RoleName: models.GetRoleName(models.CustomerRole),
	}

	if err := database.DB.Create(&newRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorDatabaseCreateFailed,
			constants.ErrorText(constants.ErrorDatabaseCreateFailed), nil))
		return
	}

	token, err := utils.NewTokenWithClaims(constants.JWT_SECRET, utils.CustomClaims{
		Role:        models.GetRoleName(newAccount.RoleID),
		IsPartial:   false,
		AccountUUID: newAccount.UUID,
	}, time.Now().Add(5*time.Minute))

	if err != nil {
		c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorTokenGenerationFailed,
			constants.ErrorText(constants.ErrorTokenGenerationFailed), nil))
		return
	}

	otpCode := utils.GenerateOTP()

	// Create OTP instance
	otp := models.OTP{
		AccountUUID: newAccount.UUID,
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
		"message":      "User registered successfully",
		"user_id":      newAccount.UUID,
		"access_token": token,
	})
}

func Login(c *gin.Context) {
	var req struct {
		PhoneNumber *string `json:"phone_number"`
		Email       string  `json:"email"`
		Password    string  `json:"password" validate:"required"`
	}
	var (
		userAccountRepo = models.InitAccountRepo(database.DB)
		emailsRepo      = models.InitEmailRepo(database.DB)
	)

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

	if req.Email == "" && req.PhoneNumber == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Either phone number or email is required",
		})
		return
	}

	if req.Email != "" {
		existingAccount, err := emailsRepo.Get(&models.Email{
			Email: req.Email,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "User does not exist",
			})
			return
		}

		var accountWithEmail *models.Account
		accountWithEmail, err = userAccountRepo.Get(&models.Account{
			PrimaryEmailID: &existingAccount.ID,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "User does not exist",
			})
			return
		}

		err = utils.CompareHashAndPasswordWithSecret(accountWithEmail.Password, req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Incorrect password!",
			})
			return
		}

		token, err := utils.NewTokenWithClaims(constants.JWT_SECRET, utils.CustomClaims{
			Role:        models.GetRoleName(accountWithEmail.RoleID),
			IsPartial:   false,
			AccountUUID: accountWithEmail.UUID,
		}, time.Now().Add(5*time.Minute))

		if err != nil {
			c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorTokenGenerationFailed,
				constants.ErrorText(constants.ErrorTokenGenerationFailed), nil))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"goto":         "continue",
			"access_token": token,
		})
		return
	}

	if req.PhoneNumber != nil {
		existingAccount, err := userAccountRepo.Get(&models.Account{
			PhoneNumber: req.PhoneNumber,
		})

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "User does not exist",
			})
			return
		}

		err = utils.CompareHashAndPasswordWithSecret(existingAccount.Password, req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Incorrect password!",
			})
			return
		}

		token, err := utils.NewTokenWithClaims(constants.JWT_SECRET, utils.CustomClaims{
			Role:        models.GetRoleName(existingAccount.RoleID),
			IsPartial:   false,
			AccountUUID: existingAccount.UUID,
		}, time.Now().Add(5*time.Minute))

		if err != nil {
			c.JSON(http.StatusInternalServerError, errResponse.Generate(constants.ErrorTokenGenerationFailed,
				constants.ErrorText(constants.ErrorTokenGenerationFailed), nil))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"goto":         "continue",
			"access_token": token,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Please check your credentials",
	})
}

func GetUserProfile(c *gin.Context) {

}

func VerifyOTP(c *gin.Context) {
	var (
		req          = verifyRequest{}
		customerRepo = models.InitCustomerRepo(database.DB)
		merchantRepo = models.InitMerchantRepo(database.DB)
		otpRepo      = models.InitOTPRepo(database.DB)
	)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse request body",
		})
		return
	}

	role, ok := c.Get(middleware.AuthorizedUserRoleContextKey)

	utils.Info("getting role from context ", role)

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get role || invalid role",
		})
		return
	}

	roleStr, ok := role.(string)
	if !ok {
		utils.Error("Role is not of type string")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get role || invalid string",
		})
		return
	}

	utils.Info("getting role from string ", roleStr)

	account, ok := c.Get(middleware.AccountUUIDContextKey)
	utils.Info("getting customer account ", account)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Failed to get CustomerID",
		})
		return
	}
	accountiD, ok := account.(string)
	if !ok {
		utils.Error("Role is not of type string")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get role || unable to get role",
		})
		return
	}

	if roleStr == models.GetRoleName(models.CustomerRole) {
		customer, err := customerRepo.Get(&models.Customer{
			AccountUUID: accountiD,
		})

		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Failed to get customer details",
			})
			return
		}

		otp, err := otpRepo.Get(&models.OTP{
			AccountUUID: customer.AccountUUID,
		})
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Failed to get OTP",
			})
		}

		if otp.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "OTP expired",
			})
			return
		}

		if otp.Code == req.OTP {
			err = otpRepo.Update(&models.OTP{
				ExpiresAt: time.Now(),
			}, otp)

			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Failed to update OTP",
				})
				return
			}

			token, err := utils.NewTokenWithClaims(constants.JWT_SECRET, utils.CustomClaims{
				Role:        role.(string),
				IsPartial:   false,
				AccountUUID: customer.AccountUUID,
			}, time.Now().Add(60*60*time.Minute))

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate token",
				})
				return
			}

			c.JSON(http.StatusAccepted, gin.H{
				"message":      "OTP verified successfully",
				"access_token": token,
			})
			return
		}

	} else if roleStr == models.GetRoleName(models.MerchantRole) {
		merchant, err := merchantRepo.Get(&models.Merchant{
			AccountUUID: accountiD,
		})

		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Failed to get merchant details",
			})
			return
		}

		otp, err := otpRepo.Get(&models.OTP{
			AccountUUID: merchant.AccountUUID,
		})
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Failed to get OTP",
			})
		}

		if otp.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "OTP expired",
			})
			return
		}

		if otp.Code == req.OTP {
			token, err := utils.NewTokenWithClaims(constants.JWT_SECRET, utils.CustomClaims{
				Role:        role.(string),
				IsPartial:   false,
				AccountUUID: merchant.AccountUUID,
			}, time.Now().Add(60*60*time.Minute))

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate token",
				})
				return
			}

			c.JSON(http.StatusAccepted, gin.H{
				"message":      "OTP verified successfully",
				"access_token": token,
			})
			return
		}
	}

	c.JSON(http.StatusForbidden, gin.H{
		"message": "Invalid OTP",
	})
}
