<template>
  <div>
    <ProPageHeader title="Messagerie" subtitle="Échanges sécurisés avec vos clients." />
    <div class="pro-chat">
      <aside class="pro-chat__threads">
        <h3 class="pro-card__title">Conversations</h3>
        <ProEmptyState
          v-if="!threads.length"
          title="Aucune conversation"
          description="Les threads apparaîtront lorsque vos clients vous écriront."
        />
        <button
          v-for="t in threads"
          :key="t.id"
          type="button"
          class="pro-chat__thread-btn"
          :class="{ 'pro-chat__thread-btn--active': active?.id === t.id }"
          @click="select(t)"
        >
          <strong>{{ t.clientName || `Client ${t.clientUserId.slice(0, 8)}…` }}</strong>
          <span v-if="t.lastMessagePreview" class="pro-chat__thread-preview">{{ t.lastMessagePreview }}</span>
          <ProBadge v-if="t.unreadCount > 0" variant="warning">{{ t.unreadCount }} non lu{{ t.unreadCount > 1 ? 's' : '' }}</ProBadge>
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
              placeholder="Votre message…"
              required
            />
            <ProButton type="submit" :disabled="!draft.trim()">Envoyer</ProButton>
          </form>
        </template>
        <ProEmptyState
          v-else
          title="Sélectionnez une conversation"
          description="Choisissez un thread dans la liste pour afficher les messages."
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const threads = ref<any[]>([])
const messages = ref<any[]>([])
const active = ref<any>(null)
const draft = ref('')

onMounted(async () => {
  const res: any = await $fetch('/api/messaging/threads')
  threads.value = res.data ?? res ?? []
})

async function select(t: any) {
  active.value = t
  const res: any = await $fetch(`/api/messaging/threads/${t.id}/messages`)
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
