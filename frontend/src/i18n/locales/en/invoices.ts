export default {
  title: 'Invoice Management', apply: 'Request an Invoice', submit: 'Submit Request', refresh: 'Refresh', retry: 'Retry',
  sequence: 'No.', applicant: 'Applicant (Account)', created_at: 'Requested At', company_name: 'Invoice Title (Company Name)',
  invoice_type: 'Invoice Type', tax_number: 'Tax ID', address: 'Address', phone: 'Phone',
  bank_name: 'Bank Name', bank_account: 'Bank Account', amount: 'Invoice Amount', status: 'Status', actions: 'Actions',
  tax_amount: 'Estimated Tax Deduction', current_balance: 'Current Account Balance', current_email: 'Registered Email',
  reviewed_at: 'Reviewed At', issued_at: 'Issued At', optional: 'Optional', details: 'Request Details',
  approve: 'Approve', reject: 'Reject', issue: 'Mark as Issued', confirmAction: 'Confirm', balanceUnits: 'balance credits',
  amountHint: 'Minimum invoice amount: CNY 200. Whole numbers only.',
  notice: 'The invoice amount must be at least CNY 200. Special VAT invoices incur an additional 1% tax charge. Please ensure your account balance covers the charge to avoid review failure. Within 5 business days after approval, an administrator will send the invoice to your registered email. Please check your inbox.',
  noConversion: 'The charge is 1% of the invoice amount, deducted as the same number of balance credits without currency conversion. An administrator handles the deduction manually.',
  manualDeduction: 'Please deduct the tax charge manually. This action only changes the review status and does not deduct any balance.',
  failed: 'The operation failed. Please try again.',
  types: { NORMAL: 'Standard VAT Invoice', SPECIAL: 'Special VAT Invoice' },
  statuses: { PENDING: 'Pending Review', APPROVED: 'Approved', REJECTED: 'Rejected', ISSUED: 'Issued' },
  confirm: {
    approve: 'Approve this invoice request? Complete your manual checks first.',
    reject: 'Reject this request? The user can submit a new request afterwards.',
    issue: 'Mark this invoice as issued? Send it to the registered email manually. This action does not send email.'
  },
  errors: {
    INVOICE_INVALID_AMOUNT: 'The amount must be a whole number between 200 and 9007199254740991.',
    INVOICE_INVALID_FIELDS: 'Complete the required invoice fields and check their length limits.',
    INVOICE_NOT_FOUND: 'The invoice request was not found or is not accessible.',
    INVOICE_STATE_CONFLICT: 'The request status has changed. Refresh and try again.'
  }
}
