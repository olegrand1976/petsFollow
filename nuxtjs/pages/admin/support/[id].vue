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
      <ProCard class="pro-mb-lg support-banner">
        <div class="support-meta">
          <ProBadge :variant="supportStatusBadgeVariant(ticket.status)">
            {{ statusLabel(ticket.status) }}
          </ProBadge>
          <ProBadge variant="neutral">{{ sourceLabel(ticket.source) }}</ProBadge>
          <span class="text-muted">{{ formatDate(ticket.createdAt) }}</span>
        </div>

        <div
          class="support-workflow"
          role="group"
          :aria-label="$t('admin.support.statusWorkflowAria')"
          data-testid="admin-support-status"
        >
          <button
            v-for="st in SUPPORT_WORKFLOW_ORDER"
            :key="st"
            type="button"
            class="support-workflow__step"
            :class="{
              'is-current': ticket.status === st,
              'is-reachable': canAdvanceSupportStatus(ticket.status, st),
            }"
            :aria-current="ticket.status === st ? 'step' : undefined"
            :aria-pressed="ticket.status === st"
            :disabled="statusSaving || ticket.status === st || !canAdvanceSupportStatus(ticket.status, st)"
            :title="stepTitle(st)"
            :data-testid="`admin-support-status-${st}`"
            @click="saveStatus(st)"
          >
            {{ statusLabel(st) }}
          </button>
        </div>
        <p v-if="statusHint" class="pro-hint" data-testid="admin-support-status-hint">
          {{ statusHint }}
        </p>
        <p
          v-if="statusMsg"
          :class="statusMsgError ? 'pro-error' : 'pro-hint'"
          data-testid="admin-support-status-msg"
        >
          {{ statusMsg }}
        </p>
      </ProCard>

      <div class="support-layout">
        <div class="support-layout__main">
          <ProCard class="pro-mb-lg">
            <h3 class="pro-mb-sm">{{ $t('admin.support.message') }}</h3>
            <p class="support-message" data-testid="admin-support-message">{{ ticket.message }}</p>
          </ProCard>

          <ProCard class="pro-mb-lg">
            <h3 class="pro-mb-md">{{ $t('admin.support.comments') }}</h3>
            <ProEmptyState
              v-if="!(ticket.replies || []).length"
              :title="$t('admin.support.noComments')"
            />
            <ul v-else class="support-replies" data-testid="admin-support-replies">
              <li v-for="r in ticket.replies" :key="r.id" class="support-reply">
                <div class="support-reply__head">
                  <strong>{{ r.authorName || 'Admin' }}</strong>
                  <span class="text-muted">{{ formatDate(r.createdAt) }}</span>
                </div>
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
        </div>

        <aside class="support-layout__side">
          <ProCard class="pro-mb-lg">
            <h3 class="pro-mb-md">{{ $t('admin.support.attachments') }}</h3>
            <ProEmptyState
              v-if="!(ticket.attachments || []).length"
              :title="$t('admin.support.noAttachments')"
            />
            <ul v-else class="support-attachments" data-testid="admin-support-attachments">
              <li v-for="att in ticket.attachments" :key="att.id" class="support-attach-row">
                <a
                  :href="`/api/admin/support/tickets/${ticket.id}/attachments/${att.id}/download`"
                  target="_blank"
                  rel="noopener"
                  data-testid="admin-support-attachment-link"
                >
                  {{ att.fileName }}
                </a>
                <span class="text-muted">
                  {{ formatSize(att.sizeBytes) }}
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
              <p
                v-if="attachMsg"
                :class="attachMsgError ? 'pro-error' : 'pro-hint'"
                data-testid="admin-support-attach-msg"
              >
                {{ attachMsg }}
              </p>
              <ProButton type="submit" test-id="admin-support-attach-submit" :disabled="attaching || !pickedFile">
                {{ $t('admin.support.attachmentSubmit') }}
              </ProButton>
            </form>
          </ProCard>

          <ProCard class="pro-mb-lg">
            <details class="support-diag" :open="Boolean(ticket.route)">
              <summary class="support-diag__summary">
                <span>{{ $t('admin.support.diagnostics') }}</span>
                <ProButton
                  type="button"
                  variant="secondary"
                  test-id="admin-support-download-diag"
                  @click.stop.prevent="downloadDiag"
                >
                  {{ $t('admin.support.downloadDiag') }}
                </ProButton>
              </summary>
              <p
                v-if="ticket.route"
                class="support-origin-route pro-mb-md"
                data-testid="admin-support-origin-route"
              >
                <span class="pro-label">{{ $t('admin.support.originRoute') }}</span>
                <code>{{ ticket.route }}</code>
              </p>
              <pre class="support-diag-pre" data-testid="admin-support-diagnostics">{{ diagnosticsPretty }}</pre>
            </details>
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
        </aside>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import {
  SUPPORT_WORKFLOW_ORDER,
  canAdvanceSupportStatus,
  supportStatusBadgeVariant,
} from '~/utils/support-status'

definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

const route = useRoute()
const { t } = useI18n()
const ticket = ref<any>(null)
const loading = ref(true)
const loadError = ref('')
const statusMsg = ref('')
const statusMsgError = ref(false)
const statusSaving = ref(false)
const replyBody = ref('')
const replyMsg = ref('')
const replyMsgError = ref(false)
const replying = ref(false)
const attachMsg = ref('')
const attachMsgError = ref(false)
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

