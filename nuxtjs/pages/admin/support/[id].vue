<template>
  <div data-testid="admin-support-detail-page">
    <p class="pro-mb-md">
      <NuxtLink to="/admin/support" data-testid="admin-support-back">← {{ $t('admin.support.back') }}</NuxtLink>
    </p>
    <ProPageHeader
      :title="ticket?.subject || $t('admin.support.detailTitle')"
      :subtitle="ticketMeta"
    />

    <div v-if="loading" class="text-muted">{{ $t('common.loading') }}</div>
    <template v-else-if="ticket">
      <ProCard class="pro-mb-lg">
        <div class="support-meta pro-mb-md">
          <ProBadge :variant="statusVariant(ticket.status)">{{ statusLabel(ticket.status) }}</ProBadge>
          <ProBadge variant="neutral">{{ sourceLabel(ticket.source) }}</ProBadge>
          <span class="text-muted">{{ formatDate(ticket.createdAt) }}</span>
        </div>
        <div class="pro-field pro-mb-md">
          <label class="pro-label" for="support-status">{{ $t('admin.support.status') }}</label>
          <select
            id="support-status"
            v-model="statusDraft"
            class="pro-select"
            data-testid="admin-support-status"
            @change="saveStatus"
          >
            <option value="open">{{ $t('admin.support.status_open') }}</option>
            <option value="in_progress">{{ $t('admin.support.status_in_progress') }}</option>
            <option value="resolved">{{ $t('admin.support.status_resolved') }}</option>
            <option value="closed">{{ $t('admin.support.status_closed') }}</option>
          </select>
          <p v-if="statusMsg" class="pro-hint">{{ statusMsg }}</p>
        </div>
        <h3 class="pro-mb-sm">{{ $t('admin.support.message') }}</h3>
        <p class="support-message" data-testid="admin-support-message">{{ ticket.message }}</p>
      </ProCard>

      <ProCard class="pro-mb-lg">
        <div class="support-diag-head">
          <h3>{{ $t('admin.support.diagnostics') }}</h3>
          <ProButton type="button" variant="secondary" test-id="admin-support-download-diag" @click="downloadDiag">
            {{ $t('admin.support.downloadDiag') }}
          </ProButton>
        </div>
        <pre class="support-diag-pre" data-testid="admin-support-diagnostics">{{ diagnosticsPretty }}</pre>
      </ProCard>

      <ProCard class="pro-mb-lg">
        <h3 class="pro-mb-md">{{ $t('admin.support.replies') }}</h3>
        <ProEmptyState
          v-if="!(ticket.replies || []).length"
          :title="$t('admin.support.noReplies')"
        />
        <ul v-else class="support-replies" data-testid="admin-support-replies">
          <li v-for="r in ticket.replies" :key="r.id">
            <strong>{{ r.authorName || 'Admin' }}</strong>
            <span class="text-muted"> · {{ formatDate(r.createdAt) }}</span>
            <p>{{ r.body }}</p>
          </li>
        </ul>
        <form class="pro-form pro-mt-md" @submit.prevent="sendReply">
          <div class="pro-field">
            <label class="pro-label" for="support-reply">{{ $t('admin.support.replyLabel') }}</label>
            <textarea
              id="support-reply"
              v-model="replyBody"
              class="pro-textarea"
              rows="4"
              required
              data-testid="admin-support-reply"
            />
          </div>
          <p v-if="replyMsg" class="pro-hint" data-testid="admin-support-reply-msg">{{ replyMsg }}</p>
          <ProButton type="submit" test-id="admin-support-reply-submit" :disabled="replying">
            {{ $t('admin.support.replySubmit') }}
          </ProButton>
        </form>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const route = useRoute()
const { t } = useI18n()
const ticket = ref<any>(null)
const loading = ref(true)
const statusDraft = ref('open')
const statusMsg = ref('')
const replyBody = ref('')
const replyMsg = ref('')
const replying = ref(false)

const ticketMeta = computed(() => {
  if (!ticket.value) return ''
  const name = ticket.value.creatorName || ticket.value.creatorEmail || ''
  const email = ticket.value.creatorEmail || ''
  const role = ticket.value.creatorRole || ''
  return [name, email, role].filter(Boolean).join(' · ')
})

const diagnosticsPretty = computed(() => {
  const d = ticket.value?.diagnostics
  if (!d) return '{}'
  try {
    return JSON.stringify(typeof d === 'string' ? JSON.parse(d) : d, null, 2)
  } catch {
    return String(d)
  }
})

function statusLabel(status: string) {
  return t(`admin.support.status_${status}`)
}
function sourceLabel(source: string) {
  return t(`admin.support.source_${source}`)
}
function statusVariant(status: string): 'neutral' | 'success' | 'warning' | 'danger' {
  switch (status) {
    case 'open': return 'danger'
    case 'in_progress': return 'warning'
    case 'resolved': return 'success'
    case 'closed': return 'neutral'
    default: return 'neutral'
  }
}
function formatDate(iso: string) {
  if (!iso) return '—'
  try { return new Date(iso).toLocaleString() } catch { return iso }
}

async function load() {
  loading.value = true
  try {
    const id = route.params.id as string
    const res: any = await $fetch(`/api/admin/support/tickets/${id}`)
    ticket.value = res.data ?? res
    statusDraft.value = ticket.value?.status || 'open'
  } finally {
    loading.value = false
  }
}

async function saveStatus() {
  statusMsg.value = ''
  const id = route.params.id as string
  const res: any = await $fetch(`/api/admin/support/tickets/${id}`, {
    method: 'PATCH',
    body: { status: statusDraft.value },
  })
  const data = res.data ?? res
  if (ticket.value) ticket.value.status = data.status
  statusMsg.value = t('admin.support.statusUpdated')
}

async function sendReply() {
  if (!replyBody.value.trim() || replying.value) return
  replying.value = true
  replyMsg.value = ''
  try {
    const id = route.params.id as string
    await $fetch(`/api/admin/support/tickets/${id}/replies`, {
      method: 'POST',
      body: { body: replyBody.value.trim() },
    })
    replyBody.value = ''
    replyMsg.value = t('admin.support.replySent')
    await load()
  } finally {
    replying.value = false
  }
}

function downloadDiag() {
  const blob = new Blob([diagnosticsPretty.value], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `support-${route.params.id}-diagnostics.json`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(() => { void load() })
</script>

<style scoped>
.support-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.support-message {
  white-space: pre-wrap;
  margin: 0;
}
.support-diag-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}
.support-diag-pre {
  max-height: 24rem;
  overflow: auto;
  font-size: 0.75rem;
  background: var(--pf-vet-bg);
  padding: 0.75rem;
  border-radius: var(--pf-vet-radius);
  white-space: pre-wrap;
  word-break: break-word;
}
.support-replies {
  list-style: none;
  margin: 0;
  padding: 0;
}
.support-replies li {
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.support-replies p {
  margin: 0.35rem 0 0;
  white-space: pre-wrap;
}
.pro-textarea {
  width: 100%;
  padding: 0.6rem 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  background: var(--pf-vet-surface);
  color: var(--pf-vet-text);
  font: inherit;
  resize: vertical;
}
.text-muted {
  color: var(--pf-vet-text-muted);
}
</style>
