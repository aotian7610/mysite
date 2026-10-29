package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerUserInvoiceRoutes(group *gin.RouterGroup, h *handler.InvoiceHandler) {
	invoices := group.Group("/invoices")
	invoices.GET("", h.List)
	invoices.POST("", h.Create)
	invoices.GET("/:id", h.Get)
}
func registerAdminInvoiceRoutes(group *gin.RouterGroup, h *handler.InvoiceHandler) {
	invoices := group.Group("/invoices")
	invoices.GET("", h.AdminList)
	invoices.GET("/:id", h.AdminGet)
	invoices.POST("/:id/approve", h.Approve)
	invoices.POST("/:id/reject", h.Reject)
	invoices.POST("/:id/issue", h.Issue)
}
