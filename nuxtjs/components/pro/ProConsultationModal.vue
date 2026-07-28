<template>
  <ProModal
    :open="open"
    size="lg"
    :title="$t('clients.consultation.title')"
    test-id="consultation-modal"
    :prevent-close="reportBusy || closing"
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
      <p v-if="actionError" class="pro-error" role="alert">{{ actionError }}</p>
      <ProVisitReportPanel
        :visit-id="visitId"
        @saved="onReportSaved"
        @finalized="onReportSaved"
        @busy="onReportBusy"
      />
    </div>

    <template #footer>
      <ProButton
        variant="ghost"
        test-id="consultation-cancel"
        :disabled="reportBusy || closing"
        @click="close"
      >
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
          v-if="showDafCta"
          variant="secondary"
          test-id="consultation-cta-daf"
          :disabled="closing"
          @click="goDaf"
        >
          {{ $t('clients.consultation.ctaDafInvoice') }}
        </ProButton>
        <ProButton
          v-if="showInvoiceCta"
          test-id="consultation-cta-invoice"
          :disabled="closing"
          @click="goInvoice"
        >
          {{ $t('clients.consultation.ctaInvoice') }}
        </ProButton>
        <ProButton
          variant="secondary"
          test-id="consultation-cta-done"
          :disabled="closing"
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
  billitEnabled,
  invoicingUiEnabled,
  visitId,
  starting,
  reportSaved,
  reportBusy,
  error: flowError,
  reset,
  startConsultation,
  afterSaved,
  setReportBusy,
  discardIfUnsaved,
  markDone,
  dafWizardPath,
  invoicingPath,
  loadClientPets,
} = useConsultationFlow()

const { canPractice } = usePracticePerms()
const canWritePharmacy = computed(() => canPractice('pharmacy.write'))
const showDafCta = computed(() => pharmacyEnabled.value && canWritePharmacy.value)
const showInvoiceCta = computed(() => invoicingUiEnabled && billitEnabled.value)

const pets = ref<ConsultationPet[]>([])
const petsLoading = ref(false)
const loadError = ref('')
const actionError = ref('')
const selectedPetId = ref('')
const notes = ref('')
const closing = ref(false)

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
    actionError.value = ''
    closing.value = false
    void loadPets()
  },
)

async function onOpenUpdate(v: boolean) {
  if (!v) {
    // Keep modal open while CR save is in flight (same as Cancel disabled).
    if (reportBusy.value || closing.value) return
    closing.value = true
    try {
      await discardIfUnsaved()
      reset()
      emit('update:open', false)
    }
    finally {
      closing.value = false
    }
    return
  }
  emit('update:open', v)
}

/** Unique close path — discard/reset only via onOpenUpdate. */
function close() {
  if (reportBusy.value || closing.value) return
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

function onReportBusy(busy: boolean) {
  setReportBusy(busy)
}

/** Close modal then navigate — avoid aborting navigateTo when the modal unmounts mid-flight. */
async function closeAndNavigate(path: string) {
  const target = path
  closing.value = true
  try {
    await discardIfUnsaved()
    reset()
    emit('update:open', false)
  }
  finally {
    closing.value = false
  }
  await nextTick()
  await navigateTo(target)
}

async function goDaf() {
  await closeAndNavigate(dafWizardPath())
}

async function goInvoice() {
  await closeAndNavigate(invoicingPath({ mode: 'direct' }))
}

async function finishDone() {
  actionError.value = ''
  try {
    await markDone()
    closing.value = true
    try {
      await discardIfUnsaved()
      reset()
      emit('update:open', false)
    }
    finally {
      closing.value = false
    }
  }
  catch (e: any) {
    actionError.value = mapError(e) || t('clients.consultation.markDoneError')
  }
}
</script>
