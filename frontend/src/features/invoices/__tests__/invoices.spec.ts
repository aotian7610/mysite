import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent } from 'vue'
import InvoiceApplicationDialog from '../InvoiceApplicationDialog.vue'
import InvoiceManagementView from '../InvoiceManagementView.vue'
import { invoiceTaxPreview, validInvoiceAmount, formatInvoiceTax } from '../form'
import zh from '@/i18n/locales/zh'

const api = vi.hoisted(() => ({ list: vi.fn(), create: vi.fn(), get: vi.fn(), adminGet: vi.fn(), transition: vi.fn() }))
vi.mock('../api', () => ({ invoiceAPI: api }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages = (await import('@/i18n/locales/zh')).default
  const lookup = (key: string) => key.split('.').reduce<any>((value, part) => value?.[part], messages)
  return { ...actual, useI18n: () => ({ t: (key: string) => lookup(key) ?? key, te: (key: string) => typeof lookup(key) === 'string' }) }
})
const Dialog = defineComponent({ props: ['show', 'title'], template: '<section v-if="show" :data-title="title"><slot /><slot name="footer" /></section>' })
const Table = defineComponent({ props: ['data', 'columns'], template: '<div><div v-for="row in data" :key="row.id"><slot name="cell-company_name" :row="row" /><slot name="cell-actions" :row="row" /></div></div>' })
function options() {
  return { global: {
    stubs: { BaseDialog: Dialog, AppLayout: { template: '<main><slot /></main>' }, DataTable: Table, Pagination: true }
  } }
}
const item = {
  id: 1, user_id: 1, user_email: 'user@example.com', current_email: 'user@example.com', current_balance: '-10',
  invoice_type: 'SPECIAL', company_name: 'Company', tax_number: '001', address: 'A', phone: '001', bank_name: 'B', bank_account: '0001',
  amount: '200', tax_rate: '0.01', tax_amount: '2.00', status: 'PENDING',
  reviewed_by: null, reviewed_at: null, issued_by: null, issued_at: null, created_at: '2026-10-29T10:00:00Z', updated_at: '2026-10-29T10:00:00Z'
}
beforeEach(() => {
  vi.resetAllMocks()
  api.list.mockResolvedValue({ items: [item], total: 1, page: 1, page_size: 20, pages: 1 })
  api.adminGet.mockResolvedValue({ ...item })
  api.create.mockResolvedValue(item)
  api.transition.mockResolvedValue({ ...item, status: 'APPROVED' })
})

describe('invoice amount', () => {
  it.each(['200', '201', '1000'])('accepts %s', amount => expect(validInvoiceAmount(amount)).toBe(true))
  it.each(['199', '0', '-200', '200.01', '200.000000000000000001', '', 'Infinity', '9007199254740992'])('rejects %s', amount => expect(validInvoiceAmount(amount)).toBe(false))
  it('computes tax without floating point rounding', () => {
    expect(invoiceTaxPreview('200')).toBe('2.00')
    expect(invoiceTaxPreview('201')).toBe('2.01')
    expect(invoiceTaxPreview('1000')).toBe('10.00')
    expect(formatInvoiceTax('2')).toBe('2.00')
  })
})

describe('invoice request dialog', () => {
  it('submits a normal invoice with all four optional fields empty and preserves leading zeros', async () => {
    const wrapper = mount(InvoiceApplicationDialog, options())
    expect(wrapper.text()).toContain(zh.invoices.notice)
    await wrapper.get('#invoice-company_name').setValue('Company')
    await wrapper.get('#invoice-tax_number').setValue('000123')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(api.create).toHaveBeenCalledWith(expect.objectContaining({ amount: 200, tax_number: '000123', address: '', phone: '', bank_name: '', bank_account: '' }))
    expect(wrapper.emitted('created')).toHaveLength(1)
  })
  it('requires special fields and clears required errors without losing values on switching back', async () => {
    const wrapper = mount(InvoiceApplicationDialog, options())
    await wrapper.get('#invoice-company_name').setValue('Company')
    await wrapper.get('#invoice-tax_number').setValue('001')
    await wrapper.get('#invoice-type').setValue('SPECIAL')
    expect(wrapper.get('#invoice-address').attributes()).toHaveProperty('required')
    await wrapper.get('#invoice-bank_account').setValue('000123')
    await wrapper.get('form').trigger('submit')
    expect(api.create).not.toHaveBeenCalled()
    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    await wrapper.get('#invoice-type').setValue('NORMAL')
    expect(wrapper.get('#invoice-address').attributes()).not.toHaveProperty('required')
    expect(wrapper.find('[role="alert"]').exists()).toBe(false)
    expect((wrapper.get('#invoice-bank_account').element as HTMLInputElement).value).toBe('000123')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(api.create).toHaveBeenCalledTimes(1)
  })
  it('rejects fractional amounts even if browser validation is bypassed', async () => {
    const wrapper = mount(InvoiceApplicationDialog, options())
    await wrapper.get('#invoice-amount').setValue('200.01')
    await wrapper.get('form').trigger('submit')
    expect(api.create).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain(zh.invoices.errors.INVOICE_INVALID_AMOUNT)
  })
})

describe('invoice management', () => {
  it('loads a fresh balance for each approval dialog and permits approval with a negative balance', async () => {
    const wrapper = mount(InvoiceManagementView, { ...options(), props: { admin: true } })
    await flushPromises()
    const approve = () => wrapper.findAll('button').find(b => b.text() === '通过')!
    await approve().trigger('click'); await flushPromises()
    expect(api.adminGet).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('-10')
    expect(wrapper.text()).toContain('2.00')
    expect(wrapper.text()).toContain(zh.invoices.manualDeduction)
    await wrapper.findAll('button').find(b => b.text() === '取消')!.trigger('click')
    api.adminGet.mockResolvedValueOnce({ ...item, current_balance: '1.25' })
    await approve().trigger('click'); await flushPromises()
    expect(api.adminGet).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('1.25')
    const confirm = wrapper.findAll('button').find(b => b.text() === '确认')!
    expect(confirm.attributes('disabled')).toBeUndefined()
    await confirm.trigger('click'); await flushPromises()
    expect(api.transition).toHaveBeenCalledWith(1, 'approve')
    expect(api.list).toHaveBeenCalledTimes(2)
  })
  it('requires successful detail loading, offers retry and rejects stale transitions', async () => {
    api.adminGet.mockRejectedValueOnce(new Error('offline'))
    const wrapper = mount(InvoiceManagementView, { ...options(), props: { admin: true } })
    await flushPromises()
    await wrapper.findAll('button').find(b => b.text() === '通过')!.trigger('click'); await flushPromises()
    expect(wrapper.findAll('button').find(b => b.text() === '确认')!.attributes()).toHaveProperty('disabled')
    api.adminGet.mockResolvedValueOnce({ ...item, status: 'REJECTED' })
    await wrapper.findAll('button').find(b => b.text() === '重试')!.trigger('click'); await flushPromises()
    expect(wrapper.text()).toContain(zh.invoices.errors.INVOICE_STATE_CONFLICT)
    expect(api.transition).not.toHaveBeenCalled()
  })
  it('resets administrator data when switching to the user route', async () => {
    const wrapper = mount(InvoiceManagementView, { ...options(), props: { admin: true } })
    await flushPromises()
    api.list.mockResolvedValueOnce({ items: [], total: 0 })
    await wrapper.setProps({ admin: false }); await flushPromises()
    expect(api.list).toHaveBeenLastCalledWith(false, 1, 20)
    expect(wrapper.text()).toContain('发起开票申请')
    expect(wrapper.text()).not.toContain('Company')
  })
})
