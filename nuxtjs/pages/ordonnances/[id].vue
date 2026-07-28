<template>
  <div data-testid="ordonnances-detail-page">
    <ProPageHeader
      :title="doc ? `${doc.petName || $t('prescriptions.title')}` : $t('prescriptions.title')"
      :subtitle="statusLabel(doc?.status)"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="ordonnances-detail-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton variant="secondary" test-id="ordonnances-open-pdf" :disabled="loading || !doc" @click="openPdf">
          {{ $t('prescriptions.openPdf') }}
        </ProButton>
        <ProButton
          v-if="canEditDraft"
          variant="primary"
          test-id="ordonnances-save"
          :disabled="busy"
          @click="save"
        >
          {{ busy ? $t('prescriptions.loading') : $t('prescriptions.save') }}
        </ProButton>
        <ProButton
          v-if="canEditDraft"
          variant="ghost"
          test-id="ordonnances-delete"
          :disabled="busy"
          @click="remove"
        >
          {{ $t('prescriptions.delete') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p v-if="loading" class="pro-hint" data-testid="ordonnances-detail-loading">{{ $t('prescriptions.loading') }}</p>
    <p v-if="error" class="pro-alert">{{ error }}</p>

    <ProCard v-if="doc" class="pro-mb-lg">
      <p class="pro-hint">
        {{ doc.practiceName }} · {{ doc.veterinaryName }} · {{ doc.ownerName }}
        · {{ doc.countryCode }} · {{ doc.paperFormat }}
      </p>
      <div v-if="canEditDraft" class="rx-edit">
        <label class="pro-label">{{ $t('prescriptions.paperFormat') }}</label>
        <select v-model="paperFormat" class="pro-input" style="max-width:8rem" data-testid="ordonnances-format">
          <option value="A4">A4</option>
          <option value="A5">A5</option>
        </select>
        <label class="pro-label">{{ $t('prescriptions.validUntil') }}</label>
        <input v-model="validUntil" type="date" class="pro-input" data-testid="ordonnances-valid-until" />
        <label class="pro-label">{{ $t('prescriptions.notes') }}</label>
        <textarea v-model="notes" class="pro-input" rows="2" />
      </div>
      <div v-else>
        <p v-if="doc.validUntil" class="pro-hint">{{ $t('prescriptions.validUntil') }}: {{ doc.validUntil?.slice?.(0, 10) }}</p>
        <p v-if="doc.notes" class="pro-hint">{{ doc.notes }}</p>
      </div>
    </ProCard>

    <ProCard v-if="doc">
      <div v-for="(line, idx) in lines" :key="idx" class="rx-line" :data-testid="`ordonnances-line-${idx}`">
        <template v-if="canEditDraft">
          <input v-model="line.name" class="pro-input" :placeholder="$t('prescriptions.medName')" />
          <input v-model="line.dosage" class="pro-input" :placeholder="$t('prescriptions.dosage')" />
          <input v-model="line.form" class="pro-input" :placeholder="$t('prescriptions.form')" />
          <input v-model="line.quantity" class="pro-input" :placeholder="$t('prescriptions.quantity')" />
          <input v-model="line.posology" class="pro-input" :placeholder="$t('prescriptions.posology')" />
          <input v-model="line.withdrawal_period" class="pro-input" :placeholder="$t('prescriptions.withdrawal')" />
          <div v-if="lines.length > 1" class="rx-line__full">
            <ProButton variant="ghost" :test-id="`ordonnances-remove-line-${idx}`" @click="removeLine(idx)">
              {{ $t('prescriptions.removeLine') }}
            </ProButton>
          </div>
        </template>
        <template v-else>
          <p><strong>{{ line.name }}</strong> — {{ line.dosage }} · {{ line.form }} · {{ line.quantity }}</p>
          <p class="pro-hint">{{ line.posology }}</p>
        </template>
      </div>
      <ProButton
        v-if="canEditDraft"
        variant="secondary"
        test-id="ordonnances-add-line"
        @click="lines.push(emptyLine())"
      >
        {{ $t('prescriptions.addLine') }}
      </ProButton>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pets.read' })
const { t } = useI18n()
const { canPractice } = usePracticePerms()
const { mapError } = usePrescriptionError()
const route = useRoute()
const busy = ref(false)
const loading = ref(true)
const error = ref('')
const doc = ref<any>(null)
const notes = ref('')
const paperFormat = ref('A4')
const validUntil = ref('')
const lines = ref<any[]>([])
const canEditDraft = computed(() => doc.value?.status === 'draft' && canPractice('pets.write_clinical'))

function unwrap(res: any) {
  return res?.data ?? res
}

function emptyLine() {
  return { name: '', dosage: '', form: '', quantity: '', posology: '', withdrawal_period: '' }
}

function removeLine(idx: number) {
  if (lines.value.length <= 1) return
  lines.value.splice(idx, 1)
}

function statusLabel(s?: string) {
  if (s === 'draft') return t('prescriptions.statusDraft')
  if (s === 'signed') return t('prescriptions.statusSigned')
  if (s === 'sent') return t('prescriptions.statusSent')
  if (s === 'archived') return t('prescriptions.statusArchived')
  return s || ''
}

function applyDoc(d: any) {
  doc.value = d
  notes.value = d.notes || ''
  paperFormat.value = d.paperFormat || 'A4'
  validUntil.value = d.validUntil ? String(d.validUntil).slice(0, 10) : ''
  const meds = Array.isArray(d.medications) ? d.medications : []
  lines.value = meds.length
    ? meds.map((m: any) => ({
        name: m.name || '',
        dosage: m.dosage || '',
        form: m.form || '',
        quantity: m.quantity || '',
        posology: m.posology || '',
        withdrawal_period: m.withdrawal_period || '',
      }))
    : [emptyLine()]
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await $fetch<any>(`/api/vet/prescriptions/${route.params.id}`)
    applyDoc(unwrap(res))
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
}

function openPdf() {
  window.open(`/api/vet/prescriptions/${route.params.id}/pdf`, '_blank')
}

async function save() {
  if (busy.value || doc.value?.status !== 'draft') return
  busy.value = true
  error.value = ''
  try {
    const body: any = {
      notes: notes.value,
      paperFormat: paperFormat.value,
      validUntil: validUntil.value || '',
      medications: lines.value
        .filter(l => l.name.trim())
        .map(l => ({
          name: l.name.trim(),
          dosage: l.dosage.trim(),
          form: l.form.trim(),
          quantity: l.quantity.trim(),
          posology: l.posology.trim(),
          withdrawal_period: l.withdrawal_period.trim(),
        })),
    }
    const res = await $fetch<any>(`/api/vet/prescriptions/${route.params.id}`, {
      method: 'PATCH',
      body,
    })
    applyDoc(unwrap(res))
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    busy.value = false
  }
}

async function remove() {
  if (busy.value || doc.value?.status !== 'draft') return
  if (!confirm(t('prescriptions.deleteConfirm'))) return
  busy.value = true
  try {
    await $fetch(`/api/vet/prescriptions/${route.params.id}`, { method: 'DELETE' })
    await navigateTo('/ordonnances')
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    busy.value = false
  }
}

onMounted(() => { load() })
</script>

<style scoped>
.rx-line {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.rx-line__full { grid-column: 1 / -1; }
.rx-edit { display: grid; gap: 0.5rem; margin-top: 0.75rem; }
.pro-mb-lg { margin-bottom: 1.25rem; }
.pro-hint { color: var(--pf-vet-muted, #667); }
</style>
