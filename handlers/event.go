package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"go-api-practice-10/database"
	"go-api-practice-10/models"

	"github.com/gin-gonic/gin"
)

// GetEvents 取得活動列表。
// 支援 query 參數：available（"true" 或 "false"）篩選是否還有票。
func GetEvents(c *gin.Context) {
	// TODO: 用 c.Query("available") 取出 query 參數
	// TODO: 動態組出 SQL 的 WHERE 條件（參考 practice-9 的 GetFlowers）
	// TODO: 用 database.DB.Query(query, args...) 查詢
	// TODO: defer rows.Close()
	// TODO: 用 for rows.Next() 迴圈，每一列用 rows.Scan() 掃進 models.Event
	// TODO: 回傳 200 與 []models.Event（空陣列也沒關係）
	_ = database.DB  // 避免 import 報錯，實作後可移除
	_ = strconv.Itoa // 避免 import 報錯，實作後可移除
	_ = strings.Join // 避免 import 報錯，實作後可移除
	c.JSON(http.StatusOK, []models.Event{})
}

// GetEventByID 依網址上的 id 取得單一筆活動。
// 若該 id 不存在，回傳 404；找到就回傳 200。
func GetEventByID(c *gin.Context) {
	// TODO: 用 parseID(c, "id") 取出 id
	// TODO: 用 database.DB.QueryRow() 查詢單筆 event
	// TODO: Scan 進 models.Event 的所有欄位
	// TODO: 若 err == sql.ErrNoRows → 回傳 404 gin.H{"error": "活動不存在"}
	// TODO: 其他錯誤用 respondError(c, err)
	// TODO: 成功回傳 200 與 Event
	_ = sql.ErrNoRows // 避免 import 報錯，實作後可移除
	c.JSON(http.StatusOK, gin.H{})
}

// CreateEvent 新增一筆活動（此 API 需帶 token）。
// 請求 body 需提供 name、venue、price、total_stock、event_date。
// 新增成功後回傳 201 與完整資料。
func CreateEvent(c *gin.Context) {
	// TODO: 用 c.ShouldBindJSON(&input) 綁定 body 到 models.Event
	// TODO: 失敗時用 formatValidationError(err) 回傳 400
	// TODO: INSERT 一筆新活動，注意 stock 初始值要等於 total_stock
	// TODO: 用 RETURNING 取回完整資料，Scan 進新的 models.Event
	// TODO: 成功回傳 201
	c.JSON(http.StatusCreated, gin.H{})
}

// UpdateEvent 依網址上的 id 更新一筆活動（此 API 需帶 token）。
// 若該 id 不存在，回傳 404；更新成功後回傳 200 與更新後完整資料。
func UpdateEvent(c *gin.Context) {
	// TODO: 取出 id、用 ShouldBindJSON 綁定 body
	// TODO: UPDATE 該筆活動的 name, venue, price, total_stock, event_date，記得更新 updated_at
	// TODO: 用 RETURNING 取回完整資料
	// TODO: 若 sql.ErrNoRows 回傳 404
	// TODO: 成功回傳 200
	c.JSON(http.StatusOK, gin.H{})
}

// DeleteEvent 依網址上的 id 刪除一筆活動（此 API 需帶 token）。
// 若該 id 不存在，回傳 404；刪除成功回傳 200 與成功訊息。
func DeleteEvent(c *gin.Context) {
	// TODO: 取出 id
	// TODO: DELETE 該筆活動
	// TODO: 用 result.RowsAffected() 判斷有沒有真的刪到；0 筆 → 404
	// TODO: 刪除成功回傳 200, gin.H{"message": "活動刪除成功"}
	c.JSON(http.StatusOK, gin.H{})
}
