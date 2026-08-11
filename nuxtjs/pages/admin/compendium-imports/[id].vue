<template>
  <div data-testid="admin-compendium-import-detail">
    <ProPageHeader
      :title="$t('admin.compendium.detailTitle')"
      :subtitle="job ? `${job.filename} · p.${job.pageStart}–${job.pageEnd}` : ''"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
        <NuxtLink to="/admin/compendium-imports" class="pro-link">{{ $t('admin.compendium.back') }}</NuxtLink>
      </template>
    </ProPageHeader>

    <div v-if="loading" class="pro-hint">{{ $t('common.loading') }}</div>
    <template v-else-if="job">
      <div class="pro-grid-kpi pro-mb-lg">
        <ProKpi :value="`${job.extractPct ?? 0}%`" :label="$t('admin.compendium.kpiExtract')" />
        <ProKpi :value="`${job.reviewPct ?? 0}%`" :label="$t('admin.compendium.kpiReview')" />
        <ProKpi :value="job.readyCount" :label="$t('admin.compendium.kpiReady')" />
        <ProKpi :value="job.upsertedCount" :label="$t('admin.compendium.kpiUpserted')" />
      </div>

      <ProCard class="pro-mb-lg" data-testid="admin-compendium-progress">
        <h3 class="pro-mb-md">{{ $t('admin.compendium.progressTitle') }}</h3>
        <p class="pro-hint">{{ $t('admin.compendium.progressExtract', { pct: job.extractPct ?? 0, done: job.extractDone, total: job.extractTotal }) }}</p>
        <div class="compendium-bar" role="progressbar" :aria-valuenow="job.extractPct ?? 0" aria-valuemin="0" aria-valuemax="100">
          <div class="compendium-bar__fill" :style="{ width: `${job.extractPct ?? 0}%` }" />
        </div>
        <p class="pro-hint pro-mt-md">{{ $t('admin.compendium.progressReview', { pct: job.reviewPct ?? 0, done: job.reviewedCount, total: job.rowCount }) }}</p>
        <div class="compendium-bar" role="progressbar" :aria-valuenow="job.reviewPct ?? 0" aria-valuemin="0" aria-valuemax="100">
          <div class="compendium-bar__fill compendium-bar__fill--review" :style="{ width: `${job.reviewPct ?? 0}%` }" />
        </div>
        <p v-if="job.errorMessage" class="pro-hint pro-hint--error pro-mt-md">{{ job.errorMessage }}</p>
        <div class="pro-flex-gap pro-mt-md">
          <ProButton
            v-if="job.status === 'uploaded' || job.status === 'failed' || job.status === 'extracting'"
            test-id="admin-compendium-extract"
            :disabled="busy"
            @click="startExtract"
          >
            {{ job.status === 'uploaded' ? $t('admin.compendium.startExtract') : $t('admin.compendium.retryExtract') }}
          </ProButton>
          <ProBadge :variant="statusVariant(job.status)">{{ statusLabel(job.status) }}</ProBadge>
        </div>
      </ProCard>

      <ProCard class="pro-mb-lg" data-testid="admin-compendium-preview">
        <h3 class="pro-mb-md">{{ $t('admin.compendium.previewTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.previewHint') }}</p>
        <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.manualHint') }}</p>
        <p
          v-if="refCatalogCount === 0 && job.status === 'extracted'"
          class="pro-hint pro-hint--error pro-mb-md"
          data-testid="admin-compendium-catalog-empty"
        >
          {{ $t('admin.compendium.catalogEmpty') }}
        </p>
        <div v-if="pendingCount > 0 && job.status === 'extracted'" class="pro-flex-gap pro-mb-md">
          <ProButton
            test-id="admin-compendium-confirm-ready"
            variant="secondary"
            :disabled="busy"
            @click="confirmReady"
          >
            {{ $t('admin.compendium.confirmReady', { count: pendingCount }) }}
          </ProButton>
        </div>
        <ProTable :empty="!rows.length" :empty-title="$t('admin.compendium.emptyRows')">
          <thead>
            <tr>
              <th>#</th>
              <th>{{ $t('admin.compendium.colPage') }}</th>
              <th>{{ $t('admin.compendium.colCnk') }}</th>
              <th>{{ $t('admin.compendium.colSuggestedCnk') }}</th>
              <th>{{ $t('admin.compendium.colName') }}</th>
              <th>{{ $t('admin.compendium.colLab') }}</th>
              <th>{{ $t('admin.compendium.colForm') }}</th>
              <th>{{ $t('admin.compendium.colStatus') }}</th>
              <th>{{ $t('admin.compendium.colActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in rows"
              :key="row.id"
              :class="{ 'compendium-row--needs-manual': row.status === 'error' }"
              :data-testid="row.status === 'error' ? 'admin-compendium-row-error' : undefined"
            >
              <td>{{ row.rowNumber }}</td>
              <td>
                <span
                  class="compendium-page"
                  data-testid="admin-compendium-row-page"
                  :title="$t('admin.compendium.pageHint', { page: row.sourcePage ?? '—' })"
                >
                  {{ row.sourcePage ?? '—' }}
                </span>
              </td>
              <td>
                <input
                  v-if="rowEditable(row)"
                  v-model="row.cnk"
                  class="pro-input"
                  :placeholder="$t('admin.compendium.colCnk')"
                  @change="patchRow(row, { cnk: row.cnk })"
                >
                <span v-else>{{ row.cnk }}</span>
              </td>
              <td>
                <template v-if="rowEditable(row) && candidates(row).length">
                  <select
                    class="pro-input"
                    data-testid="admin-compendium-cnk-candidates"
                    :value="row.cnk || row.suggestedCnk || ''"
                    @change="onPickCandidate(row, ($event.target as HTMLSelectElement).value)"
                  >
                    <option value="">{{ $t('admin.compendium.pickCnk') }}</option>
                    <option
                      v-for="c in candidates(row)"
                      :key="c.cnk"
                      :value="c.cnk"
                    >
                      {{ c.cnk }} — {{ c.name }} ({{ Math.round((c.score || 0) * 100) }}%)
                    </option>
                  </select>
                </template>
                <span v-else-if="row.suggestedCnk">{{ row.suggestedCnk }}</span>
                <span v-else>—</span>
              </td>
              <td>
                <input
                  v-if="rowEditable(row)"
                  v-model="row.name"
                  class="pro-input"
                  :placeholder="$t('admin.compendium.colName')"
                  @change="patchRow(row, { name: row.name })"
                >
                <span v-else>{{ row.name }}</span>
              </td>
              <td>
                <input
                  v-if="rowEditable(row)"
                  v-model="row.manufacturer"
                  class="pro-input"
                  :placeholder="$t('admin.compendium.colLab')"
                  @change="patchRow(row, { manufacturer: row.manufacturer })"
                >
                <span v-else>{{ row.manufacturer || '—' }}</span>
              </td>
              <td>
                <input
                  v-if="rowEditable(row)"
                  v-model="row.pharmaceuticalForm"
                  class="pro-input"
                  :placeholder="$t('admin.compendium.colForm')"
                  @change="patchRow(row, { pharmaceuticalForm: row.pharmaceuticalForm })"
                >
                <span v-else>{{ row.pharmaceuticalForm || '—' }}</span>
              </td>
              <td>
                <ProBadge :variant="rowStatusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge>
                <span v-if="row.errorCode" class="pro-hint"> {{ row.errorCode }}</span>
                <p v-if="row.status === 'error' && row.sourcePage" class="pro-hint">
                  {{ $t('admin.compendium.fixFromPage', { page: row.sourcePage }) }}
                </p>
              </td>
              <td>
                <ProButton
                  v-if="row.status === 'pending' || row.status === 'error'"
                  variant="ghost"
                  test-id="admin-compendium-lookup-cnk"
                  :disabled="busy || lookingUpId === row.id || !row.name"
                  @click="lookupCnk(row)"
                >
                  {{ $t('admin.compendium.lookupCnk') }}
                </ProButton>
                <ProButton
                  v-if="(row.status === 'pending' || row.status === 'error') && row.suggestedCnk && !row.cnk"
                  variant="ghost"
                  @click="patchRow(row, { cnk: row.suggestedCnk })"
                >
                  {{ $t('admin.compendium.acceptSuggested') }}
                </ProButton>
                <ProButton
                  v-if="canConfirmRow(row)"
                  variant="ghost"
                  test-id="admin-compendium-confirm-row"
                  :disabled="busy"
                  @click="patchRow(row, { name: row.name, cnk: row.cnk })"
                >
                  {{ $t('admin.compendium.confirmRow') }}
                </ProButton>
                <ProButton
                  v-if="rowEditable(row)"
                  variant="ghost"
                  @click="patchRow(row, { excluded: true })"
                >
                  {{ $t('admin.compendium.exclude') }}
                </ProButton>
                <ProButton
                  v-else-if="row.status === 'excluded'"
                  variant="ghost"
                  @click="patchRow(row, { excluded: false })"
                >
                  {{ $t('admin.compendium.include') }}
                </ProButton>
                <p v-if="lookupMsg[row.id]" class="pro-hint" data-testid="admin-compendium-lookup-msg">{{ lookupMsg[row.id] }}</p>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>

      <ProCard data-testid="admin-compendium-commit">
        <h3 class="pro-mb-md">{{ $t('admin.compendium.commitTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.commitHint') }}</p>
        <ProButton
          test-id="admin-compendium-commit"
          :disabled="busy || job.status !== 'extracted' || !job.readyCount"
          @click="commit"
        >
          {{ $t('admin.compendium.commit') }}
        </ProButton>
        <p v-if="commitMsg" class="pro-hint pro-mt-md" data-testid="admin-compendium-commit-msg">{{ commitMsg }}</p>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const route = useRoute()
const id = computed(() => String(route.params.id))
const loading = ref(true)
const busy = ref(false)
const job = ref<any>(null)
const rows = ref<any[]>([])
const refCatalogCount = ref(0)
const commitMsg = ref('')
const lookingUpId = ref('')
const lookupMsg = ref<Record<string, string>>({})
let pollTimer: ReturnType<typeof setInterval> | null = null

const pendingCount = computed(() => rows.value.filter(r => r.status === 'pending' && r.cnk).length)

function rowEditable (row: any) {
  return row?.status !== 'upserted' && row?.status !== 'excluded'
}

function canConfirmRow (row: any) {
  if (!rowEditable(row)) return false
  if (row.status !== 'pending' && row.status !== 'error') return false
  return Boolean(String(row.cnk || '').trim() && String(row.name || '').trim())
}

function candidates (row: any): Array<{ cnk: string; name: string; score?: number }> {
  const raw = row?.matchCandidates
  if (Array.isArray(raw)) return raw
  if (typeof raw === 'string') {
    try {
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}

function onPickCandidate (row: any, cnk: string) {
  if (!cnk) return
  row.cnk = cnk
  void patchRow(row, { cnk })
}

function statusVariant (status: string) {
  switch (status) {
    case 'completed': case 'ready': case 'upserted': return 'success'
    case 'failed': case 'error': return 'danger'
    case 'extracting': case 'committing': case 'pending': return 'warning'
    default: return 'neutral'
  }
}
function rowStatusVariant (status: string) {
  return statusVariant(status)
}
function statusLabel (status: string) {
  return t(`admin.compendium.status.${status}`, status)
}

async function load () {
  const res: any = await $fetch(`/api/admin/compendium-imports/${id.value}`)
  const data = res?.data ?? res
  job.value = data?.job ?? data
  rows.value = data?.rows ?? []
  refCatalogCount.value = Number(data?.refCatalogCount ?? 0)
  loading.value = false
  if (job.value?.status === 'extracting') {
    startPoll()
  } else {
    stopPoll()
  }
}

function startPoll () {
  if (pollTimer) return
  pollTimer = setInterval(() => { void load() }, 1500)
}
function stopPoll () {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function startExtract () {
  busy.value = true
  try {
    await $fetch(`/api/admin/compendium-imports/${id.value}/extract`, { method: 'POST' })
    startPoll()
    await load()
  } finally {
    busy.value = false
  }
}

async function patchRow (row: any, body: Record<string, unknown>) {
  busy.value = true
  try {
    const res: any = await $fetch(`/api/admin/compendium-imports/${id.value}/rows/${row.id}`, {
      method: 'PATCH',
      body,
    })
    const updated = res?.data ?? res
    const idx = rows.value.findIndex(r => r.id === row.id)
    if (idx >= 0) rows.value[idx] = { ...rows.value[idx], ...updated }
    await load()
  } finally {
    busy.value = false
  }
}

async function lookupCnk (row: any) {
  if (!row?.id || !row.name) return
  lookingUpId.value = row.id
  lookupMsg.value = { ...lookupMsg.value, [row.id]: '' }
  busy.value = true
  try {
    const body: Record<string, string> = {}
    if (row.cnk) body.cnk = String(row.cnk).trim()
    const res: any = await $fetch(`/api/admin/compendium-imports/${id.value}/rows/${row.id}/lookup-cnk`, {
      method: 'POST',
      body,
    })
    const data = res?.data ?? res
    const updated = data?.row ?? data
    const idx = rows.value.findIndex(r => r.id === row.id)
    if (idx >= 0 && updated) rows.value[idx] = { ...rows.value[idx], ...updated }
    if (typeof data?.cnkFound === 'boolean') {
      lookupMsg.value = {
        ...lookupMsg.value,
        [row.id]: data.cnkFound
          ? t('admin.compendium.lookupCnkFound')
          : t('admin.compendium.lookupCnkNotFound'),
      }
    } else {
      lookupMsg.value = {
        ...lookupMsg.value,
        [row.id]: t('admin.compendium.lookupCnkDone'),
      }
    }
    await load()
  } catch (e: any) {
    lookupMsg.value = {
      ...lookupMsg.value,
      [row.id]: e?.data?.error?.message ?? e?.statusMessage ?? t('admin.compendium.lookupCnkFailed'),
    }
  } finally {
    lookingUpId.value = ''
    busy.value = false
  }
}

async function confirmReady () {
  busy.value = true
  try {
    await $fetch(`/api/admin/compendium-imports/${id.value}/confirm-ready`, { method: 'POST' })
    await load()
  } finally {
    busy.value = false
  }
}

async function commit () {
  if (!confirm(t('admin.compendium.commitConfirm'))) return
  busy.value = true
  commitMsg.value = ''
  try {
    const res: any = await $fetch(`/api/admin/compendium-imports/${id.value}/commit`, { method: 'POST' })
    const result = res?.data?.result ?? res?.result
    commitMsg.value = t('admin.compendium.commitOk', {
      upserted: result?.upserted ?? 0,
      skipped: result?.skipped ?? 0,
    })
    await load()
  } catch (e: any) {
    commitMsg.value = e?.data?.error?.message ?? t('admin.compendium.commitFailed')
  } finally {
    busy.value = false
  }
}

onMounted(() => { void load() })
onBeforeUnmount(() => stopPoll())
</script>

<style scoped>
.compendium-bar {
  height: 8px;
  border-radius: 4px;
  background: var(--pf-vet-border);
  overflow: hidden;
}
.compendium-bar__fill {
  height: 100%;
  background: var(--pf-vet-accent);
  transition: width 0.3s ease;
}
.compendium-bar__fill--review {
  background: var(--pf-vet-primary);
}
.compendium-page {
  font-family: var(--pf-font-mono, ui-monospace, monospace);
  font-weight: 600;
}
.compendium-row--needs-manual td {
  background: color-mix(in srgb, var(--pf-vet-alert) 8%, transparent);
}
</style>
