import apiClient from '@/api/client'
import type { BasePaginationResponse } from '@/types'

export type InvoiceType = 'NORMAL' | 'SPECIAL'
export type InvoiceStatus = 'PENDING' | 'APPROVED' | 'REJECTED' | 'ISSUED'
export type InvoiceAction = 'approve' | 'reject' | 'issue'
export interface InvoiceTitle {
  invoice_type: InvoiceType
  company_name: string
  tax_number: string
  address: string
  phone: string
  bank_name: string
  bank_account: string
}
export interface InvoiceApplication extends InvoiceTitle {
  id: number
  user_id: number
  user_email: string
  amount: string
  tax_rate: string
  tax_amount: string
  status: InvoiceStatus
  reviewed_by: number | null
  reviewed_at: string | null
  issued_by: number | null
  issued_at: string | null
  created_at: string
  updated_at: string
}
export interface AdminInvoiceDetail extends InvoiceApplication {
  current_email: string
  current_balance: string
}
export const invoiceAPI = {
  async list(admin: boolean, page: number, pageSize: number) {
    return (await apiClient.get<BasePaginationResponse<InvoiceApplication>>(
      admin ? '/admin/invoices' : '/invoices', { params: { page, page_size: pageSize } }
    )).data
  },
  async create(input: InvoiceTitle & { amount: number }) {
    return (await apiClient.post<InvoiceApplication>('/invoices', input)).data
  },
  async get(id: number) {
    return (await apiClient.get<InvoiceApplication>(`/invoices/${id}`)).data
  },
  async adminGet(id: number) {
    return (await apiClient.get<AdminInvoiceDetail>(`/admin/invoices/${id}`)).data
  },
  async transition(id: number, action: InvoiceAction) {
    return (await apiClient.post<InvoiceApplication>(`/admin/invoices/${id}/${action}`)).data
  }
}
