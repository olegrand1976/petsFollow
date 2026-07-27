<template>
  <div data-testid="invoicing-page">
    <ProPageHeader
      :title="$t('invoicing.title')"
      :subtitle="$t('invoicing.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <p v-if="error" class="pro-alert pro-mb-md" data-testid="invoicing-error">{{ error }}</p>

    <ProCard class="pro-mb-lg" data-testid="invoicing-connection">
      <div class="invoicing-conn">
        <div>
          <h3 class="pro-mb-sm">{{ $t('invoicing.connectionTitle') }}</h3>
          <p class="pro-hint">
            {{ $t('invoicing.status') }}:
            <ProBadge :variant="statusVariant">{{ connection?.status || '…' }}</ProBadge>
          </p>
          <p v-if="connection?.billitPartyId" class="pro-hint">
            PartyID: {{ connection.billitPartyId }}
          </p>
          <p class="pro-hint">
            {{ $t('invoicing.usage', {
              used: connection?.usageThisMonth ?? 0,
              max: connection?.docsIncludedMonthly ?? 50,
            }) }}
          </p>
        </div>
        <div class="invoicing-conn__actions">
          <ProButton
            v-if="!isActive"
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
            v-if="isActive || pendingKyc"
            variant="secondary"
            test-id="invoicing-refresh"
            :disabled="busy"
            @click="refreshConnection"
          >
            {{ $t('invoicing.refresh') }}
          </ProButton>
        </div>
      </div>

      <form
        v-if="showCompleteForm"
        class="pro-form pro-mt-md"
        data-testid="invoicing-complete-form"
        @submit.prevent="completeConnect"
      >
        <p class="pro-hint pro-mb-md">{{ $t('invoicing.completeHint') }}</p>
        <ProInput v-model="completeForm.state" :label="$t('invoicing.state')" required />
        <ProInput v-model="completeForm.partyId" :label="$t('invoicing.partyId')" required />
        <ProInput
          v-model="completeForm.apiKey"
          type="password"
          :label="$t('invoicing.apiKey')"
          required
        />
        <ProButton type="submit" test-id="invoicing-complete-submit" :disabled="busy">
          {{ $t('invoicing.complete') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard v-if="isActive" data-testid="invoicing-documents">
      <div class="invoicing-docs-head pro-mb-md">
        <h3>{{ $t('invoicing.documentsTitle') }}</h3>
        <ProButton test-id="invoicing-new-doc" :disabled="busy" @click="createSampleInvoice">
          {{ $t('invoicing.newInvoice') }}
        </ProButton>
      </div>
      <ProEmptyState v-if="!documents.length" :title="$t('invoicing.emptyDocs')" />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('invoicing.colType') }}</th>
            <th>{{ $t('invoicing.colStatus') }}</th>
            <th>{{ $t('invoicing.colTotal') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="doc in documents" :key="doc.id" :data-testid="`invoicing-doc-${doc.id}`">
            <td>{{ doc.type }}</td>
            <td>
              <ProBadge variant="neutral">{{ doc.status }}</ProBadge>
              <span v-if="doc.peppolStatus" class="pro-hint"> · {{ doc.peppolStatus }}</span>
            </td>
            <td>{{ formatMoney(doc.totalInclCents) }}</td>
            <td>
              <ProButton
                v-if="doc.status === 'draft' || doc.status === 'issued'"
                variant="secondary"
                :disabled="busy"
                @click="sendDoc(doc.id)"
              >
                {{ $t('invoicing.send') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard v-else>
      <p class="pro-hint" data-testid="invoicing-wip">{{ $t('invoicing.wip') }}</p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
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
}

const { t } = useI18n()
const connection = ref<Connection | null>(null)
const documents = ref<Document[]>([])
const busy = ref(false)
const error = ref('')
const resellerUrl = ref('')
const showCompleteForm = ref(false)
const completeForm = reactive({
  state: '',
  partyId: '',
  apiKey: '',
})

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

function unwrap<T>(res: any): T {
  return (res?.data ?? res) as T
}

function formatMoney(cents: number) {
  return `${(cents / 100).toFixed(2)} €`
}

async function loadConnection() {
  const res = await $fetch('/api/invoicing/connection')
  connection.value = unwrap<Connection>(res)
}

async function loadDocuments() {
  if (!isActive.value) {
    documents.value = []
    return
  }
  const res = await $fetch('/api/invoicing/documents')
  documents.value = unwrap<Document[]>(res) || []
}

async function startConnect() {
  busy.value = true
  error.value = ''
  try {
    const res = await $fetch('/api/invoicing/connect/start', { method: 'POST' })
    const data = unwrap<{ resellerUrl: string, state: string }>(res)
    resellerUrl.value = data.resellerUrl
    completeForm.state = data.state
    showCompleteForm.value = true
    await loadConnection()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('invoicing.errorGeneric')
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
  try {
    const res = await $fetch('/api/invoicing/connect/complete', {
      method: 'POST',
      body: { ...completeForm },
    })
    connection.value = unwrap<Connection>(res)
    showCompleteForm.value = false
    await loadDocuments()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('invoicing.errorGeneric')
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
    error.value = e?.data?.error?.message || e?.message || t('invoicing.errorGeneric')
  } finally {
    busy.value = false
  }
}

async function createSampleInvoice() {
  busy.value = true
  error.value = ''
  try {
    await $fetch('/api/invoicing/documents', {
      method: 'POST',
      body: {
        type: 'invoice',
        counterparty: {
          name: 'Client démo',
          vatNumber: 'BE0123456789',
          street: 'Rue de la Loi 1',
          city: 'Bruxelles',
          postal: '1000',
          country: 'BE',
        },
        lines: [
          {
            description: 'Consultation',
            quantity: 1,
            unitPriceExclCents: 5000,
            vatPercent: 21,
          },
        ],
      },
    })
    await loadDocuments()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('invoicing.errorGeneric')
  } finally {
    busy.value = false
  }
}

async function sendDoc(id: string) {
  busy.value = true
  error.value = ''
  try {
    await $fetch(`/api/invoicing/documents/${id}/send`, { method: 'POST' })
    await loadDocuments()
    await loadConnection()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('invoicing.errorGeneric')
  } finally {
    busy.value = false
  }
}

onMounted(async () => {
  try {
    await loadConnection()
    await loadDocuments()
  } catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('invoicing.errorGeneric')
  }
})
</script>

<style scoped>
.invoicing-conn {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  justify-content: space-between;
  align-items: flex-start;
}
.invoicing-conn__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.invoicing-docs-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}
.pro-mt-md { margin-top: 1rem; }
</style>
