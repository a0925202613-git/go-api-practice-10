package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		fmt.Printf("[Logger] 收到請求: %s %s\n", method, path)
		c.Next()
		fmt.Printf("[Logger] 回應狀態: %d\n", c.Writer.Status())
	}
}

const DemoToken = "demo-token-123"

func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "請在 Header 帶上 Authorization: Bearer <你的token>",
			})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization 格式須為: Bearer <token>",
			})
			c.Abort()
			return
		}
		token := parts[1]
		if token != DemoToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token 不正確",
			})
			c.Abort()
			return
		}
		c.Set("token", token)
		c.Next()
	}
}
