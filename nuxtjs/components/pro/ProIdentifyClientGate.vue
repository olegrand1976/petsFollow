<template>
  <div class="identify-gate pro-form" data-testid="identify-client-gate">
    <p class="pro-hint">{{ $t('clients.walkin.identifyHint') }}</p>
    <p v-if="error" class="pro-error" role="alert">{{ error }}</p>

    <div class="identify-gate__modes" role="tablist">
      <button
        type="button"
        class="identify-gate__mode"
        :class="{ 'identify-gate__mode--active': mode === 'create' }"
        data-testid="identify-mode-create"
        role="tab"
        :aria-selected="mode === 'create'"
        @click="mode = 'create'"
      >
        {{ $t('clients.walkin.modeCreate') }}
      </button>
      <button
        type="button"
        class="identify-gate__mode"
        :class="{ 'identify-gate__mode--active': mode === 'existing' }"
        data-testid="identify-mode-existing"
        role="tab"
        :aria-selected="mode === 'existing'"
        @click="mode = 'existing'"
      >
        {{ $t('clients.walkin.modeExisting') }}
      </button>
    </div>

    <template v-if="mode === 'create'">
      <ProInput
        v-model="createForm.firstName"
        test-id="identify-first-name"
        :label="$t('clients.create.firstName')"
        required
      />
      <ProInput
        v-model="createForm.lastName"
        test-id="identify-last-name"
        :label="$t('clients.create.lastName')"
        required
      />
      <ProInput
        v-model="createForm.contactPhone"
        test-id="identify-phone"
        type="tel"
        :label="$t('clients.create.contactPhone')"
        required
        :maxlength="40"
      />
      <ProInput
        v-model="createForm.email"
        test-id="identify-email"
        type="email"
        :label="$t('clients.walkin.emailOptional')"
      />
      <ProInput
        v-model="createForm.petName"
        test-id="identify-pet-name"
        :label="$t('clients.walkin.petName')"
        required
      />
      <div class="pro-field">
        <label class="pro-label" for="identify-pet-species">{{ $t('clients.walkin.petSpecies') }}</label>
        <select
          id="identify-pet-species"
          v-model="createForm.species"
          class="pro-select"
          required
          data-testid="identify-pet-species"
        >
          <option value="dog">{{ $t('clients.walkin.speciesDog') }}</option>
          <option value="cat">{{ $t('clients.walkin.speciesCat') }}</option>
          <option value="horse">{{ $t('clients.walkin.speciesHorse') }}</option>
          <option value="other">{{ $t('clients.walkin.speciesOther') }}</option>
        </select>
      </div>
    </template>

    <template v-else>
      <div class="pro-field">
        <label class="pro-label" for="identify-client-search">{{ $t('clients.search') }}</label>
        <input
          id="identify-client-search"
          v-model="search"
          type="search"
          class="pro-input"
          :placeholder="$t('clients.searchPlaceholder')"
          data-testid="identify-client-search"
        >
      </div>
      <div class="pro-field">
        <label class="pro-label" for="identify-client-select">{{ $t('calendar.columnClient') }}</label>
        <select
          id="identify-client-select"
          v-model="existingClientId"
          class="pro-select"
          data-testid="identify-client-select"
        >
          <option value="">{{ $t('calendar.selectClient') }}</option>
          <option
            v-for="c in filteredClients"
            :key="c.userId"
            :value="c.userId"
          >
            {{ c.fullName }}{{ c.contactPhone ? ` · ${c.contactPhone}` : '' }}
          </option>
        </select>
      </div>

      <div
        v-if="selectedClient"
        class="identify-gate__confirm-card"
        data-testid="identify-confirm-card"
      >
        <p><strong>{{ selectedClient.fullName }}</strong></p>
        <p class="pro-hint">{{ selectedClient.email }}</p>
        <p v-if="selectedClient.contactPhone" class="pro-hint">{{ selectedClient.contactPhone }}</p>
        <p class="pro-hint">{{ $t('clients.walkin.petCount', { n: selectedClient.petCount }) }}</p>
        <label class="pro-checkbox-row">
          <input
            v-model="confirmed"
            type="checkbox"
            data-testid="identify-confirm-checkbox"
          >
          <span>{{ $t('clients.walkin.confirmExisting') }}</span>
        </label>
      </div>

      <div v-if="existingClientId && confirmed" class="pro-field">
        <label class="pro-label" for="identify-existing-pet">{{ $t('clients.consultation.pet') }}</label>
        <select
          id="identify-existing-pet"
          v-model="existingPetMode"
          class="pro-select"
          data-testid="identify-existing-pet-mode"
        >
          <option value="new">{{ $t('clients.walkin.createNewPet') }}</option>
          <option
            v-for="p in existingPets"
            :key="p.id"
            :value="p.id"
          >
            {{ p.name }}
          </option>
        </select>
      </div>
      <template v-if="existingClientId && confirmed && existingPetMode === 'new'">
        <ProInput
          v-model="existingNewPetName"
          test-id="identify-existing-pet-name"
          :label="$t('clients.walkin.petName')"
          required
        />
        <div class="pro-field">
          <label class="pro-label" for="identify-existing-species">{{ $t('clients.walkin.petSpecies') }}</label>
          <select
            id="identify-existing-species"
            v-model="existingNewPetSpecies"
            class="pro-select"
            data-testid="identify-existing-species"
          >
            <option value="dog">{{ $t('clients.walkin.speciesDog') }}</option>
            <option value="cat">{{ $t('clients.walkin.speciesCat') }}</option>
            <option value="horse">{{ $t('clients.walkin.speciesHorse') }}</option>
            <option value="other">{{ $t('clients.walkin.speciesOther') }}</option>
          </select>
        </div>
      </template>
    </template>

    <div class="identify-gate__actions">
      <ProButton
        test-id="identify-submit"
        :loading="busy"
        :disabled="!canSubmit || busy"
        @click="submit"
      >
        {{ $t('clients.walkin.submit') }}
      </ProButton>
    </div>
  </div>
