package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

type invoiceHandlerRepo struct {
	service.InvoiceRepository
	owner int64
	item  *service.InvoiceApplication
	actor int64
}

func (r *invoiceHandlerRepo) Create(_ context.Context, i *service.InvoiceApplication) error {
	r.item = i
	return nil
}
func (r *invoiceHandlerRepo) Get(_ context.Context, id, userID int64) (*service.InvoiceApplication, error) {
	r.owner = userID
	if id != 1 || userID != 7 {
		return nil, service.ErrInvoiceNotFound
	}
	return &service.InvoiceApplication{ID: 1, UserID: 7, Status: "PENDING"}, nil
}
func (r *invoiceHandlerRepo) List(_ context.Context, userID int64, _ pagination.PaginationParams) ([]service.InvoiceApplication, int64, error) {
	r.owner = userID
	return []service.InvoiceApplication{}, 0, nil
}
func (r *invoiceHandlerRepo) Transition(_ context.Context, id, actor int64, status string) (*service.InvoiceApplication, error) {
	r.actor = actor
	return &service.InvoiceApplication{ID: id, Status: status}, nil
}
func TestInvoiceHandlerUsesAuthenticatedOwnerAndServerValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &invoiceHandlerRepo{}
	h := NewInvoiceHandler(service.NewInvoiceService(repo))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
		c.Next()
	})
	router.POST("/invoices", h.Create)
	router.GET("/invoices", h.List)
	router.GET("/invoices/:id", h.Get)
	req := httptest.NewRequest("POST", "/invoices", strings.NewReader(`{"user_id":99,"amount":200,"invoice_type":"NORMAL","company_name":"C","tax_number":"0001","status":"ISSUED","tax_amount":99}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusCreated, res.Code)
	require.EqualValues(t, 7, repo.item.UserID)
	require.Equal(t, "PENDING", repo.item.Status)
	require.True(t, repo.item.TaxAmount.Equal(decimal.Zero))
	for _, tc := range []struct {
		path   string
		status int
	}{{"/invoices?user_id=99", 200}, {"/invoices/1", 200}, {"/invoices/2", 404}, {"/invoices/invalid", 400}} {
		res = httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest("GET", tc.path, nil))
		require.Equal(t, tc.status, res.Code)
		require.NotContains(t, res.Body.String(), "current_balance")
	}
	require.EqualValues(t, 7, repo.owner)
	unauthorized := gin.New()
	unauthorized.POST("/invoices", h.Create)
	res = httptest.NewRecorder()
	unauthorized.ServeHTTP(res, httptest.NewRequest("POST", "/invoices", nil))
	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestInvoiceHandlerReviewHasNoBalanceDependency(t *testing.T) {
	// The embedded repository panics for any unimplemented method, including
	// AdminDetail. Approval must not call it to check the user's balance.
	repo := &invoiceHandlerRepo{}
	h := NewInvoiceHandler(service.NewInvoiceService(repo))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		c.Next()
	})
	router.POST("/invoices/:id/approve", h.Approve)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httptest.NewRequest("POST", "/invoices/1/approve", nil))
	require.Equal(t, http.StatusOK, res.Code)
	require.EqualValues(t, 42, repo.actor)
	require.Contains(t, res.Body.String(), "APPROVED")
}
