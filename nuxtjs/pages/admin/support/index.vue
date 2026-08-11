<template>
  <div data-testid="admin-support-page">
    <ProPageHeader :title="$t('admin.support.title')" :subtitle="$t('admin.support.subtitle')" />

    <div v-if="stats" class="pro-grid-kpi pro-mb-md" data-testid="admin-support-stats">
      <ProKpi :value="statusCount('open')" :label="$t('admin.support.statsOpen')" />
      <ProKpi :value="statusCount('in_progress')" :label="$t('admin.support.statsInProgress')" />
      <ProKpi :value="statusCount('to_test')" :label="$t('admin.support.statsToTest')" />
      <ProKpi :value="statusCount('done')" :label="$t('admin.support.statsDone')" />
      <ProKpi :value="statusCount('closed')" :label="$t('admin.support.statsClosed')" />
      <ProKpi :value="stats.openOlderThan24h ?? 0" :label="$t('admin.support.statsOpen24h')" />
      <ProKpi :value="stats.openOlderThan7d ?? 0" :label="$t('admin.support.statsOpen7d')" />
    </div>

    <ProCard>
      <ProListToolbar v-model:view-mode="viewMode">
        <template #filters>
          <div class="support-filters">
            <div class="support-filters__field">
              <label class="pro-label" for="support-search">{{ $t('admin.support.searchLabel') }}</label>
              <input
                id="support-search"
                v-model="searchQ"
                type="search"
                class="pro-input"
                data-testid="admin-support-search"
                :placeholder="$t('admin.support.searchPlaceholder')"
                @keydown.enter.prevent="onFilterChange"
              >
            </div>
            <div class="support-filters__field">
              <label class="pro-label" for="support-status-filter">{{ $t('admin.support.colStatus') }}</label>
              <select
                id="support-status-filter"
                v-model="statusFilter"
                class="pro-select"
                data-testid="admin-support-status-filter"
                @change="onFilterChange"
              >
                <option value="">{{ $t('admin.support.filterAll') }}</option>
                <option value="open">{{ $t('admin.support.filterOpen') }}</option>
                <option value="in_progress">{{ $t('admin.support.filterInProgress') }}</option>
                <option value="to_test">{{ $t('admin.support.filterToTest') }}</option>
                <option value="done">{{ $t('admin.support.filterDone') }}</option>
                <option value="closed">{{ $t('admin.support.filterClosed') }}</option>
              </select>
            </div>
            <div class="support-filters__field">
              <label class="pro-label" for="support-source-filter">{{ $t('admin.support.colSource') }}</label>
              <select
                id="support-source-filter"
                v-model="sourceFilter"
                class="pro-select"
                data-testid="admin-support-source-filter"
                @change="onFilterChange"
              >
                <option value="">{{ $t('admin.support.filterAllSources') }}</option>
                <option value="nuxt_pro">{{ $t('admin.support.source_nuxt_pro') }}</option>
                <option value="flutter_client">{{ $t('admin.support.source_flutter_client') }}</option>
                <option value="flutter_pro_light">{{ $t('admin.support.source_flutter_pro_light') }}</option>
                <option value="system">{{ $t('admin.support.source_system') }}</option>
              </select>
            </div>
            <ProButton type="button" variant="secondary" test-id="admin-support-search-btn" :disabled="loading" @click="onFilterChange">
              {{ $t('admin.support.searchSubmit') }}
            </ProButton>
          </div>
        </template>
      </ProListToolbar>

      <p v-if="patchError" class="pro-error" role="alert" data-testid="admin-support-patch-error">
        {{ patchError }}
      </p>

      <p
        v-if="viewMode === 'kanban' && total > KANBAN_LIMIT"
        class="text-muted"
        data-testid="admin-support-kanban-truncated"
      >
        {{ $t('admin.support.kanbanTruncated', { shown: rows.length, total, limit: KANBAN_LIMIT }) }}
      </p>

      <ProTable v-if="viewMode === 'table'" :empty="!rows.length && !loadError" :empty-title="$t('admin.support.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.support.colSubject') }}</th>
            <th>{{ $t('admin.support.colUser') }}</th>
            <th>{{ $t('admin.support.colSource') }}</th>
            <th>{{ $t('admin.support.colStatus') }}</th>
            <th>{{ $t('admin.support.colReplies') }}</th>
            <th>{{ $t('admin.support.colCreated') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="ticket in rows"
            :key="ticket.id"
            class="support-row"
            data-testid="admin-support-row"
            @click="navigateTo(`/admin/support/${ticket.id}`)"
          >
            <td>{{ ticket.subject }}</td>
            <td>
              <div>{{ ticket.creatorName || '—' }}</div>
              <div class="text-muted">{{ ticket.creatorEmail }}</div>
            </td>
            <td><ProBadge variant="neutral">{{ sourceLabel(ticket.source) }}</ProBadge></td>
            <td><ProBadge :variant="statusVariant(ticket.status)">{{ statusLabel(ticket.status) }}</ProBadge></td>
            <td>{{ ticket.replyCount ?? 0 }}</td>
            <td>{{ formatDate(ticket.createdAt) }}</td>
          </tr>
        </tbody>
      </ProTable>

      <ProKanban v-else data-testid="admin-support-kanban">
        <ProKanbanColumn
          v-for="col in kanbanColumns"
          :key="col.status"
          :title="col.title"
          :count="col.items.length"
          :empty="!col.items.length"
          :empty-title="$t('common.none')"
        >
          <article
            v-for="ticket in col.items"
            :key="ticket.id"
            class="pro-kanban-card support-kanban-card"
            data-testid="admin-support-kanban-card"
          >
            <button
              type="button"
              class="support-kanban-card__title"
              @click="navigateTo(`/admin/support/${ticket.id}`)"
            >
              <strong>{{ ticket.subject }}</strong>
            </button>
            <p class="pro-kanban-card__meta">{{ ticket.creatorEmail || '—' }}</p>
            <div class="pro-flex-gap">
              <ProBadge variant="neutral">{{ sourceLabel(ticket.source) }}</ProBadge>
              <span class="text-muted">{{ formatDate(ticket.createdAt) }}</span>
            </div>
            <label class="support-kanban-card__status">
              <span class="visually-hidden">{{ $t('admin.support.colStatus') }}</span>
              <select
                class="pro-select"
                :value="ticket.status"
                data-testid="admin-support-kanban-status"
                @change="onKanbanStatus(ticket, ($event.target as HTMLSelectElement).value)"
              >
                <option
                  v-for="st in SUPPORT_STATUSES"
                  :key="st"
                  :value="st"
                  :disabled="st !== ticket.status && !canAdvanceSupportStatus(ticket.status, st)"
                >
                  {{ statusLabel(st) }}
                </option>
              </select>
            </label>
          </article>
        </ProKanbanColumn>
      </ProKanban>

      <p v-if="loadError" class="pro-error" role="alert" data-testid="admin-support-load-error">
        {{ $t('admin.support.loadError') }}
      </p>
      <div v-if="viewMode === 'table'" class="support-pager" data-testid="admin-support-pager">
        <p class="pro-hint">{{ pageLabel }}</p>
        <div class="support-pager__actions">
          <ProButton
            type="button"
            variant="secondary"
            test-id="admin-support-prev"
            :disabled="offset <= 0 || loading"
            @click="prevPage"
          >
            {{ $t('admin.support.prevPage') }}
          </ProButton>
          <ProButton
            type="button"
            variant="secondary"
            test-id="admin-support-next"
            :disabled="!hasNext || loading"
            @click="nextPage"
          >
            {{ $t('admin.support.nextPage') }}
          </ProButton>
        </div>
      </div>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import {
  SUPPORT_STATUSES,
  canAdvanceSupportStatus,
  supportStatusBadgeVariant,
} from '~/utils/support-status'

definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

const PAGE_SIZE = 25
const KANBAN_LIMIT = 100

const { t } = useI18n()
const { viewMode } = useListView('pf-admin-support-view', 'table')
const statusFilter = ref('')
const sourceFilter = ref('')
const searchQ = ref('')
const rows = ref<any[]>([])
const total = ref(0)
const offset = ref(0)
const limit = ref(PAGE_SIZE)
const loading = ref(false)
const loadError = ref(false)
const stats = ref<any>(null)
const patchError = ref('')

const hasNext = computed(() => offset.value + limit.value < total.value)
const pageLabel = computed(() => {
  if (total.value === 0) return t('admin.support.pageEmpty')
  const from = offset.value + 1
  const to = Math.min(offset.value + rows.value.length, total.value)
  return t('admin.support.pageInfo', { from, to, total: total.value })
})

const kanbanColumns = computed(() =>
  SUPPORT_STATUSES.map((status) => ({
    status,
    title: statusLabel(status),
    items: rows.value.filter((ticket) => ticket.status === status),
  })),
)

function statusCount(status: string) {
  return Number(stats.value?.byStatus?.[status] ?? 0)
}

function statusLabel(status: string) {
  return t(`admin.support.status_${status}`)
}

function sourceLabel(source: string) {
  return t(`admin.support.source_${source}`)
}

function statusVariant(status: string): 'neutral' | 'success' | 'warning' | 'danger' {
  return supportStatusBadgeVariant(status)
}

function formatDate(iso: string) {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function onFilterChange() {
  offset.value = 0
  void load()
}

async function prevPage() {
  const prev = Math.max(0, offset.value - PAGE_SIZE)
  if (prev === offset.value) return
  const ok = await loadAt(prev)
  if (ok) offset.value = prev
}

async function nextPage() {
  if (!hasNext.value) return
  const next = offset.value + PAGE_SIZE
  const ok = await loadAt(next)
  if (ok) offset.value = next
}

async function load() {
  await Promise.all([loadAt(offset.value), loadStats()])
}

async function loadStats() {
  try {
    const res: any = await $fetch('/api/admin/support/stats')
    stats.value = res.data ?? res
  } catch {
    stats.value = { byStatus: {}, openOlderThan24h: 0, openOlderThan7d: 0 }
  }
}

async function loadAt(targetOffset: number): Promise<boolean> {
  loading.value = true
  loadError.value = false
  try {
    const q = searchQ.value.trim()
    const pageLimit = viewMode.value === 'kanban' ? KANBAN_LIMIT : PAGE_SIZE
    const res: any = await $fetch('/api/admin/support/tickets', {
      query: {
        limit: pageLimit,
        offset: viewMode.value === 'kanban' ? 0 : targetOffset,
        ...(statusFilter.value ? { status: statusFilter.value } : {}),
        ...(sourceFilter.value ? { source: sourceFilter.value } : {}),
        ...(q ? { q } : {}),
      },
    })
    const data = res.data ?? res
    rows.value = data.items ?? []
    total.value = Number(data.total ?? 0)
    limit.value = Number(data.limit ?? pageLimit)
    return true
  } catch {
    loadError.value = true
    return false
  } finally {
    loading.value = false
  }
}

async function onKanbanStatus(ticket: any, status: string) {
  if (!status || status === ticket.status) return
  if (!canAdvanceSupportStatus(ticket.status, status)) {
    patchError.value = t('admin.support.statusTransitionBlocked')
    await load()
    return
  }
  patchError.value = ''
  try {
    await $fetch(`/api/admin/support/tickets/${ticket.id}`, {
      method: 'PATCH',
      body: { status },
    })
    await load()
  } catch {
    patchError.value = t('admin.support.patchError')
    await load()
  }
}

watch(viewMode, () => {
  offset.value = 0
  void load()
})

onMounted(() => { void load() })
</script>

<style scoped>
.support-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: 0.75rem;
}
.support-filters__field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 12rem;
  flex: 1;
}
.support-row {
  cursor: pointer;
}
.support-row:hover td {
  background: var(--pf-vet-bg);
}
.text-muted {
  color: var(--pf-vet-text-muted);
  font-size: 0.85em;
}
.support-pager {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 1rem;
}
.support-pager__actions {
  display: flex;
  gap: 0.5rem;
}
.support-kanban-card {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.support-kanban-card__title {
  background: none;
  border: 0;
  padding: 0;
  text-align: left;
  cursor: pointer;
  color: inherit;
}
.support-kanban-card__status {
  display: block;
  margin-top: 0.25rem;
}
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