</template>

<script setup lang="ts">
type ClientRow = {
  userId: string
  fullName: string
  email: string
  contactPhone?: string
  petCount: number
  isWalkinPlaceholder?: boolean
}

type PetRow = { id: string, name: string }

const props = defineProps<{
  visitId: string
  initialPhone?: string
}>()

const emit = defineEmits<{
  identified: [payload: { clientId: string, petId: string, clientName: string, petName: string }]
}>()

const { t } = useI18n()
const { mapError } = useApiError()

const mode = ref<'create' | 'existing'>('create')
const busy = ref(false)
const error = ref('')
const clients = ref<ClientRow[]>([])
const search = ref('')
const existingClientId = ref('')
const confirmed = ref(false)
const existingPets = ref<PetRow[]>([])
const existingPetMode = ref('new')
const existingNewPetName = ref('')
const existingNewPetSpecies = ref('dog')

const createForm = reactive({
  firstName: '',
  lastName: '',
  contactPhone: props.initialPhone?.trim() || '',
  email: '',
  petName: '',
  species: 'dog',
})

watch(
  () => props.initialPhone,
  (phone) => {
    if (phone?.trim() && !createForm.contactPhone.trim()) {
      createForm.contactPhone = phone.trim()
    }
  },
)

const filteredClients = computed(() => {
  const q = search.value.trim().toLowerCase()
  return clients.value.filter((c) => {
    if (c.isWalkinPlaceholder) return false
    if (!q) return true
    return [c.fullName, c.email, c.contactPhone || ''].some(s => s.toLowerCase().includes(q))
  })
})

const selectedClient = computed(() =>
  clients.value.find(c => c.userId === existingClientId.value) || null,
)

const canSubmit = computed(() => {
  if (mode.value === 'create') {
    return Boolean(
      createForm.firstName.trim()
      && createForm.lastName.trim()
      && createForm.contactPhone.trim()
      && createForm.petName.trim()
      && createForm.species,
    )
  }
  if (!existingClientId.value || !confirmed.value) return false
  if (existingPetMode.value === 'new') {
    return Boolean(existingNewPetName.value.trim() && existingNewPetSpecies.value)
  }
  return Boolean(existingPetMode.value)
})

onMounted(async () => {
  try {
    const res: any = await $fetch('/api/clients')
    const list = res?.data ?? res ?? []
    clients.value = (Array.isArray(list) ? list : []).filter((c: any) => c?.userId).map((c: any) => ({
      userId: String(c.userId),
      fullName: String(c.fullName || c.displayName || c.email || ''),
      email: String(c.email || ''),
      contactPhone: c.contactPhone || '',
      petCount: Number(c.petCount || 0),
      isWalkinPlaceholder: !!c.isWalkinPlaceholder,
    }))
  }
  catch {
    clients.value = []
  }
})

watch(existingClientId, async (id) => {
  confirmed.value = false
  existingPets.value = []
  existingPetMode.value = 'new'
  if (!id) return
  try {
    const res: any = await $fetch(`/api/clients/${id}/pets`)
    const list = res?.data ?? res ?? []
    existingPets.value = (Array.isArray(list) ? list : [])
      .filter((p: any) => p?.id && !p.isWalkinPlaceholder)
      .map((p: any) => ({ id: String(p.id), name: String(p.name || p.id) }))
  }
  catch {
    existingPets.value = []
  }
})

async function submit() {
  if (!canSubmit.value || busy.value) return
  busy.value = true
  error.value = ''
  try {
    const body: Record<string, unknown> = { mode: mode.value }
    if (mode.value === 'create') {
      body.client = {
        firstName: createForm.firstName.trim(),
        lastName: createForm.lastName.trim(),
        contactPhone: createForm.contactPhone.trim(),
        email: createForm.email.trim() || undefined,
      }
      body.newPet = {
        name: createForm.petName.trim(),
        species: createForm.species,
      }
    }
    else {
      body.clientUserId = existingClientId.value
      if (existingPetMode.value === 'new') {
        body.newPet = {
          name: existingNewPetName.value.trim(),
          species: existingNewPetSpecies.value,
        }
      }
      else {
        body.petId = existingPetMode.value
      }
    }
    const res: any = await $fetch(`/api/visits/${props.visitId}/identify-client`, {
      method: 'POST',
      body,
    })
    const data = res?.data ?? res
    const clientId = String(data?.client?.userId || '')
    const petId = String(data?.pet?.id || '')
    if (!clientId || !petId) throw new Error('identify_incomplete')
    emit('identified', {
      clientId,
      petId,
      clientName: String(data?.client?.fullName || ''),
      petName: String(data?.pet?.name || ''),
    })
  }
  catch (e: any) {
    error.value = mapError(e) || t('clients.walkin.identifyFailed')
  }
  finally {
    busy.value = false
  }
}
</script>

<style scoped>
.identify-gate__modes {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.identify-gate__mode {
  flex: 1;
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface);
  border-radius: 8px;
  padding: 0.65rem 0.75rem;
  cursor: pointer;
  font: inherit;
  color: inherit;
}
.identify-gate__mode--active {
  border-color: var(--pf-vet-accent);
  box-shadow: inset 0 0 0 1px var(--pf-vet-accent);
}
.identify-gate__confirm-card {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  margin-bottom: 0.75rem;
  background: var(--pf-vet-bg);
}
.identify-gate__actions {
  margin-top: 1rem;
  display: flex;
  justify-content: flex-end;
}
</style>
