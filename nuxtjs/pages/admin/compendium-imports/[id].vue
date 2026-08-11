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
            @click="startExtract(false)"
          >
            {{ extractPrimaryLabel }}
          </ProButton>
          <ProButton
            v-if="canForceRestartExtract"
            variant="secondary"
            test-id="admin-compendium-extract-restart"
            :disabled="busy"
            @click="startExtract(true)"
          >
            {{ $t('admin.compendium.restartExtract') }}
          </ProButton>
          <ProBadge :variant="statusVariant(job.status)">{{ statusLabel(job.status) }}</ProBadge>
        </div>
        <p v-if="canResumeExtract" class="pro-hint pro-mt-md" data-testid="admin-compendium-resume-hint">
          {{ $t('admin.compendium.resumeHint', { done: job.extractDone, total: job.extractTotal }) }}
        </p>
      </ProCard>

      <div class="compendium-review pro-mb-lg" data-testid="admin-compendium-preview">
        <ProCard class="compendium-review__table">
          <h3 class="pro-mb-md">{{ $t('admin.compendium.previewTitle') }}</h3>
          <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.previewHint') }}</p>
          <p v-if="errorRowCount > 0" class="pro-hint pro-mb-md" data-testid="admin-compendium-manual-hint">
            {{ $t('admin.compendium.manualHint') }}
          </p>
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
                <th>{{ $t('admin.compendium.colSubstance') }}</th>
                <th>{{ $t('admin.compendium.colStrength') }}</th>
                <th>{{ $t('admin.compendium.colAtc') }}</th>
                <th>{{ $t('admin.compendium.colForm') }}</th>
                <th>{{ $t('admin.compendium.colPack') }}</th>
                <th>{{ $t('admin.compendium.colStatus') }}</th>
                <th>{{ $t('admin.compendium.colActions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in rows"
                :key="row.id"
                :class="{
                  'compendium-row--needs-manual': row.status === 'error',
                  'compendium-row--focused': focusedRowId === row.id,
                }"
                :data-testid="row.status === 'error' ? 'admin-compendium-row-error' : undefined"
                @click="focusRow(row)"
              >
                <td>{{ row.rowNumber }}</td>
                <td>
                  <input
                    v-if="rowEditable(row)"
                    v-model.number="row.sourcePage"
                    type="number"
                    min="1"
                    class="pro-input compendium-page-input"
                    data-testid="admin-compendium-row-page"
                    :title="$t('admin.compendium.pageHint', { page: row.sourcePage ?? '—' })"
                    :disabled="isRowBusy(row)"
                    @click.stop
                    @focus="focusRow(row)"
                    @change="onSourcePageChange(row)"
                  >
                  <button
                    v-else
                    type="button"
                    class="compendium-page compendium-page-btn"
                    data-testid="admin-compendium-row-page"
                    @click.stop="focusRow(row)"
                  >{{ row.sourcePage ?? '—' }}</button>
                </td>
                <td>
                  <input
                    v-if="rowEditable(row)"
                    v-model="row.cnk"
                    class="pro-input"
                    :placeholder="$t('admin.compendium.colCnk')"
                    :disabled="isRowBusy(row)"
                    @click.stop
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
                      :disabled="isRowBusy(row)"
                      @click.stop
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
                    :disabled="isRowBusy(row)"
                    @click.stop
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
                    :disabled="isRowBusy(row)"
                    @click.stop
                    @change="patchRow(row, { manufacturer: row.manufacturer })"
                  >
                  <span v-else>{{ row.manufacturer || '—' }}</span>
                </td>
                <td>
                  <input
                    v-if="rowEditable(row)"
                    v-model="row.activeSubstance"
                    class="pro-input"
                    :placeholder="$t('admin.compendium.colSubstance')"
                    :disabled="isRowBusy(row)"
                    @click.stop
                    @change="patchRow(row, { activeSubstance: row.activeSubstance })"
                  >
                  <span v-else>{{ row.activeSubstance || '—' }}</span>
                </td>
                <td>
                  <input
                    v-if="rowEditable(row)"
                    v-model="row.strength"
                    class="pro-input"
                    data-testid="admin-compendium-row-strength"
                    :placeholder="$t('admin.compendium.colStrength')"
                    :disabled="isRowBusy(row)"
                    @click.stop
                    @change="patchRow(row, { strength: row.strength })"
                  >
                  <span v-else>{{ row.strength || '—' }}</span>
                </td>
                <td>
                  <input
                    v-if="rowEditable(row)"
                    v-model="row.atcCode"
                    class="pro-input"
                    data-testid="admin-compendium-row-atc"
                    :placeholder="$t('admin.compendium.colAtc')"
                    :disabled="isRowBusy(row)"
                    @click.stop
                    @change="patchRow(row, { atcCode: row.atcCode })"
                  >
                  <span v-else>{{ row.atcCode || '—' }}</span>
                </td>
                <td>
                  <input
                    v-if="rowEditable(row)"
                    v-model="row.pharmaceuticalForm"
                    class="pro-input"
                    :placeholder="$t('admin.compendium.colForm')"
                    :disabled="isRowBusy(row)"
                    @click.stop
                    @change="patchRow(row, { pharmaceuticalForm: row.pharmaceuticalForm })"
                  >
                  <span v-else>{{ row.pharmaceuticalForm || '—' }}</span>
                </td>
                <td>
                  <input
                    v-if="rowEditable(row)"
                    v-model="row.packSize"
                    class="pro-input"
                    :placeholder="$t('admin.compendium.colPack')"
                    :disabled="isRowBusy(row)"
                    @click.stop
                    @change="patchRow(row, { packSize: row.packSize })"
                  >
                  <span v-else>{{ row.packSize || '—' }}</span>
                </td>
                <td>
                  <ProBadge :variant="rowStatusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge>
                  <span v-if="row.errorCode" class="pro-hint"> {{ row.errorCode }}</span>
                  <p v-if="row.status === 'error' && row.sourcePage" class="pro-hint">
                    {{ $t('admin.compendium.fixFromPage', { page: row.sourcePage }) }}
                  </p>
                </td>
                <td @click.stop>
                  <ProButton
                    v-if="row.status === 'pending' || row.status === 'error'"
                    variant="ghost"
                    test-id="admin-compendium-lookup-cnk"
                    :disabled="busy || isRowBusy(row) || !row.name"
                    @click="lookupCnk(row)"
                  >
                    {{ $t('admin.compendium.lookupCnk') }}
                  </ProButton>
                  <ProButton
                    v-if="(row.status === 'pending' || row.status === 'error') && row.suggestedCnk && !row.cnk"
                    variant="ghost"
                    :disabled="isRowBusy(row)"
                    @click="patchRow(row, { cnk: row.suggestedCnk })"
                  >
                    {{ $t('admin.compendium.acceptSuggested') }}
                  </ProButton>
                  <ProButton
                    v-if="canConfirmRow(row)"
                    variant="ghost"
                    test-id="admin-compendium-confirm-row"
                    :disabled="busy || isRowBusy(row)"
                    @click="patchRow(row, { name: row.name, cnk: row.cnk })"
                  >
                    {{ $t('admin.compendium.confirmRow') }}
                  </ProButton>
                  <ProButton
                    v-if="rowEditable(row)"
                    variant="ghost"
                    :disabled="isRowBusy(row)"
                    @click="patchRow(row, { excluded: true })"
                  >
                    {{ $t('admin.compendium.exclude') }}
                  </ProButton>
                  <ProButton
                    v-else-if="row.status === 'excluded'"
                    variant="ghost"
                    :disabled="isRowBusy(row)"
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

        <ProCard class="compendium-review__viewer" data-testid="admin-compendium-pdf-viewer">
          <h3 class="pro-mb-md">{{ $t('admin.compendium.viewerTitle') }}</h3>
          <p class="pro-hint pro-mb-md">{{ $t('admin.compendium.viewerHint') }}</p>
          <div class="pro-flex-gap pro-mb-md">
            <label class="compendium-viewer-page">
              <span class="pro-hint">{{ $t('admin.compendium.colPage') }}</span>
              <input
                v-model.number="viewerPage"
                type="number"
                min="1"
                class="pro-input compendium-page-input"
                data-testid="admin-compendium-viewer-page"
                @change="applyViewerPage"
                @keyup.enter="applyViewerPage"
              >
            </label>
            <ProButton
              variant="secondary"
              test-id="admin-compendium-open-pdf"
              @click="openPdfTab"
            >
              {{ $t('admin.compendium.openPdfTab') }}
            </ProButton>
          </div>
          <iframe
            v-if="pdfIframeSrc"
            class="compendium-pdf-frame"
            :src="pdfIframeSrc"
            :title="$t('admin.compendium.viewerTitle')"
            data-testid="admin-compendium-pdf-frame"
          />
        </ProCard>
      </div>

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
const patchingId = ref('')
const lookupMsg = ref<Record<string, string>>({})
const focusedRowId = ref('')
const viewerPage = ref(0)
/** Page applied to the iframe (not every keystroke — avoids re-download of large PDFs). */
const appliedViewerPage = ref(0)
let pollTimer: ReturnType<typeof setInterval> | null = null

const pendingCount = computed(() => rows.value.filter(r => canConfirmRow(r)).length)
const errorRowCount = computed(() => rows.value.filter(r => r.status === 'error').length)
const pdfBaseUrl = computed(() => `/api/admin/compendium-imports/${id.value}/pdf`)
const pdfIframeSrc = computed(() => {
  if (!job.value) return ''
  const page = Number(appliedViewerPage.value) > 0
    ? Number(appliedViewerPage.value)
    : (Number(job.value.pageStart) || 1)
  return `${pdfBaseUrl.value}#page=${page}`
})
const canResumeExtract = computed(() => {
  if (!job.value) return false
  if (job.value.status !== 'failed' && job.value.status !== 'extracting') return false
  const done = Number(job.value.extractDone || 0)
  const total = Number(job.value.extractTotal || 0)
  return done > 0 && total > done
})
const canForceRestartExtract = computed(() => canResumeExtract.value)
const extractPrimaryLabel = computed(() => {
  if (!job.value) return ''
  if (job.value.status === 'uploaded') return t('admin.compendium.startExtract')
  if (canResumeExtract.value) return t('admin.compendium.resumeExtract')
  return t('admin.compendium.retryExtract')
})

function rowEditable (row: any) {
  return row?.status !== 'upserted' && row?.status !== 'excluded'
}

function isRowBusy (row: any) {
  return patchingId.value === row?.id || lookingUpId.value === row?.id
}

function canConfirmRow (row: any) {
  if (!rowEditable(row)) return false
  if (row.status !== 'pending' && row.status !== 'error') return false
  return Boolean(String(row.cnk || '').trim() && String(row.name || '').trim())
}

function setViewerPage (page: number) {
  const n = Math.trunc(Number(page))
  if (!Number.isFinite(n) || n < 1) return
  viewerPage.value = n
  appliedViewerPage.value = n
}

function applyViewerPage () {
  setViewerPage(viewerPage.value)
}

function focusRow (row: any) {
  if (!row?.id) return
  focusedRowId.value = row.id
  const page = Number(row.sourcePage)
  if (page > 0) setViewerPage(page)
  else if (job.value?.pageStart) setViewerPage(Number(job.value.pageStart) || 1)
}

function onSourcePageChange (row: any) {
  focusRow(row)
  void patchSourcePage(row)
}

function openPdfTab () {
  const page = Number(appliedViewerPage.value) > 0
    ? Number(appliedViewerPage.value)
    : (Number(viewerPage.value) > 0 ? Number(viewerPage.value) : (Number(job.value?.pageStart) || 1))
  window.open(`${pdfBaseUrl.value}#page=${page}`, '_blank', 'noopener,noreferrer')
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
  if (!viewerPage.value && job.value?.pageStart) {
    setViewerPage(Number(job.value.pageStart) || 1)
  }
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

async function startExtract (forceRestart = false) {
  busy.value = true
  try {
    const q = forceRestart ? '?restart=1' : ''
    await $fetch(`/api/admin/compendium-imports/${id.value}/extract${q}`, { method: 'POST' })
    startPoll()
    await load()
  } finally {
    busy.value = false
  }
}

async function patchSourcePage (row: any) {
  const n = Number(row.sourcePage)
  if (!Number.isFinite(n) || n < 1) {
    row.sourcePage = null
    // 0 clears source_page server-side (*int null vs omit is ambiguous).
    await patchRow(row, { sourcePage: 0 })
    return
  }
  row.sourcePage = Math.trunc(n)
  await patchRow(row, { sourcePage: row.sourcePage })
}

async function patchRow (row: any, body: Record<string, unknown>) {
  patchingId.value = row.id
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
    patchingId.value = ''
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
.compendium-review {
  display: grid;
  gap: 1rem;
  align-items: start;
}
@media (min-width: 1100px) {
  .compendium-review {
    grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.9fr);
  }
  .compendium-review__viewer {
    position: sticky;
    top: 1rem;
  }
}
.compendium-viewer-page {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}
.compendium-pdf-frame {
  width: 100%;
  min-height: 70vh;
  border: 1px solid var(--pf-vet-border);
  border-radius: 6px;
  background: var(--pf-vet-surface, #fff);
}
.compendium-page {
  font-family: var(--pf-font-mono, ui-monospace, monospace);
  font-weight: 600;
}
.compendium-page-btn {
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
  text-underline-offset: 2px;
}
.compendium-page-input {
  width: 5.5rem;
  font-family: var(--pf-font-mono, ui-monospace, monospace);
  font-weight: 600;
}
.compendium-row--needs-manual td {
  background: color-mix(in srgb, var(--pf-vet-alert) 8%, transparent);
}
.compendium-row--focused td {
  outline: 1px solid color-mix(in srgb, var(--pf-vet-primary) 35%, transparent);
}
</style>
