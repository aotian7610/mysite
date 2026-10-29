//go:build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestInvoicePostgresFullMigrationChain(t *testing.T) {
	dsn := os.Getenv("INVOICE_TEST_DSN")
	if dsn == "" {
		t.Skip("INVOICE_TEST_DSN required")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	var name string
	require.NoError(t, db.QueryRow(`SELECT current_database()`).Scan(&name))
	require.True(t, strings.HasSuffix(name, "_test"), "full migration test requires a dedicated test database")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))
	var count int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM schema_migrations WHERE filename='249_mysite_invoice_applications.sql'`).Scan(&count))
	require.Equal(t, 1, count)
}

func invoiceTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("INVOICE_TEST_DSN")
	if dsn == "" {
		t.Skip("INVOICE_TEST_DSN must name an isolated test database")
	}
	base, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := fmt.Sprintf("invoice_test_%d", time.Now().UnixNano())
	_, err = base.Exec(`CREATE SCHEMA ` + schema)
	require.NoError(t, err)
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close(); _, _ = base.Exec(`DROP SCHEMA ` + schema + ` CASCADE`); _ = base.Close() })
	_, err = db.Exec(`CREATE TABLE users(id BIGINT PRIMARY KEY,email TEXT NOT NULL,balance NUMERIC(20,8) NOT NULL,deleted_at TIMESTAMPTZ);
        INSERT INTO users VALUES(1,'user@example.com',-10,NULL),(2,'other@example.com',0,NULL),(3,'admin@example.com',100,NULL)`)
	require.NoError(t, err)
	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "249_mysite_invoice_applications.sql"))
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	// Migration can also be applied to an already initialized schema.
	_, err = db.Exec(string(migration))
	require.NoError(t, err)
	return db
}

func TestInvoicePostgresWorkflowAndIsolation(t *testing.T) {
	db := invoiceTestDB(t)
	repo := NewInvoiceRepository(db)
	svc := service.NewInvoiceService(repo)
	ctx := context.Background()
	input := service.CreateInvoiceInput{InvoiceTitle: service.InvoiceTitle{InvoiceType: "SPECIAL", CompanyName: "Company", TaxNumber: "001", Address: "A", Phone: "001", BankName: "B", BankAccount: "0001"}, Amount: "200"}
	first, err := svc.Create(ctx, 1, input)
	require.NoError(t, err)
	other, err := svc.Create(ctx, 2, input)
	require.NoError(t, err)
	second, err := svc.Create(ctx, 1, input)
	require.NoError(t, err)
	_, err = repo.Get(ctx, other.ID, 1)
	require.ErrorIs(t, err, service.ErrInvoiceNotFound)
	list, total, err := repo.List(ctx, 1, pagination.PaginationParams{Page: 1, PageSize: 1})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, list, 1)
	require.Equal(t, second.ID, list[0].ID)
	_, err = repo.Transition(ctx, first.ID, 3, "ISSUED")
	require.ErrorIs(t, err, service.ErrInvoiceState)
	for _, balance := range []string{"100", "2", "1", "0", "-10"} {
		_, err = db.Exec(`UPDATE users SET balance=$1 WHERE id=1`, balance)
		require.NoError(t, err)
		item, err := svc.Create(ctx, 1, input)
		require.NoError(t, err)
		detail, err := repo.AdminDetail(ctx, item.ID)
		require.NoError(t, err)
		require.Equal(t, balance, detail.CurrentBalance.String())
		approved, err := repo.Transition(ctx, item.ID, 3, "APPROVED")
		require.NoError(t, err)
		require.NotNil(t, approved.ReviewedAt)
		retry, err := repo.Transition(ctx, item.ID, 3, "APPROVED")
		require.NoError(t, err)
		require.Equal(t, approved.ReviewedAt, retry.ReviewedAt)
		_, err = repo.Transition(ctx, item.ID, 3, "REJECTED")
		require.ErrorIs(t, err, service.ErrInvoiceState)
		issued, err := repo.Transition(ctx, item.ID, 3, "ISSUED")
		require.NoError(t, err)
		require.NotNil(t, issued.IssuedAt)
		var after decimal.Decimal
		require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&after))
		require.Equal(t, balance, after.String())
	}
	_, err = repo.Transition(ctx, first.ID, 3, "REJECTED")
	require.NoError(t, err)
	_, err = repo.Transition(ctx, first.ID, 3, "APPROVED")
	require.ErrorIs(t, err, service.ErrInvoiceState)
	normal := input
	normal.InvoiceType = "NORMAL"
	normal.Address = ""
	normal.Phone = ""
	normal.BankName = ""
	normal.BankAccount = ""
	item, err := svc.Create(ctx, 1, normal)
	require.NoError(t, err)
	require.True(t, item.TaxAmount.IsZero())
	for _, amount := range []string{"199", "200.01"} {
		_, err = db.Exec(`UPDATE invoice_applications SET amount=$1 WHERE id=$2`, amount, item.ID)
		require.Error(t, err)
	}
	_, err = db.Exec(`UPDATE invoice_applications SET invoice_type='SPECIAL',tax_rate=0.01,tax_amount=2 WHERE id=$1`, item.ID)
	require.Error(t, err)
	_, err = db.Exec(`UPDATE users SET email='updated@example.com',balance=42 WHERE id=1`)
	require.NoError(t, err)
	detail, err := repo.AdminDetail(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, "updated@example.com", detail.CurrentEmail)
	require.Equal(t, "user@example.com", detail.UserEmail)
	require.Equal(t, "42", detail.CurrentBalance.String())
}

func TestInvoicePostgresConcurrentReview(t *testing.T) {
	db := invoiceTestDB(t)
	repo := NewInvoiceRepository(db)
	svc := service.NewInvoiceService(repo)
	ctx := context.Background()
	input := service.CreateInvoiceInput{InvoiceTitle: service.InvoiceTitle{InvoiceType: "NORMAL", CompanyName: "C", TaxNumber: "001"}, Amount: "200"}
	item, err := svc.Create(ctx, 1, input)
	require.NoError(t, err)
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, target := range []string{"APPROVED", "REJECTED"} {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			<-start
			_, err := repo.Transition(ctx, item.ID, 3, target)
			errs <- err
		}(target)
	}
	close(start)
	wg.Wait()
	close(errs)
	successes, conflicts := 0, 0
	for err := range errs {
		if err == nil {
			successes++
		} else {
			require.ErrorIs(t, err, service.ErrInvoiceState)
			conflicts++
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)
	var balance decimal.Decimal
	require.NoError(t, db.QueryRow(`SELECT balance FROM users WHERE id=1`).Scan(&balance))
	require.Equal(t, "-10", balance.String())
}
