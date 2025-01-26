package errResponse

import "github.com/gin-gonic/gin"

// Generate generates a standardized error response
func Generate(code int, message string, data interface{}) gin.H {
	return gin.H{
		"code":    code,
		"message": message,
		"data":    data,
	}
}
