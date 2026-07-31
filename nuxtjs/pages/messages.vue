<template>
  <div data-testid="messages-page">
    <ProPageHeader :title="$t('messages.title')" :subtitle="$t('messages.subtitle')" />
    <p v-if="actionError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ actionError }}</p>
    <div class="pro-chat">
      <aside class="pro-chat__threads">
        <h3 class="pro-card__title">{{ $t('messages.threads') }}</h3>
        <div class="pro-chat__search" data-testid="messages-client-search">
          <label class="pro-sr-only" for="messages-client-search-input">
            {{ $t('messages.searchPlaceholder') }}
          </label>
          <ProCombobox
            v-model="clientSearch"
            input-id="messages-client-search-input"
            :placeholder="$t('messages.searchPlaceholder')"
            :min-chars="1"
            :disabled="searchLocked"
            :search-fn="searchClients"
            @select="onClientSelect"
          />
          <p v-if="startingClientId && !petPickerOpen" class="pro-hint" role="status">{{ $t('messages.loadingPets') }}</p>
        </div>
        <ProEmptyState
          v-if="!visibleThreads.length"
          :title="$t('messages.emptyTitle')"
          :description="$t('messages.emptyDescription')"
        />
        <button
          v-for="t in visibleThreads"
          :key="t.id"
          type="button"
          class="pro-chat__thread-btn"
          :class="{ 'pro-chat__thread-btn--active': active?.id === t.id }"
          :data-testid="`thread-${t.id}`"
          @click="select(t)"
        >
          <span class="pro-chat__thread-top">
            <strong>{{ clientLabel(t) }}</strong>
            <span v-if="t.petName" class="pro-chat__thread-pet">{{ t.petName }}</span>
          </span>
          <span v-if="t.lastMessagePreview" class="pro-chat__thread-preview">{{ t.lastMessagePreview }}</span>
          <span class="pro-chat__thread-meta">
            <span v-if="t.lastMessageAt" class="pro-chat__thread-date">
              {{ $t('messages.lastExchangeAt', { date: formatThreadDate(t.lastMessageAt) }) }}
            </span>
            <span v-else class="pro-chat__thread-date">{{ $t('messages.newConversation') }}</span>
            <ProBadge v-if="t.unreadCount > 0" variant="warning">{{ unreadLabel(t.unreadCount) }}</ProBadge>
          </span>
        </button>
      </aside>
      <section class="pro-chat__messages">
        <template v-if="active">
          <header class="pro-chat__header">
            <strong>{{ threadLabel(active) }}</strong>
          </header>
          <div class="pro-chat__list">
            <div
              v-for="m in messages"
              :key="m.id"
              class="pro-chat__bubble"
              :class="isVetMessage(m) ? 'pro-chat__bubble--vet' : 'pro-chat__bubble--client'"
            >
              <small>{{ senderLabel(m) }}</small>
              <template v-if="m.mediaUrl">
                <a
                  v-if="m.mediaType === 'video'"
                  class="pro-chat__media-link"
                  :href="m.mediaUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <ProIcon name="play_circle" />
                  {{ $t('messages.video') }}
                </a>
                <a
                  v-else
                  class="pro-chat__media-link"
                  :href="m.mediaUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <img :src="m.mediaUrl" :alt="$t('messages.image')" class="pro-chat__media-img" />
                </a>
              </template>
              <p v-if="m.body">{{ m.body }}</p>
            </div>
          </div>
          <form class="pro-chat__composer" @submit.prevent="send">
            <div class="pro-chat__composer-row">
              <textarea
                v-model="draft"
                class="pro-textarea"
                rows="3"
                :placeholder="$t('messages.placeholder')"
              />
              <div class="pro-chat__composer-actions">
                <input
                  ref="fileInput"
                  type="file"
                  class="pro-sr-only"
                  accept="image/jpeg,image/png,image/webp,video/mp4,video/quicktime,video/webm"
                  @change="onFileSelected"
                >
                <ProButton
                  type="button"
                  variant="secondary"
                  data-testid="messages-attach"
                  :disabled="sending || !active"
                  @click="fileInput?.click()"
                >
                  <ProIcon name="attach_file" />
                  {{ $t('messages.attach') }}
                </ProButton>
                <ProButton type="submit" :disabled="sending || !draft.trim()">
                  {{ sending ? $t('messages.sending') : $t('common.send') }}
                </ProButton>
              </div>
            </div>
          </form>
        </template>
        <ProEmptyState
          v-else
          :title="$t('messages.selectTitle')"
          :description="$t('messages.selectDescription')"
        />
      </section>
    </div>

    <ProModal
      :open="petPickerOpen"
      :title="$t('messages.pickPetTitle')"
      test-id="messages-pet-picker"
      :prevent-close="startingClientId !== ''"
      @update:open="(v) => { if (!v) closePetPicker() }"
    >
      <p class="pro-hint">{{ $t('messages.pickPetHint', { name: pendingClientLabel }) }}</p>
      <div class="pro-field">
        <label class="pro-label" for="messages-pet-select">{{ $t('messages.pickPet') }}</label>
        <select
          id="messages-pet-select"
          v-model="selectedPetId"
          class="pro-select"
          data-testid="messages-pet-select"
        >
          <option value="" disabled>{{ $t('messages.pickPetPlaceholder') }}</option>
          <option v-for="p in petPickerPets" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </div>
      <template #footer>
        <ProButton variant="secondary" type="button" :disabled="startingClientId !== ''" @click="closePetPicker">
          {{ $t('common.cancel') }}
        </ProButton>
        <ProButton
          type="button"
          data-testid="messages-pet-picker-confirm"
          :disabled="!selectedPetId || startingClientId !== ''"
          @click="confirmPetPick"
        >
          {{ $t('messages.startConversation') }}
        </ProButton>
      </template>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'messaging' })

