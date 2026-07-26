<template>
  <div data-testid="admin-support-page">
    <ProPageHeader :title="$t('admin.support.title')" :subtitle="$t('admin.support.subtitle')" />
    <ProCard>
      <div class="pro-mb-md support-filters">
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
            <option value="resolved">{{ $t('admin.support.filterResolved') }}</option>
            <option value="closed">{{ $t('admin.support.filterClosed') }}</option>
          </select>
        </div>
        <ProButton type="button" variant="secondary" test-id="admin-support-search-btn" :disabled="loading" @click="onFilterChange">
          {{ $t('admin.support.searchSubmit') }}
        </ProButton>
      </div>
      <ProTable :empty="!rows.length && !loadError" :empty-title="$t('admin.support.empty')">
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
            v-for="t in rows"
            :key="t.id"
            class="support-row"
            data-testid="admin-support-row"
            @click="navigateTo(`/admin/support/${t.id}`)"
          >
            <td>{{ t.subject }}</td>
            <td>
              <div>{{ t.creatorName || '—' }}</div>
              <div class="text-muted">{{ t.creatorEmail }}</div>
            </td>
            <td><ProBadge variant="neutral">{{ sourceLabel(t.source) }}</ProBadge></td>
            <td><ProBadge :variant="statusVariant(t.status)">{{ statusLabel(t.status) }}</ProBadge></td>
            <td>{{ t.replyCount ?? 0 }}</td>
            <td>{{ formatDate(t.createdAt) }}</td>
          </tr>
        </tbody>
      </ProTable>
      <p v-if="loadError" class="pro-error" role="alert" data-testid="admin-support-load-error">
        {{ $t('admin.support.loadError') }}
      </p>
      <div class="support-pager" data-testid="admin-support-pager">
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
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const PAGE_SIZE = 25

const { t } = useI18n()
const statusFilter = ref('')
const searchQ = ref('')
const rows = ref<any[]>([])
const total = ref(0)
const offset = ref(0)
const limit = ref(PAGE_SIZE)
const loading = ref(false)
const loadError = ref(false)

const hasNext = computed(() => offset.value + limit.value < total.value)
const pageLabel = computed(() => {
  if (total.value === 0) return t('admin.support.pageEmpty')
  const from = offset.value + 1
  const to = Math.min(offset.value + rows.value.length, total.value)
  return t('admin.support.pageInfo', { from, to, total: total.value })
})

function statusLabel(status: string) {
  return t(`admin.support.status_${status}`)
}

function sourceLabel(source: string) {
  return t(`admin.support.source_${source}`)
}

function statusVariant(status: string): 'neutral' | 'success' | 'warning' | 'danger' {
  switch (status) {
    case 'open':
      return 'danger'
    case 'in_progress':
      return 'warning'
    case 'resolved':
      return 'success'
    case 'closed':
      return 'neutral'
    default:
      return 'neutral'
  }
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
  await loadAt(offset.value)
}

async function loadAt(targetOffset: number): Promise<boolean> {
  loading.value = true
  loadError.value = false
  try {
    const q = searchQ.value.trim()
    const res: any = await $fetch('/api/admin/support/tickets', {
      query: {
        limit: PAGE_SIZE,
        offset: targetOffset,
        ...(statusFilter.value ? { status: statusFilter.value } : {}),
        ...(q ? { q } : {}),
      },
    })
    const data = res.data ?? res
    rows.value = data.items ?? []
    total.value = Number(data.total ?? 0)
    limit.value = Number(data.limit ?? PAGE_SIZE)
    return true
  } catch {
    loadError.value = true
    return false
  } finally {
    loading.value = false
  }
}

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
</style>
