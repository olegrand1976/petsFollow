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
    <p v-else-if="loadError" class="pro-error" role="alert" data-testid="admin-support-detail-error">
      {{ loadError }}
    </p>
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
            <option v-for="st in STATUSES" :key="st" :value="st">{{ statusLabel(st) }}</option>
          </select>
          <p
            v-if="statusMsg"
            :class="statusMsgError ? 'pro-error' : 'pro-hint'"
            data-testid="admin-support-status-msg"
          >
            {{ statusMsg }}
          </p>
        </div>
        <h3 class="pro-mb-sm">{{ $t('admin.support.message') }}</h3>
        <p class="support-message" data-testid="admin-support-message">{{ ticket.message }}</p>
      </ProCard>

      <ProCard class="pro-mb-lg">
        <h3 class="pro-mb-md">{{ $t('admin.support.statusHistory') }}</h3>
        <ProEmptyState
          v-if="!(ticket.statusHistory || []).length"
          :title="$t('admin.support.noStatusHistory')"
        />
        <ul v-else class="support-history" data-testid="admin-support-status-history">
          <li v-for="ev in ticket.statusHistory" :key="ev.id">
            <strong>{{ statusTransitionLabel(ev) }}</strong>
            <span class="text-muted">
              · {{ formatDate(ev.createdAt) }}
              <template v-if="ev.changedName"> · {{ ev.changedName }}</template>
            </span>
          </li>
        </ul>
      </ProCard>

      <ProCard class="pro-mb-lg">
        <div class="support-diag-head">
          <h3>{{ $t('admin.support.diagnostics') }}</h3>
          <ProButton type="button" variant="secondary" test-id="admin-support-download-diag" @click="downloadDiag">
            {{ $t('admin.support.downloadDiag') }}
          </ProButton>
        </div>
        <p
          v-if="ticket.route"
          class="support-origin-route pro-mb-md"
          data-testid="admin-support-origin-route"
        >
          <span class="pro-label">{{ $t('admin.support.originRoute') }}</span>
          <code>{{ ticket.route }}</code>
        </p>
        <pre class="support-diag-pre" data-testid="admin-support-diagnostics">{{ diagnosticsPretty }}</pre>
      </ProCard>

      <ProCard class="pro-mb-lg">
        <h3 class="pro-mb-md">{{ $t('admin.support.attachments') }}</h3>
        <ProEmptyState
          v-if="!(ticket.attachments || []).length"
          :title="$t('admin.support.noAttachments')"
        />
        <ul v-else class="support-attachments" data-testid="admin-support-attachments">
          <li v-for="att in ticket.attachments" :key="att.id">
            <a
              :href="`/api/admin/support/tickets/${ticket.id}/attachments/${att.id}/download`"
              target="_blank"
              rel="noopener"
              data-testid="admin-support-attachment-link"
            >
              {{ att.fileName }}
            </a>
            <span class="text-muted">
              · {{ formatSize(att.sizeBytes) }}
              · {{ formatDate(att.createdAt) }}
              <template v-if="att.uploaderName"> · {{ att.uploaderName }}</template>
            </span>
          </li>
        </ul>
        <form class="pro-form pro-mt-md" @submit.prevent="uploadAttachment">
          <div class="pro-field">
            <label class="pro-label" for="support-attachment">{{ $t('admin.support.attachmentLabel') }}</label>
            <input
              id="support-attachment"
              ref="fileInput"
              type="file"
              accept=".pdf,image/jpeg,image/png,image/webp,.jpg,.jpeg,.png,.webp"
              class="pro-input"
              data-testid="admin-support-attachment-input"
              @change="onFilePicked"
            >
          </div>
          <p v-if="attachMsg" class="pro-hint" data-testid="admin-support-attach-msg">{{ attachMsg }}</p>
          <ProButton type="submit" test-id="admin-support-attach-submit" :disabled="attaching || !pickedFile">
            {{ $t('admin.support.attachmentSubmit') }}
          </ProButton>
        </form>
      </ProCard>

      <ProCard class="pro-mb-lg">
        <h3 class="pro-mb-md">{{ $t('admin.support.comments') }}</h3>
        <ProEmptyState
          v-if="!(ticket.replies || []).length"
          :title="$t('admin.support.noComments')"
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
            <label class="pro-label" for="support-reply">{{ $t('admin.support.commentLabel') }}</label>
            <textarea
              id="support-reply"
              v-model="replyBody"
              class="pro-textarea"
              rows="4"
              required
              data-testid="admin-support-reply"
            />
          </div>
          <p
            v-if="replyMsg"
            :class="replyMsgError ? 'pro-error' : 'pro-hint'"
            data-testid="admin-support-reply-msg"
          >
            {{ replyMsg }}
          </p>
          <ProButton type="submit" test-id="admin-support-reply-submit" :disabled="replying">
            {{ $t('admin.support.commentSubmit') }}
          </ProButton>
        </form>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

