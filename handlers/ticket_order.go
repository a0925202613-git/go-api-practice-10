package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"go-api-practice-10/database"
	"go-api-practice-10/models"

	"github.com/gin-gonic/gin"
)

// GetTicketOrders 取得所有訂單列表，JOIN events 帶出活動名稱，依 id 排序。
func GetTicketOrders(c *gin.Context) {
	// TODO: 寫一個 JOIN 查詢，從 ticket_orders 和 events 兩張表取資料
	query := `
		SELECT t.id, t.event_id, t.customer_name, t.quantity, t.total_price, t.status, 
		t.ordered_at, t.cancelled_at, t.created_at, t.updated_at, e.name AS event_name
		FROM ticket_orders t JOIN events e
		ON e.id = t.event_id
		ORDER BY t.id
		`

	rows, err := database.DB.Query(query)
	if err != nil {
		respondError(c, err)
		return
	}
	defer rows.Close()

	orders := []models.TicketOrderWithEvent{}
	for rows.Next() {
		var o models.TicketOrderWithEvent
		if err := rows.Scan(&o.ID, &o.EventID, &o.CustomerName, &o.Quantity, &o.TotalPrice, &o.Status,
			&o.OrderedAt, &o.CancelledAt, &o.CreatedAt, &o.UpdatedAt, &o.EventName); err != nil {
			respondError(c, err)
			return
		}
		orders = append(orders, o)
	}
	//       要取出訂單的所有欄位，加上活動名稱（e.name AS event_name）
	// TODO: 用 database.DB.Query() 執行查詢
	// TODO: defer rows.Close()
	// TODO: 用 for rows.Next() 迴圈，Scan 進 models.TicketOrderWithEvent
	//       注意：cancelled_at 是 *time.Time，可以是 NULL
	// TODO: 回傳 200 與 []models.TicketOrderWithEvent
	c.JSON(http.StatusOK, orders)
}

// GetTicketOrderByID 依網址上的 id 取得單一筆訂單，JOIN events 帶出活動名稱。
func GetTicketOrderByID(c *gin.Context) {
	// TODO: 取出 id
	id, ok := parseID(c, "id")
	if !ok {
		return // parseID 已經回傳錯誤訊息了，這裡直接結束
	}
	// TODO: 跟 GetTicketOrders 類似的 JOIN 查詢，但加上 WHERE 條件篩選單筆
	query := `
    SELECT t.id, t.event_id, t.customer_name, t.quantity, t.total_price, 
    t.status, t.ordered_at, t.cancelled_at, t.created_at, t.updated_at, e.name AS event_name
    FROM ticket_orders t JOIN events e
    ON e.id = t.event_id
	WHERE id = $1
`

	var o models.TicketOrderWithEvent
	// TODO: 用 QueryRow 查詢，Scan 進 models.TicketOrderWithEvent
	if err := database.DB.QueryRow(query, id).Scan(&o.ID, &o.EventID, &o.CustomerName, &o.Quantity, &o.TotalPrice, &o.Status,
		&o.OrderedAt, &o.CancelledAt, &o.CreatedAt, &o.UpdatedAt, &o.EventName); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "訂單不存在"})
		} else {
			respondError(c, err)
		}
		return
	}
	// TODO: 若 sql.ErrNoRows → 404 "訂單不存在"
	// TODO: 成功回傳 200
	c.JSON(http.StatusOK, o)
}

