<template>
  <div data-testid="messages-page">
    <ProPageHeader :title="$t('messages.title')" :subtitle="$t('messages.subtitle')" />
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
              <p>{{ m.body }}</p>
            </div>
          </div>
          <form class="pro-chat__composer" @submit.prevent="send">
            <textarea
              v-model="draft"
              class="pro-textarea"
              rows="3"
              :placeholder="$t('messages.placeholder')"
              required
            />
            <ProButton type="submit" :disabled="!draft.trim()">{{ $t('common.send') }}</ProButton>
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
const { user, fetchUser } = useProUser()
const { refresh: refreshNotif } = useProNotifications()

const threads = ref<any[]>([])
const messages = ref<any[]>([])
const active = ref<any>(null)
const draft = ref('')

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

async function loadThreads() {
  const res: any = await $fetch('/api/messaging/threads')
  threads.value = res.data ?? res ?? []
  if (active.value) {
    active.value = threads.value.find((item) => item.id === active.value.id) ?? active.value
  }
}

async function select(thread: any) {
  active.value = thread
  const res: any = await $fetch(`/api/messaging/threads/${thread.id}/messages`)
  messages.value = res.data ?? res ?? []
  if ((thread.unreadCount ?? 0) > 0) {
    await $fetch(`/api/messaging/threads/${thread.id}/read`, { method: 'POST' })
    await loadThreads()
    await refreshNotif()
  }
}

async function openThreadFromQuery() {
  const threadId = String(route.query.thread || '')
  if (!threadId) return
  const thread = threads.value.find((item) => item.id === threadId)
  if (thread) await select(thread)
}

onMounted(async () => {
  await fetchUser()
  await loadThreads()
  await openThreadFromQuery()
})

watch(
  () => route.query.thread,
  async () => {
    if (threads.value.length) await openThreadFromQuery()
  },
)

async function send() {
  if (!active.value || !draft.value.trim()) return
  await $fetch(`/api/messaging/threads/${active.value.id}/messages`, {
    method: 'POST',
    body: { body: draft.value },
  })
  draft.value = ''
  const res: any = await $fetch(`/api/messaging/threads/${active.value.id}/messages`)
  messages.value = res.data ?? res ?? []
  await loadThreads()
}
</script>

<style scoped>
.pro-chat__composer {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
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
</style>
