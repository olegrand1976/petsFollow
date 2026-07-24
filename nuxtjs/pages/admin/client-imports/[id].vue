<template>
  <div data-testid="admin-client-import-detail">
    <ProPageHeader
      :title="$t('admin.clientImport.detailTitle')"
      :subtitle="job ? `${job.filename} — ${job.vetFullName}` : ''"
    >
      <template #actions>
        <NuxtLink to="/admin/client-imports" class="pro-link">{{ $t('admin.clientImport.back') }}</NuxtLink>
      </template>
    </ProPageHeader>

    <div v-if="loading" class="pro-hint">{{ $t('common.loading') }}</div>
    <template v-else-if="job">
      <div class="pro-grid-kpi pro-mb-lg">
        <ProKpi :value="job.rowCount" :label="$t('admin.clientImport.kpiRows')" />
        <ProKpi :value="job.okCount" :label="$t('admin.clientImport.kpiReady')" />
        <ProKpi :value="job.errorCount" :label="$t('admin.clientImport.kpiErrors')" />
        <ProKpi :value="job.createdCount" :label="$t('admin.clientImport.kpiCreated')" />
      </div>

      <ProCard class="pro-mb-lg" data-testid="admin-import-mapping">
        <h3 class="pro-mb-md">{{ $t('admin.clientImport.mappingTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('admin.clientImport.mappingHint') }}</p>
        <div class="pro-form pro-form--inline">
          <div class="pro-field">
            <label class="pro-label">{{ $t('admin.clientImport.mapEmail') }}</label>
            <select v-model="mapping.email" class="pro-select" data-testid="admin-import-map-email">
              <option value="">—</option>
              <option v-for="h in job.headers" :key="`e-${h}`" :value="h">{{ h }}</option>
            </select>
          </div>
          <div class="pro-field">
            <label class="pro-label">{{ $t('admin.clientImport.mapFullName') }}</label>
            <select v-model="mapping.fullName" class="pro-select" data-testid="admin-import-map-fullname">
              <option value="">—</option>
              <option v-for="h in job.headers" :key="`n-${h}`" :value="h">{{ h }}</option>
            </select>
          </div>
          <div class="pro-field">
            <label class="pro-label">{{ $t('admin.clientImport.mapLocale') }}</label>
            <select v-model="mapping.locale" class="pro-select" data-testid="admin-import-map-locale">
              <option value="">{{ $t('admin.clientImport.mapLocaleNone') }}</option>
              <option v-for="h in job.headers" :key="`l-${h}`" :value="h">{{ h }}</option>
            </select>
          </div>
        </div>
        <div class="pro-flex-gap pro-mt-md">
          <ProButton
            variant="secondary"
            test-id="admin-import-suggest"
            :disabled="busy || !canMap"
            @click="suggestMapping"
          >
            {{ $t('admin.clientImport.suggestGemini') }}
          </ProButton>
          <ProButton
            test-id="admin-import-apply-mapping"
            :disabled="busy || !mapping.email || !mapping.fullName || !canMap"
            @click="applyMapping"
          >
            {{ $t('admin.clientImport.applyMapping') }}
          </ProButton>
        </div>
        <p v-if="msg" class="pro-hint pro-mt-md" data-testid="admin-import-msg">{{ msg }}</p>
      </ProCard>

      <ProCard class="pro-mb-lg" data-testid="admin-import-preview">
        <h3 class="pro-mb-md">{{ $t('admin.clientImport.previewTitle') }}</h3>
        <ProTable :empty="!rows.length" :empty-title="$t('admin.clientImport.emptyRows')">
          <thead>
            <tr>
              <th>#</th>
              <th>{{ $t('admin.clientImport.colEmail') }}</th>
              <th>{{ $t('admin.clientImport.colName') }}</th>
              <th>{{ $t('admin.clientImport.colLocale') }}</th>
              <th>{{ $t('admin.clientImport.colStatus') }}</th>
              <th>{{ $t('admin.clientImport.colActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>{{ row.rowNumber }}</td>
              <td>
                <input
                  v-if="row.status !== 'created' && row.status !== 'excluded'"
                  v-model="row.email"
                  class="pro-input"
                  @change="patchRow(row, { email: row.email })"
                >
                <span v-else>{{ row.email }}</span>
              </td>
              <td>
                <input
                  v-if="row.status !== 'created' && row.status !== 'excluded'"
                  v-model="row.fullName"
                  class="pro-input"
                  @change="patchRow(row, { fullName: row.fullName })"
                >
                <span v-else>{{ row.fullName }}</span>
              </td>
              <td>{{ row.locale || '—' }}</td>
              <td>
                <ProBadge :variant="rowStatusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge>
                <span v-if="row.errorCode" class="pro-hint"> {{ row.errorCode }}</span>
              </td>
              <td>
                <ProButton
                  v-if="row.status !== 'created' && row.status !== 'excluded'"
                  variant="ghost"
                  @click="patchRow(row, { excluded: true })"
                >
                  {{ $t('admin.clientImport.exclude') }}
                </ProButton>
                <ProButton
                  v-else-if="row.status === 'excluded'"
                  variant="ghost"
                  @click="patchRow(row, { excluded: false })"
                >
                  {{ $t('admin.clientImport.include') }}
                </ProButton>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>

      <ProCard data-testid="admin-import-commit">
        <h3 class="pro-mb-md">{{ $t('admin.clientImport.commitTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('admin.clientImport.commitHint') }}</p>
        <div class="pro-flex-gap">
          <ProButton
            test-id="admin-import-commit-btn"
            :disabled="busy || !canCommit"
            @click="commit"
          >
            {{ $t('admin.clientImport.commit') }}
          </ProButton>
          <ProButton
            v-if="credentialsToken"
            variant="secondary"
            test-id="admin-import-credentials"
            @click="downloadCredentials"
          >
            {{ $t('admin.clientImport.downloadCredentials') }}
          </ProButton>
        </div>
        <p v-if="commitMsg" class="pro-hint pro-mt-md" data-testid="admin-import-commit-msg">{{ commitMsg }}</p>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const route = useRoute()
const id = computed(() => String(route.params.id))

type ImportJob = {
  id: string
  filename: string
  status: string
  headers: string[]
  rowCount: number
  okCount: number
  errorCount: number
  createdCount: number
  vetFullName?: string
  columnMapping?: { email?: string | null, fullName?: string | null, locale?: string | null }
}

type ImportRow = {
  id: string
  rowNumber: number
  email: string
  fullName: string
  locale: string
  status: string
  errorCode?: string
}

const loading = ref(true)
const busy = ref(false)
const job = ref<ImportJob | null>(null)
const rows = ref<ImportRow[]>([])
const mapping = reactive({ email: '', fullName: '', locale: '' })
const msg = ref('')
const commitMsg = ref('')
const credentialsToken = ref('')

const canMap = computed(() => {
  const s = job.value?.status
  return s === 'uploaded' || s === 'mapping_ready' || s === 'preview_ready'
})

const readyRowCount = computed(() => rows.value.filter(r => r.status === 'ready').length)

const canCommit = computed(() => {
  const s = job.value?.status
  return (s === 'preview_ready' || s === 'failed') && readyRowCount.value > 0
})

function statusLabel(status: string) {
  return t(`admin.clientImport.status.${status}`, status)
}

function rowStatusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'created':
    case 'ready':
      return 'success'
    case 'error':
      return 'danger'
    case 'excluded':
      return 'neutral'
    default:
      return 'warning'
  }
}

function applyMappingFromJob(j: ImportJob) {
  mapping.email = j.columnMapping?.email ?? ''
  mapping.fullName = j.columnMapping?.fullName ?? ''
  mapping.locale = j.columnMapping?.locale ?? ''
}

async function load() {
  loading.value = true
  try {
    const res: any = await $fetch(`/api/admin/client-imports/${id.value}`)
    const detail = res?.data ?? res
    job.value = detail.job
    rows.value = detail.rows ?? []
    if (job.value) applyMappingFromJob(job.value)
  } finally {
    loading.value = false
  }
}

async function suggestMapping() {
  busy.value = true
  msg.value = ''
  try {
    const res: any = await $fetch(`/api/admin/client-imports/${id.value}/suggest-mapping`, { method: 'POST' })
    const data = res?.data ?? res
    job.value = data.job
    rows.value = data.rows ?? rows.value
    if (job.value) applyMappingFromJob(job.value)
    msg.value = t('admin.clientImport.suggestOk')
  } catch (e: any) {
    const code = e?.data?.error?.code ?? e?.data?.error?.message
    msg.value = code === 'gemini_not_configured'
      ? t('admin.clientImport.geminiMissing')
      : (e?.data?.error?.message ?? t('admin.clientImport.suggestFailed'))
  } finally {
    busy.value = false
  }
}

async function applyMapping() {
  busy.value = true
  msg.value = ''
  try {
    const body: Record<string, string | null> = {
      email: mapping.email,
      fullName: mapping.fullName,
      locale: mapping.locale || null,
    }
    const res: any = await $fetch(`/api/admin/client-imports/${id.value}/mapping`, {
      method: 'PUT',
      body,
    })
    const detail = res?.data ?? res
    job.value = detail.job
    rows.value = detail.rows ?? []
    msg.value = t('admin.clientImport.mappingApplied')
  } catch (e: any) {
    msg.value = e?.data?.error?.message ?? t('admin.clientImport.mappingFailed')
  } finally {
    busy.value = false
  }
}

async function patchRow(row: ImportRow, patch: Record<string, unknown>) {
  busy.value = true
  try {
    const res: any = await $fetch(`/api/admin/client-imports/${id.value}/rows/${row.id}`, {
      method: 'PATCH',
      body: patch,
    })
    const updated = res?.data ?? res
    const idx = rows.value.findIndex(r => r.id === row.id)
    if (idx >= 0) rows.value[idx] = { ...rows.value[idx], ...updated }
    await load()
  } catch (e: any) {
    msg.value = e?.data?.error?.message ?? t('admin.clientImport.rowPatchFailed')
  } finally {
    busy.value = false
  }
}

async function commit() {
  if (!confirm(t('admin.clientImport.commitConfirm'))) return
  busy.value = true
  commitMsg.value = ''
  try {
    const res: any = await $fetch(`/api/admin/client-imports/${id.value}/commit`, { method: 'POST' })
    const data = res?.data ?? res
    job.value = data.job
    credentialsToken.value = data.credentialsToken ?? ''
    commitMsg.value = t('admin.clientImport.commitOk', {
      created: data.createdCount ?? 0,
      errors: data.errorCount ?? 0,
    })
    await load()
  } catch (e: any) {
    commitMsg.value = e?.data?.error?.message ?? t('admin.clientImport.commitFailed')
  } finally {
    busy.value = false
  }
}

function downloadCredentials() {
  if (!credentialsToken.value) return
  window.location.href = `/api/admin/client-imports/${id.value}/credentials?token=${encodeURIComponent(credentialsToken.value)}`
}

onMounted(() => load())
</script>

<style scoped>
.pro-form--inline {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
}
</style>
