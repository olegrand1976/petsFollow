<template>
  <div data-testid="consultation-detail-page" class="consultation-detail-redirect">
    <p class="pro-hint" data-testid="consultation-detail-loading">
      {{ $t('common.loading') }}
    </p>
  </div>
</template>

<script setup lang="ts">
/**
 * Deep-link /consultations/:id → ouvre la coque modale globale (même que Nouvelle consultation)
 * puis revient sur la liste /consultations.
 */
definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'consultations.history.read' })

const route = useRoute()
const { mapError } = useApiError()
const { t } = useI18n()
const active = useActiveConsultation()
const loadError = ref('')

const visitId = computed(() => {
  const raw = route.params.id
  return typeof raw === 'string' ? raw.trim() : ''
})

onMounted(async () => {
  const id = visitId.value
  if (!id) {
    await navigateTo('/consultations', { replace: true })
    return
  }
  try {
    const res: any = await $fetch(`/api/vet/consultations/${encodeURIComponent(id)}`)
    const row = (res?.data ?? res) as {
      id?: string
      clientId?: string
      petId?: string
      consultationSession?: boolean
    }
    if (row?.id && row.clientId) {
      active.openForVisit({
        visitId: row.id,
        clientId: row.clientId,
        petId: row.petId,
        keepVisit: !row.consultationSession,
      })
    }
  }
  catch (e: any) {
    loadError.value = mapError(e) || t('consultations.detailLoadError')
  }
  await navigateTo('/consultations', { replace: true })
})
</script>

<style scoped>
.consultation-detail-redirect {
  min-height: 4rem;
}
</style>
