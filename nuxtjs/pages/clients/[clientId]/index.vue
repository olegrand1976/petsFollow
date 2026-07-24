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
          variant="secondary"
          class="pro-btn--icon"
          test-id="client-app-invite-open"
          :aria-label="$t('clients.appInvite.open')"
          @click="appInviteOpen = true"
        >
          <ProIcon name="qr_code_2" :size="22" />
        </ProButton>
        <ProButton
          v-if="client"
          :disabled="sendingAppLink"
          :loading="sendingAppLink"
          data-testid="send-app-link"
          @click="sendAppLink"
        >
          <ProIcon name="smartphone" />
          {{ $t('clients.detail.sendAppLink') }}
        </ProButton>
      </template>
    </ProPageHeader>
    <p v-if="appLinkFeedback" class="pro-inline-feedback" role="status">{{ appLinkFeedback }}</p>
    <p v-if="petsLoadError" class="pro-field-error" role="alert">{{ petsLoadError }}</p>
    <ProAppInviteModal v-model:open="appInviteOpen" />
    <ProCard v-if="client" :title="$t('clients.detail.identity')">
      <p><strong>{{ client.fullName }}</strong></p>
      <p class="text-muted">{{ client.email }}</p>
      <ProBadge variant="neutral">{{ client.petCount }} {{ petLabel(client.petCount) }}</ProBadge>
      <div class="pro-mt-md client-app-actions">
        <ProButton
          variant="secondary"
          :disabled="sendingAppLink"
          :loading="sendingAppLink"
          @click="sendAppLink"
        >
          <ProIcon name="mail" />
          {{ $t('clients.detail.sendAppLink') }}
        </ProButton>
        <ProButton variant="secondary" test-id="client-detail-app-invite" @click="appInviteOpen = true">
          <ProIcon name="qr_code_2" />
          {{ $t('clients.appInvite.open') }}
        </ProButton>
        <p class="text-muted pro-hint">{{ $t('clients.detail.sendAppLinkHint') }}</p>
      </div>
    </ProCard>
    <ProCard :title="$t('share.clientTitle')" data-testid="client-shares-card">
      <p class="pro-hint pro-mb-md">{{ $t('share.clientHint') }}</p>
      <form class="pro-pet-inline-form" @submit.prevent="addClientShare">
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
              <ProButton variant="ghost" :disabled="shareBusy" @click="revokeClientShare(s.granteeUserId)">
                {{ $t('share.revoke') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
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
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in pets" :key="p.id">
            <td>{{ p.name }}</td>
            <td>{{ p.species }}</td>
            <td>{{ p.breed || $t('common.dash') }}</td>
            <td>
              <NuxtLink :to="`/clients/${clientId}/pets/${p.id}`">{{ $t('common.detail') }}</NuxtLink>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
    <div v-if="pets.length" class="pro-grid-2">
      <NuxtLink
        v-for="p in pets"
        :key="`card-${p.id}`"
        :to="`/clients/${clientId}/pets/${p.id}`"
        class="pro-pet-card-link"
      >
        <ProCard>
          <strong>{{ p.name }}</strong>
          <p class="text-muted">{{ p.species }} · {{ p.breed || $t('common.unknownBreed') }}</p>
        </ProCard>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

type ClientRow = {
  userId: string
  email: string
  fullName: string
  petCount: number
}

const { t } = useI18n()
const { mapError } = useApiError()

const route = useRoute()
const clientId = route.params.clientId as string
const client = ref<ClientRow | null>(null)
const pets = ref<any[]>([])
const petsLoadError = ref('')
const sendingAppLink = ref(false)
const appLinkFeedback = ref('')
const appInviteOpen = ref(false)
const clientShares = ref<any[]>([])
const shareEmail = ref('')
const sharePermission = ref('write_notes')
const shareExpiresDays = ref('')
const shareBusy = ref(false)
const shareError = ref('')

function petLabel(count: number) {
  return count > 1 ? t('common.pets') : t('common.pet')
}

const clientSubtitle = computed(() => client.value?.email || t('clients.detail.subtitle'))

async function loadClientShares() {
  const res: any = await $fetch(`/api/clients/${clientId}/shares`)
  clientShares.value = res.data ?? res ?? []
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
  } catch {
    client.value = null
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
    await loadClientShares()
  } catch {
    clientShares.value = []
  }
})
</script>

<style scoped>
.pro-pet-card-link {
  text-decoration: none;
  color: inherit;
}
.pro-pet-card-link:hover .pro-card {
  border-color: var(--pf-vet-accent);
}
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
.pro-mt-md {
  margin-top: 1rem;
}
.client-app-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-start;
}
.client-app-actions .pro-hint {
  flex-basis: 100%;
}
</style>
