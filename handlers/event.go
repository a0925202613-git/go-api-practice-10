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
	available := c.Query("available")
	// TODO: 動態組出 SQL 的 WHERE 條件（參考 practice-9 的 GetFlowers）
	query := "SELECT id, name, venue, price, total_stock, stock, available, event_date, created_at, updated_at FROM events"

	var conditions []string // 儲存 WHERE 條件
	var args []interface{}  // 儲存對應的參數值

	// 動態組裝 WHERE 條件
	if available != "" {
		if parsedBool, err := strconv.ParseBool(available); err == nil {
			// 使用 len(args)+1 動態產生 $1, $2...
			conditions = append(conditions, "available = $"+strconv.Itoa(len(args)+1))
			args = append(args, parsedBool)
		}
	}

	// 如果有收集到條件，才加上 WHERE 並用 AND 串接
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// 確保輸出順序一致
	query += " ORDER BY id ASC"

	// TODO: 用 database.DB.Query(query, args...) 查詢
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		respondError(c, err)
		return
	}
	// TODO: defer rows.Close()
	defer rows.Close()
	// TODO: 用 for rows.Next() 迴圈，每一列用 rows.Scan() 掃進 models.Event
	events := []models.Event{}
	for rows.Next() {
		var e models.Event
		err := rows.Scan(&e.ID, &e.Name, &e.Venue, &e.Price, &e.TotalStock, &e.Stock, &e.Available, &e.EventDate, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			respondError(c, err)
			return
		}
		events = append(events, e)
	}

	// TODO: 回傳 200 與 []models.Event（空陣列也沒關係）
	c.JSON(http.StatusOK, events)
}

// GetEventByID 依網址上的 id 取得單一筆活動。
// 若該 id 不存在，回傳 404；找到就回傳 200。
func GetEventByID(c *gin.Context) {
	// TODO: 用 parseID(c, "id") 取出 id
	id, ok := parseID(c, "id")
	if !ok {
		return // parseID 已經回傳錯誤訊息了，這裡直接結束
	}
	// TODO: 用 database.DB.QueryRow() 查詢單筆 event
	var event models.Event

	query := "SELECT id, name, venue, price, total_stock, stock, available, event_date, created_at, updated_at FROM events WHERE id = $1"

	// TODO: Scan 進 models.Event 的所有欄位
	err := database.DB.QueryRow(query, id).Scan(
		&event.ID,
		&event.Name,
		&event.Venue,
		&event.Price,
		&event.TotalStock,
		&event.Stock,
		&event.Available,
		&event.EventDate,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	// TODO: 若 err == sql.ErrNoRows → 回傳 404 gin.H{"error": "活動不存在"}
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "活動不存在"})
		return
	}
	// TODO: 其他錯誤用 respondError(c, err)
	if err != nil {
		respondError(c, err)
		return
	}
	// TODO: 成功回傳 200 與 Event
	c.JSON(http.StatusOK, event)
}

// CreateEvent 新增一筆活動（此 API 需帶 token）。
// 請求 body 需提供 name、venue、price、total_stock、event_date。
// 新增成功後回傳 201 與完整資料。
func CreateEvent(c *gin.Context) {
	// TODO: 用 c.ShouldBindJSON(&input) 綁定 body 到 models.Event
	// TODO: 失敗時用 formatValidationError(err) 回傳 400
	var input models.Event // 準備用來裝前端傳來的資料
	if err := c.ShouldBindJSON(&input); err != nil {
		status, body := formatValidationError(err)
		c.JSON(status, body)
		return
	}
	// TODO: INSERT 一筆新活動，注意 stock 初始值要等於 total_stock
	// TODO: 用 RETURNING 取回完整資料，Scan 進新的 models.Event
	var event models.Event

	query := `
		INSERT INTO events (name, venue, price, total_stock, stock, available, event_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, name, venue, price, total_stock, stock, available, event_date, created_at, updated_at
	`

	err := database.DB.QueryRow(
		query,
		input.Name,
		input.Venue,
		input.Price,
		input.TotalStock,
		input.TotalStock,
		input.Available,
		input.EventDate,
	).Scan(
		&event.ID,
		&event.Name,
		&event.Venue,
		&event.Price,
		&event.TotalStock,
		&event.Stock,
		&event.Available,
		&event.EventDate,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		respondError(c, err)
		return
	}

	// TODO: 成功回傳 201
	c.JSON(http.StatusCreated, event)
}

// UpdateEvent 依網址上的 id 更新一筆活動（此 API 需帶 token）。
// 若該 id 不存在，回傳 404；更新成功後回傳 200 與更新後完整資料。
func UpdateEvent(c *gin.Context) {
	// TODO: 取出 id、用 ShouldBindJSON 綁定 body
	id, ok := parseID(c, "id")
	if !ok {
		return // parseID 已經回傳錯誤訊息了，這裡直接結束
	}
	
	var input models.Event // 準備用來裝前端傳來的資料
	if err := c.ShouldBindJSON(&input); err != nil {
		status, body := formatValidationError(err)
		c.JSON(status, body)
		return
	}
	// TODO: UPDATE 該筆活動的 name, venue, price, total_stock, event_date，記得更新 updated_at
	query := "UPDATE "
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
