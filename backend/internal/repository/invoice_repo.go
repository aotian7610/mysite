package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type invoiceRepository struct{ db *sql.DB }

func NewInvoiceRepository(db *sql.DB) service.InvoiceRepository { return &invoiceRepository{db: db} }

const invoiceColumns = `id,user_id,user_email,invoice_type,company_name,tax_number,address,phone,bank_name,bank_account,amount,tax_rate,tax_amount,status,reviewed_by,reviewed_at,issued_by,issued_at,created_at,updated_at`

func scanInvoice(row interface{ Scan(...any) error }) (*service.InvoiceApplication, error) {
	var i service.InvoiceApplication
	err := row.Scan(&i.ID, &i.UserID, &i.UserEmail, &i.InvoiceType, &i.CompanyName, &i.TaxNumber, &i.Address, &i.Phone, &i.BankName, &i.BankAccount, &i.Amount, &i.TaxRate, &i.TaxAmount, &i.Status, &i.ReviewedBy, &i.ReviewedAt, &i.IssuedBy, &i.IssuedAt, &i.CreatedAt, &i.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrInvoiceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *invoiceRepository) Create(ctx context.Context, i *service.InvoiceApplication) error {
	return r.db.QueryRowContext(ctx, `INSERT INTO invoice_applications
        (user_id,user_email,invoice_type,company_name,tax_number,address,phone,bank_name,bank_account,amount,tax_rate,tax_amount)
        SELECT id,email,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11 FROM users WHERE id=$1 AND deleted_at IS NULL
        RETURNING id,user_email,created_at,updated_at`, i.UserID, i.InvoiceType, i.CompanyName, i.TaxNumber, i.Address, i.Phone, i.BankName, i.BankAccount, i.Amount, i.TaxRate, i.TaxAmount).
		Scan(&i.ID, &i.UserEmail, &i.CreatedAt, &i.UpdatedAt)
}

func (r *invoiceRepository) List(ctx context.Context, userID int64, p pagination.PaginationParams) ([]service.InvoiceApplication, int64, error) {
	// Filter before pagination; userID=0 is used only by the authenticated admin route.
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM invoice_applications WHERE ($1::bigint=0 OR user_id=$1)`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+invoiceColumns+` FROM invoice_applications WHERE ($1::bigint=0 OR user_id=$1) ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3`, userID, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.InvoiceApplication, 0)
	for rows.Next() {
		item, err := scanInvoice(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (r *invoiceRepository) Get(ctx context.Context, id, userID int64) (*service.InvoiceApplication, error) {
	return scanInvoice(r.db.QueryRowContext(ctx, `SELECT `+invoiceColumns+` FROM invoice_applications WHERE id=$1 AND ($2::bigint=0 OR user_id=$2)`, id, userID))
}
func (r *invoiceRepository) AdminDetail(ctx context.Context, id int64) (*service.AdminInvoiceDetail, error) {
	item, err := r.Get(ctx, id, 0)
	if err != nil {
		return nil, err
	}
	detail := &service.AdminInvoiceDetail{InvoiceApplication: item}
	err = r.db.QueryRowContext(ctx, `SELECT email,balance FROM users WHERE id=$1`, item.UserID).Scan(&detail.CurrentEmail, &detail.CurrentBalance)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (r *invoiceRepository) Transition(ctx context.Context, id, actorID int64, target string) (*service.InvoiceApplication, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	item, err := scanInvoice(tx.QueryRowContext(ctx, `SELECT `+invoiceColumns+` FROM invoice_applications WHERE id=$1 FOR UPDATE`, id))
	if err != nil {
		return nil, err
	}
	if item.Status == target {
		return item, tx.Commit()
	}
	from := "PENDING"
	if target == "ISSUED" {
		from = "APPROVED"
	}
	if item.Status != from || (target != "APPROVED" && target != "REJECTED" && target != "ISSUED") {
		return nil, service.ErrInvoiceState
	}
	update := `UPDATE invoice_applications SET status=$2,reviewed_by=$3,reviewed_at=NOW(),updated_at=NOW() WHERE id=$1 RETURNING `
	if target == "ISSUED" {
		update = `UPDATE invoice_applications SET status=$2,issued_by=$3,issued_at=NOW(),updated_at=NOW() WHERE id=$1 RETURNING `
	}
	item, err = scanInvoice(tx.QueryRowContext(ctx, update+invoiceColumns, id, target, actorID))
	if err != nil {
		return nil, err
	}
	// No balance read/write: approval records the administrator's manual decision only.
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}
