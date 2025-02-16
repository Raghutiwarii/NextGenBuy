package controllers

import (
	"ecom/backend/database"
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

// CompleteCheckout handles checkout completion after successful payment
func CompleteCheckout(c *gin.Context) {
	var (
		reqeust   = CompleteCheckoutRequest{}
		orderRepo = models.InitOrdersrepo(database.DB)
	)
	if err := c.ShouldBindJSON(&reqeust); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	db := database.DB
	var checkout models.Checkout

	// Fetch the checkout record
	if err := db.Preload("CheckoutItems").Where("checkout_id = ?", reqeust.CheckoutID).First(&checkout).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Checkout not found"})
		return
	}

	// Ensure checkout is still pending
	if checkout.Status != models.CheckoutStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Checkout is not in a valid state for completion"})
		return
	}

	// Validate payment reference ID (mock verification here)
	if reqeust.PaymentReferenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment reference ID"})
		return
	}

	txx := db.Begin()
	// Use transaction to ensure atomicity
	err := db.Transaction(func(tx *gorm.DB) error {
		// Deduct stock for each product
		for _, item := range checkout.CheckoutItems {
			if err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				txx.Rollback()
				utils.Error("unable to update stock", err)
				return err
			}
		}

		// Update checkout status to completed
		if err := tx.Model(&checkout).Update("status", models.CheckoutStatusCompleted).Error; err != nil {
			txx.Rollback()
			utils.Error("unable to update checkout status", err)
			return err
		}

		return nil
	})

	if err != nil {
		txx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Checkout completion failed", "details": err.Error()})
		return
	}

	err = orderRepo.Create(&models.Order{
		CheckoutID:       checkout.CheckoutID,
		UserID:           checkout.UserID,
		TotalOrderAmount: checkout.TotalAmount,
		PaymentID:        reqeust.PaymentReferenceID,
	})

	if err != nil {
		txx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Order creation failed", "details": err.Error()})
		return
	}

	txx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Checkout completed successfully", "checkout_id": checkout.CheckoutID})
}
