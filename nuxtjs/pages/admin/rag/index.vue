<template>
  <div data-testid="admin-rag-page">
    <ProPageHeader
      :title="$t('admin.rag.title')"
      :subtitle="$t('admin.rag.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="admin-rag-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton
          v-if="aiOn"
          variant="ghost"
          test-id="admin-rag-reindex"
          :disabled="reindexing"
          @click="reindex"
        >
          {{ $t('admin.rag.reindex') }}
        </ProButton>
        <ProButton v-if="aiOn" test-id="admin-rag-new" @click="showUpload = true">
          {{ $t('admin.rag.newUpload') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <ProEmptyState
      v-if="!aiOn"
      data-testid="admin-rag-disabled"
      :title="$t('admin.rag.disabledTitle')"
      :description="$t('admin.rag.disabledHint')"
    />

    <template v-else>
      <ProCard class="pro-mb-lg" data-testid="admin-rag-improve-stats">
        <h3 class="pro-mb-sm">{{ $t('admin.rag.improveStatsTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('admin.rag.improveStatsHint') }}</p>
        <div v-if="improveStats" class="pro-flex-gap" style="flex-wrap:wrap;gap:1rem">
          <div data-testid="admin-rag-stats-total">
            <strong>{{ improveStats.total }}</strong>
            <span class="pro-hint"> {{ $t('admin.rag.statsTotal') }}</span>
          </div>
          <div data-testid="admin-rag-stats-p50">
            <strong>{{ Math.round(improveStats.latencyP50Ms) }}ms</strong>
            <span class="pro-hint"> p50</span>
          </div>
          <div data-testid="admin-rag-stats-p95">
            <strong>{{ Math.round(improveStats.latencyP95Ms) }}ms</strong>
            <span class="pro-hint"> p95</span>
          </div>
          <div data-testid="admin-rag-stats-cancel">
            <strong>{{ Math.round((improveStats.cancelRate || 0) * 100) }}%</strong>
            <span class="pro-hint"> {{ $t('admin.rag.statsCancel') }}</span>
          </div>
          <div data-testid="admin-rag-stats-error">
            <strong>{{ Math.round((improveStats.errorRate || 0) * 100) }}%</strong>
            <span class="pro-hint"> {{ $t('admin.rag.statsError') }}</span>
          </div>
        </div>
        <p v-else class="pro-hint">{{ $t('admin.rag.statsLoading') }}</p>
      </ProCard>

      <ProCard v-if="showUpload" class="pro-mb-lg" data-testid="admin-rag-upload-card">
        <h3 class="pro-mb-md">{{ $t('admin.rag.uploadTitle') }}</h3>
        <form class="pro-form" @submit.prevent="upload">
          <div class="pro-field">
            <label class="pro-label" for="rag-title">{{ $t('admin.rag.fieldTitle') }}</label>
            <input id="rag-title" v-model="title" type="text" class="pro-input" data-testid="admin-rag-title">
          </div>
          <div class="pro-field">
            <label class="pro-label" for="rag-file">{{ $t('admin.rag.file') }}</label>
            <input
              id="rag-file"
              type="file"
              accept=".pdf,.txt,.md,application/pdf,text/plain,text/markdown"
              class="pro-input"
              data-testid="admin-rag-file"
              required
              @change="onFile"
            >
          </div>
          <p v-if="uploadError" class="pro-hint pro-hint--error" data-testid="admin-rag-upload-error">{{ uploadError }}</p>
          <div class="pro-flex-gap">
            <ProButton type="submit" test-id="admin-rag-upload" :disabled="uploading || !file">
              {{ $t('admin.rag.uploadSubmit') }}
            </ProButton>
            <ProButton variant="ghost" type="button" @click="showUpload = false">{{ $t('common.cancel') }}</ProButton>
          </div>
        </form>
      </ProCard>

      <div class="pro-mb-md pro-flex-gap">
        <label class="pro-label" for="rag-status-filter">{{ $t('admin.rag.filterStatus') }}</label>
        <select id="rag-status-filter" v-model="statusFilter" class="pro-input" data-testid="admin-rag-status-filter" @change="load">
          <option value="">{{ $t('admin.rag.statusAll') }}</option>
          <option value="pending">{{ $t('admin.rag.status.pending') }}</option>
          <option value="indexing">{{ $t('admin.rag.status.indexing') }}</option>
          <option value="ready">{{ $t('admin.rag.status.ready') }}</option>
          <option value="failed">{{ $t('admin.rag.status.failed') }}</option>
          <option value="rejected">{{ $t('admin.rag.status.rejected') }}</option>
        </select>
      </div>

      <p v-if="actionOk" class="pro-hint pro-mb-md" data-testid="admin-rag-action-ok">{{ actionOk }}</p>
      <p v-if="actionError" class="pro-hint pro-hint--error pro-mb-md" data-testid="admin-rag-action-error">{{ actionError }}</p>
      <ProCard>
        <ProTable :empty="!docs.length" :empty-title="$t('admin.rag.empty')">
          <thead>
            <tr>
              <th>{{ $t('admin.rag.colDate') }}</th>
              <th>{{ $t('admin.rag.colTitle') }}</th>
              <th>{{ $t('admin.rag.colScope') }}</th>
              <th>{{ $t('admin.rag.colStatus') }}</th>
              <th>{{ $t('admin.rag.colChunks') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="d in docs" :key="d.id" data-testid="admin-rag-row">
              <td>{{ d.createdAt?.substring(0, 16)?.replace('T', ' ') }}</td>
              <td>{{ d.title || d.filename }}</td>
              <td>{{ scopeLabel(d.scope) }}</td>
              <td><ProBadge :variant="statusVariant(d.status)">{{ statusLabel(d.status) }}</ProBadge></td>
              <td>{{ d.chunkCount ?? 0 }}</td>
              <td class="pro-flex-gap">
                <ProButton
                  v-if="canDownload(d)"
                  variant="ghost"
                  test-id="admin-rag-download"
                  @click="download(d)"
                >
                  {{ $t('admin.rag.download') }}
                </ProButton>
                <ProButton
                  v-if="d.status === 'pending'"
                  variant="primary"
                  test-id="admin-rag-approve"
                  :disabled="busyId === d.id"
                  @click="approve(d)"
                >
                  {{ $t('admin.rag.approve') }}
                </ProButton>
                <ProButton
                  v-if="d.status === 'pending'"
                  variant="ghost"
                  test-id="admin-rag-reject"
                  :disabled="busyId === d.id"
                  @click="reject(d)"
                >
                  {{ $t('admin.rag.reject') }}
                </ProButton>
                <ProButton
                  variant="ghost"
                  test-id="admin-rag-delete"
                  :disabled="busyId === d.id || d.status === 'indexing'"
                  @click="remove(d)"
                >
                  {{ $t('common.delete') }}
                </ProButton>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import { isPublicFlagOn } from '~/utils/public-feature-flag'

definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const runtimeConfig = useRuntimeConfig()
const aiOn = computed(() => isPublicFlagOn(runtimeConfig.public.aiCrAdvancedEnabled))

type RagDoc = {
  id: string
  scope: string
  title?: string
  filename?: string
  status: string
  chunkCount?: number
  createdAt?: string
  errorMessage?: string
}

const docs = ref<RagDoc[]>([])
const improveStats = ref<{
  total: number
  latencyP50Ms: number
  latencyP95Ms: number
  cancelRate: number
  errorRate: number
} | null>(null)
const showUpload = ref(false)
const file = ref<File | null>(null)
const title = ref('')
const uploading = ref(false)
const uploadError = ref('')
const actionError = ref('')
const actionOk = ref('')
const busyId = ref('')
const reindexing = ref(false)
const statusFilter = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null

function onFile(e: Event) {
  const input = e.target as HTMLInputElement
  file.value = input.files?.[0] ?? null
}

/** Reject purges the source blob — hide download when no file remains. */
function canDownload(d: RagDoc) {
  return d.status === 'pending'
    || d.status === 'indexing'
    || d.status === 'ready'
    || d.status === 'failed'
}

function download(d: RagDoc) {
  const a = document.createElement('a')
  a.href = `/api/admin/rag/documents/${d.id}/download`
  a.rel = 'noopener'
  a.download = d.filename || d.title || 'document'
  document.body.appendChild(a)
  a.click()
  a.remove()
}

function statusVariant(s: string) {
  switch (s) {
    case 'ready': return 'success'
    case 'pending':
    case 'indexing': return 'warning'
    case 'failed':
    case 'rejected': return 'danger'
    default: return 'neutral'
  }
}

function statusLabel(s: string) {
  const key = `admin.rag.status.${s}`
  const translated = t(key)
  return translated === key ? s : translated
}

function scopeLabel(s: string) {
  if (s === 'platform') return t('admin.rag.scopePlatform')
  if (s === 'practice') return t('admin.rag.scopePractice')
  return s
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function maybePoll() {
  stopPoll()
  if (!aiOn.value) return
  if (!docs.value.some(d => d.status === 'indexing')) return
  pollTimer = setInterval(() => { void load() }, 2000)
}

async function load() {
  if (!aiOn.value) return
  const q = statusFilter.value ? `?status=${encodeURIComponent(statusFilter.value)}` : ''
  const res: any = await $fetch(`/api/admin/rag/documents${q}`)
  docs.value = res.data ?? res ?? []
  maybePoll()
  void loadImproveStats()
}

async function loadImproveStats() {
  try {
    const res: any = await $fetch('/api/admin/rag/improve-stats?days=7')
    improveStats.value = res.data ?? res
  } catch {
    improveStats.value = null
  }
}

async function upload() {
  if (!file.value) return
  uploading.value = true
  uploadError.value = ''
  try {
    const fd = new FormData()
    fd.append('file', file.value)
    if (title.value.trim()) fd.append('title', title.value.trim())
    await $fetch('/api/admin/rag/documents', { method: 'POST', body: fd })
    showUpload.value = false
    file.value = null
    title.value = ''
    await load()
  } catch (e: any) {
    uploadError.value = e?.data?.error?.message || e?.message || t('admin.rag.uploadFailed')
  } finally {
    uploading.value = false
  }
}

async function approve(d: RagDoc) {
  busyId.value = d.id
  actionError.value = ''
  actionOk.value = ''
  try {
    await $fetch(`/api/admin/rag/documents/${d.id}/approve`, { method: 'POST' })
    await load()
  } catch (e: any) {
    actionError.value = e?.data?.error?.message || e?.message || t('admin.rag.actionFailed')
  } finally {
    busyId.value = ''
  }
}

async function reject(d: RagDoc) {
  busyId.value = d.id
  actionError.value = ''
  actionOk.value = ''
  try {
    await $fetch(`/api/admin/rag/documents/${d.id}/reject`, {
      method: 'POST',
      body: { reason: 'rejected_by_admin' },
    })
    await load()
  } catch (e: any) {
    actionError.value = e?.data?.error?.message || e?.message || t('admin.rag.actionFailed')
  } finally {
    busyId.value = ''
  }
}

async function remove(d: RagDoc) {
  if (!confirm(t('admin.rag.confirmDelete'))) return
  busyId.value = d.id
  actionError.value = ''
  actionOk.value = ''
  try {
    await $fetch(`/api/admin/rag/documents/${d.id}`, { method: 'DELETE' })
    await load()
  } catch (e: any) {
    actionError.value = e?.data?.error?.message || e?.message || t('admin.rag.actionFailed')
  } finally {
    busyId.value = ''
  }
}

async function reindex() {
  reindexing.value = true
  actionError.value = ''
  actionOk.value = ''
  try {
    const res: any = await $fetch('/api/admin/rag/reindex', { method: 'POST' })
    const stats = res.data ?? res ?? {}
    actionOk.value = t('admin.rag.reindexDone', {
      indexed: stats.indexed ?? 0,
      failed: stats.failed ?? 0,
    })
    await load()
  } catch (e: any) {
    actionError.value = e?.data?.error?.message || e?.message || t('admin.rag.actionFailed')
  } finally {
    reindexing.value = false
  }
}

onMounted(() => {
  if (aiOn.value) void load()
})

onBeforeUnmount(() => stopPoll())
</script>
