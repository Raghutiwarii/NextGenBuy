package controllers

import (
	"ecom/backend/constants"
	"ecom/backend/database"
	"ecom/backend/errResponse"
	"ecom/backend/middleware"
	"ecom/backend/models"
	"ecom/backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProdctRequest struct {
	UUID        string `gorm:"unique" json:"uuid,omitempty"`
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description,omitempty"`
	Price       uint   `json:"price" gorm:"not null"`
	Stock       uint   `json:"stock" gorm:"default:0"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url,omitempty"`
	IsActive    *bool  `json:"is_active" gorm:"default:false"`
}

type ProductDetailsResponse struct {
	UUID        string `gorm:"unique" json:"uuid,omitempty"`
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description,omitempty"`
	Price       uint   `json:"price" gorm:"not null"`
	Stock       uint   `json:"stock" gorm:"default:0"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty" gorm:"default:false"`
}

func CreateProduct(c *gin.Context) {
	var (
		request      = ProdctRequest{}
		productRepo  = models.InitProductsRepo(database.DB)
		merchantRepo = models.InitMerchantRepo(database.DB)
	)
	// Get merchant UUID from the context (set by AuthMiddleware)
	AccountUUIDStr, ok := c.Get(middleware.AccountUUIDContextKey)
	if !ok || AccountUUIDStr == "" {
		utils.Error("role is not merchant")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get valid role",
		})
		return
	}

	utils.Info("merchant uuid ", AccountUUIDStr.(string), " ", ok)

	// Bind the request body to the Product model
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	merchantInfo, err := merchantRepo.Get(&models.Merchant{
		AccountUUID: AccountUUIDStr.(string),
	})

	if err != nil || merchantInfo == nil {
		utils.Error("error in getting merchant || err: ", err)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "error in getting merchant",
		})
		return
	}

	// Create a new product
	product := models.Product{
		Title:       request.Title,
		Description: request.Description,
		Price:       request.Price,
		Stock:       request.Stock,
		Category:    request.Category,
		ImageURL:    request.ImageURL,
		MerchantID:  merchantInfo.UUID,
		IsActive:    request.IsActive,
	}

	// Create the product
	if err := productRepo.Create(&product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully added",
	})
}

func UpdateProduct(c *gin.Context) {
	var (
		request      = ProdctRequest{}
		productRepo  = models.InitProductsRepo(database.DB)
		merchantRepo = models.InitMerchantRepo(database.DB)
	)
	productId := c.Param("product_id")
	if productId == "" {
		utils.Error("ach uuid cannot be empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, errResponse.Generate(
			constants.ErrorBadRequest,
			"product id cannot be empty",
			constants.ErrorBadRequest,
		))
		return
	}

	// Get merchant UUID from the context (set by AuthMiddleware)
	AccountUUIDStr, ok := c.Get(middleware.AccountUUIDContextKey)
	if !ok || AccountUUIDStr == "" {
		utils.Error("failed to get account uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get account uuid",
		})
		return
	}

	// Bind the request body to the Product model
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	merchantInfo, err := merchantRepo.Get(&models.Merchant{
		AccountUUID: AccountUUIDStr.(string),
	})

	if err != nil || merchantInfo == nil {
		utils.Error("error in getting merchant || err: ", err)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "error in getting merchant",
		})
		return
	}

	// Create a new product
	product := models.Product{
		Title:       request.Title,
		Description: request.Description,
		Price:       request.Price,
		Stock:       request.Stock,
		Category:    request.Category,
		ImageURL:    request.ImageURL,
		IsActive:    request.IsActive,
	}

	// Create the product
	if err := productRepo.Update(&models.Product{
		UUID:       productId,
		MerchantID: merchantInfo.UUID,
	}, &product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully updated",
	})
}
func GetProductDetails(c *gin.Context) {
	var (
		productRepo  = models.InitProductsRepo(database.DB)
		merchantRepo = models.InitMerchantRepo(database.DB)
	)
	productId := c.Param("product_id")
	if productId == "" {
		utils.Error("ach uuid cannot be empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, errResponse.Generate(
			constants.ErrorBadRequest,
			"product id cannot be empty",
			constants.ErrorBadRequest,
		))
		return
	}

	// Get merchant UUID from the context (set by AuthMiddleware)
	AccountUUIDStr, ok := c.Get(middleware.AccountUUIDContextKey)
	if !ok || AccountUUIDStr == "" {
		utils.Error("failed to get account uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get account uuid",
		})
		return
	}

	merchantInfo, err := merchantRepo.Get(&models.Merchant{
		AccountUUID: AccountUUIDStr.(string),
	})

	if err != nil || merchantInfo == nil {
		utils.Error("error in getting merchant || err: ", err)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "error in getting merchant",
		})
		return
	}

	product, err := productRepo.Get(&models.Product{
		UUID:       productId,
		MerchantID: merchantInfo.UUID,
	})

	if err != nil || product == nil {
		utils.Error("error in getting product || err: ", err)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "error in getting product",
		})
		return
	}

	productDetails := ProductDetailsResponse{
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Category:    product.Category,
		ImageURL:    product.ImageURL,
		IsActive:    product.IsActive,
	}

	c.JSON(http.StatusOK, productDetails)
}

func ListFilteredActiveProducts(c *gin.Context) {
	var products []models.Product
	query := database.DB.Where("is_active = ?", true)

	// Get filters from query parameters
	filters := c.Request.URL.Query()

	for key, values := range filters {
		switch key {
		case "category":
			query = query.Where("category = ?", values[0])
		case "min_price":
			if minPrice, err := strconv.ParseUint(values[0], 10, 32); err == nil {
				query = query.Where("price >= ?", minPrice)
			}
		case "max_price":
			if maxPrice, err := strconv.ParseUint(values[0], 10, 32); err == nil {
				query = query.Where("price <= ?", maxPrice)
			}
		case "min_stock":
			if minStock, err := strconv.ParseUint(values[0], 10, 32); err == nil {
				query = query.Where("stock >= ?", minStock)
			}
		}
	}

	// Execute query
	if err := query.Find(&products).Error; err != nil {
		utils.Error("error fetching products:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}

	// Convert to response format
	var productResponses []ProductDetailsResponse
	for _, product := range products {
		productResponses = append(productResponses, ProductDetailsResponse{
			Title:       product.Title,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			Category:    product.Category,
			ImageURL:    product.ImageURL,
		})
	}

	// Return response
	c.JSON(http.StatusOK, productResponses)
}
