<template>
  <div data-testid="prescriptions-page">
    <ProPageHeader
      :title="$t('prescriptions.title')"
      :subtitle="$t('prescriptions.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="prescriptions-page-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton
          v-if="canWriteClinical"
          variant="primary"
          test-id="prescriptions-new"
          @click="navigateTo('/prescriptions/nouveau')"
        >
          {{ $t('prescriptions.new') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p class="pro-hint">{{ $t('prescriptions.v1Hint') }}</p>
    <p v-if="loadError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ loadError }}</p>
    <p v-if="actionError" class="pro-inline-feedback pro-inline-feedback--error" role="alert" data-testid="prescriptions-pdf-error">{{ actionError }}</p>

    <ProCard>
      <ProListToolbar :show-view-toggle="false">
        <template #filters>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="rx-search">{{ $t('prescriptions.search') }}</label>
            <input
              id="rx-search"
              v-model="query"
              type="search"
              class="pro-input"
              :placeholder="$t('prescriptions.searchPlaceholder')"
              data-testid="prescriptions-search"
            >
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="rx-status">{{ $t('prescriptions.statusFilter') }}</label>
            <select
              id="rx-status"
              v-model="statusFilter"
              class="pro-select"
              data-testid="prescriptions-status-filter"
            >
              <option value="">{{ $t('prescriptions.statusAll') }}</option>
              <option value="draft">{{ $t('prescriptions.statusDraft') }}</option>
              <option value="signed">{{ $t('prescriptions.statusSigned') }}</option>
              <option value="sent">{{ $t('prescriptions.statusSent') }}</option>
              <option value="archived">{{ $t('prescriptions.statusArchived') }}</option>
            </select>
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="rx-from">{{ $t('prescriptions.from') }}</label>
            <input
              id="rx-from"
              v-model="fromDate"
              type="date"
              class="pro-input"
              data-testid="prescriptions-from"
            >
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="rx-to">{{ $t('prescriptions.to') }}</label>
            <input
              id="rx-to"
              v-model="toDate"
              type="date"
              class="pro-input"
              data-testid="prescriptions-to"
            >
          </div>
        </template>
      </ProListToolbar>

      <ProTable
        :empty="!rows.length"
        :empty-title="$t('prescriptions.emptyTitle')"
        :empty-description="$t('prescriptions.emptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('prescriptions.columnDate') }}</th>
            <th>{{ $t('prescriptions.columnClient') }}</th>
            <th>{{ $t('prescriptions.columnPet') }}</th>
            <th>{{ $t('prescriptions.columnStatus') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in rows"
            :key="row.id"
            :data-testid="`prescriptions-row-${row.id}`"
          >
            <td>{{ formatDate(row.createdAt) }}</td>
            <td>
              <NuxtLink
                v-if="row.ownerId"
                :to="`/clients/${row.ownerId}`"
              >
                {{ row.ownerName || $t('common.dash') }}
              </NuxtLink>
              <template v-else>{{ row.ownerName || $t('common.dash') }}</template>
            </td>
            <td>
              <NuxtLink
                v-if="row.ownerId && row.petId"
                :to="`/clients/${row.ownerId}/pets/${row.petId}`"
              >
                {{ row.petName || $t('common.dash') }}
              </NuxtLink>
              <template v-else>{{ row.petName || $t('common.dash') }}</template>
            </td>
            <td>
              <ProBadge :variant="statusVariant(row.status)">
                {{ statusLabel(row.status) }}
              </ProBadge>
            </td>
            <td>
              <div class="pro-flex-gap">
                <ProIconAction
                  icon="description"
                  :label="$t('prescriptions.openDetail')"
                  :to="`/prescriptions/${row.id}`"
                  :test-id="`prescriptions-open-${row.id}`"
                />
                <ProIconAction
                  icon="picture_as_pdf"
                  :label="$t('prescriptions.openPdf')"
                  :test-id="`prescriptions-pdf-${row.id}`"
                  @click="openPdf(row)"
                />
                <ProIconAction
                  v-if="row.ownerId && row.petId"
                  icon="pets"
                  :label="$t('common.profile')"
                  :to="`/clients/${row.ownerId}/pets/${row.petId}`"
                  :test-id="`prescriptions-pet-link-${row.id}`"
                />
                <ProIconAction
                  v-if="canDeleteDraft(row)"
                  icon="delete"
                  :label="$t('prescriptions.delete')"
                  :test-id="`prescriptions-delete-${row.id}`"
                  @click="askDelete(row)"
                />
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
      <p
        v-if="rows.length >= 200"
        class="pro-hint"
        data-testid="prescriptions-limit-hint"
      >
        {{ $t('prescriptions.limitHint') }}
      </p>
    </ProCard>

    <ProModal
      v-model:open="deleteOpen"
      size="md"
      :title="$t('prescriptions.deleteConfirmTitle')"
      test-id="prescriptions-delete-modal"
    >
      <p>{{ $t('prescriptions.deleteConfirmBody') }}</p>
      <p v-if="deleteError" class="pro-error" role="alert">{{ deleteError }}</p>
      <template #footer>
        <ProButton variant="secondary" test-id="prescriptions-delete-cancel" @click="deleteOpen = false">
          {{ $t('common.cancel') }}
        </ProButton>
        <ProButton
          variant="primary"
          :disabled="deleteBusy"
          test-id="prescriptions-delete-confirm"
          @click="confirmDelete"
        >
          {{ $t('prescriptions.deleteConfirmCta') }}
        </ProButton>
      </template>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
import { openPrescriptionPdfBlob } from '~/utils/prescription-pdf'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pets.read' })

type PrescriptionRow = {
  id: string
  petId: string
  ownerId?: string
  petName?: string
  ownerName?: string
  status: string
  createdAt?: string
}

const { t, locale } = useI18n()
const { mapError } = usePrescriptionError()
const { canPractice } = usePracticePerms()
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))

const rows = ref<PrescriptionRow[]>([])
const loadError = ref('')
const query = ref('')
const statusFilter = ref('')
const fromDate = ref('')
const toDate = ref('')
const loading = ref(false)

const deleteOpen = ref(false)
const deleteBusy = ref(false)
const deleteError = ref('')
const pendingDeleteId = ref('')
const pdfBusyId = ref('')
const actionError = ref('')

function canDeleteDraft(row: PrescriptionRow) {
  return canWriteClinical.value && row.status === 'draft'
}

function askDelete(row: PrescriptionRow) {
  pendingDeleteId.value = row.id
  deleteError.value = ''
  deleteOpen.value = true
}

async function confirmDelete() {
  const id = pendingDeleteId.value
  if (!id || deleteBusy.value) return
  deleteBusy.value = true
  deleteError.value = ''
  try {
    await $fetch(`/api/vet/prescriptions/${id}`, { method: 'DELETE' })
    deleteOpen.value = false
    pendingDeleteId.value = ''
    rows.value = rows.value.filter(r => r.id !== id)
  }
  catch (e: any) {
    deleteError.value = mapError(e)
  }
  finally {
    deleteBusy.value = false
  }
}

async function openPdf(row: PrescriptionRow) {
  if (!row.id || pdfBusyId.value) return
  pdfBusyId.value = row.id
  actionError.value = ''
  try {
    await openPrescriptionPdfBlob(row.id)
  }
  catch (e: any) {
    actionError.value = mapError(e) || t('prescriptions.errorPdf')
  }
  finally {
    pdfBusyId.value = ''
  }
}

function formatDate(iso?: string) {
  if (!iso) return t('common.dash')
  try {
    return new Intl.DateTimeFormat(locale.value || 'fr', {
      dateStyle: 'short',
      timeStyle: 'short',
    }).format(new Date(iso))
  }
  catch {
    return iso
  }
}

function statusLabel(s?: string) {
  if (s === 'draft') return t('prescriptions.statusDraft')
  if (s === 'signed') return t('prescriptions.statusSigned')
  if (s === 'sent') return t('prescriptions.statusSent')
  if (s === 'archived') return t('prescriptions.statusArchived')
  return s || ''
}

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'draft': return 'warning'
    case 'signed': return 'success'
    case 'sent': return 'success'
    case 'archived': return 'neutral'
    default: return 'neutral'
  }
}

