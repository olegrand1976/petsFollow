<template>
  <div>
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
          @click="select(t)"
        >
          <strong>{{ threadLabel(t) }}</strong>
          <span v-if="t.lastMessagePreview" class="pro-chat__thread-preview">{{ t.lastMessagePreview }}</span>
          <ProBadge v-if="t.unreadCount > 0" variant="warning">{{ unreadLabel(t.unreadCount) }}</ProBadge>
        </button>
      </aside>
      <section class="pro-chat__messages">
        <template v-if="active">
          <div class="pro-chat__list">
            <div v-for="m in messages" :key="m.id" class="pro-chat__bubble">
              <small>{{ m.senderUserId.slice(0, 8) }}…</small>
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

const { t } = useI18n()

const threads = ref<any[]>([])
const messages = ref<any[]>([])
const active = ref<any>(null)
const draft = ref('')

function threadLabel(thread: any) {
  return thread.clientName || t('common.clientFallback', { id: thread.clientUserId.slice(0, 8) })
}

function unreadLabel(count: number) {
  return count > 1 ? t('messages.unreadPlural', { count }) : t('messages.unread', { count })
}

onMounted(async () => {
  const res: any = await $fetch('/api/messaging/threads')
  threads.value = res.data ?? res ?? []
})

async function select(thread: any) {
  active.value = thread
  const res: any = await $fetch(`/api/messaging/threads/${thread.id}/messages`)
  messages.value = res.data ?? res ?? []
}

async function send() {
  if (!active.value || !draft.value.trim()) return
  await $fetch(`/api/messaging/threads/${active.value.id}/messages`, {
    method: 'POST',
    body: { body: draft.value },
  })
  draft.value = ''
  const res: any = await $fetch(`/api/messaging/threads/${active.value.id}/messages`)
  messages.value = res.data ?? res ?? []
}
</script>

<style scoped>
.pro-chat__composer {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.pro-chat__thread-preview {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.8125rem;
  color: var(--pf-vet-text-muted);
  font-weight: 400;
}
</style>