type ClientRow = {
  userId: string
  email?: string
  fullName?: string
  avatarUrl?: string
  petCount?: number
}

type PetRow = {
  id: string
  name: string
  species?: string
}

const route = useRoute()
const { t } = useI18n()
const { mapError } = useApiError()
const { user, fetchUser } = useProUser()
const { refresh: refreshNotif } = useProNotifications()
const { formatDate } = useFormatters()

const threads = ref<any[]>([])
const clients = ref<ClientRow[]>([])
const messages = ref<any[]>([])
const active = ref<any>(null)
const clientSearch = ref<ProComboboxItem | null>(null)
const draft = ref('')
const sending = ref(false)
const actionError = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const refreshing = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const petPickerOpen = ref(false)
const petPickerPets = ref<PetRow[]>([])
const petPickerClient = ref<ClientRow | null>(null)
const selectedPetId = ref('')
const startingClientId = ref('')

const POLL_MS = 4_000
const MAX_MEDIA_BYTES = 25 * 1024 * 1024

const vetUserId = computed(() => user.value?.userId ?? user.value?.id ?? '')
const pendingClientLabel = computed(() => petPickerClient.value ? clientDisplayName(petPickerClient.value) : '')
const searchLocked = computed(() => Boolean(startingClientId.value) || petPickerOpen.value)

function clientLabel(thread: any) {
  return thread.clientName || t('common.clientFallback', { id: thread.clientUserId?.slice(0, 8) ?? '' })
}

function threadLabel(thread: any) {
  const name = clientLabel(thread)
  return thread.petName ? `${name} · ${thread.petName}` : name
}

function formatThreadDate(value?: string | null) {
  if (!value) return ''
  return formatDate(value)
}

function isInProgressThread(thread: any) {
  return Boolean(thread?.lastMessageAt)
    || Boolean(String(thread?.lastMessagePreview ?? '').trim())
}

const visibleThreads = computed(() => {
  const activeId = active.value?.id
  const filtered = threads.value.filter(
    (t) => isInProgressThread(t) || (activeId != null && t.id === activeId),
  )
  return [...filtered].sort((a, b) => {
    const aUnread = (a.unreadCount ?? 0) > 0 ? 1 : 0
    const bUnread = (b.unreadCount ?? 0) > 0 ? 1 : 0
    if (aUnread !== bUnread) return bUnread - aUnread
    const aTime = a.lastMessageAt ? new Date(a.lastMessageAt).getTime() : 0
    const bTime = b.lastMessageAt ? new Date(b.lastMessageAt).getTime() : 0
    if (aTime !== bTime) return bTime - aTime
    return String(a.id ?? '').localeCompare(String(b.id ?? ''))
  })
})

function unreadLabel(count: number) {
  return count > 1 ? t('messages.unreadPlural', { count }) : t('messages.unread', { count })
}

