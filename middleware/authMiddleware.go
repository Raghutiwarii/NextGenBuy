package middleware

import (
	"context"
	"ecom/backend/database"
	"ecom/backend/models"
	"ecom/backend/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
)

var (
	AuthorizedUserUUIDContextKey = "user_uuid"
	AuthorizedUserRoleContextKey = "role"
	IsPartialContextKey          = "is_partial"
	MerchantUUIDKey              = "merchant_uuid"
	MerchantKey                  = "merchant"
	UploadCategory               = "upload_category"
	Account                      = "account"
	ApprovalID                   = "approval_id"
	ProductID                    = "pid"
	Email                        = "email"
	AccountUUID                  = "account_uuid"
)

func AuthMiddleware(secretKey []byte, allowPartial bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			customerRepo = models.InitCustomerRepo(database.DB)
			merchantRepo = models.InitMerchantRepo(database.DB)
		)
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" || !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		token := strings.TrimPrefix(authHeader, BearerPrefix)

		// Parse and validate the token
		claims, err := utils.ParseToken(token, secretKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token", "details": err.Error()})
			return
		}

		if claims.IsPartial && !allowPartial {
			utils.Error("cannot mix tokentype partial with full auth scoped token")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":       "invalid token",
				"description": "cannot mix tokentype partial with full auth scoped token",
			})
			return
		}

		if !claims.IsPartial && allowPartial {
			utils.Error("cannot mix tokentype partial with full auth scoped token")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":       "invalid token",
				"description": "cannot mix tokentype partial with full auth scoped token",
			})
			return
		}

		if c.GetString(MerchantUUIDKey) == "" && claims.MerchantUUID != "" &&
			claims.Role == models.MerchantRole {
			merchantUUID, err := merchantRepo.Get(&models.Merchant{
				UUID: claims.MerchantUUID})
			if err != nil {
				utils.Error("error in getting user ", err)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "invalid merchant",
				})
				return
			}
			c.Set(MerchantUUIDKey, merchantUUID.UUID)
		}

		if c.GetString(AccountUUID) == "" && claims.AccountUUID != "" &&
			claims.Role == models.CustomerRole {
			Customer, err := customerRepo.Get(&models.Customer{
				AccountUUID: claims.AccountUUID})
			if err != nil {
				utils.Error("error in getting customer ", err)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "invalid customer",
				})
				return
			}
			c.Set(AccountUUID, Customer.AccountUUID)
		}

		type contextKey string

		const userUUIDKey contextKey = "user_uuid"

		ctx := context.WithValue(c.Request.Context(), userUUIDKey, claims.AccountUUID)
		ctx = context.WithValue(ctx, contextKey(AuthorizedUserRoleContextKey), claims.Role)
		ctx = context.WithValue(ctx, contextKey(IsPartialContextKey), claims.IsPartial)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
