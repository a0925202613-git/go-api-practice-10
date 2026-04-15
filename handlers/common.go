package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var ErrNotFound = errors.New("not found")

// respondError 依錯誤類型回傳對應的 HTTP 狀態與 body。
// 若 err 為 ErrNotFound，回傳 404；其他錯誤回傳 500。
func respondError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

// parseID 從網址參數（例如 :id）讀取整數 id。
// 若不是合法數字，回傳 400；若合法則回傳該數字與 true。
func parseID(c *gin.Context, param string) (int, bool) {
	id, err := strconv.Atoi(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}
