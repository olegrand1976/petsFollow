<template>
  <div data-testid="daf-detail-page">
    <ProPageHeader
      :title="doc?.displayNumber || $t('pharmacy.daf.title')"
      :subtitle="statusLabel(doc?.status)"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton
          v-if="doc?.status === 'finalized' || doc?.status === 'cancelled'"
          variant="secondary"
          test-id="daf-open-pdf"
          @click="openPdf"
        >
          {{ $t('pharmacy.daf.openPdf') }}
        </ProButton>
        <ProButton
          v-if="doc?.status === 'finalized' && canWritePharmacy"
          variant="ghost"
          test-id="daf-cancel-btn"
          :disabled="busy"
          @click="cancel"
        >
          {{ $t('pharmacy.daf.cancel') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p v-if="error" class="pro-alert">{{ error }}</p>

    <ProCard v-if="doc">
      <p class="pro-hint">{{ doc.practiceName }} · {{ doc.prescriberName }}</p>
      <p v-if="doc.hasAntibiotic" class="pro-hint">VAMReg: {{ doc.vamregStatus }}</p>
      <table class="pro-table">
        <thead>
          <tr>
            <th>{{ $t('pharmacy.stock.colMed') }}</th>
            <th>{{ $t('pharmacy.daf.amm') }}</th>
            <th>{{ $t('pharmacy.stock.lot') }}</th>
            <th>{{ $t('pharmacy.stock.expiresOn') }}</th>
            <th>{{ $t('pharmacy.stock.qty') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="it in doc.items || []" :key="it.id">
            <td>{{ it.medicationName }} <span class="pro-hint">{{ it.medicationCnk }}</span></td>
            <td>{{ it.ammNumber }}</td>
            <td>{{ it.lotNumber || '—' }}</td>
            <td>{{ it.expiresOn || '—' }}</td>
            <td>{{ it.qty }} {{ it.unit }}</td>
          </tr>
        </tbody>
      </table>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['auth', 'vet-only', 'practice-perm'], practicePerm: 'pharmacy.read' })
const { t } = useI18n()
const { canPractice } = usePracticePerms()
const canWritePharmacy = computed(() => canPractice('pharmacy.write'))
const route = useRoute()
const busy = ref(false)
const error = ref('')
const doc = ref<any>(null)

function unwrap(res: any) {
  return res?.data ?? res
}

function statusLabel(s?: string) {
  if (s === 'draft') return t('pharmacy.daf.statusDraft')
  if (s === 'finalized') return t('pharmacy.daf.statusFinalized')
  if (s === 'cancelled') return t('pharmacy.daf.statusCancelled')
  return s || ''
}

async function load() {
  const res = await $fetch<any>(`/api/vet/pharmacy/daf/${route.params.id}`)
  doc.value = unwrap(res)
}

function openPdf() {
  window.open(`/api/vet/pharmacy/daf/${route.params.id}/pdf`, '_blank')
}

async function cancel() {
  if (!confirm(t('pharmacy.daf.cancelConfirm'))) return
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>(`/api/vet/pharmacy/daf/${route.params.id}/cancel`, {
      method: 'POST',
      body: { reason: 'manual', restock: true },
    })
    doc.value = unwrap(res)
  }
  catch (e: any) {
    error.value = e?.data?.error?.code || t('pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

onMounted(async () => {
  try {
    await load()
  }
  catch (e: any) {
    error.value = e?.data?.error?.message || t('pharmacy.daf.error')
  }
})
</script>
