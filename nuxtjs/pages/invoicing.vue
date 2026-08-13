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
      <div
        v-if="prefillHint"
        class="invoicing-prefill"
        data-testid="invoicing-consultation-context"
        role="status"
      >
        <ProIcon name="receipt_long" class="invoicing-prefill__icon" aria-hidden="true" />
        <p class="invoicing-prefill__text">{{ prefillHint }}</p>
      </div>

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
      <ProCard v-if="isActive && canWriteDocs" class="pro-mb-lg invoicing-create" data-testid="invoicing-create">
        <h3 class="invoicing-create__title">{{ $t('invoicing.newDocument') }}</h3>
        <form class="pro-form invoicing-create__form" data-testid="invoicing-create-form" @submit.prevent="createDocument">
          <!-- 1. Document -->
          <section class="invoicing-section" aria-labelledby="invoicing-section-doc">
            <h4 id="invoicing-section-doc" class="pro-section-title">{{ $t('invoicing.sectionDocument') }}</h4>
            <div class="pro-field">
              <span class="pro-field__label" id="invoicing-doc-type-label">{{ $t('invoicing.docType') }}</span>
              <div
                class="pro-toggle"
                role="group"
                aria-labelledby="invoicing-doc-type-label"
                data-testid="invoicing-doc-type"
                :data-value="docForm.type"
              >
                <button
                  type="button"
                  class="pro-toggle-btn"
                  :class="{ 'pro-toggle-btn--active': docForm.type === 'invoice' }"
                  data-testid="invoicing-doc-type-invoice"
                  :aria-pressed="docForm.type === 'invoice'"
                  @click="docForm.type = 'invoice'"
                >
                  {{ $t('invoicing.typeInvoice') }}
                </button>
                <button
                  type="button"
                  class="pro-toggle-btn"
                  :class="{ 'pro-toggle-btn--active': docForm.type === 'credit_note' }"
                  data-testid="invoicing-doc-type-credit_note"
                  :aria-pressed="docForm.type === 'credit_note'"
                  @click="docForm.type = 'credit_note'"
                >
                  {{ $t('invoicing.typeCreditNote') }}
                </button>
                <button
                  type="button"
                  class="pro-toggle-btn"
                  :class="{ 'pro-toggle-btn--active': docForm.type === 'proforma' }"
                  data-testid="invoicing-doc-type-proforma"
                  :aria-pressed="docForm.type === 'proforma'"
                  @click="docForm.type = 'proforma'"
                >
                  {{ $t('invoicing.typeProforma') }}
                </button>
              </div>
            </div>
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
          </section>

          <!-- 2. Client -->
          <section class="invoicing-section" aria-labelledby="invoicing-section-client">
            <h4 id="invoicing-section-client" class="pro-section-title">{{ $t('invoicing.sectionClient') }}</h4>
            <div class="pro-field" data-testid="invoicing-client-search">
              <label class="pro-field__label" for="invoicing-client-search-input">{{ $t('invoicing.clientSearch') }}</label>
              <ProCombobox
                v-model="clientSearch"
                input-id="invoicing-client-search-input"
                :placeholder="$t('invoicing.clientSearchPlaceholder')"
                :min-chars="1"
                :search-fn="searchClients"
                @select="onClientSelect"
              />
            </div>
            <div class="pro-field">
              <span class="pro-field__label" id="invoicing-cp-kind-label">{{ $t('invoicing.counterparty.kind') }}</span>
              <div
                class="pro-toggle"
                role="group"
                aria-labelledby="invoicing-cp-kind-label"
                data-testid="invoicing-cp-kind"
                :data-value="docForm.customerKind"
              >
                <button
                  type="button"
                  class="pro-toggle-btn"
                  :class="{ 'pro-toggle-btn--active': docForm.customerKind === 'individual' }"
                  data-testid="invoicing-cp-kind-individual"
                  :aria-pressed="docForm.customerKind === 'individual'"
                  @click="docForm.customerKind = 'individual'"
                >
                  {{ $t('invoicing.counterparty.kindIndividual') }}
                </button>
                <button
                  type="button"
                  class="pro-toggle-btn"
                  :class="{ 'pro-toggle-btn--active': docForm.customerKind === 'business' }"
                  data-testid="invoicing-cp-kind-business"
                  :aria-pressed="docForm.customerKind === 'business'"
                  @click="docForm.customerKind = 'business'"
                >
                  {{ $t('invoicing.counterparty.kindBusiness') }}
                </button>
              </div>
              <p class="pro-hint">
                {{ isIndividual ? $t('invoicing.counterparty.kindIndividualHint') : $t('invoicing.counterparty.kindBusinessHint') }}
              </p>
            </div>
            <div class="invoicing-grid invoicing-grid--2">
              <ProInput
                v-model="docForm.name"
                test-id="invoicing-cp-name"
                :label="$t('invoicing.counterparty.name')"
                required
              />
              <div>
                <ProInput
                  v-model="docForm.email"
                  test-id="invoicing-cp-email"
                  type="email"
                  :label="$t('invoicing.counterparty.email')"
                  :required="docForm.type === 'proforma' || isIndividual"
                />
                <p v-if="docForm.type === 'proforma'" class="pro-hint">{{ $t('invoicing.counterparty.emailProformaHint') }}</p>
                <p v-else-if="isIndividual" class="pro-hint">{{ $t('invoicing.counterparty.emailIndividualHint') }}</p>
              </div>
            </div>
            <label class="pro-field invoicing-country">
              <span class="pro-field__label">{{ $t('invoicing.counterparty.country') }}</span>
              <select v-model="docForm.country" class="pro-input" data-testid="invoicing-country" required>
                <option value="BE">BE</option>
                <option value="FR">FR</option>
                <option value="IT">IT</option>
                <option value="ES">ES</option>
              </select>
            </label>
            <div v-if="!isIndividual" class="invoicing-grid invoicing-grid--2">
              <ProInput
                v-if="vatRequired || docForm.country === 'ES'"
                v-model="docForm.vatNumber"
                test-id="invoicing-cp-vat"
                :label="$t('invoicing.counterparty.vatNumber')"
                :required="vatRequired"
              />
              <ProInput
                v-if="docForm.country === 'BE'"
                v-model="docForm.companyNumber"
                test-id="invoicing-cp-company"
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
            </div>
            <div class="invoicing-address">
              <p class="invoicing-address__label">{{ $t('invoicing.sectionAddress') }}</p>
              <ProInput
                v-model="docForm.street"
                test-id="invoicing-cp-street"
                :label="$t('invoicing.counterparty.street')"
                :required="addressRequired"
              />
              <div class="invoicing-grid invoicing-grid--postal">
                <ProInput
                  v-model="docForm.postal"
                  test-id="invoicing-cp-postal"
                  :label="$t('invoicing.counterparty.postal')"
                  :required="addressRequired"
                />
                <ProInput
                  v-model="docForm.city"
                  test-id="invoicing-cp-city"
                  :label="$t('invoicing.counterparty.city')"
                  :required="addressRequired"
                />
              </div>
            </div>
          </section>

          <!-- 3. Lignes -->
          <section class="invoicing-section invoicing-lines" data-testid="invoicing-lines" aria-labelledby="invoicing-section-lines">
            <div class="invoicing-lines__head">
              <h4 id="invoicing-section-lines" class="pro-section-title">{{ $t('invoicing.sectionLines') }}</h4>
              <ProButton type="button" variant="secondary" test-id="invoicing-add-line" @click="addLine">
                <ProIcon name="add" aria-hidden="true" />
                {{ $t('invoicing.addLine') }}
              </ProButton>
            </div>
            <div
              v-for="(line, idx) in docLines"
              :key="line.key"
              class="invoicing-line"
              :data-testid="`invoicing-line-${idx}`"
            >
              <div class="invoicing-line__head">
                <span class="invoicing-line__index">{{ $t('invoicing.lines.lineN', { n: idx + 1 }) }}</span>
                <span class="invoicing-line__subtotal">
                  {{ $t('invoicing.lines.lineSubtotal') }}
                  <strong>{{ formatMoney(lineInclCents(line)) }}</strong>
                </span>
                <ProButton
                  v-if="docLines.length > 1"
                  type="button"
                  variant="ghost"
                  class="invoicing-line__remove"
                  :test-id="`invoicing-remove-line-${idx}`"
                  :aria-label="$t('invoicing.removeLine')"
                  @click="removeLine(idx)"
                >
                  <ProIcon name="delete" aria-hidden="true" />
                  {{ $t('invoicing.removeLine') }}
                </ProButton>
              </div>
              <ProInput
                v-model="line.description"
                :test-id="`invoicing-line-${idx}-desc`"
                :label="$t('invoicing.lines.description')"
                required
              />
              <div class="invoicing-line__row">
                <ProInput
                  v-model="line.quantity"
                  :test-id="`invoicing-line-${idx}-qty`"
                  type="number"
                  :label="$t('invoicing.lines.qty')"
                  required
                />
                <ProInput
                  v-model="line.unitPriceExcl"
                  :test-id="`invoicing-line-${idx}-unit`"
                  type="number"
                  :label="$t('invoicing.lines.unitPrice')"
                  required
                />
                <label class="pro-field">
                  <span class="pro-field__label">{{ $t('invoicing.lines.vat') }}</span>
                  <select
                    v-model.number="line.vatPercent"
                    class="pro-input"
                    :data-testid="`invoicing-line-${idx}-vat`"
                  >
                    <option v-for="rate in vatPresets" :key="rate" :value="rate">{{ rate }}</option>
                  </select>
                </label>
              </div>
              <div v-if="pharmacyEnabled" class="invoicing-line__med" :data-testid="`invoicing-line-${idx}-med`">
                <label class="pro-field__label" :for="`invoicing-med-${idx}`">{{ $t('invoicing.pickMedication') }}</label>
                <ProCombobox
                  v-model="line.med"
                  :input-id="`invoicing-med-${idx}`"
                  :placeholder="$t('invoicing.medSearchPlaceholder')"
                  :min-chars="2"
                  :search-fn="searchMedications"
                  @select="(item) => onMedicationSelect(idx, item)"
                />
              </div>
            </div>
          </section>

          <div class="invoicing-sticky-bar" data-testid="invoicing-totals">
            <div class="invoicing-totals">
              <div class="invoicing-totals__item">
                <span class="invoicing-totals__label">{{ $t('invoicing.totalExcl') }}</span>
                <span class="invoicing-totals__value">{{ formatMoney(lineTotals.excl) }}</span>
              </div>
              <div class="invoicing-totals__item">
                <span class="invoicing-totals__label">{{ $t('invoicing.totalVat') }}</span>
                <span class="invoicing-totals__value">{{ formatMoney(lineTotals.vat) }}</span>
              </div>
              <div class="invoicing-totals__item invoicing-totals__item--incl">
                <span class="invoicing-totals__label">{{ $t('invoicing.totalIncl') }}</span>
                <span class="invoicing-totals__value">{{ formatMoney(lineTotals.incl) }}</span>
              </div>
            </div>
            <ProButton type="submit" test-id="invoicing-new-doc" :disabled="busy">
              {{ $t('invoicing.createDocument') }}
            </ProButton>
          </div>
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
                <span
                  v-if="deliveryLabel(doc)"
                  class="pro-hint"
                  :title="doc.peppolStatus"
                  :data-testid="`invoicing-delivery-${doc.id}`"
                > · {{ deliveryLabel(doc) }}</span>
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
                  {{ doc.type === 'proforma' ? $t('invoicing.sendToClient') : $t('invoicing.send') }}
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
import { counterpartyKind } from '~/utils/invoicing-counterparty'
import { isRedundantDelivery, parseDeliveryStatus } from '~/utils/invoicing-delivery-status'
import {
  buildInvoicePayloadLines,
  defaultVatPercent,
  formatInvoiceMoney,
  isInvoiceAddressRequired,
  isInvoiceVatRequired,
  lineAmountsCents,
  sumInvoiceLineTotals,
} from '~/utils/invoicing-lines'
import { prefillDraft } from '~/utils/invoicing-prefill'
import { isPublicFlagOn } from '~/utils/public-feature-flag'
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'

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
const pharmacyEnabled = computed(() => isPublicFlagOn(runtimeConfig.public.pharmacyEnabled))

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
    customerKind?: string
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
    email?: string
  }
  number?: string
}

