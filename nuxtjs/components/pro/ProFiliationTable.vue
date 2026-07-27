<template>
  <div data-testid="filiation-table">
    <ProListToolbar>
      <template #filters>
        <ProInput
          v-model="search"
          test-id="filiation-search"
          :label="t('filiation.search')"
          :placeholder="t('filiation.searchPlaceholder')"
        />
        <div v-if="showBranchFilter" class="pro-field">
          <label class="pro-label" for="filiation-branch">{{ t('filiation.branch') }}</label>
          <select
            id="filiation-branch"
            v-model="branchId"
            class="pro-select"
            data-testid="filiation-branch"
          >
            <option value="">{{ t('filiation.branchAll') }}</option>
            <option v-for="b in branches" :key="b.id" :value="b.id">
              {{ b.name }} ({{ b.code }})
            </option>
          </select>
        </div>
      </template>
    </ProListToolbar>

    <p v-if="loading" class="pro-hint" data-testid="filiation-loading">{{ t('common.loading') }}</p>
    <p v-else-if="loadError" class="pro-error" data-testid="filiation-error">{{ loadError }}</p>

    <ProTable
      v-else
      :empty="!rows.length"
      :empty-title="t('filiation.empty')"
    >
      <thead>
        <tr>
          <th v-if="showOrgCols">{{ t('filiation.branch') }}</th>
          <th v-if="showOrgCols">{{ t('filiation.commercial') }}</th>
          <th>{{ t('filiation.vet') }}</th>
          <th>{{ t('filiation.practice') }}</th>
          <th>{{ t('filiation.client') }}</th>
          <th>{{ t('filiation.inviteCode') }}</th>
          <th>{{ t('filiation.sponsor') }}</th>
          <th>{{ t('filiation.effective') }}</th>
          <th>{{ t('filiation.linkedAt') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(r, idx) in rows"
          :key="rowKey(r, idx)"
          :data-testid="`filiation-row-${idx}`"
        >
          <td v-if="showOrgCols">{{ r.branchName || '—' }}</td>
          <td v-if="showOrgCols">
            <div>{{ r.commercialName }}</div>
            <div class="pro-hint">{{ r.commercialEmail }}</div>
          </td>
          <td>
            <template v-if="r.vetName">
              <div>{{ r.vetName }}</div>
              <div class="pro-hint">{{ r.vetEmail }}</div>
            </template>
            <template v-else>—</template>
          </td>
          <td>{{ r.practiceName || '—' }}</td>
          <td>
            <template v-if="r.clientName">
              <div>{{ r.clientName }}</div>
              <div class="pro-hint">{{ r.clientEmail }}</div>
            </template>
            <template v-else>—</template>
          </td>
          <td>{{ r.inviteCode || '—' }}</td>
          <td>{{ r.sponsorClientName || '—' }}</td>
          <td>
            <ProBadge :variant="sourceVariant(r.effectiveSource)" data-testid="filiation-effective-badge">
              {{ sourceLabel(r.effectiveSource) }}
            </ProBadge>
            <div v-if="r.effectiveCommercialName" class="pro-hint">{{ r.effectiveCommercialName }}</div>
          </td>
          <td>{{ formatDt(r.linkedAt) }}</td>
        </tr>
      </tbody>
    </ProTable>
  </div>
</template>

<script setup lang="ts">
export type FiliationRow = {
  branchId?: string
  branchName?: string
  commercialUserId: string
  commercialName: string
  commercialEmail: string
  vetUserId?: string
  vetName?: string
  vetEmail?: string
  practiceName?: string
  clientUserId?: string
  clientName?: string
  clientEmail?: string
  inviteCode?: string
  sponsorClientName?: string
  effectiveCommercialId?: string
  effectiveCommercialName?: string
  effectiveSource: string
  linkedAt?: string
}

type BranchOpt = { id: string, name: string, code: string }

const props = withDefaults(defineProps<{
  apiPath: string
  showOrgCols?: boolean
  showBranchFilter?: boolean
}>(), {
  showOrgCols: false,
  showBranchFilter: false,
})

const { t } = useI18n()
const rows = ref<FiliationRow[]>([])
const branches = ref<BranchOpt[]>([])
const search = ref('')
const branchId = ref('')
const loading = ref(false)
const loadError = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

function rowKey(r: FiliationRow, idx: number) {
  return `${r.commercialUserId}-${r.vetUserId || ''}-${r.clientUserId || ''}-${idx}`
}

function sourceLabel(source: string) {
  if (source === 'vet_assignment') return t('filiation.source.vet_assignment')
  if (source === 'client_referral') return t('filiation.source.client_referral')
  if (source === 'none') return t('filiation.source.none')
  return source
}

function sourceVariant(source: string): 'success' | 'warning' | 'neutral' {
  if (source === 'vet_assignment') return 'success'
  if (source === 'client_referral') return 'warning'
  return 'neutral'
}

function formatDt(raw?: string) {
  if (!raw) return '—'
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString()
}

async function loadBranches() {
  if (!props.showBranchFilter) return
  try {
    const res: any = await $fetch('/api/admin/sales-branches')
    const list = res.data ?? res ?? []
    branches.value = Array.isArray(list) ? list : []
  } catch {
    branches.value = []
  }
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const query: Record<string, string> = {}
    const q = search.value.trim()
    if (q) query.q = q
    if (branchId.value.trim()) query.branchId = branchId.value.trim()
    const res: any = await $fetch(props.apiPath, { query })
    rows.value = res.data ?? res ?? []
  } catch (e: any) {
    rows.value = []
    loadError.value = e?.data?.error?.message || e?.message || t('filiation.loadError')
  } finally {
    loading.value = false
  }
}

function scheduleSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { void load() }, 300)
}

watch(search, () => { scheduleSearch() })
watch(branchId, () => { void load() })

onMounted(async () => {
  await loadBranches()
  await load()
})

onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
</script>