// CreateTicketOrder 建立一筆購票訂單（此 API 需帶 token）。
// 流程：
//  1. 綁定 body 並驗證（event_id, customer_name, quantity 必填，quantity 最多 4 張）
//  2. 呼叫 validateEventForOrder 檢查活動存在、available、庫存足夠
//  3. 查詢票價，計算 total_price = 單價 × 數量
//  4. 開 transaction
//  5. INSERT 訂單（用 RETURNING 取回完整資料）
//  6. UPDATE events 扣庫存，庫存歸零時把 available 設為 false
//  7. Commit
func CreateTicketOrder(c *gin.Context) {
	// TODO: 用 c.ShouldBindJSON 綁定 body 到 models.TicketOrder
	// TODO: 失敗時用 formatValidationError 回傳 400
	var input models.TicketOrder // 準備用來裝前端傳來的資料
	if err := c.ShouldBindJSON(&input); err != nil {
		status, body := formatValidationError(err)
		c.JSON(status, body)
		return
	}
	// TODO: 呼叫 validateEventForOrder(c, input.EventID, input.Quantity)
	//       若 err != nil → 回傳 400 然後 return
	if err := validateEventForOrder(c, input.EventID, input.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO: 從 events 表查出該活動的 price，計算 totalPrice
	query := "SELECT price FROM events WHERE id = $1"
	var price int

	err := database.DB.QueryRow(query, input.EventID).Scan(&price)
	if err != nil {
		respondError(c, err)
		return
	}
	input.TotalPrice = price * input.Quantity
	// TODO: 開 transaction（database.DB.Begin()），記得 defer tx.Rollback()

	// TODO: 用 tx.QueryRow INSERT 訂單，RETURNING 取回完整欄位，Scan 進 models.TicketOrder

	// TODO: 用 tx.Exec 扣庫存，庫存歸零時同時把 available 設為 false
	//       提示：可以用 CASE WHEN ... THEN ... ELSE ... END 來判斷

	// TODO: tx.Commit()，成功回傳 201 與訂單資料
	c.JSON(http.StatusCreated, gin.H{})
}

// CancelTicketOrder 將訂單狀態從 pending 改為 cancelled，並還原庫存（此 API 需帶 token）。
// 流程：
//  1. 驗證訂單可以取消（用 validateOrderCanCancel）
//  2. 查出 event_id 和 quantity（還原庫存用）
//  3. 開 transaction
//  4. UPDATE 訂單狀態為 cancelled，設定 cancelled_at
//  5. UPDATE events 還原庫存，把 available 設回 true
//  6. Commit
func CancelTicketOrder(c *gin.Context) {
	// TODO: 取出 id

	// TODO: 呼叫 validateOrderCanCancel(c, id)
	//       若 err != nil → 回傳 400 然後 return

	// TODO: 從 ticket_orders 查出 event_id 和 quantity（等等還原庫存要用）

	// TODO: 開 transaction，defer tx.Rollback()

	// TODO: 用 tx.QueryRow 更新訂單狀態為 cancelled，設定 cancelled_at=NOW()
	//       RETURNING 取回完整欄位，Scan 進 models.TicketOrder

	// TODO: 用 tx.Exec 還原庫存（stock + quantity），把 available 設回 true

	// TODO: tx.Commit()，成功回傳 200 與更新後的訂單
	c.JSON(http.StatusOK, gin.H{})
}

// ─── 內部驗證函式 ────────────────────────────────────────────

// validateEventForOrder 在建立訂單前確認：
//  1. 活動存在（若查不到 → errors.New("此活動不存在")）
//  2. available = true（若 false → errors.New("此活動已售完")）
//  3. stock >= quantity（若不足 → errors.New("票券庫存不足")）
//     全部通過回傳 nil
func validateEventForOrder(c *gin.Context, eventID, quantity int) error {
	// TODO: 從 events 表查出 available 和 stock
	// TODO: 依序檢查三個條件，不符合就回傳對應的 error
	// TODO: 全部通過 return nil
	_ = database.DB // 避免 import 報錯，實作後可移除
	_ = sql.ErrNoRows
	_ = errors.New
	return nil
}

// validateOrderCanCancel 確認訂單存在且可以取消。
//  1. 查 status（若查不到 → errors.New("訂單不存在")）
//  2. 若 status == "cancelled" → errors.New("此訂單已取消")
//     全部通過回傳 nil
func validateOrderCanCancel(c *gin.Context, orderID int) error {
	// TODO: 從 ticket_orders 查出 status
	// TODO: 檢查條件，不符合就回傳對應的 error
	// TODO: 全部通過 return nil
	return nil
}