const STATUSES = ['open', 'in_progress', 'to_test', 'done', 'closed'] as const

const route = useRoute()
const { t } = useI18n()
const ticket = ref<any>(null)
const loading = ref(true)
const loadError = ref('')
const statusDraft = ref('open')
const statusMsg = ref('')
const statusMsgError = ref(false)
const replyBody = ref('')
const replyMsg = ref('')
const replyMsgError = ref(false)
const replying = ref(false)
const attachMsg = ref('')
const attaching = ref(false)
const pickedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

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
    case 'to_test': return 'neutral'
    case 'done': return 'success'
    case 'closed': return 'neutral'
    default: return 'neutral'
  }
}
function statusTransitionLabel(ev: { fromStatus?: string | null, toStatus: string }) {
  const to = statusLabel(ev.toStatus)
  if (!ev.fromStatus) return to
  return `${statusLabel(ev.fromStatus)} → ${to}`
}
function formatDate(iso: string) {
  if (!iso) return '—'
  try { return new Date(iso).toLocaleString() } catch { return iso }
}
function formatSize(bytes: number) {
  if (!bytes || bytes < 0) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function load(opts?: { quiet?: boolean }) {
  const quiet = Boolean(opts?.quiet)
  if (!quiet) loading.value = true
  loadError.value = ''
  try {
    const id = route.params.id as string
    const res: any = await $fetch(`/api/admin/support/tickets/${id}`)
    ticket.value = res.data ?? res
    statusDraft.value = ticket.value?.status || 'open'
  } catch {
    loadError.value = t('admin.support.loadError')
    if (!quiet) ticket.value = null
  } finally {
    if (!quiet) loading.value = false
  }
}

async function saveStatus() {
  statusMsg.value = ''
  statusMsgError.value = false
  const previous = ticket.value?.status || 'open'
  try {
    const id = route.params.id as string
    await $fetch(`/api/admin/support/tickets/${id}`, {
      method: 'PATCH',
      body: { status: statusDraft.value },
    })
    statusMsg.value = t('admin.support.statusUpdated')
    await load({ quiet: true })
  } catch {
    statusDraft.value = previous
    statusMsgError.value = true
    statusMsg.value = t('admin.support.patchError')
  }
}

async function sendReply() {
  if (!replyBody.value.trim() || replying.value) return
  replying.value = true
  replyMsg.value = ''
  replyMsgError.value = false
  try {
    const id = route.params.id as string
    await $fetch(`/api/admin/support/tickets/${id}/replies`, {
      method: 'POST',
      body: { body: replyBody.value.trim() },
    })
    replyBody.value = ''
    replyMsg.value = t('admin.support.commentSent')
    await load({ quiet: true })
  } catch {
    replyMsgError.value = true
    replyMsg.value = t('admin.support.commentError')
  } finally {
    replying.value = false
  }
}

function onFilePicked(ev: Event) {
  const input = ev.target as HTMLInputElement
  pickedFile.value = input.files?.[0] ?? null
  attachMsg.value = ''
}

async function uploadAttachment() {
  if (!pickedFile.value || attaching.value) return
  attaching.value = true
  attachMsg.value = ''
  try {
    const id = route.params.id as string
    const fd = new FormData()
    fd.append('file', pickedFile.value)
    await $fetch(`/api/admin/support/tickets/${id}/attachments`, {
      method: 'POST',
      body: fd,
    })
    pickedFile.value = null
    if (fileInput.value) fileInput.value.value = ''
    attachMsg.value = t('admin.support.attachmentUploaded')
    await load({ quiet: true })
  } catch {
    attachMsg.value = t('admin.support.attachmentError')
  } finally {
    attaching.value = false
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
.support-origin-route {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem 0.75rem;
}
.support-diag-pre {
  margin: 0;
  padding: 0.75rem;
  background: var(--pf-vet-bg);
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-sm, 6px);
  overflow: auto;
  max-height: 24rem;
  font-size: 0.8rem;
}
.support-replies,
.support-history,
.support-attachments {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.support-replies p {
  margin: 0.25rem 0 0;
  white-space: pre-wrap;
}
.text-muted {
  color: var(--pf-vet-text-muted);
  font-size: 0.85em;
}
</style>
