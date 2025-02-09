package controllers

import (
	"ecom/backend/constants"
	"ecom/backend/database"
	"ecom/backend/errResponse"
	"ecom/backend/middleware"
	"ecom/backend/models"
	"ecom/backend/utils"
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/datatypes"
)

type ProdctRequest struct {
	UUID           string         `gorm:"unique" json:"uuid,omitempty"`
	Title          string         `json:"title" gorm:"not null"`
	Description    string         `json:"description,omitempty"`
	Price          uint           `json:"price" gorm:"not null"`
	Stock          uint           `json:"stock" gorm:"default:0"`
	Category       string         `json:"category"`
	ImageURL       string         `json:"image_url,omitempty"`
	IsActive       *bool          `json:"is_active" gorm:"default:false"`
	Specifications datatypes.JSON `json:"specifications" gorm:"type:jsonb"`
}

type ProductDetailsResponse struct {
	UUID           string         `gorm:"unique" json:"uuid,omitempty"`
	Title          string         `json:"title" gorm:"not null"`
	Description    string         `json:"description,omitempty"`
	Price          uint           `json:"price" gorm:"not null"`
	Stock          uint           `json:"stock" gorm:"default:0"`
	Category       string         `json:"category"`
	ImageURL       string         `json:"image_url,omitempty"`
	IsActive       *bool          `json:"is_active,omitempty" gorm:"default:false"`
	Specifications datatypes.JSON `json:"specifications" gorm:"type:jsonb"`
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
		Title:          request.Title,
		Description:    request.Description,
		Price:          request.Price,
		Stock:          request.Stock,
		Category:       request.Category,
		ImageURL:       request.ImageURL,
		MerchantID:     merchantInfo.UUID,
		IsActive:       request.IsActive,
		Specifications: request.Specifications,
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
		Title:          request.Title,
		Description:    request.Description,
		Price:          request.Price,
		Stock:          request.Stock,
		Category:       request.Category,
		ImageURL:       request.ImageURL,
		IsActive:       request.IsActive,
		Specifications: request.Specifications,
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
		productRepo = models.InitProductsRepo(database.DB)
	)
	productId := c.Param("product_id")
	if productId == "" {
		utils.Error("product_id cannot be empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, errResponse.Generate(
			constants.ErrorBadRequest,
			"product id cannot be empty",
			constants.ErrorBadRequest,
		))
		return
	}

	product, err := productRepo.Get(&models.Product{
		UUID: productId,
	})

	if err != nil || product == nil {
		utils.Error("error in getting product || err: ", err)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "error in getting product",
		})
		return
	}

	productDetails := ProductDetailsResponse{
		Title:          product.Title,
		Description:    product.Description,
		Price:          product.Price,
		Stock:          product.Stock,
		Category:       product.Category,
		ImageURL:       product.ImageURL,
		IsActive:       product.IsActive,
		Specifications: product.Specifications,
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
			UUID:           product.UUID,
			Title:          product.Title,
			Description:    product.Description,
			Price:          product.Price,
			Stock:          product.Stock,
			Category:       product.Category,
			ImageURL:       product.ImageURL,
			Specifications: product.Specifications,
		})
	}

	// Return response
	c.JSON(http.StatusOK, productResponses)
}

// BulkUploadProducts handles bulk product upload via CSV or Excel
func BulkUploadProducts(c *gin.Context) {
	var (
		merchantRepo = models.InitMerchantRepo(database.DB)
		productRepo  = models.InitProductsRepo(database.DB)
	)

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

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file"})
		return
	}

	defer file.Close()

	var products []models.Product
	var failedRecords []map[string]string

	switch {
	case strings.HasSuffix(header.Filename, ".csv"):
		utils.Info("get csv file: ", header.Filename)
		products, failedRecords = parseCSV(file)
	case strings.HasSuffix(header.Filename, ".xlsx"):
		utils.Info("get xlsx file: ", header.Filename)
		products, failedRecords = parseExcel(file)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only CSV or Excel files are supported"})
		return
	}

	if len(products) > 0 {
		// Convert the slice of Product structs to a slice of pointers
		var productPtrs []*models.Product
		for i := range products {
			productPtrs = append(productPtrs, &products[i])
		}

		// Now pass the converted slice to CreateInBatches
		result := productRepo.CreateInBatches(productPtrs, 50, merchantInfo.UUID)
		if result != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert products"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        fmt.Sprintf("%d products uploaded successfully", len(products)),
		"failed_records": failedRecords,
	})
}

// parseCSV reads and processes a CSV file
func parseCSV(file multipart.File) ([]models.Product, []map[string]string) {
	// Reset the file cursor before reading
	file.Seek(0, 0)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		utils.Error("CSV Read Error:", err)
		return nil, []map[string]string{{"error": "Failed to read CSV file"}}
	}

	if len(records) < 2 {
		return nil, []map[string]string{{"error": "CSV file must have at least one product row"}}
	}

	headers := records[0]
	var products []models.Product
	var failedRecords []map[string]string

	for _, row := range records[1:] {
		product, err := mapRowToProduct(headers, row)
		if err != nil {
			failedRecords = append(failedRecords, map[string]string{"error": err.Error()})
			continue
		}
		products = append(products, product)
	}

	return products, failedRecords
}

// parseExcel reads and processes an Excel file
func parseExcel(file multipart.File) ([]models.Product, []map[string]string) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, []map[string]string{{"error": "Failed to read Excel file"}}
	}

	sheetName := f.GetSheetName(1)
	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) < 2 {
		return nil, []map[string]string{{"error": "Invalid Excel format"}}
	}

	headers := rows[0]
	var products []models.Product
	var failedRecords []map[string]string

	for _, row := range rows[1:] {
		product, err := mapRowToProduct(headers, row)
		if err != nil {
			failedRecords = append(failedRecords, map[string]string{"error": err.Error()})
			continue
		}
		products = append(products, product)
	}

	return products, failedRecords
}

// mapRowToProduct maps a row of data to a Product struct
func mapRowToProduct(headers, row []string) (models.Product, error) {
	if len(headers) != len(row) {
		return models.Product{}, fmt.Errorf("invalid row length")
	}

	productMap := make(map[string]string)
	for i, header := range headers {
		productMap[strings.ToLower(header)] = row[i]
	}

	price, err := strconv.ParseUint(productMap["price"], 10, 32)
	if err != nil {
		return models.Product{}, fmt.Errorf("invalid price value")
	}

	stock, err := strconv.ParseUint(productMap["stock"], 10, 32)
	if err != nil {
		return models.Product{}, fmt.Errorf("invalid stock value")
	}

	isActive := productMap["is_active"] == "true" || productMap["is_active"] == "TRUE"

	return models.Product{
		MerchantID:     productMap["merchant_id"],
		Title:          productMap["title"],
		Description:    productMap["description"],
		Price:          uint(price),
		Stock:          uint(stock),
		Category:       productMap["category"],
		ImageURL:       productMap["imageurl"],
		Specifications: datatypes.JSON([]byte(productMap["specifications"])),
		IsActive:       &isActive,
	}, nil
}
