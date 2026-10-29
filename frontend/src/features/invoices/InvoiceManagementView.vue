<template>
  <AppLayout>
    <div class="space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <h1 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('invoices.title') }}</h1>
        <div class="flex gap-2">
          <button class="btn btn-secondary" :disabled="loading" @click="fetchList">{{ t('invoices.refresh') }}</button>
          <button v-if="!admin" class="btn btn-primary" @click="showApply = true">{{ t('invoices.apply') }}</button>
        </div>
      </div>
      <p v-if="listError" role="alert" class="text-sm text-red-600 dark:text-red-400">{{ listError }}</p>
      <div class="card overflow-hidden">
        <DataTable :columns="columns" :data="rows" :loading="loading">
          <template #cell-company_name="{ row }"><button class="text-primary-600 hover:underline dark:text-primary-400" @click="openDetail(row.id)">{{ row.company_name }}</button></template>
          <template #cell-user_email="{ row }"><button class="text-primary-600 hover:underline dark:text-primary-400" @click="openDetail(row.id)">{{ row.user_email }}</button></template>
          <template #cell-created_at="{ value }">{{ new Date(value).toLocaleString() }}</template>
          <template #cell-amount="{ value }">¥{{ value }}</template>
          <template #cell-status="{ value }"><span class="rounded-full px-2 py-1 text-xs font-medium" :class="statusClass[value as InvoiceStatus]">{{ t(`invoices.statuses.${value}`) }}</span></template>
          <template #cell-actions="{ row }">
            <div class="flex flex-wrap gap-2">
              <button class="btn btn-secondary btn-sm" @click="openDetail(row.id)">{{ t('invoices.details') }}</button>
              <template v-if="row.status === 'PENDING'">
                <button class="btn btn-primary btn-sm" :disabled="actionBusy" @click="openAction(row.id, 'approve')">{{ t('invoices.approve') }}</button>
                <button class="btn btn-danger btn-sm" :disabled="actionBusy" @click="openAction(row.id, 'reject')">{{ t('invoices.reject') }}</button>
              </template>
              <button v-if="row.status === 'APPROVED'" class="btn btn-primary btn-sm" :disabled="actionBusy" @click="openAction(row.id, 'issue')">{{ t('invoices.issue') }}</button>
            </div>
          </template>
        </DataTable>
      </div>
      <Pagination v-if="total > 0" :page="page" :total="total" :page-size="pageSize" @update:page="changePage" @update:pageSize="changeSize" />
    </div>
    <InvoiceApplicationDialog v-if="showApply && !admin" @close="showApply = false" @created="onCreated" />
    <BaseDialog :show="detailID !== null" :title="t('invoices.details')" @close="closeDetail">
      <p v-if="detailLoading">{{ t('common.loading') }}</p>
      <p v-else-if="detailError" role="alert" class="text-red-600">{{ detailError }} <button class="btn btn-secondary" @click="openDetail(detailID!)">{{ t('invoices.retry') }}</button></p>
      <InvoiceDetail v-else-if="detail" :item="detail" />
    </BaseDialog>
    <BaseDialog :show="actionTarget !== null" :title="t(`invoices.${actionTarget?.action || 'approve'}`)" @close="closeAction">
      <p v-if="actionLoading">{{ t('common.loading') }}</p>
      <template v-else-if="actionDetail && actionTarget">
        <p class="mb-4">{{ t(`invoices.confirm.${actionTarget.action}`) }}</p>
        <dl class="space-y-2 text-sm">
          <div>{{ t('invoices.applicant') }}: <strong>{{ actionDetail.current_email }}</strong></div>
          <div>{{ t('invoices.amount') }}: <strong>¥{{ actionDetail.amount }}</strong></div>
          <template v-if="actionTarget.action === 'approve' && actionDetail.invoice_type === 'SPECIAL'">
            <div>{{ t('invoices.tax_amount') }} (1%): <strong>{{ formatInvoiceTax(actionDetail.tax_amount) }} {{ t('invoices.balanceUnits') }}</strong></div>
            <div>{{ t('invoices.current_balance') }}: <strong>{{ actionDetail.current_balance }} {{ t('invoices.balanceUnits') }}</strong></div>
          </template>
        </dl>
        <p v-if="actionTarget.action === 'approve' && actionDetail.invoice_type === 'SPECIAL'" class="mt-4 rounded-lg bg-amber-50 p-3 text-sm font-medium text-amber-900 dark:bg-amber-950 dark:text-amber-200">{{ t('invoices.manualDeduction') }}</p>
      </template>
      <p v-if="actionError" role="alert" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ actionError }}</p>
      <button v-if="actionError && !actionDetail" class="btn btn-secondary mt-3" @click="loadActionDetail">{{ t('invoices.retry') }}</button>
      <template #footer><div class="flex justify-end gap-3">
        <button class="btn btn-secondary" :disabled="actionBusy" @click="closeAction">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="actionLoading || actionBusy || !actionAllowed" @click="confirmAction">{{ actionBusy ? t('common.processing') : t('invoices.confirmAction') }}</button>
      </div></template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { extractApiErrorCode, extractApiErrorMessage } from '@/utils/apiError'
