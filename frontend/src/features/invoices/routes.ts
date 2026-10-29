import type { RouteRecordRaw } from 'vue-router'

export const invoiceRoutes: RouteRecordRaw[] = [
  {
    path: '/invoices', name: 'Invoices',
    component: () => import('./InvoiceManagementView.vue'),
    props: { admin: false },
    meta: { requiresAuth: true, requiresAdmin: false, title: 'Invoices', titleKey: 'invoices.title' }
  },
  {
    path: '/admin/invoices', name: 'AdminInvoices',
    component: () => import('./InvoiceManagementView.vue'),
    props: { admin: true },
    meta: { requiresAuth: true, requiresAdmin: true, title: 'Invoice Management', titleKey: 'invoices.title' }
  }
]
