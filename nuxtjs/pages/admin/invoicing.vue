<template>
  <div data-testid="admin-invoicing-page">
    <ProPageHeader
      :title="$t('admin.invoicing.title')"
      :subtitle="$t('admin.invoicing.subtitle')"
    >
      <template #actions>
        <ProButton
          variant="secondary"
          test-id="admin-invoicing-export-csv"
          :disabled="!pendingPartner.length"
          @click="exportPendingCsv"
        >
          {{ $t('admin.invoicing.exportCsv') }}
        </ProButton>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <p class="pro-hint pro-mb-lg" data-testid="admin-invoicing-hint">
      {{ $t('admin.invoicing.hint') }}
    </p>

    <div v-if="!loading && rows.length" class="pro-field-inline-wrap pro-mb-lg" data-testid="admin-invoicing-kpis">
      <ProBadge variant="neutral" data-testid="admin-invoicing-kpi-total">
        {{ $t('admin.invoicing.kpiTotal', { n: rows.length }) }}
      </ProBadge>
      <ProBadge
        v-if="overduePartner.length"
        variant="danger"
        data-testid="admin-invoicing-kpi-overdue"
      >
        {{ $t('admin.invoicing.kpiOverdue', { n: overduePartner.length }) }}
      </ProBadge>
      <ProBadge
        v-if="pendingPartner.length"
        variant="warning"
        data-testid="admin-invoicing-kpi-pending"
      >
        {{ $t('admin.invoicing.kpiPending', { n: pendingPartner.length }) }}
      </ProBadge>
      <ProBadge
        v-if="quotaWarn.length"
        variant="warning"
        data-testid="admin-invoicing-kpi-quota"
      >
        {{ $t('admin.invoicing.kpiQuota', { n: quotaWarn.length }) }}
      </ProBadge>
    </div>

    <p v-if="error" class="pro-alert pro-mb-md" data-testid="admin-invoicing-error">{{ error }}</p>
    <p v-if="msg" class="pro-hint pro-mb-md" data-testid="admin-invoicing-msg">{{ msg }}</p>

    <ProCard data-testid="admin-invoicing-list">
      <ProTable
        :empty="!rows.length && !loading"
        :empty-title="$t('admin.invoicing.empty')"
      >
        <thead>
          <tr>
            <th>{{ $t('admin.invoicing.colPractice') }}</th>
            <th>{{ $t('admin.invoicing.colParty') }}</th>
            <th>{{ $t('admin.invoicing.colStatus') }}</th>
            <th>{{ $t('admin.invoicing.colUsage') }}</th>
            <th>{{ $t('admin.invoicing.colPartner') }}</th>
            <th>{{ $t('admin.invoicing.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="c in rows"
            :key="c.practiceId"
            :data-testid="`admin-invoicing-row-${c.practiceId}`"
            :class="{
              'pf-inv-row--overdue': isPartnerOverdue(c),
              'pf-inv-row--quota': isQuotaWarn(c) && !isPartnerOverdue(c),
            }"
          >
            <td>
              <strong>{{ c.practiceName || shortId(c.practiceId) }}</strong>
              <div v-if="c.contactEmail" class="pro-hint">{{ c.contactEmail }}</div>
              <div class="pro-hint"><code>{{ shortId(c.practiceId) }}</code></div>
            </td>
            <td>{{ c.billitPartyId || '—' }}</td>
            <td>
              <ProBadge :variant="statusVariant(c.status)">{{ c.status }}</ProBadge>
            </td>
            <td>
              <span :data-testid="`admin-invoicing-usage-${c.practiceId}`">
                {{ c.usageThisMonth ?? 0 }} / {{ c.docsIncludedMonthly ?? '—' }}
              </span>
              <ProBadge v-if="isQuotaWarn(c)" variant="warning" class="pro-ml-sm">
                {{ $t('admin.invoicing.quotaNear') }}
              </ProBadge>
            </td>
            <td>
              <span v-if="c.partnerListedAt" data-testid="admin-invoicing-listed">
                {{ formatDate(c.partnerListedAt) }}
              </span>
              <template v-else>
                <span class="pro-hint">{{ $t('admin.invoicing.notListed') }}</span>
                <ProBadge
                  v-if="isPartnerOverdue(c)"
                  variant="danger"
                  class="pro-ml-sm"
                  :data-testid="`admin-invoicing-overdue-${c.practiceId}`"
                >
                  {{ $t('admin.invoicing.partnerOverdue') }}
                </ProBadge>
              </template>
            </td>
            <td>
              <div class="pro-field-inline-wrap">
                <ProButton
                  v-if="isPartnerPending(c)"
                  variant="secondary"
                  :disabled="busyId === c.practiceId"
                  :test-id="`admin-invoicing-mark-${c.practiceId}`"
                  @click="markPartner(c)"
                >
                  {{ $t('admin.invoicing.markPartner') }}
                </ProButton>
                <span v-else-if="c.partnerListedAt" class="pro-hint">{{ $t('admin.invoicing.alreadyListed') }}</span>
                <span v-else class="pro-hint" data-testid="admin-invoicing-mark-disabled">
                  {{ $t('admin.invoicing.markUnavailable') }}
                </span>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard
      class="pro-mt-xl"
      :title="$t('admin.invoicing.saasTitle')"
      data-testid="admin-invoicing-saas-list"
    >
      <p class="pro-hint pro-mb-md">{{ $t('admin.invoicing.saasSubtitle') }}</p>
      <ProTable
        :empty="!saasTargets.length && !loading"
        :empty-title="$t('admin.invoicing.saasEmpty')"
      >
        <thead>
          <tr>
            <th>{{ $t('admin.invoicing.colPractice') }}</th>
            <th>{{ $t('admin.invoicing.colVat') }}</th>
            <th>{{ $t('admin.invoicing.colConnect') }}</th>
            <th>{{ $t('admin.invoicing.colSaas') }}</th>
            <th>{{ $t('admin.invoicing.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="c in saasTargets"
            :key="`saas-${c.practiceId}`"
            :data-testid="`admin-invoicing-saas-row-${c.practiceId}`"
          >
            <td>
              <strong>{{ c.practiceName || shortId(c.practiceId) }}</strong>
              <div v-if="c.contactEmail" class="pro-hint">{{ c.contactEmail }}</div>
              <div class="pro-hint"><code>{{ shortId(c.practiceId) }}</code></div>
            </td>
            <td>{{ c.vatNumber || '—' }}</td>
            <td>
              <ProBadge :variant="c.hasBillitConnect ? 'success' : 'neutral'">
                {{ c.hasBillitConnect ? $t('admin.invoicing.connectYes') : $t('admin.invoicing.connectNo') }}
              </ProBadge>
            </td>
            <td>
              <ProBadge
                v-if="c.saasDocument"
                :variant="saasStatusVariant(c.saasDocument.status)"
                :data-testid="`admin-invoicing-saas-status-${c.practiceId}`"
              >
                {{ $t('admin.invoicing.saasStatus', { status: c.saasDocument.status }) }}
              </ProBadge>
              <span v-else class="pro-hint">—</span>
            </td>
            <td>
              <div class="pro-field-inline-wrap">
                <ProButton
                  v-if="!c.saasBillingEnabled"
                  variant="secondary"
                  :disabled="busyId === c.practiceId"
                  :test-id="`admin-invoicing-saas-enable-${c.practiceId}`"
                  @click="enableSaasBilling(c)"
                >
                  {{ $t('admin.invoicing.saasEnable') }}
                </ProButton>
                <ProButton
                  v-if="canCreateSaasDraft(c)"
                  variant="secondary"
                  :disabled="busyId === c.practiceId"
                  :test-id="`admin-invoicing-saas-draft-${c.practiceId}`"
                  @click="createSaasDraft(c)"
                >
                  {{ $t('admin.invoicing.saasDraft') }}
                </ProButton>
                <ProButton
                  v-if="canSendSaas(c)"
                  variant="primary"
                  :disabled="busyId === c.practiceId"
                  :test-id="`admin-invoicing-saas-send-${c.practiceId}`"
                  @click="sendSaas(c)"
                >
                  {{ $t('admin.invoicing.saasSend') }}
                </ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const PARTNER_OVERDUE_MS = 7 * 24 * 60 * 60 * 1000
const QUOTA_WARN_RATIO = 0.8

type SaasDoc = {
  id: string
  status: string
  billitOrderId?: string
  totalInclCents?: number
  peppolStatus?: string
}

type Conn = {
  practiceId: string
  practiceName?: string
  contactEmail?: string
  billitPartyId?: string
  status: string
  docsIncludedMonthly?: number
  usageThisMonth?: number
  partnerListedAt?: string
  connectedAt?: string
  invoiceToPartner?: boolean
}

type SaasTarget = {
  practiceId: string
  practiceName?: string
  contactEmail?: string
  vatNumber?: string
  hasBillitConnect?: boolean
  saasBillingEnabled?: boolean
  saasDraftEnabled?: boolean
  saasDocument?: SaasDoc
}

const { t } = useI18n()
const rows = ref<Conn[]>([])
const saasTargets = ref<SaasTarget[]>([])
const loading = ref(true)
const busyId = ref('')
const error = ref('')
const msg = ref('')

function unwrap<T>(res: any): T {
  return (res?.data ?? res) as T
}

function shortId(id: string) {
  return id?.length > 12 ? `${id.slice(0, 8)}…` : id
}

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'active': return 'success'
    case 'pending_kyc':
    case 'pending_registration': return 'warning'
    case 'error':
    case 'suspended': return 'danger'
    default: return 'neutral'
  }
}

function isPartnerPending(c: Conn) {
  return c.status === 'active'
    && c.invoiceToPartner !== false
    && !c.partnerListedAt
    && Boolean(c.billitPartyId)
}

function isPartnerOverdue(c: Conn) {
  if (!isPartnerPending(c) || !c.connectedAt) return false
  const t0 = Date.parse(c.connectedAt)
  if (Number.isNaN(t0)) return false
  return Date.now() - t0 >= PARTNER_OVERDUE_MS
}

function isQuotaWarn(c: Conn) {
  const max = c.docsIncludedMonthly ?? 0
  const used = c.usageThisMonth ?? 0
  if (max <= 0) return false
  return used / max >= QUOTA_WARN_RATIO
}

const pendingPartner = computed(() => rows.value.filter(isPartnerPending))
const overduePartner = computed(() => rows.value.filter(isPartnerOverdue))
const quotaWarn = computed(() => rows.value.filter(isQuotaWarn))

function exportPendingCsv() {
  const lines = [
    'practiceId,practiceName,contactEmail,billitPartyId,connectedAt,status',
    ...pendingPartner.value.map((c) => [
      c.practiceId,
      csvEscape(c.practiceName || ''),
      csvEscape(c.contactEmail || ''),
      c.billitPartyId || '',
      c.connectedAt || '',
      c.status,
    ].join(',')),
  ]
  const blob = new Blob([lines.join('\n')], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `billit-partner-pending-${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
  msg.value = t('admin.invoicing.exportOk', { n: pendingPartner.value.length })
}

function csvEscape(v: string) {
  if (/[",\n]/.test(v)) return `"${v.replace(/"/g, '""')}"`
  return v
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [connRes, saasRes] = await Promise.all([
      $fetch('/api/admin/invoicing/connections'),
      $fetch('/api/admin/invoicing/saas-targets').catch((e: any) => {
        if (e?.statusCode === 404 || e?.status === 404) return null
        throw e
      }),
    ])
    rows.value = unwrap<Conn[]>(connRes) || []
    saasTargets.value = saasRes ? (unwrap<SaasTarget[]>(saasRes) || []) : []
  } catch (e: any) {
    if (e?.statusCode === 404 || e?.status === 404) {
      error.value = t('admin.invoicing.disabled')
      rows.value = []
      saasTargets.value = []
    } else {
      error.value = e?.data?.error?.message || e?.message || t('admin.invoicing.loadError')
    }
  } finally {
    loading.value = false
  }
}

async function markPartner(c: Conn) {
  if (!isPartnerPending(c)) {
    error.value = t('admin.invoicing.markUnavailable')
    return
  }
  const label = c.practiceName || shortId(c.practiceId)
  if (!confirm(t('admin.invoicing.markConfirm', { name: label }))) return
  busyId.value = c.practiceId
  error.value = ''
  msg.value = ''
  try {
    await $fetch(`/api/admin/invoicing/connections/${encodeURIComponent(c.practiceId)}/mark-partner-invoiced`, {
      method: 'POST',
    })
    msg.value = t('admin.invoicing.markOk')
    await load()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('admin.invoicing.markError')
  } finally {
    busyId.value = ''
  }
}

async function createSaasDraft(c: SaasTarget) {
  if (!c.saasDraftEnabled) {
    error.value = t('admin.invoicing.saasUnavailable')
    return
  }
  if (!c.saasBillingEnabled) {
    error.value = t('admin.invoicing.saasBillingOff')
    return
  }
  const label = c.practiceName || shortId(c.practiceId)
  if (!confirm(t('admin.invoicing.saasConfirm', { name: label }))) return
  busyId.value = c.practiceId
  error.value = ''
  msg.value = ''
  try {
    const res = await $fetch(`/api/admin/invoicing/connections/${encodeURIComponent(c.practiceId)}/saas-draft`, {
      method: 'POST',
    })
    const doc = unwrap<{ id?: string; totalInclCents?: number; billitOrderId?: string }>(res)
    msg.value = t('admin.invoicing.saasOk', {
      id: shortId(doc?.id || ''),
      total: ((doc?.totalInclCents ?? 0) / 100).toFixed(2),
    })
    await load()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('admin.invoicing.saasError')
  } finally {
    busyId.value = ''
  }
}

function canCreateSaasDraft(c: SaasTarget) {
  if (!c.saasDraftEnabled || !c.saasBillingEnabled) return false
  const st = c.saasDocument?.status
  if (!st) return true
  return st === 'draft' || st === 'rejected'
}

function canSendSaas(c: SaasTarget) {
  if (!c.saasDraftEnabled || !c.saasBillingEnabled || !c.saasDocument?.id) return false
  return c.saasDocument.status === 'draft' || c.saasDocument.status === 'rejected'
}

function saasStatusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'delivered': return 'success'
    case 'sending':
    case 'issued': return 'warning'
    case 'rejected':
    case 'cancelled': return 'danger'
    default: return 'neutral'
  }
}

