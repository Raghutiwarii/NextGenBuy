package middleware

import (
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
	AuthorizedUserRoleContextKey = "role"
	IsPartialContextKey          = "is_partial"
	MerchantUUIDKey              = "merchant_uuid"
	AccountUUIDContextKey        = "account_uuid"
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

		utils.Info("getting context role from token ", claims.Role)

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
			claims.Role == models.GetRoleName(models.MerchantRole) {
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

		if c.GetString(AccountUUIDContextKey) == "" && claims.AccountUUID != "" &&
			claims.Role == models.GetRoleName(models.CustomerRole) {
			Customer, err := customerRepo.Get(&models.Customer{
				AccountUUID: claims.AccountUUID})
			if err != nil {
				utils.Error("error in getting customer ", err)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "invalid customer",
				})
				return
			}
			c.Set(AccountUUIDContextKey, Customer.AccountUUID)
		}

		c.Set(AuthorizedUserRoleContextKey, claims.Role)
		c.Set(IsPartialContextKey, claims.IsPartial)
		c.Set(MerchantUUIDKey, claims.MerchantUUID)
		c.Set(AccountUUIDContextKey, claims.AccountUUID)

		c.Next()
	}
}
