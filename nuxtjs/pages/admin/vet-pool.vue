<template>
  <div data-testid="admin-vet-pool-page">
    <ProPageHeader :title="$t('admin.vetPool.title')" :subtitle="$t('admin.vetPool.subtitle')" />

    <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
    <p v-if="actionMsg" class="pro-hint">{{ actionMsg }}</p>

    <ProCard>
      <ProTable :empty="!rows.length" :empty-title="$t('admin.vetPool.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.vetPool.colName') }}</th>
            <th>{{ $t('admin.vetPool.colEmail') }}</th>
            <th>{{ $t('admin.vetPool.colPractice') }}</th>
            <th>{{ $t('admin.vetPool.colCity') }}</th>
            <th>{{ $t('admin.vetPool.colPostal') }}</th>
            <th>{{ $t('admin.vetPool.colCreated') }}</th>
            <th>{{ $t('admin.vetPool.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="v in rows" :key="v.userId" :data-testid="`vet-pool-row-${v.userId}`">
            <td>{{ v.fullName }}</td>
            <td>{{ v.email }}</td>
            <td>{{ v.practiceName || '—' }}</td>
            <td>{{ v.city || '—' }}</td>
            <td>{{ v.postalCode || '—' }}</td>
            <td>{{ formatDate(v.createdAt) }}</td>
            <td>
              <ProButton
                type="button"
                variant="secondary"
                :loading="suggestLoading === v.userId"
                :test-id="`vet-pool-suggest-${v.userId}`"
                @click="loadSuggestions(v)"
              >
                {{ $t('admin.vetPool.suggest') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard v-if="selectedVet" class="pro-mt-lg" data-testid="vet-pool-suggestions">
      <h3 class="pro-mb-md">
        {{ $t('admin.vetPool.suggestionsTitle', { name: selectedVet.fullName }) }}
      </h3>
      <p v-if="!suggestions.length" class="pro-hint">{{ $t('admin.vetPool.suggestionsEmpty') }}</p>
      <ul v-else class="vet-pool-suggestions">
        <li v-for="s in suggestions" :key="s.userId" class="vet-pool-suggestions__item">
          <div>
            <strong>{{ s.fullName }}</strong>
            <span class="pro-hint"> — {{ s.email }}</span>
            <p class="pro-hint">{{ s.note }}</p>
            <p class="pro-hint">
              {{ $t('admin.vetPool.score', { score: s.score }) }}
              <template v-if="s.distanceKm != null">
                · {{ $t('admin.vetPool.distance', { km: s.distanceKm }) }}
              </template>
              · {{ $t('admin.vetPool.assignedLoad', { n: s.assignedVets }) }}
            </p>
          </div>
          <ProButton
            type="button"
            :loading="assignLoading === s.userId"
            :test-id="`vet-pool-assign-${s.userId}`"
            @click="assign(s.userId)"
          >
            {{ $t('admin.vetPool.assign') }}
          </ProButton>
        </li>
      </ul>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

const { t } = useI18n()
const { mapError } = useApiError()

type VetRow = {
  userId: string
  fullName: string
  email: string
  practiceName?: string
  city?: string
  postalCode?: string
  createdAt: string
}

type Suggestion = {
  userId: string
  fullName: string
  email: string
  score: number
  note: string
  distanceKm?: number
  assignedVets: number
}

const rows = ref<VetRow[]>([])
const selectedVet = ref<VetRow | null>(null)
const suggestions = ref<Suggestion[]>([])
const error = ref('')
const actionMsg = ref('')
const suggestLoading = ref('')
const assignLoading = ref('')

function formatDate(raw: string) {
  if (!raw) return '—'
  try {
    return new Date(raw).toLocaleDateString()
  } catch {
    return raw
  }
}

async function load() {
  error.value = ''
  try {
    const res: any = await $fetch('/api/admin/vets/unassigned')
    rows.value = res.data ?? res ?? []
  } catch (e) {
    error.value = mapError(e)
  }
}

async function loadSuggestions(v: VetRow) {
  actionMsg.value = ''
  error.value = ''
  selectedVet.value = v
  suggestions.value = []
  suggestLoading.value = v.userId
  try {
    const res: any = await $fetch(`/api/admin/vets/${v.userId}/assign-suggestions`)
    suggestions.value = res.data ?? res ?? []
  } catch (e) {
    error.value = mapError(e)
  } finally {
    suggestLoading.value = ''
  }
}

async function assign(commercialId: string) {
  if (!selectedVet.value) return
  actionMsg.value = ''
  error.value = ''
  assignLoading.value = commercialId
  try {
    await $fetch(`/api/admin/commercials/${commercialId}/assign`, {
      method: 'PATCH',
      body: { vetUserId: selectedVet.value.userId },
    })
    actionMsg.value = t('admin.vetPool.assignOk')
    selectedVet.value = null
    suggestions.value = []
    await load()
  } catch (e) {
    error.value = mapError(e)
  } finally {
    assignLoading.value = ''
  }
}

onMounted(load)
</script>

<style scoped>
.vet-pool-suggestions {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 1rem;
}
.vet-pool-suggestions__item {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--pf-vet-border);
}
.vet-pool-suggestions__item:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}
</style>
