package routes

import (
	"net/http"

	"go-api-practice-10/handlers"
	"go-api-practice-10/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(middleware.Logger())

	// 公開：活動列表（可加 ?available=true|false）、單筆、訂單列表、單筆訂單
	api.GET("/events", handlers.GetEvents)
	api.GET("/events/:id", handlers.GetEventByID)
	api.GET("/ticket-orders", handlers.GetTicketOrders)
	api.GET("/ticket-orders/:id", handlers.GetTicketOrderByID)

	// 需 token：活動管理、建立訂單、取消訂單
	protected := api.Group("").Use(middleware.TokenAuth())
	{
		protected.POST("/events", handlers.CreateEvent)
		protected.PUT("/events/:id", handlers.UpdateEvent)
		protected.DELETE("/events/:id", handlers.DeleteEvent)
		protected.POST("/ticket-orders", handlers.CreateTicketOrder)
		protected.POST("/ticket-orders/:id/cancel", handlers.CancelTicketOrder)
	}

	// 搶票模擬（需 token）— 比較有鎖 vs 沒鎖的差異
	rush := api.Group("/rush").Use(middleware.TokenAuth())
	{
		rush.POST("/without-lock", handlers.RushWithoutLock)
		rush.POST("/with-lock", handlers.RushWithLock)
	}

	api.GET("/me", middleware.TokenAuth(), func(c *gin.Context) {
		token, _ := c.Get("token")
		c.JSON(http.StatusOK, gin.H{
			"message": "你有帶正確的 token！",
			"token":   token,
		})
	})
}
