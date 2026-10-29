package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestInvoiceAuditOmitsTitleButPreservesHandlerBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &auditCaptureRepository{}
	svc := service.NewAuditLogService(repo, nil)
	svc.Start()
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(string(ContextKeyUser), AuthSubject{UserID: 1}); c.Next() })
	router.Use(gin.HandlerFunc(NewAuditLogMiddleware(svc)))
	router.POST("/api/v1/invoices", func(c *gin.Context) {
		var body map[string]string
		require.NoError(t, c.ShouldBindJSON(&body))
		require.Equal(t, "000123-private", body["bank_account"])
		c.Status(http.StatusCreated)
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoices", bytes.NewBufferString(`{"bank_account":"000123-private","tax_number":"tax-private"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusCreated, recorder.Code)
	svc.Stop()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	require.Len(t, repo.logs, 1)
	require.Equal(t, "<credential-bearing body omitted>", repo.logs[0].RequestBody)
}
