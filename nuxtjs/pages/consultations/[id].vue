<template>
  <div data-testid="consultation-detail-page" class="consultation-detail">
    <ProPageHeader :title="pageTitle" :subtitle="pageSubtitle">
      <template #actions>
        <ProButton
          variant="secondary"
          test-id="consultation-detail-back"
          @click="navigateTo('/consultations')"
        >
          {{ $t('consultations.backToList') }}
        </ProButton>
        <ProButton
          v-if="canReadPets && meta?.clientId && meta?.petId"
          variant="ghost"
          test-id="consultation-detail-pet"
          @click="navigateTo(`/clients/${meta.clientId}/pets/${meta.petId}`)"
        >
          {{ $t('common.profile') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p v-if="loadError" class="pro-inline-feedback pro-inline-feedback--error" role="alert" data-testid="consultation-detail-error">
      {{ loadError }}
    </p>
    <p v-else-if="loading" class="pro-hint" data-testid="consultation-detail-loading">
      {{ $t('common.loading') }}
    </p>

    <ProCard v-else-if="visitId && meta && canReadPets" class="consultation-detail__card">
      <ProVisitReportPanel
        fill-height
        :visit-id="visitId"
        :visit-scheduled-at="meta.scheduledAt || meta.createdAt || undefined"
        :readonly="!canWriteClinical"
      />
    </ProCard>
    <p v-else-if="!canReadPets" class="pro-inline-feedback pro-inline-feedback--error" role="alert">
      {{ $t('consultations.detailForbidden') }}
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'consultations.history.read' })

type VisitMeta = {
  id: string
  petId?: string
  clientId?: string
  petName?: string
  clientName?: string
  scheduledAt?: string
  createdAt?: string
  status?: string
}

const route = useRoute()
const { t } = useI18n()
const { mapError } = useApiError()
const { formatDate } = useFormatters()
const { canPractice } = usePracticePerms()
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))
const canReadPets = computed(() => canPractice('pets.read'))

const visitId = computed(() => {
  const raw = route.params.id
  return typeof raw === 'string' ? raw.trim() : ''
})

const loading = ref(true)
const loadError = ref('')
const meta = ref<VisitMeta | null>(null)

const pageTitle = computed(() => {
  if (meta.value?.petName) {
    return t('consultations.detailTitleNamed', { pet: meta.value.petName })
  }
  return t('consultations.detailTitle')
})

const pageSubtitle = computed(() => {
  const parts: string[] = []
  if (meta.value?.clientName) parts.push(meta.value.clientName)
  const when = meta.value?.scheduledAt || meta.value?.createdAt
  if (when) parts.push(formatDate(when))
  return parts.length ? parts.join(' · ') : t('consultations.detailSubtitle')
})

async function loadMeta() {
  loading.value = true
  loadError.value = ''
  meta.value = null
  const id = visitId.value
  if (!id) {
    loadError.value = t('consultations.detailNotFound')
    loading.value = false
    return
  }
  try {
    const res: any = await $fetch(`/api/vet/consultations/${encodeURIComponent(id)}`)
    const row = (res?.data ?? res) as VisitMeta
    if (!row?.id) {
      loadError.value = t('consultations.detailNotFound')
      return
    }
    meta.value = row
  }
  catch (e: any) {
    const status = e?.statusCode ?? e?.status ?? e?.data?.statusCode
    if (status === 404) {
      loadError.value = t('consultations.detailNotFound')
    }
    else if (status === 403) {
      loadError.value = t('consultations.detailForbidden')
    }
    else {
      loadError.value = mapError(e) || t('consultations.detailLoadError')
    }
  }
  finally {
    loading.value = false
  }
}

watch(visitId, () => { void loadMeta() }, { immediate: true })
</script>

<style scoped>
.consultation-detail__card {
  min-height: min(70vh, 720px);
  display: flex;
  flex-direction: column;
}
.consultation-detail__card :deep(.pro-visit-report) {
  flex: 1;
  min-height: 0;
}
</style>