function clientDisplayName(c: ClientRow) {
  return c.fullName || c.email || ''
}

function asClientList(raw: unknown): ClientRow[] {
  const list = Array.isArray((raw as any)?.data) ? (raw as any).data : Array.isArray(raw) ? raw : []
  return list.filter((c: any) => c?.userId) as ClientRow[]
}

async function loadClients() {
  try {
    const res: any = await $fetch('/api/clients')
    clients.value = asClientList(res)
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
      return name.includes(needle) || email.includes(needle)
    })
    .slice(0, 20)
    .map((c) => ({
      id: c.userId,
      label: clientDisplayName(c),
      hint: c.email,
      raw: c,
    }))
}

function asPetList(raw: unknown): PetRow[] {
  const list = Array.isArray((raw as any)?.data) ? (raw as any).data : Array.isArray(raw) ? raw : []
  return list.filter((p: any) => p?.id) as PetRow[]
}

async function onClientSelect(item: ProComboboxItem) {
  if (searchLocked.value) return
  const client = (item.raw as ClientRow | undefined) ?? clients.value.find((c) => c.userId === item.id)
  clientSearch.value = null
  if (!client) return
  await startConversation(client)
}

async function startConversation(client: ClientRow) {
  actionError.value = ''
  startingClientId.value = client.userId
  try {
    const res: any = await $fetch(`/api/clients/${client.userId}/pets`)
    const pets = asPetList(res?.data ?? res)
    if (pets.length === 0) {
      actionError.value = t('messages.noPets')
      return
    }
    if (pets.length === 1) {
      await ensureAndOpenThread(client.userId, pets[0]!.id)
      return
    }
    petPickerClient.value = client
    petPickerPets.value = pets
    selectedPetId.value = ''
    petPickerOpen.value = true
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    startingClientId.value = ''
  }
}

function closePetPicker() {
  if (startingClientId.value) return
  petPickerOpen.value = false
  petPickerClient.value = null
  petPickerPets.value = []
  selectedPetId.value = ''
}

async function confirmPetPick() {
  const client = petPickerClient.value
  if (!client || !selectedPetId.value) return
  const petId = selectedPetId.value
  startingClientId.value = client.userId
  actionError.value = ''
  try {
    await ensureAndOpenThread(client.userId, petId)
    petPickerOpen.value = false
    petPickerClient.value = null
    petPickerPets.value = []
    selectedPetId.value = ''
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    startingClientId.value = ''
  }
}

async function ensureAndOpenThread(clientUserId: string, petId: string) {
  const res: any = await $fetch('/api/messaging/threads', {
    method: 'POST',
    body: { clientUserId, petId },
  })
  const created = res?.data ?? res
  await loadThreads()
  const thread = threads.value.find((item) => item?.id === created?.id) ?? created
  if (thread) await select(thread)
}

function isVetMessage(msg: any) {
  return msg.senderUserId === vetUserId.value
}

function senderLabel(msg: any) {
  if (isVetMessage(msg)) {
    return user.value?.fullName || t('messages.you')
  }
  return active.value?.clientName || t('common.clientFallback', { id: active.value?.clientUserId?.slice(0, 8) ?? '' })
}

function asThreadList(raw: unknown): any[] {
  const list = Array.isArray(raw) ? raw : []
  return list.filter((item) => item != null && item.id != null)
}

function asMessageList(raw: unknown): any[] {
  const list = Array.isArray(raw) ? raw : []
  return list.filter((item) => item != null && item.id != null)
}

async function loadThreads() {
  const res: any = await $fetch('/api/messaging/threads')
  threads.value = asThreadList(res?.data ?? res)
  const currentId = active.value?.id
  if (currentId) {
    active.value = threads.value.find((item) => item?.id === currentId) ?? active.value
  }
}

async function select(thread: any) {
  if (!thread?.id) return
  active.value = thread
  const res: any = await $fetch(`/api/messaging/threads/${thread.id}/messages`)
  messages.value = asMessageList(res?.data ?? res)
  if ((thread.unreadCount ?? 0) > 0) {
    await $fetch(`/api/messaging/threads/${thread.id}/read`, { method: 'POST' })
    await loadThreads()
    await refreshNotif()
  }
}

async function openThreadFromQuery() {
  const threadId = String(route.query.thread || '')
  if (!threadId) return
  const thread = threads.value.find((item) => item?.id === threadId)
  if (thread) await select(thread)
}

