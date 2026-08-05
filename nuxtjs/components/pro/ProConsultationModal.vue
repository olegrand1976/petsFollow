<template>
  <ProModal
    :open="open"
    :size="modalSize"
    :title="modalTitle"
    test-id="consultation-modal"
    :prevent-close="workspaceBusy || closing || starting || leavePromptBlocking"
    :expandable="modalExpandable"
    v-model:expanded="modalExpanded"
    contain-scroll
    @update:open="onOpenUpdate"
  >
    <!-- Étape A : choisir l'animal -->
    <div v-if="!visitId" class="consultation-setup pro-form" data-testid="consultation-setup">
      <p class="pro-hint">{{ $t('clients.consultation.hint') }}</p>
      <p v-if="loadError" class="pro-error" role="alert">{{ loadError }}</p>
      <p v-if="flowError" class="pro-error" role="alert">{{ flowError }}</p>

      <div class="pro-field">
        <label class="pro-label" for="consultation-pet">{{ $t('clients.consultation.pet') }}</label>
        <select
          id="consultation-pet"
          v-model="selectedPetId"
          class="pro-select"
          data-testid="consultation-pet-select"
          :disabled="petsLoading || starting || !!resumeVisitId"
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

    <!-- Étape B/C : workspace CR + hub (même coque partout) -->
    <div
      v-else
      class="consultation-modal-shell"
      data-testid="consultation-detail-surface"
    >
      <ProConsultationWorkspace
        :key="visitId"
        ref="workspaceRef"
        :visit-id="visitId"
        :client-id="workspaceClientId"
        :pet-id="workspacePetId"
        :scheduled-at="scheduledAt || undefined"
        :preserve-visit="preserveVisit"
        :readonly="readonly"
        @stage="onStage"
        @leave-prompt="onLeavePrompt"
        @busy="onWorkspaceBusy"
        @closed="onWorkspaceClosed"
      />
    </div>

    <template v-if="!visitId" #footer>
      <ProButton
        variant="ghost"
        test-id="consultation-setup-cancel"
        :disabled="starting"
        @click="close"
      >
        {{ $t('common.cancel') }}
      </ProButton>
      <ProButton
        test-id="consultation-start"
        :loading="starting"
        :disabled="!selectedPetId || starting || petsLoading"
        @click="start"
      >
        {{ $t('clients.consultation.start') }}
      </ProButton>
    </template>
  </ProModal>
</template>

<script setup lang="ts">
import type { ConsultationPet } from '~/composables/useConsultationFlow'
import { useActiveConsultation } from '~/composables/useActiveConsultation'

type WorkspaceExpose = {
  requestLeave: () => void
  isBusy: () => boolean
}

const props = defineProps<{
  open: boolean
  clientId: string
  /** Reprise / historique / agenda : visite déjà créée. */
  resumeVisitId?: string
  resumePetId?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  closed: []
}>()

const { t } = useI18n()
const { mapError } = useApiError()
const active = useActiveConsultation()
const { user } = useProUser()
const { canPractice } = usePracticePerms()
const {
  visitId,
  petId,
  clientId: flowClientId,
  scheduledAt,
  starting,
  preserveVisit,
  error: flowError,
  reset,
  startConsultation,
  loadClientPets,
} = useConsultationFlow()

const pets = ref<ConsultationPet[]>([])
const petsLoading = ref(false)
const loadError = ref('')
const selectedPetId = ref('')
const notes = ref('')
const closing = ref(false)
const leavePromptBlocking = ref(false)
const workspaceBusy = ref(false)
const stage = ref<'report' | 'hub'>('report')
const modalExpanded = ref(false)
const workspaceRef = ref<WorkspaceExpose | null>(null)

const readonly = computed(() => !canPractice('pets.write_clinical'))

const workspaceClientId = computed(() => flowClientId.value || props.clientId)
const workspacePetId = computed(() => petId.value || selectedPetId.value || props.resumePetId || '')

/** CR (après start ou reprise) = toujours coque full + bouton plein écran. */
const hasVisitShell = computed(() => Boolean(visitId.value || props.resumeVisitId))
const modalSize = computed<'md' | 'full'>(() => (hasVisitShell.value ? 'full' : 'md'))
const modalExpandable = computed(() => hasVisitShell.value)

const modalTitle = computed(() => {
  if (!hasVisitShell.value) return t('clients.consultation.title')
  if (stage.value === 'hub') return t('clients.consultation.nextStepsTitle')
  return t('consultations.detailTitle')
})

