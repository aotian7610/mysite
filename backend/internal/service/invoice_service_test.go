package service

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInvoiceValidation(t *testing.T) {
	for _, tc := range []struct {
		amount string
		valid  bool
	}{
		{"200", true}, {"201", true}, {"1000", true}, {"200.0", true},
		{"199", false}, {"0", false}, {"-200", false}, {"200.01", false}, {"", false},
		{"9007199254740992", false}, {"1e999999999", false},
	} {
		t.Run(tc.amount, func(t *testing.T) {
			input := CreateInvoiceInput{InvoiceTitle: InvoiceTitle{InvoiceType: "NORMAL", CompanyName: " Company ", TaxNumber: "00123"}, Amount: json.Number(tc.amount)}
			invoice, err := validateInvoice(input)
			if !tc.valid {
				require.ErrorIs(t, err, ErrInvoiceAmount)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "Company", invoice.CompanyName)
			require.Equal(t, "00123", invoice.TaxNumber)
			require.True(t, invoice.TaxAmount.IsZero())
		})
	}
}

func TestInvoiceSpecialRequiredFieldsAndTax(t *testing.T) {
	input := CreateInvoiceInput{InvoiceTitle: InvoiceTitle{InvoiceType: "SPECIAL", CompanyName: "Company", TaxNumber: "001", Address: "Address", Phone: "00123", BankName: "Bank", BankAccount: "000123"}, Amount: "201"}
	invoice, err := validateInvoice(input)
	require.NoError(t, err)
	require.Equal(t, "2.01", invoice.TaxAmount.StringFixed(2))
	for _, field := range []string{"address", "phone", "bank", "account"} {
		t.Run(field, func(t *testing.T) {
			in := input
			switch field {
			case "address":
				in.Address = "  "
			case "phone":
				in.Phone = ""
			case "bank":
				in.BankName = ""
			case "account":
				in.BankAccount = ""
			}
			_, err := validateInvoice(in)
			require.ErrorIs(t, err, ErrInvoiceFields)
			in.InvoiceType = "NORMAL"
			_, err = validateInvoice(in)
			require.NoError(t, err)
		})
	}
}