async function silentRefresh() {
  if (sending.value || refreshing.value) return
  refreshing.value = true
  try {
    await loadThreads()
    const threadId = active.value?.id
    if (!threadId) {
      await refreshNotif()
      return
    }
    const res: any = await $fetch(`/api/messaging/threads/${threadId}/messages`)
    const next = asMessageList(res?.data ?? res)
    const prevLast = messages.value[messages.value.length - 1]?.id
    const nextLast = next[next.length - 1]?.id
    if (next.length !== messages.value.length || nextLast !== prevLast) {
      messages.value = next
    }
    // Re-read active after awaits — user may have cleared selection.
    if (active.value?.id !== threadId) return
    const thread = threads.value.find((item) => item?.id === threadId)
    if (thread && (thread.unreadCount ?? 0) > 0) {
      await $fetch(`/api/messaging/threads/${threadId}/read`, { method: 'POST' })
      await loadThreads()
    }
    await refreshNotif()
  } catch {
    // ignore background poll errors
  } finally {
    refreshing.value = false
  }
}

function onVisibility() {
  if (document.visibilityState === 'visible') void silentRefresh()
}

function startPoll() {
  stopPoll()
  pollTimer = setInterval(() => {
    if (document.visibilityState === 'hidden') return
    void silentRefresh()
  }, POLL_MS)
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  try {
    await fetchUser()
  } catch { /* ignore */ }
  await Promise.all([loadThreads(), loadClients()])
  await openThreadFromQuery()
  if (import.meta.client) {
    document.addEventListener('visibilitychange', onVisibility)
    startPoll()
  }
})

onUnmounted(() => {
  stopPoll()
  if (import.meta.client) {
    document.removeEventListener('visibilitychange', onVisibility)
  }
})

watch(
  () => route.query.thread,
  async () => {
    if (threads.value.length) await openThreadFromQuery()
  },
)

async function reloadMessages() {
  const threadId = active.value?.id
  if (!threadId) return
  const res: any = await $fetch(`/api/messaging/threads/${threadId}/messages`)
  messages.value = asMessageList(res?.data ?? res)
  await loadThreads()
  await refreshNotif()
}

async function send() {
  const threadId = active.value?.id
  if (!threadId || !draft.value.trim() || sending.value) return
  sending.value = true
  actionError.value = ''
  try {
    await $fetch(`/api/messaging/threads/${threadId}/messages`, {
      method: 'POST',
      body: { body: draft.value },
    })
    draft.value = ''
    await reloadMessages()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    sending.value = false
  }
}

async function onFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  const threadId = active.value?.id
  if (!file || !threadId || sending.value) return
  if (file.size > MAX_MEDIA_BYTES) {
    actionError.value = t('errors.image_too_large')
    return
  }
  sending.value = true
  actionError.value = ''
  try {
    const form = new FormData()
    form.append('file', file)
    if (draft.value.trim()) form.append('body', draft.value.trim())
    await $fetch(`/api/messaging/threads/${threadId}/messages/media`, {
      method: 'POST',
      body: form,
    })
    draft.value = ''
    await reloadMessages()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    sending.value = false
  }
}
</script>

<style scoped>
.pro-chat__search {
  margin-bottom: 0.75rem;
}

.pro-chat__composer {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.pro-chat__composer-row {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.pro-chat__composer-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  justify-content: flex-end;
}

.pro-sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  border: 0;
}

.pro-chat__header {
  padding-bottom: 0.75rem;
  margin-bottom: 0.75rem;
  border-bottom: 1px solid var(--pf-vet-border);
}

.pro-chat__thread-top {
  display: flex;
  align-items: baseline;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.pro-chat__thread-pet {
  font-size: 0.8125rem;
  font-weight: 400;
  color: var(--pf-vet-text-muted);
}

.pro-chat__thread-preview {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.8125rem;
  color: var(--pf-vet-text-muted);
  font-weight: 400;
}

.pro-chat__thread-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-top: 0.35rem;
}

.pro-chat__thread-date {
  font-size: 0.75rem;
  color: var(--pf-vet-text-muted);
}

.pro-chat__media-img {
  display: block;
  max-width: 100%;
  max-height: 220px;
  border-radius: 12px;
  object-fit: cover;
}

.pro-chat__media-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  margin: 0.35rem 0;
  color: inherit;
  text-decoration: none;
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
