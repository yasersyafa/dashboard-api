package httpx

import "github.com/gin-gonic/gin"

func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func RespondData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}