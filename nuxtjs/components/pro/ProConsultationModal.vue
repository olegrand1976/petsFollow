<template>
  <ProModal
    :open="open"
    size="lg"
    :title="$t('clients.consultation.title')"
    test-id="consultation-modal"
    @update:open="onOpenUpdate"
  >
    <!-- Étape A : choisir l'animal puis démarrer la visite -->
    <div v-if="!visitId" class="pro-form" data-testid="consultation-setup">
      <p class="pro-hint pro-mb-md">{{ $t('clients.consultation.hint') }}</p>
      <p v-if="loadError" class="pro-error" role="alert">{{ loadError }}</p>
      <p v-if="flowError" class="pro-error" role="alert">{{ flowError }}</p>

      <div class="pro-field">
        <label class="pro-label" for="consultation-pet">{{ $t('clients.consultation.pet') }}</label>
        <select
          id="consultation-pet"
          v-model="selectedPetId"
          class="pro-select"
          data-testid="consultation-pet-select"
          :disabled="petsLoading || starting"
        >
          <option value="" disabled>{{ $t('clients.consultation.petPlaceholder') }}</option>
          <option v-for="p in pets" :key="p.id" :value="p.id">
            {{ p.name }}{{ p.species ? ` (${p.species})` : '' }}
          </option>
        </select>
      </div>

      <ProInput
        v-model="notes"
        :label="$t('clients.consultation.notes')"
        test-id="consultation-notes"
      />
    </div>

    <!-- Étape B : CR médical (dictée Gemini via ProVisitReportPanel) -->
    <div v-else data-testid="consultation-report">
      <p class="pro-hint pro-mb-md">{{ $t('clients.consultation.reportHint') }}</p>
      <ProVisitReportPanel
        :visit-id="visitId"
        @saved="onReportSaved"
        @finalized="onReportSaved"
      />
    </div>

    <template #footer>
      <ProButton variant="ghost" test-id="consultation-cancel" @click="close">
        {{ $t('common.cancel') }}
      </ProButton>

      <template v-if="!visitId">
        <ProButton
          test-id="consultation-start"
          :loading="starting"
          :disabled="!selectedPetId || starting || petsLoading"
          @click="start"
        >
          {{ $t('clients.consultation.start') }}
        </ProButton>
      </template>

      <!-- Étape C : CTA post-enregistrement -->
      <template v-else-if="reportSaved">
        <ProButton
          v-if="pharmacyEnabled"
          variant="secondary"
          test-id="consultation-cta-daf"
          @click="goDaf"
        >
          {{ $t('clients.consultation.ctaDafInvoice') }}
        </ProButton>
        <ProButton
          v-if="invoicingUiEnabled"
          test-id="consultation-cta-invoice"
          @click="goInvoice"
        >
          {{ $t('clients.consultation.ctaInvoice') }}
        </ProButton>
        <ProButton
          variant="secondary"
          test-id="consultation-cta-done"
          @click="finishDone"
        >
          {{ $t('clients.consultation.ctaDone') }}
        </ProButton>
      </template>
    </template>
  </ProModal>
</template>

<script setup lang="ts">
import type { ConsultationPet } from '~/composables/useConsultationFlow'

const props = defineProps<{
  open: boolean
  clientId: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const { t } = useI18n()
const { mapError } = useApiError()
const {
  pharmacyEnabled,
  invoicingUiEnabled,
  visitId,
  starting,
  reportSaved,
  error: flowError,
  reset,
  startConsultation,
  afterSaved,
  discardIfUnsaved,
  markDone,
  dafWizardPath,
  invoicingPath,
  loadClientPets,
} = useConsultationFlow()

const pets = ref<ConsultationPet[]>([])
const petsLoading = ref(false)
const loadError = ref('')
const selectedPetId = ref('')
const notes = ref('')

async function loadPets() {
  if (!props.clientId) return
  petsLoading.value = true
  loadError.value = ''
  try {
    pets.value = await loadClientPets(props.clientId)
    if (pets.value.length === 1) {
      selectedPetId.value = pets.value[0].id
    }
  }
  catch (e: any) {
    loadError.value = mapError(e) || t('clients.consultation.loadPetsError')
    pets.value = []
  }
  finally {
    petsLoading.value = false
  }
}

watch(
  () => [props.open, props.clientId] as const,
  ([open]) => {
    if (!open) return
    reset()
    selectedPetId.value = ''
    notes.value = ''
    void loadPets()
  },
)

async function onOpenUpdate(v: boolean) {
  if (!v) {
    await discardIfUnsaved()
    reset()
  }
  emit('update:open', v)
}

/** Unique close path — discard/reset only via onOpenUpdate. */
function close() {
  void onOpenUpdate(false)
}

async function start() {
  if (!selectedPetId.value || !props.clientId) return
  try {
    await startConsultation({
      clientId: props.clientId,
      petId: selectedPetId.value,
      notes: notes.value,
    })
  }
  catch {
    // error already on flowError
  }
}

function onReportSaved() {
  afterSaved()
}

async function goDaf() {
  const path = dafWizardPath()
  await navigateTo(path)
  void onOpenUpdate(false)
}

async function goInvoice() {
  const path = invoicingPath({ mode: 'direct' })
  await navigateTo(path)
  void onOpenUpdate(false)
}

async function finishDone() {
  try {
    await markDone()
  }
  catch {
    // ignore — still close
  }
  void onOpenUpdate(false)
}
</script>