type ClientRow = {
  userId: string
  email: string
  fullName: string
  contactPhone?: string
  billingVatNumber?: string
  billingCompanyNumber?: string
  billingStreet?: string
  billingCity?: string
  billingPostal?: string
  billingCountry?: string
  billingCustomerKind?: string
}

type DocLine = {
  key: string
  description: string
  quantity: string
  unitPriceExcl: string
  vatPercent: number
  med: ProComboboxItem | null
}

let lineKeySeq = 0
function emptyLine(vatDefault = 21): DocLine {
  lineKeySeq += 1
  return {
    key: `line-${lineKeySeq}`,
    description: '',
    quantity: '1',
    unitPriceExcl: '',
    vatPercent: vatDefault,
    med: null,
  }
}

const { t } = useI18n()
const { canPractice } = usePracticePerms()
const canManageBillit = computed(() => canPractice('practice.settings'))
const canWriteDocs = computed(() => canPractice('clients.write'))
const route = useRoute()
const connection = ref<Connection | null>(null)
const documents = ref<Document[]>([])
const clients = ref<ClientRow[]>([])
const clientSearch = ref<ProComboboxItem | null>(null)
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
  email: '',
  // Un cabinet facture majoritairement des particuliers : c'est le défaut.
  customerKind: 'individual' as 'individual' | 'business',
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
})
const docLines = ref<DocLine[]>([emptyLine(defaultVatPercent('BE'))])
const vatPresets = [0, 6, 12, 21, 22]

