package middleware

import (
	"context"
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
)

func AuthMiddleware(secretKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		ctx := context.WithValue(c.Request.Context(), AuthorizedUserUUIDContextKey, claims.AccountUUID)
		ctx = context.WithValue(ctx, AuthorizedUserRoleContextKey, claims.Role)
		ctx = context.WithValue(ctx, IsPartialContextKey, claims.IsPartial)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