async function enableSaasBilling(c: SaasTarget) {
  const label = c.practiceName || shortId(c.practiceId)
  if (!confirm(t('admin.invoicing.saasEnableConfirm', { name: label }))) return
  busyId.value = c.practiceId
  error.value = ''
  msg.value = ''
  try {
    await $fetch(`/api/admin/invoicing/practices/${encodeURIComponent(c.practiceId)}/saas-billing`, {
      method: 'POST',
      body: { enabled: true },
    })
    msg.value = t('admin.invoicing.saasEnableOk')
    await load()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('admin.invoicing.saasEnableError')
  } finally {
    busyId.value = ''
  }
}

async function sendSaas(c: SaasTarget) {
  if (!canSendSaas(c) || !c.saasDocument?.id) {
    error.value = t('admin.invoicing.saasSendUnavailable')
    return
  }
  const label = c.practiceName || shortId(c.practiceId)
  if (!confirm(t('admin.invoicing.saasSendConfirm', { name: label }))) return
  busyId.value = c.practiceId
  error.value = ''
  msg.value = ''
  try {
    const res = await $fetch(
      `/api/admin/invoicing/connections/${encodeURIComponent(c.practiceId)}/saas-documents/${encodeURIComponent(c.saasDocument.id)}/send`,
      { method: 'POST' },
    )
    const doc = unwrap<{ status?: string }>(res)
    msg.value = t('admin.invoicing.saasSendOk', { status: doc?.status || '' })
    await load()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('admin.invoicing.saasSendError')
  } finally {
    busyId.value = ''
  }
}

onMounted(load)
</script>

<style scoped>
.pf-inv-row--overdue td {
  background: color-mix(in srgb, var(--pf-danger, #b42318) 8%, transparent);
}
.pf-inv-row--quota td {
  background: color-mix(in srgb, var(--pf-warning, #b54708) 8%, transparent);
}
.pro-ml-sm {
  margin-left: 0.35rem;
}
.pro-mt-xl {
  margin-top: 2rem;
}
.pro-mt-md {
  margin-top: 1rem;
}
</style>