const queryClientId = computed(() => String(route.query.clientUserId || ''))
const queryDafId = computed(() => String(route.query.dafId || ''))
const queryVisitId = computed(() => String(route.query.visitId || ''))
const queryMode = computed(() => String(route.query.mode || ''))

const isIndividual = computed(() => docForm.customerKind === 'individual')
const vatRequired = computed(() => isInvoiceVatRequired(docForm.customerKind, docForm.country))
const addressRequired = computed(() => isInvoiceAddressRequired(docForm.customerKind, docForm.country))

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

const lineTotals = computed(() => sumInvoiceLineTotals(docLines.value))

watch(
  () => docForm.type,
  (type) => {
    if (type !== 'credit_note') {
      docForm.relatedDocumentId = ''
    }
  },
)

watch(
  () => docForm.country,
  (country, prev) => {
    const oldDef = defaultVatPercent(prev || 'BE')
    const newDef = defaultVatPercent(country)
    if (oldDef === newDef) return
    // N'écrase que le défaut pays précédent — préserve 0/6/12 et TVA pharma posée explicitement.
    for (const line of docLines.value) {
      if (line.vatPercent === oldDef) {
        line.vatPercent = newDef
      }
    }
  },
)

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
  // API Go already localizes msgKey (be_address_required, be_vat_invalid, …).
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
  return formatInvoiceMoney(cents)
}

