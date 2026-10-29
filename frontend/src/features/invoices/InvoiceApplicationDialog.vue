<template>
  <BaseDialog :show="true" :title="t('invoices.apply')" @close="close">
    <div role="note" class="mb-5 rounded-lg border border-amber-300 bg-amber-50 p-4 text-sm font-medium text-amber-900 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-200">{{ t('invoices.notice') }}</div>
    <form id="invoice-application" class="space-y-4" @submit.prevent="submit">
      <div>
        <label for="invoice-amount" class="input-label">{{ t('invoices.amount') }} (CNY) <span class="text-red-600 dark:text-red-400" aria-hidden="true">*</span></label>
        <input id="invoice-amount" :value="amount" type="number" min="200" max="9007199254740991" step="1" required class="input w-full" :disabled="saving" @input="amount = ($event.target as HTMLInputElement).value" />
        <p class="mt-1 text-xs text-gray-500">{{ t('invoices.amountHint') }}</p>
      </div>
      <div>
        <label for="invoice-type" class="input-label">{{ t('invoices.invoice_type') }} <span class="text-red-600 dark:text-red-400" aria-hidden="true">*</span></label>
        <select id="invoice-type" v-model="title.invoice_type" class="input w-full" :disabled="saving" @change="error = ''">
          <option value="NORMAL">{{ t('invoices.types.NORMAL') }}</option>
          <option value="SPECIAL">{{ t('invoices.types.SPECIAL') }}</option>
        </select>
      </div>
      <div v-for="field in titleFields" :key="field.key">
        <label :for="`invoice-${field.key}`" class="input-label">
          {{ t(`invoices.${field.key}`) }}
          <span v-if="field.alwaysRequired || title.invoice_type === 'SPECIAL'" class="text-red-600 dark:text-red-400" aria-hidden="true">*</span>
          <span v-else>({{ t('invoices.optional') }})</span>
        </label>
        <input :id="`invoice-${field.key}`" v-model="title[field.key]" type="text" class="input w-full" :maxlength="field.max" :required="field.alwaysRequired || title.invoice_type === 'SPECIAL'" :disabled="saving" />
      </div>
      <div v-if="title.invoice_type === 'SPECIAL'" class="rounded-lg bg-gray-50 p-3 text-sm dark:bg-dark-800">
        <p>{{ t('invoices.tax_amount') }}: <strong>{{ invoiceTaxPreview(amount) }}</strong> {{ t('invoices.balanceUnits') }} (1%)</p>
      </div>
      <p v-if="error" role="alert" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
    </form>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" :disabled="saving" @click="close">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" type="submit" form="invoice-application" :disabled="saving">{{ saving ? t('common.processing') : t('invoices.submit') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { extractApiErrorCode, extractApiErrorMessage } from '@/utils/apiError'
import { invoiceAPI, type InvoiceTitle } from './api'
import { titleFields, validInvoiceAmount, validInvoiceTitle, invoiceTaxPreview } from './form'

const emit = defineEmits<{ close: []; created: [] }>()
const { t, te } = useI18n()
const amount = ref('200')
const title = reactive<InvoiceTitle>({ invoice_type: 'NORMAL', company_name: '', tax_number: '', address: '', phone: '', bank_name: '', bank_account: '' })
const saving = ref(false)
const error = ref('')
function close() { if (!saving.value) emit('close') }
async function submit() {
  if (saving.value) return
  error.value = ''
  if (!validInvoiceAmount(amount.value)) { error.value = t('invoices.errors.INVOICE_INVALID_AMOUNT'); return }
  if (!validInvoiceTitle(title)) { error.value = t('invoices.errors.INVOICE_INVALID_FIELDS'); return }
  saving.value = true
  try {
    await invoiceAPI.create({ ...title, amount: Number(amount.value) })
    emit('created')
  } catch (err) {
    const key = `invoices.errors.${extractApiErrorCode(err)}`
    error.value = te(key) ? t(key) : extractApiErrorMessage(err, t('invoices.failed'))
  } finally { saving.value = false }
}
</script>
