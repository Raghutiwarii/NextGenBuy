package controllers

import (
	"ecom/backend/constants"
	"ecom/backend/database"
	"ecom/backend/errResponse"
	"ecom/backend/middleware"
	"ecom/backend/models"
	"ecom/backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CheckoutRequest represents the payload for creating a checkout
type CheckoutRequest struct {
	Items []struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  uint `json:"quantity" binding:"required"`
	} `json:"items" binding:"required"`
}

type CompleteCheckoutRequest struct {
	CheckoutID         string `json:"checkout_id" binding:"required"`
	PaymentReferenceID string `json:"payment_reference_id" binding:"required"`
}

type CheckoutDetailResponse struct {
	CheckoutID  string                `json:"checkout_id"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
	TotalAmount uint                  `json:"total_amount"`
	Status      string                `json:"status"`
	Prouducts   []models.CheckoutItem `json:"products"`
}

// CreateCheckout handles checkout creation
func CreateCheckout(c *gin.Context) {
	// Get merchant UUID from the context (set by AuthMiddleware)
	AccountUUIDStr, ok := c.Get(middleware.AccountUUIDContextKey)
	if !ok || AccountUUIDStr == "" {
		utils.Error("role is not merchant")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get valid role"})
		return
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	db := database.DB
	var totalAmount int
	var checkoutItems []models.CheckoutItem

	// Validate and process each item
	for _, item := range req.Items {
		var product models.Product
		if err := db.First(&product, item.ProductID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found", "product_id": item.ProductID})
			return
		}

		itemTotal := int(product.Price) * int(item.Quantity)
		totalAmount += itemTotal

		checkoutItems = append(checkoutItems, models.CheckoutItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Price:      int(product.Price),
			TotalPrice: itemTotal,
		})
	}

	// Create checkout record (without stock deduction)
	checkout := models.Checkout{
		UserID:        AccountUUIDStr.(string),
		TotalAmount:   totalAmount,
		Status:        models.CheckoutStatusPending, // Pending until payment is confirmed
		CheckoutItems: checkoutItems,
	}

	if err := db.Create(&checkout).Error; err != nil {
		utils.Error("unable to create checkout", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Checkout creation failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Checkout created successfully", "checkout_id": checkout.CheckoutID})
}

func GetCheckoutDetails(c *gin.Context) {
	var (
		checkoutRepo = models.InitCheckoutrepo(database.DB)
	)
	checkoutId := c.Param("checkout_id")
	if checkoutId == "" {
		utils.Error("checkout id cannot be empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, errResponse.Generate(
			constants.ErrorBadRequest,
			"checkout id cannot be empty",
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

	checkout, err := checkoutRepo.Get(&models.Checkout{
		CheckoutID: checkoutId,
	})

	if err != nil || checkout == nil {
		utils.Error("error in getting checkout || err: ", err)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "error in getting checkout",
		})
		return
	}

	response := CheckoutDetailResponse{
		CreatedAt:   checkout.CreatedAt.Local().String(),
		UpdatedAt:   checkout.UpdatedAt.Local().String(),
		CheckoutID:  checkoutId,
		TotalAmount: uint(checkout.TotalAmount),
		Status:      string(checkout.Status),
		Prouducts:   checkout.CheckoutItems,
	}

	c.JSON(http.StatusOK, response)
}

// CompleteCheckout handles checkout completion after successful payment
func CompleteCheckout(c *gin.Context) {
	var (
		request      = CompleteCheckoutRequest{}
		orderRepo    = models.InitOrdersrepo(database.DB)
		checkoutRepo = models.InitCheckoutrepo(database.DB)
	)

	// Bind the request body
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Start a transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch the checkout record
	checkout, err := checkoutRepo.GetWithTx(tx, &models.Checkout{CheckoutID: request.CheckoutID})
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Checkout not found"})
		return
	}

	// Ensure checkout is still pending
	if checkout.Status != models.CheckoutStatusPending {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Checkout is not in a valid state for completion"})
		return
	}

	// Validate payment reference ID (mock verification here)
	if request.PaymentReferenceID == "" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment reference ID"})
		return
	}

	// Deduct stock for each product in the checkout
	for _, item := range checkout.CheckoutItems {
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			utils.Error("unable to update stock", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Checkout completion failed", "details": err.Error()})
			return
		}
	}

	// Update checkout status to completed
	if err := tx.Model(&checkout).Update("status", models.CheckoutStatusCompleted).Error; err != nil {
		tx.Rollback()
		utils.Error("unable to update checkout status", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Checkout completion failed", "details": err.Error()})
		return
	}

	// Create order record
	if err := orderRepo.CreateWithTx(tx, &models.Order{
		CheckoutID:       checkout.CheckoutID,
		UserID:           checkout.UserID,
		TotalOrderAmount: checkout.TotalAmount,
		PaymentID:        request.PaymentReferenceID,
	}); err != nil {
		tx.Rollback()
		utils.Error("order creation failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Checkout completion failed", "details": err.Error()})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.Error("unable to commit transaction", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Checkout completion failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Checkout completed successfully", "checkout_id": checkout.CheckoutID})
}
