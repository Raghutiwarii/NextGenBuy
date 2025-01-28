package utils

import (
	"ecom/backend/constants"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	JWTIssuer = "raghu"
	MockOTP   = "123123"
)

type CustomClaims struct {
	Role        uint   `json:"role"`
	IsPartial   bool   `json:"is_partial,omitempty"`
	AccountUUID string `json:"account_uuid,omitempty"`
}

type JWTTokenClaims struct {
	CustomClaims
	jwt.RegisteredClaims
}

func NewTokenWithClaims(secretKey []byte, customClaims CustomClaims, expires time.Time) (*string, error) {
	claims := JWTTokenClaims{
		customClaims,
		jwt.RegisteredClaims{
			Issuer:    JWTIssuer,
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Signed string
	signedString, err := token.SignedString(secretKey)
	if err != nil {
		logger.Error("error in creating new jwt token ", err)
		return nil, err
	}
	return &signedString, nil
}

func ParseToken(token string, secretKey []byte, ignoreValidity ...bool) (*CustomClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		logger.Error("unable to parse token ", err)
		if err.Error() == "token has invalid claims: token is expired" || err.Error() == jwt.ErrTokenExpired.Error() {
			if len(ignoreValidity) > 0 && ignoreValidity[0] {
				logger.Info("overriding expiration check")
				claims, ok := parsedToken.Claims.(*JWTTokenClaims)
				if ok {
					return &claims.CustomClaims, nil
				}
			}
		}
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(*JWTTokenClaims); ok {
		if parsedToken.Valid {
			return &claims.CustomClaims, nil
		} else if !parsedToken.Valid && len(ignoreValidity) > 0 && ignoreValidity[0] {
			logger.Info("overriding validity check")
			return &claims.CustomClaims, nil
		}
	}
	return nil, errors.New("invalid token")
}

func HashPasswordWithSecret(password string) (string, error) {
	combinedPassword := password + string(constants.PASSWORD_SECRET)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(combinedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CompareHashAndPasswordWithSecret(hashedPassword, password string) error {
	combinedPassword := password + string(constants.PASSWORD_SECRET)
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(combinedPassword))
	if err != nil {
		return errors.New("invalid password or secret")
	}

	return nil
}
