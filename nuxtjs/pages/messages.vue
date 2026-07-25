<template>
  <div data-testid="messages-page">
    <ProPageHeader :title="$t('messages.title')" :subtitle="$t('messages.subtitle')" />
    <p v-if="actionError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ actionError }}</p>
    <div class="pro-chat">
      <aside class="pro-chat__threads">
        <h3 class="pro-card__title">{{ $t('messages.threads') }}</h3>
        <ProEmptyState
          v-if="!threads.length"
          :title="$t('messages.emptyTitle')"
          :description="$t('messages.emptyDescription')"
        />
        <button
          v-for="t in threads"
          :key="t.id"
          type="button"
          class="pro-chat__thread-btn"
          :class="{ 'pro-chat__thread-btn--active': active?.id === t.id }"
          :data-testid="`thread-${t.id}`"
          @click="select(t)"
        >
          <strong>{{ threadLabel(t) }}</strong>
          <span v-if="t.lastMessagePreview" class="pro-chat__thread-preview">{{ t.lastMessagePreview }}</span>
          <ProBadge v-if="t.unreadCount > 0" variant="warning">{{ unreadLabel(t.unreadCount) }}</ProBadge>
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
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const route = useRoute()
const { t } = useI18n()
const { mapError } = useApiError()
const { user, fetchUser } = useProUser()
const { refresh: refreshNotif } = useProNotifications()

const threads = ref<any[]>([])
const messages = ref<any[]>([])
const active = ref<any>(null)
const draft = ref('')
const sending = ref(false)
const actionError = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const refreshing = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const POLL_MS = 4_000
const MAX_MEDIA_BYTES = 25 * 1024 * 1024

const vetUserId = computed(() => user.value?.userId ?? user.value?.id ?? '')

function threadLabel(thread: any) {
  return thread.clientName || t('common.clientFallback', { id: thread.clientUserId?.slice(0, 8) ?? '' })
}

function unreadLabel(count: number) {
  return count > 1 ? t('messages.unreadPlural', { count }) : t('messages.unread', { count })
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
  await loadThreads()
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

.pro-chat__thread-preview {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.8125rem;
  color: var(--pf-vet-text-muted);
  font-weight: 400;
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
