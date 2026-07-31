<template>
  <div>
    <nav class="pro-breadcrumb" :aria-label="$t('common.breadcrumb')">
      <NuxtLink to="/clients">{{ $t('nav.clients') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <span>{{ client?.fullName || $t('clients.detail.title') }}</span>
    </nav>
    <ProPageHeader
      :title="client?.fullName || $t('clients.detail.title')"
      :subtitle="clientSubtitle"
    >
      <template #actions>
        <ProButton
          v-if="canWriteClients"
          variant="secondary"
          class="pro-btn--icon"
          test-id="client-app-invite-open"
          :aria-label="$t('clients.appInvite.open')"
          @click="appInviteOpen = true"
        >
          <ProIcon name="qr_code_2" :size="22" />
        </ProButton>
        <ProButton
          v-if="client && canWriteClients"
          :disabled="sendingAppLink"
          :loading="sendingAppLink"
          data-testid="send-app-link"
          @click="sendAppLink"
        >
          <ProIcon name="qr_code_2" />
          {{ $t('clients.detail.sendAppLink') }}
        </ProButton>
        <ProButton
          v-if="client && canWriteClinical"
          variant="secondary"
          test-id="client-new-consultation"
          @click="openConsultation"
        >
          <ProIcon name="medical_services" />
          {{ $t('clients.consultation.open') }}
        </ProButton>
      </template>
    </ProPageHeader>
    <p v-if="appLinkFeedback" class="pro-inline-feedback" role="status">{{ appLinkFeedback }}</p>
    <p v-if="petsLoadError" class="pro-field-error" role="alert">{{ petsLoadError }}</p>
    <ProAppInviteModal v-model:open="appInviteOpen" />

    <div v-if="overview" class="pro-grid-kpi" data-testid="client-kpi-strip">
      <ProKpi
        icon="pets"
        :value="overview.petCount"
        :label="$t('clients.detail.kpi.pets')"
        :to="`/clients/${clientId}?tab=pets`"
      />
      <ProKpi
        icon="favorite"
        :value="overview.unreadHeartrate"
        :label="$t('clients.detail.kpi.unreadHeartrate')"
        :variant="overview.unreadHeartrate > 0 ? 'alert' : 'default'"
        :to="overview.unreadHeartrate > 0 ? `/clients/${clientId}?tab=pets` : undefined"
      />
      <ProKpi
        icon="warning"
        :value="overview.alertSessions7d"
        :label="$t('clients.detail.kpi.alertSessions7d')"
        :variant="overview.alertSessions7d > 0 ? 'alert' : 'default'"
        :to="overview.alertSessions7d > 0 ? `/clients/${clientId}?tab=pets` : undefined"
      />
      <ProKpi
        icon="medical_services"
        :value="overview.overdueCareCount"
        :label="$t('clients.detail.kpi.overdueCare')"
        :variant="overview.overdueCareCount > 0 ? 'alert' : 'default'"
        :to="overview.overdueCareCount > 0 ? `/clients/${clientId}?tab=pets` : undefined"
      />
      <ProKpi
        icon="event"
        :value="overview.pendingVisits"
        :label="$t('clients.detail.kpi.pendingVisits')"
        :variant="overview.pendingVisits > 0 ? 'alert' : 'default'"
        :to="overview.pendingVisits > 0 ? `/clients/${clientId}?tab=pets` : undefined"
      />
      <ProKpi
        icon="group"
        :value="overview.shareCount"
        :label="$t('clients.detail.kpi.shares')"
        :to="canReadShares ? `/clients/${clientId}?tab=sharing` : undefined"
      />
    </div>
    <p v-if="overview?.upcomingVisitAt" class="pro-hint pro-mb-md" data-testid="client-upcoming-visit">
      {{ $t('clients.detail.kpi.upcomingVisit', { date: formatShareDate(overview.upcomingVisitAt) }) }}
    </p>

    <ProSectionTabs
      v-model="activeTab"
      :tabs="clientTabs"
      :aria-label="$t('clients.detail.tabsLabel')"
    />

    <div v-show="activeTab === 'pets'" role="tabpanel" aria-labelledby="tab-pets" data-testid="client-tab-pets">
      <ProCard :title="$t('clients.detail.petsTitle')">
        <ProTable
          :empty="!pets.length"
          :empty-title="$t('clients.detail.petsEmptyTitle')"
          :empty-description="$t('clients.detail.petsEmptyDescription')"
        >
          <thead>
            <tr>
              <th>{{ $t('clients.detail.columnName') }}</th>
              <th>{{ $t('clients.detail.columnSpecies') }}</th>
              <th>{{ $t('clients.detail.columnBreed') }}</th>
              <th>{{ $t('clients.detail.columnWeight') }}</th>
              <th>{{ $t('clients.detail.columnStatus') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in pets" :key="p.id">
              <td>{{ p.name }}</td>
              <td>{{ speciesLabel(p.species) }}</td>
              <td>{{ p.breed || $t('common.dash') }}</td>
              <td>{{ p.weightKg != null ? `${p.weightKg} kg` : $t('common.dash') }}</td>
              <td>
                <ProBadge :variant="petRowStatusVariant(p)">{{ petRowStatusLabel(p) }}</ProBadge>
              </td>
              <td>
                <NuxtLink :to="`/clients/${clientId}/pets/${p.id}`">{{ $t('common.detail') }}</NuxtLink>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>
    </div>

    <div v-show="activeTab === 'sharing'" role="tabpanel" aria-labelledby="tab-sharing" data-testid="client-tab-sharing">
      <ProCard :title="$t('share.clientTitle')" data-testid="client-shares-card">
        <p class="pro-hint pro-mb-md">{{ $t('share.clientHint') }}</p>
        <form v-if="canManageShares" class="pro-pet-inline-form" @submit.prevent="addClientShare">
          <ProInput v-model="shareEmail" type="email" :label="$t('share.email')" required />
          <select v-model="sharePermission" class="pro-input" data-testid="client-share-permission">
            <option value="read">{{ $t('share.permRead') }}</option>
            <option value="write_notes">{{ $t('share.permWriteNotes') }}</option>
            <option value="full">{{ $t('share.permFull') }}</option>
          </select>
          <select v-model="shareExpiresDays" class="pro-input" data-testid="client-share-expires">
            <option value="">{{ $t('share.expiresNever') }}</option>
            <option value="7">{{ $t('share.expiresDays', { n: 7 }) }}</option>
            <option value="30">{{ $t('share.expiresDays', { n: 30 }) }}</option>
            <option value="90">{{ $t('share.expiresDays', { n: 90 }) }}</option>
          </select>
          <ProButton type="submit" :disabled="shareBusy">{{ $t('share.add') }}</ProButton>
        </form>
        <p v-if="shareError" class="pro-error">{{ shareError }}</p>
        <ProTable v-if="clientShares.length">
          <thead>
            <tr>
              <th>{{ $t('share.columnName') }}</th>
              <th>{{ $t('share.columnEmail') }}</th>
              <th>{{ $t('share.columnPermission') }}</th>
              <th>{{ $t('share.columnExpires') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in clientShares" :key="s.id">
              <td>{{ s.granteeName }}</td>
              <td>{{ s.granteeEmail }}</td>
              <td>{{ s.permission }}</td>
              <td>{{ s.expiresAt ? formatShareDate(s.expiresAt) : $t('share.expiresNever') }}</td>
              <td>
                <ProButton
                  v-if="canManageShares"
                  variant="ghost"
                  :disabled="shareBusy"
                  @click="revokeClientShare(s.granteeUserId)"
                >
                  {{ $t('share.revoke') }}
                </ProButton>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>
    </div>

    <div v-show="activeTab === 'identity'" role="tabpanel" aria-labelledby="tab-identity" data-testid="client-tab-identity">
      <ProCard v-if="client" :title="$t('clients.detail.identity')">
        <div class="client-identity">
          <ProAvatar
            v-if="client.avatarUrl || client.fullName"
            :src="client.avatarUrl"
            :name="client.fullName"
            size="lg"
          />
          <div>
            <p><strong>{{ client.fullName }}</strong></p>
            <p class="text-muted">{{ client.email }}</p>
            <ProBadge variant="neutral">{{ client.petCount }} {{ petLabel(client.petCount) }}</ProBadge>
            <div v-if="canWriteClients" class="client-phone-edit">
              <ProInput
                v-model="phoneDraft"
                test-id="client-phone-input"
                type="tel"
                :label="$t('clients.detail.contactPhone')"
                :maxlength="40"
              />
              <div class="pro-flex-gap">
                <ProButton
                  test-id="client-phone-save"
                  :disabled="phoneSaving || phoneDraft.trim() === (client.contactPhone || '').trim()"
                  @click="saveContactPhone"
                >
                  {{ $t('clients.detail.savePhone') }}
                </ProButton>
              </div>
              <p v-if="phoneMsg" class="pro-hint" role="status" data-testid="client-phone-msg">{{ phoneMsg }}</p>
              <p v-if="phoneError" class="pro-error" role="alert">{{ phoneError }}</p>
            </div>
            <p v-else class="text-muted">{{ client.contactPhone || '—' }}</p>
            <p class="text-muted pro-hint">{{ $t('clients.detail.sendAppLinkHint') }}</p>
          </div>
        </div>
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'clients.read' })

type ClientRow = {
  userId: string
  email: string
  fullName: string
  petCount: number
  avatarUrl?: string
  contactPhone?: string
}

type ClientOverview = {
  petCount: number
  unreadHeartrate: number
  alertSessions7d: number
  overdueCareCount: number
  pendingVisits: number
  upcomingVisitAt?: string
  shareCount: number
}

const { t, te } = useI18n()

function speciesLabel(species: string | null | undefined) {
  if (!species) return t('common.dash')
  const key = `common.species.${species}`
  return te(key) ? t(key) : species
}
const { mapError } = useApiError()
const { canPractice } = usePracticePerms()
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))
const canManageShares = computed(() => canPractice('shares.manage'))
const canReadShares = computed(() => canPractice('shares.read'))
const canWriteClients = computed(() => canPractice('clients.write'))
const route = useRoute()

const clientId = route.params.clientId as string
const client = ref<ClientRow | null>(null)
const overview = ref<ClientOverview | null>(null)
const pets = ref<any[]>([])
const petsLoadError = ref('')
const sendingAppLink = ref(false)
const appLinkFeedback = ref('')
const appInviteOpen = ref(false)
const activeConsult = useActiveConsultation()
const clientShares = ref<any[]>([])
const phoneDraft = ref('')
const phoneSaving = ref(false)
const phoneMsg = ref('')
const phoneError = ref('')

async function saveContactPhone() {
  phoneSaving.value = true
  phoneMsg.value = ''
  phoneError.value = ''
  try {
    const res: any = await $fetch(`/api/clients/${clientId}`, {
      method: 'PATCH',
      body: { contactPhone: phoneDraft.value.trim() },
    })
    const data = res.data ?? res
    if (client.value) {
      client.value = { ...client.value, contactPhone: data.contactPhone || '' }
    }
    phoneDraft.value = data.contactPhone || ''
    phoneMsg.value = t('clients.detail.phoneSaved')
  } catch (e: any) {
    phoneError.value = mapError(e) || t('clients.detail.phoneSaveError')
  } finally {
    phoneSaving.value = false
  }
}

function openConsultation() {
  activeConsult.openForClient(clientId)
}
const shareEmail = ref('')
const sharePermission = ref('write_notes')
const shareExpiresDays = ref('')
const shareBusy = ref(false)
const shareError = ref('')
const activeTab = ref('pets')

function petLabel(count: number) {
  return count > 1 ? t('common.pets') : t('common.pet')
}

function petRowStatus(p: any) {
  return p?.entitlement?.status || p?.paymentStatus || ''
}

function petRowStatusLabel(p: any) {
  const status = petRowStatus(p)
  if (!status) return t('common.dash')
  const key = `clients.pet.paymentStatus.${status}`
  const translated = t(key)
  return translated !== key ? translated : status
}

function petRowStatusVariant(p: any): 'success' | 'warning' | 'danger' | 'neutral' {
  const status = petRowStatus(p)
  if (status === 'active') return 'success'
  if (status === 'pending' || status === 'pending_payment') return 'warning'
  if (status === 'expired' || status === 'cancelled') return 'danger'
  return 'neutral'
}

const clientSubtitle = computed(() => client.value?.email || t('clients.detail.subtitle'))

const clientTabs = computed(() => {
  const tabs = [
    { id: 'pets', label: t('clients.detail.tabs.pets'), count: overview.value?.petCount || pets.value.length || undefined },
    { id: 'identity', label: t('clients.detail.tabs.identity') },
  ]
  if (canReadShares.value) {
    tabs.splice(1, 0, {
      id: 'sharing',
      label: t('clients.detail.tabs.sharing'),
      count: overview.value?.shareCount || undefined,
    })
  }
  return tabs
})

async function loadClientShares() {
  const res: any = await $fetch(`/api/clients/${clientId}/shares`)
  clientShares.value = res.data ?? res ?? []
}

async function loadOverview() {
  const res: any = await $fetch(`/api/clients/${clientId}/overview`)
  overview.value = res.data ?? res
}

function formatShareDate(iso: string) {
  try {
    return new Date(iso).toLocaleDateString()
  } catch {
    return iso
  }
}

async function addClientShare() {
  if (!shareEmail.value.trim()) return
  shareBusy.value = true
  shareError.value = ''
  try {
    const body: Record<string, string> = {
      email: shareEmail.value.trim(),
      permission: sharePermission.value || 'write_notes',
    }
    if (shareExpiresDays.value) {
      const d = new Date()
      d.setDate(d.getDate() + Number(shareExpiresDays.value))
      body.expiresAt = d.toISOString()
    }
    await $fetch(`/api/clients/${clientId}/shares`, {
      method: 'POST',
      body,
    })
    shareEmail.value = ''
    sharePermission.value = 'write_notes'
    shareExpiresDays.value = ''
    await loadClientShares()
    await loadOverview().catch(() => {})
  } catch (e: any) {
    shareError.value = mapError(e)
  } finally {
    shareBusy.value = false
  }
}

async function revokeClientShare(granteeUserId: string) {
  shareBusy.value = true
  try {
    await $fetch(`/api/clients/${clientId}/shares/${granteeUserId}`, { method: 'DELETE' })
    await loadClientShares()
    await loadOverview().catch(() => {})
  } catch (e: any) {
    shareError.value = mapError(e)
  } finally {
    shareBusy.value = false
  }
}

async function sendAppLink() {
  if (!client.value || sendingAppLink.value) return
  sendingAppLink.value = true
  appLinkFeedback.value = ''
  try {
    const res: any = await $fetch(`/api/clients/${clientId}/send-app-link`, { method: 'POST' })
    const data = res.data ?? res
    appLinkFeedback.value = data.message || t('clients.detail.sendAppLinkSuccess', { email: client.value.email })
  } catch {
    appLinkFeedback.value = t('clients.detail.sendAppLinkError')
  } finally {
    sendingAppLink.value = false
  }
}

onMounted(async () => {
  try {
    const clientRes: any = await $fetch(`/api/clients/${clientId}`)
    client.value = clientRes.data ?? clientRes
    phoneDraft.value = client.value?.contactPhone || ''
  } catch {
    client.value = null
  }
  try {
    await loadOverview()
  } catch {
    overview.value = null
  }
  try {
    const petsRes: any = await $fetch(`/api/clients/${clientId}/pets`)
    pets.value = petsRes.data ?? petsRes ?? []
    petsLoadError.value = ''
  } catch (e: any) {
    pets.value = []
    petsLoadError.value = mapError(e) || t('clients.loadError')
  }
  try {
    if (canReadShares.value) {
      await loadClientShares()
    }
  } catch {
    clientShares.value = []
  }
})
</script>

<style scoped>
.pro-hint {
  margin-top: 0.5rem;
  font-size: 0.875rem;
}
.pro-inline-feedback {
  margin: 0 0 1rem;
  padding: 0.75rem 1rem;
  border-radius: var(--pf-vet-radius);
  background: color-mix(in srgb, var(--pf-vet-accent) 10%, var(--pf-vet-surface));
  border: 1px solid color-mix(in srgb, var(--pf-vet-accent) 30%, transparent);
}
.client-identity {
  display: flex;
  gap: 1.25rem;
  align-items: flex-start;
  flex-wrap: wrap;
}
.client-phone-edit {
  margin-top: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 22rem;
}
.pro-mb-md {
  margin-bottom: 1rem;
}
</style>
