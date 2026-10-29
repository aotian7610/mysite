-- Fork-local invoice workflow. Taxes are handled manually; this migration and
-- the invoice workflow never update users.balance or payment_orders.
CREATE TABLE IF NOT EXISTS invoice_applications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    user_email VARCHAR(255) NOT NULL,
    invoice_type VARCHAR(20) NOT NULL CHECK (invoice_type IN ('NORMAL', 'SPECIAL')),
    company_name VARCHAR(255) NOT NULL CHECK (btrim(company_name) <> ''),
    tax_number VARCHAR(100) NOT NULL CHECK (btrim(tax_number) <> ''),
    address VARCHAR(500) NOT NULL DEFAULT '',
    phone VARCHAR(50) NOT NULL DEFAULT '',
    bank_name VARCHAR(255) NOT NULL DEFAULT '',
    bank_account VARCHAR(100) NOT NULL DEFAULT '',
    amount NUMERIC NOT NULL CHECK (amount >= 200 AND amount <= 9007199254740991 AND amount = trunc(amount)),
    tax_rate NUMERIC(5,4) NOT NULL,
    tax_amount NUMERIC(20,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'ISSUED')),
    reviewed_by BIGINT REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    issued_by BIGINT REFERENCES users(id),
    issued_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT invoice_special_fields CHECK (invoice_type <> 'SPECIAL' OR
        (btrim(address) <> '' AND btrim(phone) <> '' AND btrim(bank_name) <> '' AND btrim(bank_account) <> '')),
    CONSTRAINT invoice_tax_consistency CHECK (
        (invoice_type = 'NORMAL' AND tax_rate = 0 AND tax_amount = 0) OR
        (invoice_type = 'SPECIAL' AND tax_rate = 0.01 AND tax_amount = round(amount * 0.01, 2)))
);
CREATE INDEX IF NOT EXISTS idx_invoice_applications_user_created ON invoice_applications(user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_invoice_applications_created ON invoice_applications(created_at DESC, id DESC);