function lineInclCents(line: DocLine): number {
  return lineAmountsCents(line)?.incl ?? 0
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

// Statut d'acheminement : traduit, avec le canal en préfixe quand la facture
// part par email. La valeur technique reste dans l'infobulle pour le support.
function deliveryLabel(doc: { status: string, peppolStatus?: string }) {
  const status = parseDeliveryStatus(doc.peppolStatus)
  if (!status || isRedundantDelivery(status, doc.status)) return ''
  const label = status.key ? t(status.key) : status.raw
  return status.byEmail ? t('invoicing.deliveryStatus.byEmail', { status: label }) : label
}

function docStatusVariant(status: string) {
  switch (status) {
    case 'delivered':
    case 'accepted': return 'success'
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
      && ['issued', 'delivered', 'rejected', 'sending'].includes(d.status)
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

function applyCounterparty(cp: {
  name?: string
  email?: string
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
  customerKind?: string
}, opts?: { keepCountryFallback?: string }) {
  docForm.name = cp.name || ''
  docForm.email = cp.email || ''
  docForm.customerKind = counterpartyKind(cp)
  docForm.country = cp.country || opts?.keepCountryFallback || 'BE'
  docForm.vatNumber = cp.vatNumber || ''
  docForm.companyNumber = cp.companyNumber || ''
  docForm.street = cp.street || ''
  docForm.city = cp.city || ''
  docForm.postal = cp.postal || ''
  docForm.siret = cp.siret || ''
  docForm.siren = cp.siren || ''
  docForm.codiceDestinatario = cp.codiceDestinatario || ''
  docForm.pec = cp.pec || ''
  docForm.taxId = cp.taxId || ''
}

function resetCounterpartyForm() {
  applyCounterparty({}, { keepCountryFallback: 'BE' })
  docForm.relatedDocumentId = ''
  clientSearch.value = null
}

function onRelatedInvoiceChange() {
  const inv = documents.value.find((d) => d.id === docForm.relatedDocumentId)
  if (!inv?.counterparty) return
  applyCounterparty(inv.counterparty)
}

function applyClientToForm(c: ClientRow) {
  // Le type vient de la fiche client quand il y est renseigné ; sinon
  // applyCounterparty le déduit des identifiants fiscaux (fiches antérieures).
  applyCounterparty({
    name: c.fullName || '',
    email: c.email || '',
    country: c.billingCountry || undefined,
    customerKind: c.billingCustomerKind || undefined,
    vatNumber: c.billingVatNumber || '',
    companyNumber: c.billingCompanyNumber || '',
    street: c.billingStreet || '',
    city: c.billingCity || '',
    postal: c.billingPostal || '',
  }, { keepCountryFallback: docForm.country || 'BE' })
}

async function loadClients() {
  try {
    const res: any = await $fetch('/api/clients')
    const list = unwrap<ClientRow[]>(res)
    clients.value = Array.isArray(list) ? list : []
  } catch {
    clients.value = []
  }
}

async function searchClients(q: string): Promise<ProComboboxItem[]> {
  const needle = q.trim().toLowerCase()
  if (!needle) return []
  return clients.value
    .filter((c) => {
      const name = String(c.fullName ?? '').toLowerCase()
      const email = String(c.email ?? '').toLowerCase()
      const phone = String(c.contactPhone ?? '').toLowerCase()
      return name.includes(needle) || email.includes(needle) || phone.includes(needle)
    })
    .slice(0, 20)
    .map((c) => ({
      id: c.userId,
      label: c.fullName || c.email,
      hint: c.email,
      raw: c,
    }))
}

function onClientSelect(item: ProComboboxItem) {
  const client = (item.raw as ClientRow | undefined) ?? clients.value.find((c) => c.userId === item.id)
  if (!client) return
  applyClientToForm(client)
}

function addLine() {
  docLines.value.push(emptyLine(defaultVatPercent(docForm.country)))
}

function removeLine(idx: number) {
  if (docLines.value.length <= 1) return
  docLines.value.splice(idx, 1)
}

async function searchMedications(q: string): Promise<ProComboboxItem[]> {
  if (!pharmacyEnabled.value) return []
  const res = await $fetch<any>('/api/vet/pharmacy/medications/search', { query: { q, limit: '20' } })
  const items = unwrap<{ items?: any[] }>(res)?.items ?? []
  return items.map((m: any) => ({
    id: String(m.id),
    label: String(m.name || ''),
    hint: m.cnk ? String(m.cnk) : undefined,
    raw: m,
  }))
}

async function onMedicationSelect(idx: number, item: ProComboboxItem) {
  const line = docLines.value[idx]
  if (!line) return
  line.description = item.label
  line.med = item
  try {
    const priceRes = await $fetch(`/api/vet/pharmacy/prices/${item.id}`)
    const price: any = unwrap(priceRes)
    const sell = Number(price?.sellPriceCents)
    if (Number.isFinite(sell) && sell > 0) {
      line.unitPriceExcl = (sell / 100).toFixed(2)
    }
    const vat = Number(price?.vatPercent)
    if (Number.isFinite(vat) && vat >= 0) {
      line.vatPercent = vat
    }
    // Sinon conserver la TVA ligne (y compris 0 %) — pas de `!vatPercent`.
  } catch {
    // price optional
  }
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

function buildPayloadLines() {
  return buildInvoicePayloadLines(docLines.value)
}

async function createDocument() {
  busy.value = true
  error.value = ''
  try {
    const lines = buildPayloadLines()
    if (!lines.length) {
      error.value = t('invoicing.linesRequired')
      return
    }
    if (docForm.type === 'credit_note' && !docForm.relatedDocumentId) {
      error.value = t('invoicing.relatedInvoiceRequired')
      return
    }
    if (addressRequired.value && (!docForm.street || !docForm.city || !docForm.postal)) {
      error.value = isIndividual.value
        ? t('invoicing.individualAddressRequired')
        : t('invoicing.beAddressRequired')
      return
    }
    if (isIndividual.value && !docForm.email.includes('@')) {
      error.value = t('invoicing.individualEmailRequired')
      return
    }
    const business = !isIndividual.value
    await $fetch('/api/invoicing/documents', {
      method: 'POST',
      body: {
        type: docForm.type,
        relatedDocumentId: docForm.type === 'credit_note' ? docForm.relatedDocumentId : undefined,
        visitId: queryVisitId.value || undefined,
        dafId: queryDafId.value || undefined,
        counterparty: {
          name: docForm.name,
          email: docForm.email || undefined,
          customerKind: docForm.customerKind,
          country: docForm.country,
          vatNumber: business ? docForm.vatNumber || undefined : undefined,
          companyNumber: business ? docForm.companyNumber || undefined : undefined,
          siret: business ? docForm.siret || undefined : undefined,
          siren: business ? docForm.siren || undefined : undefined,
          codiceDestinatario: business ? docForm.codiceDestinatario || undefined : undefined,
          pec: business ? docForm.pec || undefined : undefined,
          taxId: business ? docForm.taxId || undefined : undefined,
          street: docForm.street || undefined,
          city: docForm.city || undefined,
          postal: docForm.postal || undefined,
        },
        lines,
      },
    })
    resetCounterpartyForm()
    docLines.value = [emptyLine(defaultVatPercent(docForm.country))]
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
    await Promise.all([loadConnection(), loadClients()])
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
      if (c) {
        applyClientToForm(c as ClientRow)
        if (c.fullName) hints.push(c.fullName)
      }
    }
    catch {
      // ignore — keep form empty until user fills
    }
  }

  if (queryDafId.value || queryVisitId.value) {
    // Lignes proposées par l'API : acte au tarif du type de RDV + médicaments du
    // DAF finalisé. Le calcul est côté serveur (une requête au lieu d'un appel
    // de prix par médicament, et mêmes droits que la création du document).
    try {
      const params = new URLSearchParams()
      if (queryVisitId.value) params.set('visitId', queryVisitId.value)
      if (queryDafId.value) params.set('dafId', queryDafId.value)
      const res = await $fetch(`/api/invoicing/prefill?${params.toString()}`)
      const payload: any = unwrap(res)
      const defVat = defaultVatPercent(docForm.country)
      const draft = prefillDraft(payload, defVat)
      const nextLines: DocLine[] = draft.lines.map((l) => ({ ...emptyLine(defVat), ...l }))
      if (nextLines.length) {
        docLines.value = nextLines
        if (payload?.visitTypeName) hints.push(t('invoicing.prefillVisitType', { name: payload.visitTypeName }))
        if (draft.dafLineCount > 0) hints.push(t('invoicing.prefillDafLines', { n: draft.dafLineCount }))
        if (draft.estimatedExclCents > 0) {
          hints.push(t('invoicing.prefillEstimatedAmount', { amount: (draft.estimatedExclCents / 100).toFixed(2) }))
        }
      }
      if (payload?.clientName && !docForm.name) {
        docForm.name = payload.clientName
      }
    }
    catch {
      // Préremplissage best-effort : sans lui le véto saisit manuellement.
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
.invoicing-prefill {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding: 0.85rem 1rem;
  border: 1px solid color-mix(in srgb, var(--pf-vet-accent) 35%, var(--pf-vet-border));
  border-radius: var(--pf-vet-radius, 8px);
  background: color-mix(in srgb, var(--pf-vet-accent) 8%, var(--pf-vet-surface));
}
.invoicing-prefill__icon {
  flex-shrink: 0;
  margin-top: 0.1rem;
  color: var(--pf-vet-accent);
  font-size: 1.35rem;
}
.invoicing-prefill__text {
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.45;
  color: var(--pf-vet-text, inherit);
}
.invoicing-create__title {
  margin: 0 0 1.25rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--pf-vet-primary);
}
.invoicing-create__form {
  display: flex;
  flex-direction: column;
  gap: 0;
}
.invoicing-create__form :deep(.pro-toggle) {
  flex-wrap: wrap;
}
.invoicing-section {
  padding: 1.15rem 0 1.35rem;
  border-top: 1px solid var(--pf-vet-border, #e5e7eb);
}
.invoicing-section:first-of-type {
  padding-top: 0;
  border-top: none;
}
.pro-section-title {
  margin: 0 0 1rem;
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--pf-vet-text-muted, #64748b);
}
.invoicing-grid {
  display: grid;
  gap: 0.75rem 1rem;
  margin-bottom: 0.25rem;
}
.invoicing-grid--2 {
  grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
}
.invoicing-grid--postal {
  grid-template-columns: minmax(6rem, 8rem) minmax(0, 1fr);
}
.invoicing-country {
  max-width: 10rem;
}
.invoicing-address {
  margin-top: 0.5rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--pf-vet-border, #e5e7eb);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-bg, #f8fafc);
}
.invoicing-address__label {
  margin: 0 0 0.75rem;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--pf-vet-primary);
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
.invoicing-lines {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.invoicing-lines__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.25rem;
}
.invoicing-lines__head .pro-section-title {
  margin-bottom: 0;
}
.invoicing-line {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.9rem 1rem;
  border: 1px solid var(--pf-vet-border, #e5e7eb);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-surface);
}
.invoicing-line__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
  margin-bottom: 0.15rem;
}
.invoicing-line__index {
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
  color: var(--pf-vet-text-muted, #64748b);
}
.invoicing-line__subtotal {
  margin-left: auto;
  font-size: 0.9rem;
  color: var(--pf-vet-text-muted, #64748b);
}
.invoicing-line__subtotal strong {
  color: var(--pf-vet-primary);
  font-variant-numeric: tabular-nums;
}
.invoicing-line__remove {
  flex-shrink: 0;
}
.invoicing-line__row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
  gap: 0.75rem;
}
.invoicing-line__med {
  max-width: 28rem;
}
.invoicing-sticky-bar {
  position: sticky;
  bottom: 0;
  z-index: 2;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin: 0.5rem -0.25rem -0.25rem;
  padding: 0.9rem 1rem;
  border-top: 1px solid var(--pf-vet-border, #e5e7eb);
  border-radius: 0 0 var(--pf-vet-radius, 8px) var(--pf-vet-radius, 8px);
  background: color-mix(in srgb, var(--pf-vet-surface) 92%, white);
  box-shadow: 0 -6px 16px color-mix(in srgb, var(--pf-vet-primary) 6%, transparent);
  backdrop-filter: blur(6px);
}
.invoicing-totals {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1.5rem;
}
.invoicing-totals__item {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 5.5rem;
}
.invoicing-totals__label {
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.03em;
  text-transform: uppercase;
  color: var(--pf-vet-text-muted, #64748b);
}
.invoicing-totals__value {
  font-size: 1rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--pf-vet-text, inherit);
}
.invoicing-totals__item--incl .invoicing-totals__label {
  color: var(--pf-vet-accent);
}
.invoicing-totals__item--incl .invoicing-totals__value {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--pf-vet-primary);
}
.pro-mt-md { margin-top: 1rem; }
.pro-mt-sm { margin-top: 0.5rem; }

@media (max-width: 640px) {
  .invoicing-grid--postal {
    grid-template-columns: 1fr;
  }
  .invoicing-sticky-bar {
    flex-direction: column;
    align-items: stretch;
  }
  .invoicing-line__subtotal {
    margin-left: 0;
  }
}
</style>
