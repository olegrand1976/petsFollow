<template>
  <div data-testid="invoicing-page">
    <ProPageHeader
      :title="$t('invoicing.title')"
      :subtitle="invoicingUiEnabled ? $t('invoicing.subtitle') : undefined"
    >
      <template #actions>
        <div
          v-if="invoicingUiEnabled"
          class="invoicing-conn"
          :data-testid="canManageBillit ? 'invoicing-connection' : 'invoicing-connection-readonly'"
        >
          <span class="invoicing-conn__title">{{ $t('invoicing.connectionTitle') }}</span>
          <ProBadge :variant="statusVariant">{{ connectionStatusLabel }}</ProBadge>
          <template v-if="canManageBillit">
            <ProButton
              v-if="!isActive && !pendingRegistration"
              variant="primary"
              test-id="invoicing-connect-start"
              :disabled="busy"
              @click="startConnect"
            >
              {{ $t('invoicing.activate') }}
            </ProButton>
            <ProButton
              v-if="pendingRegistration"
              variant="secondary"
              test-id="invoicing-open-reseller"
              @click="openReseller"
            >
              {{ $t('invoicing.openReseller') }}
            </ProButton>
            <ProButton
              v-if="pendingRegistration"
              variant="primary"
              test-id="invoicing-connect-start"
              :disabled="busy"
              @click="startConnect"
            >
              {{ $t('invoicing.complete') }}
            </ProButton>
            <ProButton
              v-if="isActive || pendingKyc"
              variant="secondary"
              test-id="invoicing-refresh"
              :disabled="busy"
              @click="refreshConnection"
            >
              {{ $t('invoicing.refresh') }}
            </ProButton>
          </template>
          <ProBadge
            v-else
            variant="neutral"
            class="invoicing-conn__restricted"
            :title="$t('invoicing.connectRestricted')"
            :aria-label="$t('invoicing.connectRestricted')"
          >
            {{ $t('invoicing.connectReadonly') }}
          </ProBadge>
        </div>
        <ProBadge variant="warning" data-testid="invoicing-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <ProCard v-if="!invoicingUiEnabled" data-testid="invoicing-under-development">
      <ProEmptyState :title="$t('invoicing.underDevelopment')" />
    </ProCard>

    <template v-else>
      <p v-if="error" class="pro-alert pro-mb-md" data-testid="invoicing-error">{{ error }}</p>
      <p
        v-if="secretsMismatch"
        class="pro-alert pro-mb-md"
        data-testid="invoicing-secrets-banner"
      >
        {{ $t('invoicing.secretsMismatchHint') }}
        <ProButton
          v-if="canManageBillit"
          variant="secondary"
          class="pro-mt-sm"
          test-id="invoicing-reconnect"
          :disabled="busy"
          @click="startConnect"
        >
          {{ $t('invoicing.reconnect') }}
        </ProButton>
      </p>
      <p
        v-if="prefillHint"
        class="pro-hint pro-mb-md"
        data-testid="invoicing-consultation-context"
      >
        {{ prefillHint }}
      </p>

      <!-- Liaison compte (uniquement pendant le parcours d’activation) -->
      <ProCard v-if="canManageBillit && showCompleteForm" class="pro-mb-lg">
        <form
          class="pro-form"
          data-testid="invoicing-complete-form"
          @submit.prevent="completeConnect"
        >
          <p class="pro-hint pro-mb-md">{{ $t('invoicing.completeHint') }}</p>
          <ProInput
            v-model="completeForm.partyId"
            test-id="invoicing-complete-party"
            :label="$t('invoicing.partyId')"
            required
          />
          <ProInput
            v-model="completeForm.apiKey"
            test-id="invoicing-complete-apikey"
            type="password"
            :label="$t('invoicing.apiKey')"
            required
          />
          <ProButton type="submit" test-id="invoicing-complete-submit" :disabled="busy">
            {{ $t('invoicing.complete') }}
          </ProButton>
        </form>
      </ProCard>

      <!-- Zone 2 : création -->
      <ProCard v-if="isActive && canWriteDocs" class="pro-mb-lg" data-testid="invoicing-create">
        <h3 class="pro-mb-md">{{ $t('invoicing.newDocument') }}</h3>
        <form class="pro-form" data-testid="invoicing-create-form" @submit.prevent="createDocument">
          <label class="pro-field">
            <span class="pro-field__label">{{ $t('invoicing.docType') }}</span>
            <select v-model="docForm.type" class="pro-input" data-testid="invoicing-doc-type" required>
              <option value="invoice">{{ $t('invoicing.typeInvoice') }}</option>
              <option value="credit_note">{{ $t('invoicing.typeCreditNote') }}</option>
              <option value="proforma">{{ $t('invoicing.typeProforma') }}</option>
            </select>
          </label>
          <label v-if="docForm.type === 'credit_note'" class="pro-field">
            <span class="pro-field__label">{{ $t('invoicing.relatedInvoice') }}</span>
            <select
              v-model="docForm.relatedDocumentId"
              class="pro-input"
              data-testid="invoicing-related-invoice"
              required
              @change="onRelatedInvoiceChange"
            >
              <option value="" disabled>{{ $t('invoicing.relatedInvoicePlaceholder') }}</option>
              <option
                v-for="inv in invoiceOptions"
                :key="inv.id"
                :value="inv.id"
              >
                {{ inv.label }}
              </option>
            </select>
            <p v-if="!invoiceOptions.length" class="pro-hint">{{ $t('invoicing.relatedInvoiceEmptyBillit') }}</p>
          </label>
          <ProInput
            v-model="docForm.name"
            test-id="invoicing-cp-name"
            :label="$t('invoicing.counterparty.name')"
            required
          />
          <label class="pro-field">
            <span class="pro-field__label">{{ $t('invoicing.counterparty.country') }}</span>
            <select v-model="docForm.country" class="pro-input" data-testid="invoicing-country" required>
              <option value="BE">BE</option>
              <option value="FR">FR</option>
              <option value="IT">IT</option>
              <option value="ES">ES</option>
            </select>
          </label>
          <ProInput
            v-if="docForm.country === 'BE' || docForm.country === 'FR' || docForm.country === 'IT'"
            v-model="docForm.vatNumber"
            test-id="invoicing-cp-vat"
            :label="$t('invoicing.counterparty.vatNumber')"
            required
          />
          <ProInput
            v-if="docForm.country === 'ES'"
            v-model="docForm.vatNumber"
            test-id="invoicing-cp-vat"
            :label="$t('invoicing.counterparty.vatNumber')"
          />
          <ProInput
            v-if="docForm.country === 'BE'"
            v-model="docForm.companyNumber"
            :label="$t('invoicing.counterparty.companyNumber')"
          />
          <ProInput
            v-if="docForm.country === 'FR'"
            v-model="docForm.siret"
            test-id="invoicing-cp-siret"
            :label="$t('invoicing.counterparty.siret')"
            :required="!docForm.siren"
          />
          <ProInput
            v-if="docForm.country === 'FR'"
            v-model="docForm.siren"
            test-id="invoicing-cp-siren"
            :label="$t('invoicing.counterparty.siren')"
            :required="!docForm.siret"
          />
          <ProInput
            v-if="docForm.country === 'IT'"
            v-model="docForm.codiceDestinatario"
            test-id="invoicing-cp-codice"
            :label="$t('invoicing.counterparty.codiceDestinatario')"
            :required="!docForm.pec"
          />
          <ProInput
            v-if="docForm.country === 'IT'"
            v-model="docForm.pec"
            test-id="invoicing-cp-pec"
            :label="$t('invoicing.counterparty.pec')"
            :required="!docForm.codiceDestinatario"
          />
          <ProInput
            v-if="docForm.country === 'ES'"
            v-model="docForm.taxId"
            test-id="invoicing-cp-taxid"
            :label="$t('invoicing.counterparty.taxId')"
            :required="!docForm.vatNumber"
          />
          <ProInput
            v-model="docForm.street"
            test-id="invoicing-cp-street"
            :label="$t('invoicing.counterparty.street')"
            :required="docForm.country === 'BE'"
          />
          <ProInput
            v-model="docForm.postal"
            test-id="invoicing-cp-postal"
            :label="$t('invoicing.counterparty.postal')"
            :required="docForm.country === 'BE'"
          />
          <ProInput
            v-model="docForm.city"
            test-id="invoicing-cp-city"
            :label="$t('invoicing.counterparty.city')"
            :required="docForm.country === 'BE'"
          />
          <ProInput
            v-model="docForm.lineDesc"
            test-id="invoicing-line-desc"
            :label="$t('invoicing.lineDescription')"
            required
          />
          <ProInput
            v-model="docForm.lineAmount"
            test-id="invoicing-line-amount"
            type="number"
            :label="$t('invoicing.lineAmountExcl')"
            required
          />
          <ProButton type="submit" test-id="invoicing-new-doc" :disabled="busy">
            {{ $t('invoicing.createDocument') }}
          </ProButton>
        </form>
      </ProCard>

      <!-- Zone 3 : liste -->
      <ProCard v-if="isActive && canWriteDocs" data-testid="invoicing-documents">
        <div class="invoicing-docs-head pro-mb-md">
          <h3>{{ $t('invoicing.documentsTitle') }}</h3>
          <div class="invoicing-filters" data-testid="invoicing-filters">
            <button
              v-for="f in statusFilters"
              :key="f.id"
              type="button"
              class="invoicing-filter"
              :class="{ 'invoicing-filter--active': docFilter === f.id }"
              :data-testid="`invoicing-filter-${f.id}`"
              @click="docFilter = f.id"
            >
              {{ f.label }}
            </button>
          </div>
        </div>
        <ProEmptyState v-if="!filteredDocuments.length" :title="$t('invoicing.emptyDocs')" />
        <ProTable v-else>
          <thead>
            <tr>
              <th>{{ $t('invoicing.colType') }}</th>
              <th>{{ $t('invoicing.colNumber') }}</th>
              <th>{{ $t('invoicing.colStatus') }}</th>
              <th>{{ $t('invoicing.colTotal') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="doc in filteredDocuments" :key="doc.id" :data-testid="`invoicing-doc-${doc.id}`">
              <td>{{ typeLabel(doc.type) }}</td>
              <td>{{ doc.number || doc.counterparty?.name || doc.id.slice(0, 8) }}</td>
              <td>
                <ProBadge :variant="docStatusVariant(doc.status)">{{ statusLabel(doc.status) }}</ProBadge>
                <span v-if="doc.peppolStatus" class="pro-hint"> · {{ doc.peppolStatus }}</span>
              </td>
              <td>{{ formatMoney(doc.totalInclCents) }}</td>
              <td>
                <ProButton
                  v-if="canSend(doc)"
                  variant="secondary"
                  :test-id="`invoicing-send-${doc.id}`"
                  :disabled="busy"
                  @click="sendDoc(doc.id)"
                >
                  {{ doc.type === 'proforma' ? $t('invoicing.issue') : $t('invoicing.send') }}
                </ProButton>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>

      <ProCard v-else-if="!isActive || !canWriteDocs">
        <p class="pro-hint" data-testid="invoicing-wip">{{ $t('invoicing.wip') }}</p>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import { INVOICING_UI_ENABLED } from '~/utils/invoicing-ui'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

definePageMeta({
  middleware: ['vet-only', 'practice-perm'],
  practicePermAny: ['clients.write', 'practice.settings'],
})

const runtimeConfig = useRuntimeConfig()
if (!isPublicFlagOn(runtimeConfig.public.billitEnabled)) {
  await navigateTo('/dashboard')
}

/** Source unique — voir `utils/invoicing-ui.ts`. */
const invoicingUiEnabled = INVOICING_UI_ENABLED

type Connection = {
  practiceId: string
  billitPartyId?: string
  status: string
  docsIncludedMonthly: number
  usageThisMonth: number
}

type Document = {
  id: string
  type: string
  status: string
  peppolStatus?: string
  totalInclCents: number
  billitOrderId?: string
  counterparty?: {
    name?: string
    country?: string
    vatNumber?: string
    companyNumber?: string
    street?: string
    city?: string
    postal?: string
    siret?: string
    siren?: string
    codiceDestinatario?: string
    pec?: string
    taxId?: string
  }
  number?: string
}

const { t } = useI18n()
const { canPractice } = usePracticePerms()
const canManageBillit = computed(() => canPractice('practice.settings'))
const canWriteDocs = computed(() => canPractice('clients.write'))
const route = useRoute()
const connection = ref<Connection | null>(null)
const documents = ref<Document[]>([])
const busy = ref(false)
const error = ref('')
const secretsMismatch = ref(false)
const resellerUrl = ref('')
const showCompleteForm = ref(false)
const prefillHint = ref('')
const docFilter = ref<'all' | 'draft' | 'sending' | 'delivered'>('all')
const completeForm = reactive({
  state: '',
  partyId: '',
  apiKey: '',
})
const docForm = reactive({
  type: 'invoice' as 'invoice' | 'credit_note' | 'proforma',
  relatedDocumentId: '',
  name: '',
  country: 'BE',
  vatNumber: '',
  companyNumber: '',
  siret: '',
  siren: '',
  codiceDestinatario: '',
  pec: '',
  taxId: '',
  street: '',
  postal: '',
  city: '',
  lineDesc: '',
  lineAmount: '',
})

const queryClientId = computed(() => String(route.query.clientUserId || ''))
const queryDafId = computed(() => String(route.query.dafId || ''))
const queryVisitId = computed(() => String(route.query.visitId || ''))
const queryMode = computed(() => String(route.query.mode || ''))

const isActive = computed(() => connection.value?.status === 'active')
const pendingRegistration = computed(() => connection.value?.status === 'pending_registration')
const pendingKyc = computed(() => connection.value?.status === 'pending_kyc')
const statusVariant = computed(() => {
  switch (connection.value?.status) {
    case 'active': return 'success'
    case 'pending_kyc':
    case 'pending_registration': return 'warning'
    case 'error':
    case 'suspended': return 'danger'
    default: return 'neutral'
  }
})

const connectionStatusLabel = computed(() => {
  const s = connection.value?.status
  if (!s) return '…'
  const key = `invoicing.connStatus.${s}`
  const translated = t(key)
  return translated === key ? s : translated
})

const statusFilters = computed(() => ([
  { id: 'all' as const, label: t('invoicing.filterAll') },
  { id: 'draft' as const, label: t('invoicing.filterDraft') },
  { id: 'sending' as const, label: t('invoicing.filterSending') },
  { id: 'delivered' as const, label: t('invoicing.filterDelivered') },
]))

const filteredDocuments = computed(() => {
  if (docFilter.value === 'all') return documents.value
  if (docFilter.value === 'draft') {
    return documents.value.filter((d) => d.status === 'draft' || d.status === 'rejected')
  }
  if (docFilter.value === 'sending') {
    return documents.value.filter((d) => d.status === 'sending' || d.status === 'issued')
  }
  return documents.value.filter((d) => d.status === 'delivered')
})

function unwrap<T>(res: any): T {
  return (res?.data ?? res) as T
}

function formatApiError(e: any): string {
  const err = e?.data?.error || e?.data?.data?.error
  const msgKey = err?.msgKey as string | undefined
  const message = err?.message || e?.message
  const gateway = err?.details?.gateway as string | undefined
  if (msgKey === 'invoicing_secrets_mismatch') {
    secretsMismatch.value = true
  }
  if (msgKey && t(`errors.${msgKey}`) !== `errors.${msgKey}`) {
    const base = t(`errors.${msgKey}`)
    return gateway ? `${base} (${gateway})` : base
  }
  if (message) return gateway ? `${message} (${gateway})` : message
  return t('invoicing.errorGeneric')
}

function canSend(doc: Document) {
  if (doc.type === 'proforma') return doc.status === 'draft'
  return doc.status === 'draft' || doc.status === 'rejected'
}

function formatMoney(cents: number) {
  return `${(cents / 100).toFixed(2)} €`
}

function typeLabel(type: string) {
  switch (type) {
    case 'invoice': return t('invoicing.typeInvoice')
    case 'credit_note': return t('invoicing.typeCreditNote')
    case 'proforma': return t('invoicing.typeProforma')
    default: return type
  }
}

function statusLabel(status: string) {
  const key = `invoicing.docStatus.${status}`
  const translated = t(key)
  return translated === key ? status : translated
}

function docStatusVariant(status: string) {
  switch (status) {
    case 'delivered': return 'success'
    case 'sending':
    case 'issued': return 'warning'
    case 'rejected': return 'danger'
    default: return 'neutral'
  }
}

const invoiceOptions = computed(() =>
  documents.value
    .filter((d) =>
      d.type === 'invoice'
      && ['issued', 'delivered', 'rejected'].includes(d.status)
      && !!(d.billitOrderId || d.number),
    )
    .map((d) => ({
      id: d.id,
      label: [
        d.number || d.id.slice(0, 8),
        d.counterparty?.name,
        formatMoney(d.totalInclCents),
        statusLabel(d.status),
      ].filter(Boolean).join(' · '),
    })),
)

function onRelatedInvoiceChange() {
  const inv = documents.value.find((d) => d.id === docForm.relatedDocumentId)
  if (!inv?.counterparty) return
  const cp = inv.counterparty
  if (cp.name) docForm.name = cp.name
  if (cp.country) docForm.country = cp.country
  if (cp.vatNumber) docForm.vatNumber = cp.vatNumber
  if (cp.companyNumber) docForm.companyNumber = cp.companyNumber
  if (cp.street) docForm.street = cp.street
  if (cp.city) docForm.city = cp.city
  if (cp.postal) docForm.postal = cp.postal
  if (cp.siret) docForm.siret = cp.siret
  if (cp.siren) docForm.siren = cp.siren
  if (cp.codiceDestinatario) docForm.codiceDestinatario = cp.codiceDestinatario
  if (cp.pec) docForm.pec = cp.pec
  if (cp.taxId) docForm.taxId = cp.taxId
}

async function loadConnection() {
  const res = await $fetch('/api/invoicing/connection')
  connection.value = unwrap<Connection>(res)
}

async function loadDocuments() {
  if (!isActive.value || !canWriteDocs.value) {
    documents.value = []
    return
  }
  const res = await $fetch('/api/invoicing/documents')
  documents.value = unwrap<Document[]>(res) || []
}

async function startConnect() {
  busy.value = true
  error.value = ''
  secretsMismatch.value = false
  try {
    const res = await $fetch('/api/invoicing/connect/start', { method: 'POST' })
    const data = unwrap<{ resellerUrl: string, state: string }>(res)
    resellerUrl.value = data.resellerUrl
    completeForm.state = data.state
    showCompleteForm.value = true
    await loadConnection()
  } catch (e: any) {
    error.value = formatApiError(e)
  } finally {
    busy.value = false
  }
}

function openReseller() {
  if (resellerUrl.value) {
    window.open(resellerUrl.value, '_blank', 'noopener')
  }
}

async function completeConnect() {
  busy.value = true
  error.value = ''
  secretsMismatch.value = false
  try {
    const res = await $fetch('/api/invoicing/connect/complete', {
      method: 'POST',
      body: { ...completeForm },
    })
    connection.value = unwrap<Connection>(res)
    showCompleteForm.value = false
    await loadDocuments()
  } catch (e: any) {
    error.value = formatApiError(e)
  } finally {
    busy.value = false
  }
}

async function refreshConnection() {
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch('/api/invoicing/connect/refresh', { method: 'POST' })
    connection.value = unwrap<Connection>(res)
    await loadDocuments()
  } catch (e: any) {
    error.value = formatApiError(e)
  } finally {
    busy.value = false
  }
}

async function createDocument() {
  busy.value = true
  error.value = ''
  try {
    const excl = Math.round(Number(docForm.lineAmount) * 100)
    if (!Number.isFinite(excl) || excl <= 0) {
      error.value = t('invoicing.amountRequired')
      return
    }
    if (docForm.type === 'credit_note' && !docForm.relatedDocumentId) {
      error.value = t('invoicing.relatedInvoiceRequired')
      return
    }
    if (docForm.country === 'BE' && (!docForm.street || !docForm.city || !docForm.postal)) {
      error.value = t('invoicing.beAddressRequired')
      return
    }
    const vatPercent = docForm.country === 'IT' ? 22 : 21
    await $fetch('/api/invoicing/documents', {
      method: 'POST',
      body: {
        type: docForm.type,
        relatedDocumentId: docForm.type === 'credit_note' ? docForm.relatedDocumentId : undefined,
        visitId: queryVisitId.value || undefined,
        dafId: queryDafId.value || undefined,
        counterparty: {
          name: docForm.name,
          country: docForm.country,
          vatNumber: docForm.vatNumber || undefined,
          companyNumber: docForm.companyNumber || undefined,
          siret: docForm.siret || undefined,
          siren: docForm.siren || undefined,
          codiceDestinatario: docForm.codiceDestinatario || undefined,
          pec: docForm.pec || undefined,
          taxId: docForm.taxId || undefined,
          street: docForm.street || undefined,
          city: docForm.city || undefined,
          postal: docForm.postal || undefined,
        },
        lines: [
          {
            description: docForm.lineDesc,
            quantity: 1,
            unitPriceExclCents: excl,
            vatPercent,
          },
        ],
      },
    })
    docForm.relatedDocumentId = ''
    await loadDocuments()
  } catch (e: any) {
    error.value = formatApiError(e)
  } finally {
    busy.value = false
  }
}

async function sendDoc(id: string) {
  busy.value = true
  error.value = ''
  secretsMismatch.value = false
  try {
    await $fetch(`/api/invoicing/documents/${id}/send`, { method: 'POST' })
    await loadDocuments()
    await loadConnection()
  } catch (e: any) {
    error.value = formatApiError(e)
  } finally {
    busy.value = false
  }
}

onMounted(async () => {
  if (!invoicingUiEnabled) return
  try {
    await loadConnection()
    await loadDocuments()
    await applyConsultationPrefill()
  } catch (e: any) {
    error.value = formatApiError(e)
  }
})

async function applyConsultationPrefill() {
  if (!queryClientId.value && !queryDafId.value && !queryVisitId.value) return
  const hints: string[] = []
  if (queryMode.value === 'fromDaf') hints.push(t('invoicing.prefillFromDaf'))
  else if (queryMode.value === 'direct') hints.push(t('invoicing.prefillDirect'))
  if (queryVisitId.value) hints.push(t('invoicing.prefillVisit'))

  if (queryClientId.value) {
    try {
      const res = await $fetch(`/api/clients/${queryClientId.value}`)
      const c: any = unwrap(res)
      if (c?.fullName) {
        docForm.name = c.fullName
        hints.push(c.fullName)
      }
    }
    catch {
      // ignore — keep form empty until user fills
    }
  }

  if (queryDafId.value) {
    try {
      const res = await $fetch(`/api/vet/pharmacy/daf/${queryDafId.value}`)
      const doc: any = unwrap(res)
      const items = Array.isArray(doc?.items) ? doc.items : []
      if (items.length) {
        docForm.lineDesc = items
          .map((it: any) => `${it.medicationName || 'Médicament'} × ${it.qty}`)
          .join(', ')
          .slice(0, 200)
        hints.push(t('invoicing.prefillDafLines', { n: items.length }))

        let totalExclCents = 0
        let priced = 0
        for (const it of items) {
          const medId = String(it.medicationId || it.refMedicationId || '')
          const qty = Number(it.qty) || 0
          if (!medId || qty <= 0) continue
          try {
            const priceRes = await $fetch(`/api/vet/pharmacy/prices/${medId}`)
            const price: any = unwrap(priceRes)
            const sell = Number(price?.sellPriceCents)
            if (Number.isFinite(sell) && sell > 0) {
              totalExclCents += Math.round(sell * qty)
              priced++
            }
          }
          catch {
            // price optional
          }
        }
        if (priced > 0 && totalExclCents > 0) {
          docForm.lineAmount = (totalExclCents / 100).toFixed(2)
          hints.push(t('invoicing.prefillDafAmount', { amount: docForm.lineAmount }))
        }
      }
      if (doc?.clientName && !docForm.name) {
        docForm.name = doc.clientName
      }
    }
    catch {
      // DAF optional if pharmacy off
    }
  }

  prefillHint.value = hints.filter(Boolean).join(' · ')
}
</script>

<style scoped>
.invoicing-conn {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
  min-width: 0;
}
.invoicing-conn__title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  white-space: nowrap;
}
.invoicing-conn__restricted {
  cursor: help;
}
.invoicing-docs-head {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}
.invoicing-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
.invoicing-filter {
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface);
  color: var(--pf-vet-text, inherit);
  border-radius: 999px;
  padding: 0.25rem 0.75rem;
  font-size: 0.85rem;
  cursor: pointer;
}
.invoicing-filter--active {
  border-color: var(--pf-vet-accent);
  color: var(--pf-vet-accent);
  font-weight: 600;
}
.pro-mt-md { margin-top: 1rem; }
.pro-mt-sm { margin-top: 0.5rem; }
</style>
