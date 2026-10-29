<template>
  <dl class="grid grid-cols-1 gap-4 text-sm sm:grid-cols-2">
    <div v-for="field in fields" :key="field.label" class="min-w-0">
      <dt class="text-gray-500 dark:text-gray-400">{{ field.label }}</dt>
      <dd class="mt-1 whitespace-pre-wrap break-words font-medium text-gray-900 dark:text-gray-100">{{ field.value || '—' }}</dd>
    </div>
  </dl>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InvoiceApplication, AdminInvoiceDetail } from './api'
import { titleFields, formatInvoiceTax } from './form'
const props = defineProps<{ item: InvoiceApplication | AdminInvoiceDetail }>()
const { t } = useI18n()
const date = (value: string | null) => value ? new Date(value).toLocaleString() : '—'
const fields = computed(() => {
  const item = props.item
  const result = [
    { label: t('invoices.invoice_type'), value: t(`invoices.types.${item.invoice_type}`) },
    { label: t('invoices.status'), value: t(`invoices.statuses.${item.status}`) },
    ...titleFields.map(field => ({ label: t(`invoices.${field.key}`), value: item[field.key] })),
    { label: t('invoices.amount'), value: `¥${item.amount}` },
    { label: t('invoices.tax_amount'), value: `${formatInvoiceTax(item.tax_amount)} ${t('invoices.balanceUnits')}` },
    { label: t('invoices.created_at'), value: date(item.created_at) },
    { label: t('invoices.reviewed_at'), value: date(item.reviewed_at) },
    { label: t('invoices.issued_at'), value: date(item.issued_at) }
  ]
  if ('current_email' in item) result.push(
    { label: t('invoices.current_email'), value: item.current_email },
    { label: t('invoices.current_balance'), value: `${item.current_balance} ${t('invoices.balanceUnits')}` }
  )
  return result
})
</script>
