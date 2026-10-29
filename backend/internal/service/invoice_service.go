package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/shopspring/decimal"
)

var (
	ErrInvoiceNotFound = infraerrors.NotFound("INVOICE_NOT_FOUND", "Invoice application not found")
	ErrInvoiceState    = infraerrors.Conflict("INVOICE_STATE_CONFLICT", "Invoice status has changed; refresh and try again")
	ErrInvoiceAmount   = infraerrors.BadRequest("INVOICE_INVALID_AMOUNT", "Invoice amount must be an integer of at least 200")
	ErrInvoiceFields   = infraerrors.BadRequest("INVOICE_INVALID_FIELDS", "Please complete the required invoice fields within their length limits")
)

type InvoiceTitle struct {
	InvoiceType string `json:"invoice_type"`
	CompanyName string `json:"company_name"`
	TaxNumber   string `json:"tax_number"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	BankName    string `json:"bank_name"`
	BankAccount string `json:"bank_account"`
}

type CreateInvoiceInput struct {
	InvoiceTitle
	Amount json.Number `json:"amount"`
}

type InvoiceApplication struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	UserEmail string `json:"user_email"`
	InvoiceTitle
	Amount     decimal.Decimal `json:"amount"`
	TaxRate    decimal.Decimal `json:"tax_rate"`
	TaxAmount  decimal.Decimal `json:"tax_amount"`
	Status     string          `json:"status"`
	ReviewedBy *int64          `json:"reviewed_by"`
	ReviewedAt *time.Time      `json:"reviewed_at"`
	IssuedBy   *int64          `json:"issued_by"`
	IssuedAt   *time.Time      `json:"issued_at"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// Admin-only live information is deliberately absent from the user DTO.
type AdminInvoiceDetail struct {
	*InvoiceApplication
	CurrentEmail   string          `json:"current_email"`
	CurrentBalance decimal.Decimal `json:"current_balance"`
}

type InvoiceRepository interface {
	Create(context.Context, *InvoiceApplication) error
	List(context.Context, int64, pagination.PaginationParams) ([]InvoiceApplication, int64, error)
	Get(context.Context, int64, int64) (*InvoiceApplication, error)
	AdminDetail(context.Context, int64) (*AdminInvoiceDetail, error)
	Transition(context.Context, int64, int64, string) (*InvoiceApplication, error)
}

// There is intentionally no balance writer, billing service or payment dependency.
type InvoiceService struct{ repo InvoiceRepository }

func NewInvoiceService(repo InvoiceRepository) *InvoiceService { return &InvoiceService{repo: repo} }

var invoiceAmountPattern = regexp.MustCompile(`^[0-9]+(?:\.0+)?$`)

func validateInvoice(in CreateInvoiceInput) (*InvoiceApplication, error) {
	if len(in.Amount.String()) > 32 || !invoiceAmountPattern.MatchString(in.Amount.String()) {
		return nil, ErrInvoiceAmount
	}
	amount, err := decimal.NewFromString(in.Amount.String())
	if err != nil || amount.LessThan(decimal.NewFromInt(200)) || amount.GreaterThan(decimal.NewFromInt(9007199254740991)) || !amount.Equal(amount.Truncate(0)) {
		return nil, ErrInvoiceAmount
	}
	title := in.InvoiceTitle
	if title.InvoiceType != "NORMAL" && title.InvoiceType != "SPECIAL" {
		return nil, ErrInvoiceFields
	}
	fields := []struct {
		value    *string
		max      int
		required bool
	}{
		{&title.CompanyName, 255, true}, {&title.TaxNumber, 100, true},
		{&title.Address, 500, title.InvoiceType == "SPECIAL"}, {&title.Phone, 50, title.InvoiceType == "SPECIAL"},
		{&title.BankName, 255, title.InvoiceType == "SPECIAL"}, {&title.BankAccount, 100, title.InvoiceType == "SPECIAL"},
	}
	for _, f := range fields {
		*f.value = strings.TrimSpace(*f.value)
		if (f.required && *f.value == "") || utf8.RuneCountInString(*f.value) > f.max || strings.ContainsRune(*f.value, 0) {
			return nil, ErrInvoiceFields
		}
	}
	rate := decimal.Zero
	if title.InvoiceType == "SPECIAL" {
		rate = decimal.New(1, -2)
	}
	return &InvoiceApplication{InvoiceTitle: title, Amount: amount, TaxRate: rate, TaxAmount: amount.Mul(rate).Round(2), Status: "PENDING"}, nil
}

func (s *InvoiceService) Create(ctx context.Context, userID int64, in CreateInvoiceInput) (*InvoiceApplication, error) {
	item, err := validateInvoice(in)
	if err != nil {
		return nil, err
	}
	item.UserID = userID
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}
func (s *InvoiceService) List(ctx context.Context, userID int64, p pagination.PaginationParams) ([]InvoiceApplication, int64, error) {
	return s.repo.List(ctx, userID, p)
}
func (s *InvoiceService) Get(ctx context.Context, id, userID int64) (*InvoiceApplication, error) {
	return s.repo.Get(ctx, id, userID)
}
func (s *InvoiceService) AdminDetail(ctx context.Context, id int64) (*AdminInvoiceDetail, error) {
	return s.repo.AdminDetail(ctx, id)
}
func (s *InvoiceService) Transition(ctx context.Context, id, actorID int64, status string) (*InvoiceApplication, error) {
	if status != "APPROVED" && status != "REJECTED" && status != "ISSUED" {
		return nil, ErrInvoiceState
	}
	return s.repo.Transition(ctx, id, actorID, status)
}
