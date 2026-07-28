<template>
  <div data-testid="ordonnances-page">
    <ProPageHeader
      :title="$t('prescriptions.title')"
      :subtitle="$t('prescriptions.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="ordonnances-page-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton
          v-if="canWriteClinical"
          variant="primary"
          test-id="ordonnances-new"
          @click="navigateTo('/ordonnances/nouveau')"
        >
          {{ $t('prescriptions.new') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p class="pro-hint">{{ $t('prescriptions.v1Hint') }}</p>
    <p v-if="loading" class="pro-hint" data-testid="ordonnances-loading">{{ $t('prescriptions.loading') }}</p>
    <p v-if="error" class="pro-alert">{{ error }}</p>

    <ProCard v-if="!loading">
      <div v-if="!items.length && !error" class="pro-empty" data-testid="ordonnances-empty">
        {{ $t('prescriptions.empty') }}
      </div>
      <table v-else-if="items.length" class="pro-table" data-testid="ordonnances-table">
        <thead>
          <tr>
            <th>{{ $t('prescriptions.colPet') }}</th>
            <th>{{ $t('prescriptions.colOwner') }}</th>
            <th>{{ $t('prescriptions.colStatus') }}</th>
            <th>{{ $t('prescriptions.colDate') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in items"
            :key="row.id"
            class="rx-row"
            :data-testid="`ordonnances-row-${row.id}`"
            @click="navigateTo(`/ordonnances/${row.id}`)"
          >
            <td>{{ row.petName || '—' }}</td>
            <td>{{ row.ownerName || '—' }}</td>
            <td><ProBadge variant="warning">{{ statusLabel(row.status) }}</ProBadge></td>
            <td>{{ row.createdAt?.slice?.(0, 16) || row.createdAt }}</td>
          </tr>
        </tbody>
      </table>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['auth', 'vet-only', 'practice-perm'], practicePerm: 'pets.read' })
const { t } = useI18n()
const { canPractice } = usePracticePerms()
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))
const { mapError } = usePrescriptionError()
const error = ref('')
const loading = ref(true)
const items = ref<any[]>([])

function unwrap(res: any) {
  return res?.data ?? res
}

function statusLabel(s?: string) {
  if (s === 'draft') return t('prescriptions.statusDraft')
  if (s === 'signed') return t('prescriptions.statusSigned')
  if (s === 'sent') return t('prescriptions.statusSent')
  if (s === 'archived') return t('prescriptions.statusArchived')
  return s || ''
}

onMounted(async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/prescriptions')
    const data = unwrap(res)
    items.value = data?.items ?? []
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.rx-row { cursor: pointer; }
.rx-row:hover { background: var(--pf-vet-surface); }
.pro-hint { margin-bottom: 1rem; color: var(--pf-vet-muted, #667); }
</style>
