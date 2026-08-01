<template>
  <div data-testid="admin-afmps-import-detail">
    <ProPageHeader
      :title="$t('admin.afmps.detailTitle')"
      :subtitle="job ? job.filename : ''"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
        <NuxtLink to="/admin/afmps-imports" class="pro-link">{{ $t('admin.afmps.back') }}</NuxtLink>
      </template>
    </ProPageHeader>

    <div v-if="loading && !job" class="pro-hint">{{ $t('common.loading') }}</div>
    <template v-else-if="job">
      <div class="pro-grid-kpi pro-mb-lg">
        <ProKpi :value="job.readyCount" :label="$t('admin.afmps.kpiReady')" />
        <ProKpi :value="job.insertCount" :label="$t('admin.afmps.kpiInsert')" />
        <ProKpi :value="job.updateCount" :label="$t('admin.afmps.kpiUpdate')" />
        <ProKpi :value="job.unchangedCount" :label="$t('admin.afmps.kpiUnchanged')" />
        <ProKpi :value="job.deactivatePreview" :label="$t('admin.afmps.kpiDeactivate')" />
        <ProKpi :value="job.upsertedCount" :label="$t('admin.afmps.kpiUpserted')" />
      </div>

      <ProCard class="pro-mb-lg" data-testid="admin-afmps-gates">
        <h3 class="pro-mb-md">{{ $t('admin.afmps.gatesTitle') }}</h3>
        <p class="pro-hint">{{ $t('admin.afmps.gate1') }} → {{ $t('admin.afmps.gate2') }} → {{ $t('admin.afmps.gate3') }}</p>
        <p class="pro-mt-md">
          <ProBadge :variant="statusVariant(job.status)">{{ statusLabel(job.status) }}</ProBadge>
        </p>
        <p v-if="job.status === 'blocked'" class="pro-hint pro-hint--error pro-mt-md">{{ $t('admin.afmps.blockedHint') }}</p>
        <p v-if="job.errorMessage" class="pro-hint pro-hint--error pro-mt-md">{{ job.errorMessage }}</p>

        <div v-if="job.status === 'validated'" class="pro-mt-md" data-testid="admin-afmps-review-ack">
          <p class="pro-hint pro-mb-md">{{ $t('admin.afmps.reviewHint') }}</p>
          <label class="pro-field pro-flex-gap">
            <input v-model="reviewAck" type="checkbox" data-testid="admin-afmps-review-ack-check">
            <span>{{ $t('admin.afmps.reviewAck') }}</span>
          </label>
          <div class="pro-flex-gap pro-mt-md">
            <ProButton
              test-id="admin-afmps-mark-reviewed"
              :disabled="busy || !reviewAck"
              @click="markReviewed"
            >
              {{ $t('admin.afmps.markReviewed') }}
            </ProButton>
          </div>
        </div>

        <div v-if="job.status === 'reviewed'" class="pro-form pro-mt-md" data-testid="admin-afmps-commit-form">
          <div class="pro-field">
            <label class="pro-label" for="afmps-confirm">{{ $t('admin.afmps.confirmLabel') }}</label>
            <input
              id="afmps-confirm"
              v-model="confirmPhrase"
              class="pro-input"
              data-testid="admin-afmps-confirm"
              :placeholder="$t('admin.afmps.confirmPlaceholder')"
            >
          </div>
          <label class="pro-field pro-flex-gap">
            <input v-model="deactivateMissing" type="checkbox" data-testid="admin-afmps-deactivate">
            <span>{{ $t('admin.afmps.deactivateMissing') }}</span>
          </label>
          <p v-if="deactivateMissing" class="pro-hint pro-hint--error">
            {{ $t('admin.afmps.deactivateHint', { count: job.deactivatePreview }) }}
          </p>
          <ProButton
            test-id="admin-afmps-commit"
            :disabled="busy || !canCommit"
            @click="runCommit"
          >
            {{ $t('admin.afmps.commit') }}
          </ProButton>
        </div>
        <p v-if="actionError" class="pro-hint pro-hint--error pro-mt-md" data-testid="admin-afmps-action-error">{{ actionError }}</p>
      </ProCard>

      <ProCard data-testid="admin-afmps-preview">
        <h3 class="pro-mb-md">{{ $t('admin.afmps.previewTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('admin.afmps.previewHint') }}</p>
        <div class="pro-flex-gap pro-mb-md" data-testid="admin-afmps-collision-filters">
          <ProButton
            v-for="f in collisionFilters"
            :key="f"
            :variant="collisionFilter === f ? 'primary' : 'ghost'"
            :test-id="`admin-afmps-filter-${f}`"
            @click="setCollisionFilter(f)"
          >
            {{ $t(`admin.afmps.filter.${f}`) }}
          </ProButton>
        </div>
        <p class="pro-hint pro-mb-md" data-testid="admin-afmps-rows-meta">
          {{ $t('admin.afmps.rowsShown', { shown: rows.length, total: rowsTotal }) }}
        </p>
        <ProTable :empty="!rows.length" :empty-title="$t('admin.afmps.emptyRows')">
          <thead>
            <tr>
              <th>#</th>
              <th>{{ $t('admin.afmps.colCnk') }}</th>
              <th>{{ $t('admin.afmps.colName') }}</th>
              <th>{{ $t('admin.afmps.colForm') }}</th>
              <th>{{ $t('admin.afmps.colPack') }}</th>
              <th>{{ $t('admin.afmps.colCollision') }}</th>
              <th>{{ $t('admin.afmps.colStatus') }}</th>
              <th>{{ $t('admin.afmps.colActions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td>{{ row.rowNumber }}</td>
              <td>{{ row.cnk }}</td>
              <td>{{ row.name }}</td>
              <td>{{ row.pharmaceuticalForm }}</td>
              <td>{{ row.packSize }}</td>
              <td>{{ row.collision }}</td>
              <td><ProBadge :variant="row.status === 'error' ? 'danger' : 'neutral'">{{ statusLabel(row.status) }}</ProBadge></td>
              <td>
                <ProButton
                  v-if="row.status === 'ready' && (job.status === 'validated' || job.status === 'reviewed')"
                  variant="ghost"
                  @click="patchRow(row, true)"
                >
                  {{ $t('admin.afmps.exclude') }}
                </ProButton>
                <ProButton
                  v-else-if="row.status === 'excluded' && (job.status === 'validated' || job.status === 'reviewed')"
                  variant="ghost"
                  @click="patchRow(row, false)"
                >
                  {{ $t('admin.afmps.include') }}
                </ProButton>
              </td>
            </tr>
          </tbody>
        </ProTable>
        <div v-if="rows.length < rowsTotal" class="pro-flex-gap pro-mt-md">
          <ProButton
            test-id="admin-afmps-load-more"
            variant="secondary"
            :disabled="loadingMore"
            @click="loadMore"
          >
            {{ $t('admin.afmps.loadMore') }}
          </ProButton>
        </div>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const PAGE = 500
const collisionFilters = ['all', 'insert', 'update', 'unchanged', 'error'] as const
type CollisionFilter = typeof collisionFilters[number]

const { t } = useI18n()
const route = useRoute()
const id = computed(() => String(route.params.id || ''))
const loading = ref(true)
const loadingMore = ref(false)
const busy = ref(false)
const job = ref<any>(null)
const rows = ref<any[]>([])
const rowsTotal = ref(0)
const collisionFilter = ref<CollisionFilter>('all')
const reviewAck = ref(false)
const confirmPhrase = ref('')
const deactivateMissing = ref(false)
const actionError = ref('')

const canCommit = computed(() => {
  const phrase = confirmPhrase.value.trim()
  return phrase === 'IMPORT AFMPS' || phrase === 'IMPORT_AFMPS'
})

function statusVariant (status: string) {
  switch (status) {
    case 'completed': return 'success'
    case 'failed': case 'blocked': return 'danger'
    case 'committing': case 'reviewed': return 'warning'
    default: return 'neutral'
  }
}
function statusLabel (status: string) {
  return t(`admin.afmps.status.${status}`, status)
}

async function fetchPage (offset: number, append: boolean) {
  const res: any = await $fetch(`/api/admin/afmps-imports/${id.value}`, {
    query: { limit: PAGE, offset, collision: collisionFilter.value },
  })
  const data = res?.data ?? res
  job.value = data?.job
  rowsTotal.value = data?.rowsTotal ?? 0
  const pageRows = data?.rows ?? []
  rows.value = append ? [...rows.value, ...pageRows] : pageRows
}

async function load () {
  loading.value = true
  try {
    await fetchPage(0, false)
  } finally {
    loading.value = false
  }
}

async function setCollisionFilter (f: CollisionFilter) {
  collisionFilter.value = f
  await load()
}

async function loadMore () {
  loadingMore.value = true
  try {
    await fetchPage(rows.value.length, true)
  } finally {
    loadingMore.value = false
  }
}

async function markReviewed () {
  if (!reviewAck.value) return
  busy.value = true
  actionError.value = ''
  try {
    await $fetch(`/api/admin/afmps-imports/${id.value}/mark-reviewed`, { method: 'POST' })
    await load()
  } catch (e: any) {
    actionError.value = e?.data?.error?.message ?? e?.statusMessage ?? t('admin.afmps.markFailed')
  } finally {
    busy.value = false
  }
}

async function runCommit () {
  if (deactivateMissing.value && (job.value?.deactivatePreview ?? 0) > 100) {
    if (!window.confirm(t('admin.afmps.deactivateConfirm', { count: job.value.deactivatePreview }))) {
      return
    }
  }
  busy.value = true
  actionError.value = ''
  try {
    await $fetch(`/api/admin/afmps-imports/${id.value}/commit`, {
      method: 'POST',
      body: { confirm: confirmPhrase.value.trim(), deactivateMissing: deactivateMissing.value },
    })
    await load()
  } catch (e: any) {
    actionError.value = e?.data?.error?.message ?? e?.statusMessage ?? t('admin.afmps.commitFailed')
  } finally {
    busy.value = false
  }
}

async function patchRow (row: any, exclude: boolean) {
  busy.value = true
  try {
    await $fetch(`/api/admin/afmps-imports/${id.value}/rows/${row.id}`, {
      method: 'PATCH',
      body: { exclude },
    })
    await load()
  } finally {
    busy.value = false
  }
}

onMounted(() => { void load() })
</script>
