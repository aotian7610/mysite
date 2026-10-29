import type { InvoiceTitle } from './api'

export const titleFields = [
  { key: 'company_name', max: 255, alwaysRequired: true },
  { key: 'tax_number', max: 100, alwaysRequired: true },
  { key: 'address', max: 500, alwaysRequired: false },
  { key: 'phone', max: 50, alwaysRequired: false },
  { key: 'bank_name', max: 255, alwaysRequired: false },
  { key: 'bank_account', max: 100, alwaysRequired: false }
] as const

export function validInvoiceAmount(raw: string): boolean {
  const amount = Number(raw)
  return /^[0-9]+(?:\.0+)?$/.test(raw.trim()) && Number.isSafeInteger(amount) && amount >= 200
}
export function formatInvoiceTax(value: string): string {
  const [whole, fraction = ''] = value.split('.')
  return `${whole}.${fraction.padEnd(2, '0')}`
}
export function validInvoiceTitle(title: InvoiceTitle): boolean {
  return ['NORMAL', 'SPECIAL'].includes(title.invoice_type) && titleFields.every(field => {
    const value = title[field.key].trim()
    return (!(field.alwaysRequired || title.invoice_type === 'SPECIAL') || value.length > 0) &&
      [...value].length <= field.max && !value.includes('\0')
  })
}
export function invoiceTaxPreview(raw: string): string {
  if (!validInvoiceAmount(raw)) return '—'
  // The integer CNY amount is also the number of cents in the 1% tax.
  const cents = BigInt(Number(raw))
  return `${cents / 100n}.${String(cents % 100n).padStart(2, '0')}`
}
