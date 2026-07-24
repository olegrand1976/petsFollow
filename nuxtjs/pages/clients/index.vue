<template>
  <div data-testid="clients-page">
    <ProPageHeader :title="$t('clients.title')" :subtitle="$t('clients.subtitle')">
      <template #actions>
        <ProButton
          test-id="clients-invitations-open"
          variant="secondary"
          class="pro-btn--icon clients-invites-btn"
          :aria-label="$t('clients.invitations.open')"
          @click="invitationsOpen = true"
        >
          <ProIcon name="mail" :size="22" />
          <ProBadge
            v-if="linkRequests.length"
            variant="danger"
            class="clients-invites-badge"
            data-testid="clients-invitations-badge"
          >
            {{ linkRequests.length > 99 ? '99+' : linkRequests.length }}
          </ProBadge>
        </ProButton>
        <ProButton
          test-id="create-client-open"
          class="pro-btn--icon"
          :aria-label="$t('clients.create.open')"
          @click="openCreateModal"
        >
          <ProIcon name="add" :size="22" />
        </ProButton>
      </template>
    </ProPageHeader>
    <p v-if="loadError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ loadError }}</p>
    <p v-if="appLinkFeedback" class="pro-inline-feedback" role="status">{{ appLinkFeedback }}</p>

    <ProModal v-model:open="invitationsOpen" size="lg" :title="$t('clients.invitations.title')">
      <p v-if="inviteError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ inviteError }}</p>
      <ProEmptyState
        v-if="!linkRequests.length"
        :title="$t('clients.invitations.emptyTitle')"
        :description="$t('clients.invitations.emptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.invitations.columnClient') }}</th>
            <th>{{ $t('clients.invitations.columnEmail') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="req in linkRequests" :key="req.id" :data-testid="`link-request-${req.id}`">
            <td>{{ req.clientName }}</td>
            <td>{{ req.clientEmail }}</td>
            <td>
              <div class="pro-flex-gap">
                <ProButton :disabled="busyInviteId === req.id" @click="acceptLink(req.id)">
                  {{ $t('clients.invitations.accept') }}
                </ProButton>
                <ProButton variant="ghost" :disabled="busyInviteId === req.id" @click="rejectLink(req.id)">
                  {{ $t('clients.invitations.reject') }}
                </ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProModal>

    <ProModal v-model:open="createOpen" :title="$t('clients.create.title')">
      <p class="pro-hint pro-mb-md">{{ $t('clients.create.hint') }}</p>
      <form v-if="!linkCandidate" class="pro-form" data-testid="vet-create-client-form" @submit.prevent="createClient">
        <ProInput v-model="clientForm.fullName" test-id="create-client-name" :label="$t('clients.create.fullName')" required />
        <ProInput v-model="clientForm.email" test-id="create-client-email" type="email" :label="$t('clients.create.email')" required />
        <ProInput v-model="clientForm.password" test-id="create-client-password" type="password" :label="$t('clients.create.password')" required />
        <p v-if="clientMsg" class="pro-hint" data-testid="create-client-msg">{{ clientMsg }}</p>
        <p v-if="clientError" class="pro-error">{{ clientError }}</p>
        <div class="create-client-actions">
          <ProButton variant="secondary" type="button" @click="createOpen = false">
            {{ $t('common.cancel') }}
          </ProButton>
          <ProButton type="submit" test-id="create-client-submit" :disabled="clientSaving">
            {{ $t('clients.create.submit') }}
          </ProButton>
        </div>
      </form>
      <div v-else class="pro-form" data-testid="vet-link-existing-client">
        <p class="pro-hint">
          {{ $t('clients.create.existsHint', { name: linkCandidate.displayName || linkCandidate.email }) }}
        </p>
        <p v-if="linkCandidate.alreadyLinked" class="pro-error">{{ $t('clients.create.alreadyLinked') }}</p>
        <p v-if="clientError" class="pro-error">{{ clientError }}</p>
        <p v-if="clientMsg" class="pro-hint">{{ clientMsg }}</p>
        <div class="create-client-actions">
          <ProButton variant="secondary" type="button" @click="linkCandidate = null">
            {{ $t('common.cancel') }}
          </ProButton>
          <ProButton
            v-if="linkCandidate.linkable && !linkCandidate.alreadyLinked"
            test-id="link-existing-client"
            :disabled="clientSaving"
            @click="linkExistingClient"
          >
            {{ $t('clients.create.linkExisting') }}
          </ProButton>
        </div>
      </div>
    </ProModal>

    <ProCard>
      <ProListToolbar v-model:view-mode="viewMode">
        <template #filters>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="client-search">{{ $t('clients.search') }}</label>
            <input
              id="client-search"
              v-model="query"
              type="search"
              class="pro-input"
              :placeholder="$t('clients.searchPlaceholder')"
            />
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="pet-filter">{{ $t('clients.petsFilter') }}</label>
            <select id="pet-filter" v-model="petFilter" class="pro-select">
              <option value="all">{{ $t('clients.petsAll') }}</option>
              <option value="none">{{ $t('clients.petsNone') }}</option>
              <option value="with">{{ $t('clients.petsWith') }}</option>
            </select>
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="sort-by">{{ $t('clients.sort') }}</label>
            <select id="sort-by" v-model="sortBy" class="pro-select">
              <option value="name">{{ $t('clients.sortName') }}</option>
              <option value="pets">{{ $t('clients.sortPets') }}</option>
            </select>
          </div>
        </template>
      </ProListToolbar>

      <ProTable
        v-if="viewMode === 'table'"
        :empty="!filtered.length"
        :empty-title="$t('clients.emptyTitle')"
        :empty-description="$t('clients.emptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('clients.columnClient') }}</th>
            <th>{{ $t('clients.columnEmail') }}</th>
            <th>{{ $t('clients.columnPets') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in filtered" :key="c.userId">
            <td>
              <ProAvatar :src="c.avatarUrl" :name="c.fullName" class="client-avatar" />
              {{ c.fullName }}
            </td>
            <td>{{ c.email }}</td>
            <td>{{ c.petCount }}</td>
            <td>
              <div class="pro-client-actions">
                <NuxtLink :to="`/clients/${c.userId}`">{{ $t('common.profile') }}</NuxtLink>
                <button
                  type="button"
                  class="pro-link-btn"
                  :disabled="sendingId === c.userId"
                  :data-testid="`send-app-link-${c.userId}`"
                  @click="sendAppLink(c)"
                >
                  {{ sendingId === c.userId ? $t('clients.sendingAppLink') : $t('clients.sendAppLink') }}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>

      <ProKanban v-else>
        <ProKanbanColumn
          v-for="col in kanbanColumns"
          :key="col.key"
          :title="col.title"
          :count="col.items.length"
          :empty="!col.items.length"
          :empty-title="$t('common.empty')"
        >
          <NuxtLink
            v-for="c in col.items"
            :key="c.userId"
            :to="`/clients/${c.userId}`"
            class="pro-kanban-card"
          >
            <ProAvatar :src="c.avatarUrl" :name="c.fullName" class="client-avatar" />
            <strong>{{ c.fullName }}</strong>
            <p class="pro-kanban-card__meta">{{ c.email }}</p>
            <ProBadge variant="neutral">{{ c.petCount }} {{ petLabel(c.petCount) }}</ProBadge>
            <button
              type="button"
              class="pro-link-btn pro-kanban-card__action"
              :disabled="sendingId === c.userId"
              @click.prevent="sendAppLink(c)"
            >
              {{ sendingId === c.userId ? $t('clients.sendingAppLink') : $t('clients.sendAppLink') }}
            </button>
          </NuxtLink>
        </ProKanbanColumn>
      </ProKanban>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

type ClientRow = {
  userId: string
  email: string
  fullName: string
  avatarUrl?: string
  petCount: number
}

const route = useRoute()
const { t } = useI18n()
const { compareStrings } = useFormatters()
const { mapError } = useApiError()
const { refresh: refreshNavBadges } = useNavBadges()

const clients = ref<ClientRow[]>([])
const query = ref('')
const petFilter = ref<'all' | 'none' | 'with'>('all')
const sortBy = ref<'name' | 'pets'>('name')
const sendingId = ref('')
const appLinkFeedback = ref('')
const loadError = ref('')
const { viewMode } = useListView('pf-clients-view', 'table')

const invitationsOpen = ref(false)
const linkRequests = ref<any[]>([])
const busyInviteId = ref('')
const inviteError = ref('')

function asClientList(raw: unknown): ClientRow[] {
  const list = Array.isArray((raw as any)?.data)
    ? (raw as any).data
    : Array.isArray(raw)
      ? raw
      : []
  return list.filter((c: any) => c?.userId) as ClientRow[]
}

async function loadClients() {
  loadError.value = ''
  try {
    const res: any = await $fetch('/api/clients')
    clients.value = asClientList(res)
  } catch (e: any) {
    clients.value = []
    loadError.value = mapError(e) || t('clients.loadError')
  }
}

const createOpen = ref(false)
const clientForm = reactive({ fullName: '', email: '', password: '' })
const clientSaving = ref(false)
const clientMsg = ref('')
const clientError = ref('')
const linkCandidate = ref<{
  userId?: string
  displayName?: string
  email?: string
  linkable?: boolean
  alreadyLinked?: boolean
} | null>(null)

function openCreateModal() {
  clientMsg.value = ''
  clientError.value = ''
  linkCandidate.value = null
  createOpen.value = true
}

async function createClient() {
  clientSaving.value = true
  clientMsg.value = ''
  clientError.value = ''
  linkCandidate.value = null
  try {
    await $fetch('/api/vet/clients', { method: 'POST', body: { ...clientForm } })
    clientMsg.value = t('clients.create.success')
    Object.assign(clientForm, { fullName: '', email: '', password: '' })
    await loadClients()
  } catch (e: any) {
    const details = e?.data?.error?.details || e?.data?.details
    if (e?.statusCode === 409 || e?.status === 409 || details?.exists) {
      linkCandidate.value = {
        userId: details?.userId,
        displayName: details?.displayName,
        email: details?.email || clientForm.email,
        linkable: details?.linkable !== false && details?.role === 'client',
        alreadyLinked: !!details?.alreadyLinked,
      }
      clientError.value = t('clients.create.existsError')
    } else {
      clientError.value = t('clients.create.error')
    }
  } finally {
    clientSaving.value = false
  }
}

async function linkExistingClient() {
  if (!linkCandidate.value?.userId) return
  clientSaving.value = true
  clientError.value = ''
  clientMsg.value = ''
  try {
    await $fetch(`/api/vet/clients/${linkCandidate.value.userId}/link`, { method: 'POST' })
    clientMsg.value = t('clients.create.linkSuccess')
    linkCandidate.value = null
    Object.assign(clientForm, { fullName: '', email: '', password: '' })
    await loadClients()
  } catch {
    clientError.value = t('clients.create.linkError')
  } finally {
    clientSaving.value = false
  }
}

function petLabel(count: number) {
  return count > 1 ? t('common.pets') : t('common.pet')
}

async function sendAppLink(c: ClientRow) {
  if (sendingId.value) return
  sendingId.value = c.userId
  appLinkFeedback.value = ''
  try {
    const res: any = await $fetch(`/api/clients/${c.userId}/send-app-link`, { method: 'POST' })
    const data = res.data ?? res
    appLinkFeedback.value = data.message || t('clients.detail.sendAppLinkSuccess', { email: c.email })
  } catch {
    appLinkFeedback.value = t('clients.detail.sendAppLinkError')
  } finally {
    sendingId.value = ''
  }
}

const filtered = computed(() => {
  let list = Array.isArray(clients.value) ? [...clients.value] : []
  const q = query.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (c) =>
        c.fullName?.toLowerCase().includes(q) ||
        c.email?.toLowerCase().includes(q),
    )
  }
  if (petFilter.value === 'none') list = list.filter((c) => c.petCount === 0)
  if (petFilter.value === 'with') list = list.filter((c) => c.petCount > 0)
  if (sortBy.value === 'name') {
    list.sort((a, b) => compareStrings(a.fullName || '', b.fullName || ''))
  } else {
    list.sort((a, b) => (b.petCount || 0) - (a.petCount || 0))
  }
  return list
})

const kanbanColumns = computed(() => [
  {
    key: 'none',
    title: t('clients.kanbanNone'),
    items: filtered.value.filter((c) => c.petCount === 0),
  },
  {
    key: 'one',
    title: t('clients.kanbanOne'),
    items: filtered.value.filter((c) => c.petCount === 1),
  },
  {
    key: 'multi',
    title: t('clients.kanbanMulti'),
    items: filtered.value.filter((c) => c.petCount > 1),
  },
])

async function loadLinkRequests() {
  inviteError.value = ''
  try {
    const res: any = await $fetch('/api/vet/link-requests')
    linkRequests.value = res.data ?? res ?? []
  } catch (e: any) {
    linkRequests.value = []
    inviteError.value = mapError(e)
  }
}

async function acceptLink(id: string) {
  busyInviteId.value = id
  inviteError.value = ''
  try {
    await $fetch(`/api/vet/link-requests/${id}/accept`, { method: 'POST' })
    await Promise.all([loadLinkRequests(), loadClients(), refreshNavBadges()])
  } catch (e: any) {
    inviteError.value = mapError(e)
  } finally {
    busyInviteId.value = ''
  }
}

async function rejectLink(id: string) {
  busyInviteId.value = id
  inviteError.value = ''
  try {
    await $fetch(`/api/vet/link-requests/${id}/reject`, { method: 'POST' })
    await Promise.all([loadLinkRequests(), refreshNavBadges()])
  } catch (e: any) {
    inviteError.value = mapError(e)
  } finally {
    busyInviteId.value = ''
  }
}

onMounted(async () => {
  await Promise.all([loadClients(), loadLinkRequests()])
  if (route.query.invitations === '1' || linkRequests.value.length > 0) {
    invitationsOpen.value = true
  }
})
</script>

<style scoped>
.clients-invites-btn {
  position: relative;
}
.clients-invites-badge {
  position: absolute;
  top: -0.25rem;
  right: -0.25rem;
  min-width: 1.1rem;
  justify-content: center;
  padding: 0.1rem 0.3rem;
  font-size: 0.65rem;
}
.client-avatar {
  margin-right: 0.5rem;
  vertical-align: middle;
}
.create-client-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
:deep(.pro-btn--icon) {
  min-width: 44px;
  padding-inline: 0.65rem;
}
.pro-client-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}
.pro-link-btn {
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--pf-vet-accent);
  font: inherit;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
}
.pro-link-btn:disabled {
  opacity: 0.6;
  cursor: wait;
}
.pro-kanban-card__action {
  display: block;
  margin-top: 0.5rem;
  text-align: left;
}
.pro-inline-feedback {
  margin: 0 0 1rem;
  padding: 0.75rem 1rem;
  border-radius: var(--pf-vet-radius);
  background: color-mix(in srgb, var(--pf-vet-accent) 10%, var(--pf-vet-surface));
  border: 1px solid color-mix(in srgb, var(--pf-vet-accent) 30%, transparent);
}
.pro-inline-feedback--error {
  background: color-mix(in srgb, var(--pf-vet-danger, #b42318) 10%, var(--pf-vet-surface));
  border-color: color-mix(in srgb, var(--pf-vet-danger, #b42318) 35%, transparent);
  color: var(--pf-vet-danger, #b42318);
}
</style>