async function loadPets() {
  if (!props.clientId) return
  petsLoading.value = true
  loadError.value = ''
  try {
    pets.value = await loadClientPets(props.clientId)
    if (props.resumePetId) {
      selectedPetId.value = props.resumePetId
    }
    else if (pets.value.length === 1) {
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

function applyResume() {
  if (!props.resumeVisitId) return
  visitId.value = props.resumeVisitId
  petId.value = props.resumePetId || ''
  flowClientId.value = props.clientId
  selectedPetId.value = props.resumePetId || ''
  preserveVisit.value = active.resumeKeepVisit.value
  stage.value = 'report'
  active.syncActiveVisit({
    visitId: props.resumeVisitId,
    clientId: props.clientId,
    petId: props.resumePetId || '',
    keepVisit: preserveVisit.value,
  })
  void hydrateResumeSchedule(props.resumeVisitId, props.resumePetId || '')
}

/** No GET /visits/:id — resolve date via pet visit list when possible. */
async function hydrateResumeSchedule(id: string, petHint: string) {
  if (!petHint) return
  try {
    const res: any = await $fetch(`/api/pets/${petHint}/visits`)
    const list = Array.isArray(res?.data ?? res) ? (res.data ?? res) : []
    const visit = list.find((v: any) => v?.id === id)
    const at = visit?.scheduledAt || visit?.proposedScheduledAt || visit?.createdAt || ''
    if (at) scheduledAt.value = String(at)
  }
  catch {
    /* date label optional on resume */
  }
}

watch(
  () => [props.open, props.clientId, props.resumeVisitId] as const,
  ([isOpen]) => {
    if (!isOpen) return
    reset()
    selectedPetId.value = ''
    notes.value = ''
    closing.value = false
    leavePromptBlocking.value = false
    workspaceBusy.value = false
    stage.value = 'report'
    modalExpanded.value = false
    if (props.resumeVisitId) {
      applyResume()
    }
    else {
      void loadPets()
    }
  },
  { immediate: true },
)

watch(
  () => props.resumePetId,
  (petIdHint) => {
    if (!props.open || props.resumeVisitId || visitId.value) return
    if (petIdHint) selectedPetId.value = petIdHint
  },
)

watch(
  () => [props.open, visitId.value, props.clientId, petId.value || selectedPetId.value] as const,
  ([isOpen, vid, cid, pid]) => {
    if (!isOpen || !vid || !cid) {
      if (!isOpen) active.syncActiveVisit(null)
      return
    }
    active.syncActiveVisit({
      visitId: vid,
      clientId: cid,
      petId: String(pid || ''),
      keepVisit: preserveVisit.value,
    })
  },
)

function onStage(s: 'report' | 'hub') {
  stage.value = s
}

function onLeavePrompt(open: boolean) {
  leavePromptBlocking.value = open
}

function onWorkspaceBusy(busy: boolean) {
  workspaceBusy.value = busy
}

async function onWorkspaceClosed() {
  closing.value = true
  try {
    const email = user.value?.email
    if (email) active.clearResumeForEmail(email)
    reset()
    active.syncActiveVisit(null)
    emit('update:open', false)
    emit('closed')
  }
  finally {
    closing.value = false
  }
}

function onOpenUpdate(v: boolean) {
  if (!v) {
    if (starting.value || closing.value || workspaceBusy.value || leavePromptBlocking.value) return
    if (visitId.value) {
      workspaceRef.value?.requestLeave()
      return
    }
    reset()
    active.syncActiveVisit(null)
    emit('update:open', false)
    emit('closed')
    return
  }
  emit('update:open', v)
}

function close() {
  if (starting.value || closing.value) return
  onOpenUpdate(false)
}

async function start() {
  if (!selectedPetId.value || !props.clientId) return
  try {
    const id = await startConsultation({
      clientId: props.clientId,
      petId: selectedPetId.value,
      notes: notes.value,
    })
    if (!id) return
    preserveVisit.value = false
    active.syncActiveVisit({
      visitId: id,
      clientId: props.clientId,
      petId: selectedPetId.value,
      keepVisit: false,
    })
    stage.value = 'report'
  }
  catch {
    // error already on flowError
  }
}
</script>

<style scoped>
.consultation-setup {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.consultation-modal-shell {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.consultation-modal-shell :deep(.consultation-workspace),
.consultation-modal-shell :deep(.pro-visit-report) {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
}
</style>
