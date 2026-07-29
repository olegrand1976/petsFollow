<template>
  <div data-testid="daf-wizard-page">
    <ProPageHeader
      :title="$t('pharmacy.daf.wizardTitle')"
      :subtitle="$t('pharmacy.daf.wizardSubtitle')"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <p v-if="error" class="pro-alert" data-testid="daf-wizard-error">{{ error }}</p>

    <ProCard v-if="contextLabel" class="pro-mb-lg" data-testid="daf-consultation-context">
      <p class="pro-hint">{{ contextLabel }}</p>
    </ProCard>

    <ProCard class="pro-mb-lg">
      <div v-for="(line, idx) in lines" :key="idx" class="daf-line" :data-testid="`daf-line-${idx}`">
        <div>
          <label class="pro-label">{{ $t('pharmacy.stock.medication') }}</label>
          <ProCombobox
            :input-id="`daf-med-${idx}`"
            v-model="line.med"
            :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
            :search-fn="searchMedications"
          />
        </div>
        <div>
          <label class="pro-label">{{ $t('pharmacy.daf.amm') }}</label>
          <input v-model="line.ammNumber" class="pro-input" :data-testid="`daf-amm-${idx}`" />
        </div>
        <div>
          <label class="pro-label">{{ $t('pharmacy.stock.qty') }}</label>
          <input v-model.number="line.qty" type="number" min="0.001" step="any" class="pro-input" />
        </div>
        <template v-if="line.med?.raw?.isAntibiotic">
          <div>
            <label class="pro-label">{{ $t('pharmacy.daf.vamregSpecies') }}</label>
            <input v-model="line.species" class="pro-input" />
          </div>
          <div>
            <label class="pro-label">{{ $t('pharmacy.daf.vamregIndication') }}</label>
            <input v-model="line.indication" class="pro-input" />
          </div>
          <div>
            <label class="pro-label">{{ $t('pharmacy.daf.vamregDuration') }}</label>
            <input v-model.number="line.durationDays" type="number" min="1" class="pro-input" />
          </div>
        </template>
      </div>
      <ProButton variant="secondary" test-id="daf-add-line" @click="addLine">{{ $t('pharmacy.daf.addLine') }}</ProButton>
    </ProCard>

    <ProCard v-if="preview.length" class="pro-mb-lg" data-testid="daf-preview">
      <h3>{{ $t('pharmacy.daf.previewTitle') }}</h3>
      <ul>
        <li v-for="(p, i) in preview" :key="i">
          {{ p.medicationName }} · lot {{ p.lotNumber }} · {{ p.qty }} · DLC {{ p.expiresOn }}
        </li>
      </ul>
    </ProCard>

    <div class="daf-actions">
      <ProButton variant="secondary" test-id="daf-preview-btn" :disabled="busy" @click="runPreview">{{ $t('pharmacy.daf.preview') }}</ProButton>
      <ProButton variant="secondary" test-id="daf-save-draft" :disabled="busy || !canSubmit" @click="saveDraft">{{ $t('pharmacy.daf.saveDraft') }}</ProButton>
      <ProButton variant="primary" test-id="daf-finalize-btn" :disabled="busy || !canSubmit" @click="finalize">{{ $t('pharmacy.daf.finalize') }}</ProButton>
      <ProButton
        v-if="finalizedDafId && invoicingUiEnabled"
        test-id="daf-go-invoice"
        :disabled="busy"
        @click="goInvoice"
      >
        {{ $t('pharmacy.daf.goInvoice') }}
      </ProButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import { INVOICING_UI_ENABLED } from '~/utils/invoicing-ui'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pharmacy.write' })
const { t } = useI18n()
function pharmacyErr(e: any, fallbackKey: string): string {
  return pharmacyErrorMessage(t, e, fallbackKey)
}
const route = useRoute()
const invoicingUiEnabled = INVOICING_UI_ENABLED
const busy = ref(false)
const error = ref('')
const preview = ref<any[]>([])
const finalizedDafId = ref('')
const contextClientName = ref('')
const contextPetName = ref('')

const clientUserId = computed(() => String(route.query.clientUserId || ''))
const petId = computed(() => String(route.query.petId || ''))
const visitId = computed(() => String(route.query.visitId || ''))