import InvoiceApplicationDialog from './InvoiceApplicationDialog.vue'
import InvoiceDetail from './InvoiceDetail.vue'
import { formatInvoiceTax } from './form'
import { invoiceAPI, type InvoiceAction, type InvoiceApplication, type AdminInvoiceDetail, type InvoiceStatus } from './api'

const props = withDefaults(defineProps<{ admin?: boolean }>(), { admin: false })
const { t, te } = useI18n()
const items = ref<InvoiceApplication[]>([])
const page = ref(1), pageSize = ref(20), total = ref(0)
const loading = ref(false), listError = ref(''), showApply = ref(false)
let listVersion = 0, detailVersion = 0, actionVersion = 0
const detailID = ref<number | null>(null), detail = ref<InvoiceApplication | AdminInvoiceDetail | null>(null)
const detailLoading = ref(false), detailError = ref('')
const actionTarget = ref<{ id: number; action: InvoiceAction } | null>(null)
const actionDetail = ref<AdminInvoiceDetail | null>(null)
const actionLoading = ref(false), actionBusy = ref(false), actionError = ref('')
const actionAllowed = computed(() => actionDetail.value && actionTarget.value && actionDetail.value.status === (actionTarget.value.action === 'issue' ? 'APPROVED' : 'PENDING'))
const rows = computed(() => items.value.map((item, index) => ({ ...item, sequence: (page.value - 1) * pageSize.value + index + 1 })))
const columns = computed(() => [
  { key: 'sequence', label: t('invoices.sequence') },
  ...(props.admin ? [{ key: 'user_email', label: t('invoices.applicant') }] : []),
  { key: 'created_at', label: t('invoices.created_at') },
  ...(!props.admin ? [{ key: 'company_name', label: t('invoices.company_name') }] : []),
  { key: 'amount', label: `${t('invoices.amount')} (CNY)` },
  { key: 'status', label: t('invoices.status') },
  ...(props.admin ? [{ key: 'actions', label: t('invoices.actions') }] : [])
])
const statusClass: Record<InvoiceStatus, string> = {
  PENDING: 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200',
  APPROVED: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  ISSUED: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
  REJECTED: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
}
function errorMessage(err: unknown) {
  const key = `invoices.errors.${extractApiErrorCode(err)}`
  return te(key) ? t(key) : extractApiErrorMessage(err, t('invoices.failed'))
}
async function fetchList() {
  const version = ++listVersion
  loading.value = true; listError.value = ''
  try {
    const result = await invoiceAPI.list(props.admin, page.value, pageSize.value)
    if (version !== listVersion) return
    items.value = result.items; total.value = result.total
  } catch (err) { if (version === listVersion) listError.value = errorMessage(err) }
  finally { if (version === listVersion) loading.value = false }
}
function changePage(value: number) { page.value = value; void fetchList() }
function changeSize(value: number) { pageSize.value = value; changePage(1) }
function onCreated() { showApply.value = false; changePage(1) }
function closeDetail() { ++detailVersion; detailID.value = null; detail.value = null }
async function openDetail(id: number) {
  const version = ++detailVersion
  detailID.value = id; detail.value = null; detailError.value = ''; detailLoading.value = true
  try {
    const result = props.admin ? await invoiceAPI.adminGet(id) : await invoiceAPI.get(id)
    if (version === detailVersion) detail.value = result
  } catch (err) { if (version === detailVersion) detailError.value = errorMessage(err) }
  finally { if (version === detailVersion) detailLoading.value = false }
}
function closeAction() { if (!actionBusy.value) { ++actionVersion; actionTarget.value = null; actionDetail.value = null } }
function openAction(id: number, action: InvoiceAction) {
  if (actionBusy.value) return
  actionTarget.value = { id, action }; void loadActionDetail()
}
async function loadActionDetail() {
  const target = actionTarget.value
  if (!target) return
  const version = ++actionVersion
  actionDetail.value = null; actionError.value = ''; actionLoading.value = true
  try {
    const result = await invoiceAPI.adminGet(target.id)
    if (version !== actionVersion) return
    actionDetail.value = result
    if (!actionAllowed.value) actionError.value = t('invoices.errors.INVOICE_STATE_CONFLICT')
  } catch (err) { if (version === actionVersion) actionError.value = errorMessage(err) }
  finally { if (version === actionVersion) actionLoading.value = false }
}
async function confirmAction() {
  const target = actionTarget.value
  if (!target || !actionAllowed.value || actionBusy.value || actionLoading.value) return
  actionBusy.value = true; actionError.value = ''
  try {
    await invoiceAPI.transition(target.id, target.action)
    actionTarget.value = null; actionDetail.value = null
    await fetchList()
    if (detailID.value === target.id) await openDetail(target.id)
  } catch (err) { actionError.value = errorMessage(err) }
  finally { actionBusy.value = false }
}
watch(() => props.admin, () => {
  closeDetail(); closeAction(); showApply.value = false
  items.value = []; total.value = 0; page.value = 1
  void fetchList()
}, { immediate: true })
</script>
