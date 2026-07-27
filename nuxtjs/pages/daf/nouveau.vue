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
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'

definePageMeta({ middleware: ['auth', 'vet-only'] })
const { t } = useI18n()
const busy = ref(false)
const error = ref('')
const preview = ref<any[]>([])

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

async function runPreview() {
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/preview-fefo', { method: 'POST', body: { items: buildItems() } })
    preview.value = unwrap(res)?.lines ?? []
  }
  catch (e: any) {
    error.value = e?.data?.error?.code || t('pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

async function saveDraft() {
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf', { method: 'POST', body: { items: buildItems() } })
    const doc = unwrap(res)
    await navigateTo(`/daf/${doc.id}`)
  }
  catch (e: any) {
    error.value = e?.data?.error?.code || t('pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

async function finalize() {
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf', { method: 'POST', body: { items: buildItems() } })
    const doc = unwrap(res)
    const fin = await $fetch<any>(`/api/vet/pharmacy/daf/${doc.id}/finalize`, { method: 'POST', body: {} })
    const out = unwrap(fin)
    const id = out?.id || out?.daf?.id || doc.id
    await navigateTo(`/daf/${id}`)
  }
  catch (e: any) {
    error.value = e?.data?.error?.code || t('pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
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