const statusHint = computed(() => {
  const st = ticket.value?.status
  if (!st) return ''
  const next = SUPPORT_WORKFLOW_ORDER.filter(s => canAdvanceSupportStatus(st, s))
  if (!next.length) return ''
  return t('admin.support.statusTransitionHint', {
    next: next.map(statusLabel).join(', '),
  })
})

function statusLabel (status: string) {
  return t(`admin.support.status_${status}`)
}
function sourceLabel (source: string) {
  return t(`admin.support.source_${source}`)
}
function stepTitle (st: string) {
  if (!ticket.value) return statusLabel(st)
  if (ticket.value.status === st) return statusLabel(st)
  if (canAdvanceSupportStatus(ticket.value.status, st)) return statusLabel(st)
  return t('admin.support.statusTransitionBlocked')
}
function statusTransitionLabel (ev: { fromStatus?: string | null, toStatus: string }) {
  const to = statusLabel(ev.toStatus)
  if (!ev.fromStatus) return to
  return `${statusLabel(ev.fromStatus)} → ${to}`
}
function formatDate (iso: string) {
  if (!iso) return '—'
  try { return new Date(iso).toLocaleString() } catch { return iso }
}
function formatSize (bytes: number) {
  if (!bytes || bytes < 0) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function load (opts?: { quiet?: boolean }) {
  const quiet = Boolean(opts?.quiet)
  if (!quiet) loading.value = true
  loadError.value = ''
  try {
    const id = route.params.id as string
    const res: any = await $fetch(`/api/admin/support/tickets/${id}`)
    ticket.value = res.data ?? res
  } catch {
    loadError.value = t('admin.support.loadError')
    if (!quiet) ticket.value = null
  } finally {
    if (!quiet) loading.value = false
  }
}

async function saveStatus (next: string) {
  if (!ticket.value || statusSaving.value) return
  if (!canAdvanceSupportStatus(ticket.value.status, next)) return
  statusSaving.value = true
  statusMsg.value = ''
  statusMsgError.value = false
  try {
    const id = route.params.id as string
    await $fetch(`/api/admin/support/tickets/${id}`, {
      method: 'PATCH',
      body: { status: next },
    })
    statusMsg.value = t('admin.support.statusUpdated')
    await load({ quiet: true })
  } catch {
    statusMsgError.value = true
    statusMsg.value = t('admin.support.patchError')
  } finally {
    statusSaving.value = false
  }
}

async function sendReply () {
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

function onFilePicked (ev: Event) {
  const input = ev.target as HTMLInputElement
  pickedFile.value = input.files?.[0] ?? null
  attachMsg.value = ''
}

async function uploadAttachment () {
  if (!pickedFile.value || attaching.value) return
  attaching.value = true
  attachMsg.value = ''
  attachMsgError.value = false
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
    attachMsgError.value = true
    attachMsg.value = t('admin.support.attachmentError')
  } finally {
    attaching.value = false
  }
}

function downloadDiag () {
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
.support-banner {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}
.support-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.support-workflow {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}
.support-workflow__step {
  appearance: none;
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface);
  color: var(--pf-vet-text-muted);
  border-radius: 999px;
  padding: 0.35rem 0.75rem;
  font: inherit;
  font-size: 0.85rem;
  cursor: not-allowed;
  opacity: 0.55;
}
.support-workflow__step.is-reachable:not(:disabled) {
  cursor: pointer;
  opacity: 1;
  color: var(--pf-vet-text);
  border-color: var(--pf-vet-accent);
}
.support-workflow__step.is-reachable:not(:disabled):hover {
  background: color-mix(in srgb, var(--pf-vet-accent) 12%, var(--pf-vet-surface));
}
.support-workflow__step.is-current {
  opacity: 1;
  cursor: default;
  color: var(--pf-vet-primary);
  border-color: var(--pf-vet-primary);
  background: color-mix(in srgb, var(--pf-vet-primary) 10%, var(--pf-vet-surface));
  font-weight: 600;
}
.support-layout {
  display: grid;
  gap: 0;
  grid-template-columns: 1fr;
}
@media (min-width: 960px) {
  .support-layout {
    grid-template-columns: minmax(0, 1.4fr) minmax(18rem, 1fr);
    gap: 1rem;
    align-items: start;
  }
}
.support-message {
  white-space: pre-wrap;
  margin: 0;
  line-height: 1.5;
}
.support-reply {
  padding: 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-sm, 6px);
  background: var(--pf-vet-bg);
}
.support-reply__head {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: baseline;
  margin-bottom: 0.35rem;
}
.support-diag__summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  cursor: pointer;
  font-weight: 600;
  list-style: none;
}
.support-diag__summary::-webkit-details-marker {
  display: none;
}
.support-origin-route {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem 0.75rem;
  margin-top: 0.75rem;
}
.support-diag-pre {
  margin: 0.75rem 0 0;
  padding: 0.75rem;
  background: var(--pf-vet-bg);
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-sm, 6px);
  overflow: auto;
  max-height: 20rem;
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
  margin: 0;
  white-space: pre-wrap;
}
.support-attach-row {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}
.text-muted {
  color: var(--pf-vet-text-muted);
  font-size: 0.85em;
}
</style>