const contextLabel = computed(() => {
  if (!clientUserId.value && !petId.value && !visitId.value) return ''
  const parts: string[] = []
  if (contextClientName.value) parts.push(contextClientName.value)
  else if (clientUserId.value) parts.push(t('pharmacy.daf.contextClient'))
  if (contextPetName.value) parts.push(contextPetName.value)
  else if (petId.value) parts.push(t('pharmacy.daf.contextPet'))
  if (visitId.value) parts.push(t('pharmacy.daf.contextVisit'))
  return t('pharmacy.daf.contextBanner', { context: parts.join(' · ') })
})

type Line = {
  med: ProComboboxItem | null
  ammNumber: string
  qty: number
  species: string
  indication: string
  durationDays: number
}

const lines = ref<Line[]>([emptyLine()])

function emptyLine(): Line {
  return { med: null, ammNumber: '', qty: 1, species: '', indication: '', durationDays: 5 }
}

function addLine() {
  lines.value.push(emptyLine())
}

const canSubmit = computed(() =>
  lines.value.every(l => l.med?.id && l.ammNumber && l.qty > 0),
)

function unwrap(res: any) {
  return res?.data ?? res
}

async function loadContext() {
  if (clientUserId.value) {
    try {
      const res = await $fetch<any>(`/api/clients/${clientUserId.value}`)
      const c = unwrap(res)
      contextClientName.value = c?.fullName || c?.email || ''
    }
    catch {
      contextClientName.value = ''
    }
  }
  if (petId.value && clientUserId.value) {
    try {
      const res = await $fetch<any>(`/api/clients/${clientUserId.value}/pets`)
      const pets = unwrap(res)
      const list = Array.isArray(pets) ? pets : []
      const pet = list.find((p: any) => p.id === petId.value)
      contextPetName.value = pet?.name || ''
    }
    catch {
      contextPetName.value = ''
    }
  }
}

onMounted(() => {
  void loadContext()
})

async function searchMedications(q: string): Promise<ProComboboxItem[]> {
  const res = await $fetch<any>('/api/vet/pharmacy/medications/search', { query: { q, limit: '20' } })
  const items = unwrap(res)?.items ?? []
  return items.map((m: any) => ({
    id: m.id,
    label: m.name,
    hint: m.cnk,
    badge: m.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
    raw: m,
  }))
}

function buildItems() {
  return lines.value.map((l) => {
    const item: any = {
      medicationId: l.med!.id,
      qty: l.qty,
      ammNumber: l.ammNumber,
      unit: 'unit',
    }
    if (l.med?.raw?.isAntibiotic) {
      item.vamregPayload = {
        species: l.species,
        indication: l.indication,
        durationDays: l.durationDays,
      }
    }
    return item
  })
}

function draftBody() {
  return {
    clientUserId: clientUserId.value || undefined,
    petId: petId.value || undefined,
    visitId: visitId.value || undefined,
    items: buildItems(),
  }
}

async function runPreview() {
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/preview-fefo', { method: 'POST', body: { items: buildItems() } })
    preview.value = unwrap(res)?.lines ?? []
  }
  catch (e: any) {
    error.value = pharmacyErr(e, 'pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

async function saveDraft() {
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf', { method: 'POST', body: draftBody() })
    const doc = unwrap(res)
    await navigateTo(`/daf/${doc.id}`)
  }
  catch (e: any) {
    error.value = pharmacyErr(e, 'pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

async function finalize() {
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf', { method: 'POST', body: draftBody() })
    const doc = unwrap(res)
    const fin = await $fetch<any>(`/api/vet/pharmacy/daf/${doc.id}/finalize`, { method: 'POST', body: {} })
    const out = unwrap(fin)
    const id = out?.id || out?.daf?.id || doc.id
    finalizedDafId.value = id
  }
  catch (e: any) {
    error.value = pharmacyErr(e, 'pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

async function goInvoice() {
  const q = new URLSearchParams()
  if (clientUserId.value) q.set('clientUserId', clientUserId.value)
  if (visitId.value) q.set('visitId', visitId.value)
  if (finalizedDafId.value) q.set('dafId', finalizedDafId.value)
  q.set('mode', 'fromDaf')
  await navigateTo(`/invoicing?${q.toString()}`)
}
</script>

<style scoped>
.daf-line {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--pf-vet-border);
}
.daf-actions { display: flex; gap: 0.75rem; flex-wrap: wrap; }
</style>
