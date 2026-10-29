export default {
  title: '发票管理', apply: '发起开票申请', submit: '提交申请', refresh: '刷新', retry: '重试',
  sequence: '序号', applicant: '申请人（账号）', created_at: '申请时间', company_name: '发票抬头（企业名称）',
  invoice_type: '发票类型', tax_number: '税号', address: '详细地址', phone: '联系电话',
  bank_name: '开户银行', bank_account: '银行账号', amount: '开票金额', status: '当前状态', actions: '操作',
  tax_amount: '预计应扣税费额度', current_balance: '当前账户余额', current_email: '账号注册邮箱',
  reviewed_at: '审核时间', issued_at: '开票时间', optional: '选填', details: '申请详情',
  approve: '通过', reject: '驳回', issue: '开票', confirmAction: '确认', balanceUnits: '余额额度',
  amountHint: '最低开票金额为200元，仅支持整数。',
  notice: '开票金额不得低于200元，开具专票时需按照1%税率额外支付税费，请确保账户余额足以支付税费，防止审核失败。审核成功后5个工作日内由管理员发放至账号注册邮箱内，请注意查收。',
  noConversion: '按开票金额的1%直接扣除同数值余额额度，不进行汇率换算；由管理员手动处理。',
  manualDeduction: '请手动完成税费扣除，本操作仅修改审核状态，不会自动扣款。',
  failed: '操作失败，请重试。',
  types: { NORMAL: '普票', SPECIAL: '专票' },
  statuses: { PENDING: '待审核', APPROVED: '审核通过', REJECTED: '已驳回', ISSUED: '已开票' },
  confirm: {
    approve: '确认通过此开票申请？请先完成线下核对。',
    reject: '确认驳回此申请？驳回后用户可以重新提交新的申请。',
    issue: '确认已完成开票？请将发票发送至账号注册邮箱，此操作不会自动发送邮件。'
  },
  errors: {
    INVOICE_INVALID_AMOUNT: '开票金额必须为不低于200的整数，且不得超过9007199254740991。',
    INVOICE_INVALID_FIELDS: '请完整填写必填抬头信息，并检查各字段长度。',
    INVOICE_NOT_FOUND: '发票申请不存在或无权访问。',
    INVOICE_STATE_CONFLICT: '申请状态已发生变化，请刷新后重试。'
  }
}
