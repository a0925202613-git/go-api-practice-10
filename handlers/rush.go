package handlers

import (
	"net/http"

	"go-api-practice-10/worker"

	"github.com/gin-gonic/gin"
)

// RushWithoutLock 模擬搶票（不加鎖）— 會出現超賣問題
// 用 goroutine 同時發送多筆購票請求，觀察沒有加 database lock 時的 race condition
func RushWithoutLock(c *gin.Context) {
	var req worker.RushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := formatValidationError(err)
		c.JSON(status, body)
		return
	}
	result := worker.RunRushWithoutLock(req)
	c.JSON(http.StatusOK, result)
}

// RushWithLock 模擬搶票（有加鎖）— 使用 SELECT FOR UPDATE 防止超賣
// 用 goroutine 同時發送多筆購票請求，觀察有加 database lock 時如何正確處理
func RushWithLock(c *gin.Context) {
	var req worker.RushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := formatValidationError(err)
		c.JSON(status, body)
		return
	}
	result := worker.RunRushWithLock(req)
	c.JSON(http.StatusOK, result)
}
