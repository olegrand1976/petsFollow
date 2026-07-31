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
            v-if="job.status === 'uploaded' || job.status === 'failed'"
            test-id="admin-compendium-extract"
            :disabled="busy"
            @click="startExtract"
          >
            {{ $t('admin.compendium.startExtract') }}
          </ProButton>
          <ProBadge :variant="statusVariant(job.status)">{{ statusLabel(job.status) }}</ProBadge>
        </div>
      </ProCard>

      <ProCard class="pro-mb-lg" data-testid="admin-compendium-preview">
        <h3 class="pro-mb-md">{{ $t('admin.compendium.previewTitle') }}</h3>
        <ProTable :empty="!rows.length" :empty-title="$t('admin.compendium.emptyRows')">
          <thead>
            <tr>
              <th>#</th>
              <th>{{ $t('admin.compendium.colCnk') }}</th>
              <th>{{ $t('admin.compendium.colName') }}</th>
              <th>{{ $t('admin.compendium.colAtc') }}</th>
              <th>{{ $t('admin.compendium.colForm') }}</th>
              <th>{{ $t('admin.compendium.colAb') }}</th>
              <th>{{ $t('admin.compendium.colStatus') }}</th>
              <th>{{ $t('admin.compendium.colActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>{{ row.rowNumber }}</td>
              <td>
                <input
                  v-if="row.status !== 'upserted' && row.status !== 'excluded'"
                  v-model="row.cnk"
                  class="pro-input"
                  @change="patchRow(row, { cnk: row.cnk })"
                >
                <span v-else>{{ row.cnk }}</span>
              </td>
              <td>
                <input
                  v-if="row.status !== 'upserted' && row.status !== 'excluded'"
                  v-model="row.name"
                  class="pro-input"
                  @change="patchRow(row, { name: row.name })"
                >
                <span v-else>{{ row.name }}</span>
              </td>
              <td>{{ row.atcCode || '—' }}</td>
              <td>{{ row.pharmaceuticalForm || '—' }}</td>
              <td>{{ row.isAntibiotic ? '✓' : '—' }}</td>
              <td>
                <ProBadge :variant="rowStatusVariant(row.status)">{{ statusLabel(row.status) }}</ProBadge>
                <span v-if="row.errorCode" class="pro-hint"> {{ row.errorCode }}</span>
              </td>
              <td>
                <ProButton
                  v-if="row.status !== 'upserted' && row.status !== 'excluded'"
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
const commitMsg = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null

function statusVariant (status: string) {
  switch (status) {
    case 'completed': case 'ready': case 'upserted': return 'success'
    case 'failed': case 'error': return 'danger'
    case 'extracting': case 'committing': return 'warning'
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
</style>