/** Civil day bounds in Europe/Brussels → RFC3339 UTC. */
function dateBounds(dateStr: string, endOfDay: boolean): string | undefined {
  if (!dateStr || !/^\d{4}-\d{2}-\d{2}$/.test(dateStr)) return undefined
  const y = Number(dateStr.slice(0, 4))
  const m = Number(dateStr.slice(5, 7))
  const d = Number(dateStr.slice(8, 10))
  if (!y || !m || !d) return undefined
  const h = endOfDay ? 23 : 0
  const mi = endOfDay ? 59 : 0
  const s = endOfDay ? 59 : 0
  const ms = endOfDay ? 999 : 0
  const utcGuess = new Date(Date.UTC(y, m - 1, d, h, mi, s, ms))
  const dtf = new Intl.DateTimeFormat('en-US', {
    timeZone: 'Europe/Brussels',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hourCycle: 'h23',
  })
  const parts = Object.fromEntries(
    dtf.formatToParts(utcGuess)
      .filter(p => p.type !== 'literal')
      .map(p => [p.type, p.value]),
  ) as Record<string, string>
  const asBrusselsMs = Date.UTC(
    Number(parts.year),
    Number(parts.month) - 1,
    Number(parts.day),
    Number(parts.hour),
    Number(parts.minute),
    Number(parts.second),
    ms,
  )
  const offset = asBrusselsMs - utcGuess.getTime()
  return new Date(utcGuess.getTime() - offset).toISOString()
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const params = new URLSearchParams()
    if (query.value.trim()) params.set('q', query.value.trim())
    if (statusFilter.value) params.set('status', statusFilter.value)
    const from = dateBounds(fromDate.value, false)
    const to = dateBounds(toDate.value, true)
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    const qs = params.toString()
    const res: any = await $fetch(`/api/vet/prescriptions${qs ? `?${qs}` : ''}`)
    const data = res?.data ?? res
    rows.value = Array.isArray(data?.items) ? data.items : []
  }
  catch (e: any) {
    loadError.value = mapError(e) || t('prescriptions.loadError')
    rows.value = []
  }
  finally {
    loading.value = false
  }
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function scheduleLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { void load() }, 250)
}

watch([query, statusFilter, fromDate, toDate], scheduleLoad)

onMounted(() => { void load() })
onBeforeUnmount(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style scoped>
.pro-hint { margin-bottom: 1rem; color: var(--pf-vet-muted, #667); }
</style>
