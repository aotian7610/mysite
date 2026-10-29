package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct{ service *service.InvoiceService }

func NewInvoiceHandler(s *service.InvoiceService) *InvoiceHandler { return &InvoiceHandler{service: s} }
func invoiceID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid invoice ID")
		return 0, false
	}
	return id, true
}
func (h *InvoiceHandler) Create(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	var in service.CreateInvoiceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.ErrorFrom(c, service.ErrInvoiceFields)
		return
	}
	item, err := h.service.Create(c.Request.Context(), subject.UserID, in)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Created(c, item)
}
func (h *InvoiceHandler) list(c *gin.Context, userID int64) {
	page, size := response.ParsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), userID, pagination.PaginationParams{Page: page, PageSize: size})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Paginated(c, items, total, page, size)
}
func (h *InvoiceHandler) List(c *gin.Context) {
	subject, ok := requireAuth(c)
	if ok {
		h.list(c, subject.UserID)
	}
}
func (h *InvoiceHandler) AdminList(c *gin.Context) { h.list(c, 0) }
func (h *InvoiceHandler) Get(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, ok := invoiceID(c)
	if !ok {
		return
	}
	item, err := h.service.Get(c.Request.Context(), id, subject.UserID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, item)
}
func (h *InvoiceHandler) AdminGet(c *gin.Context) {
	id, ok := invoiceID(c)
	if !ok {
		return
	}
	item, err := h.service.AdminDetail(c.Request.Context(), id)
	if response.ErrorFrom(c, err) {
		return
	}
	c.Header("Cache-Control", "no-store")
	response.Success(c, item)
}
func (h *InvoiceHandler) transition(c *gin.Context, target string) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, ok := invoiceID(c)
	if !ok {
		return
	}
	item, err := h.service.Transition(c.Request.Context(), id, subject.UserID, target)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, item)
}
func (h *InvoiceHandler) Approve(c *gin.Context) { h.transition(c, "APPROVED") }
func (h *InvoiceHandler) Reject(c *gin.Context)  { h.transition(c, "REJECTED") }
func (h *InvoiceHandler) Issue(c *gin.Context)   { h.transition(c, "ISSUED") }
