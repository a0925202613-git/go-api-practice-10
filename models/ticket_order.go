package models

import "time"

type TicketOrder struct {
	ID           int        `json:"id"`
	EventID      int        `json:"event_id"      binding:"required,gte=1"`
	CustomerName string     `json:"customer_name"  binding:"required,max=255"`
	Quantity     int        `json:"quantity"       binding:"required,gte=1,lte=4"`
	TotalPrice   int        `json:"total_price"`
	Status       string     `json:"status"`
	OrderedAt    time.Time  `json:"ordered_at"`
	CancelledAt  *time.Time `json:"cancelled_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TicketOrderWithEvent 是查詢訂單時附帶活動名稱的回應結構（JOIN 用）
type TicketOrderWithEvent struct {
	TicketOrder
	EventName string `json:"event_name"`
}
