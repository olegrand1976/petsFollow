<template>
  <div data-testid="admin-compendium-imports-page">
    <ProPageHeader
      :title="$t('admin.compendium.title')"
      :subtitle="$t('admin.compendium.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="compendium-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton test-id="admin-compendium-new" @click="showUpload = true">
          {{ $t('admin.compendium.newImport') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <ProCard v-if="showUpload" class="pro-mb-lg" data-testid="admin-compendium-upload-card">
      <h3 class="pro-mb-md">{{ $t('admin.compendium.uploadTitle') }}</h3>
      <form class="pro-form" @submit.prevent="upload">
        <div class="pro-field">
          <label class="pro-label" for="compendium-file">{{ $t('admin.compendium.file') }}</label>
          <input
            id="compendium-file"
            type="file"
            accept="application/pdf,.pdf"
            class="pro-input"
            data-testid="admin-compendium-file"
            required
            @change="onFile"
          >
        </div>
        <div class="pro-form pro-form--inline">
          <div class="pro-field">
            <label class="pro-label" for="page-start">{{ $t('admin.compendium.pageStart') }}</label>
            <input id="page-start" v-model.number="pageStart" type="number" min="1" class="pro-input" required data-testid="admin-compendium-page-start">
          </div>
          <div class="pro-field">
            <label class="pro-label" for="page-end">{{ $t('admin.compendium.pageEnd') }}</label>
            <input id="page-end" v-model.number="pageEnd" type="number" min="1" class="pro-input" required data-testid="admin-compendium-page-end">
          </div>
        </div>
        <p v-if="uploadError" class="pro-hint pro-hint--error" data-testid="admin-compendium-upload-error">{{ uploadError }}</p>
        <div class="pro-flex-gap">
          <ProButton type="submit" test-id="admin-compendium-upload" :disabled="uploading || !file">
            {{ $t('admin.compendium.uploadSubmit') }}
          </ProButton>
          <ProButton variant="ghost" type="button" @click="showUpload = false">{{ $t('common.cancel') }}</ProButton>
        </div>
      </form>
    </ProCard>

    <p v-if="deleteError" class="pro-hint pro-hint--error pro-mb-md" data-testid="admin-compendium-delete-error">{{ deleteError }}</p>
    <ProCard>
      <ProTable :empty="!jobs.length" :empty-title="$t('admin.compendium.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.compendium.colDate') }}</th>
            <th>{{ $t('admin.compendium.colFile') }}</th>
            <th>{{ $t('admin.compendium.colPages') }}</th>
            <th>{{ $t('admin.compendium.colStatus') }}</th>
            <th>{{ $t('admin.compendium.colExtract') }}</th>
            <th>{{ $t('admin.compendium.colReview') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="j in jobs" :key="j.id" data-testid="admin-compendium-row">
            <td>{{ j.createdAt?.substring(0, 16)?.replace('T', ' ') }}</td>
            <td>{{ j.filename }}</td>
            <td>{{ j.pageStart }}–{{ j.pageEnd }}</td>
            <td><ProBadge :variant="statusVariant(j.status)">{{ statusLabel(j.status) }}</ProBadge></td>
            <td>{{ j.extractPct ?? 0 }}%</td>
            <td>{{ j.reviewPct ?? 0 }}%</td>
            <td class="pro-flex-gap">
              <NuxtLink :to="`/admin/compendium-imports/${j.id}`" class="pro-link">
                {{ $t('admin.compendium.open') }}
              </NuxtLink>
              <ProButton
                variant="ghost"
                test-id="admin-compendium-delete"
                :disabled="deletingId === j.id || j.status === 'extracting' || j.status === 'committing'"
                @click="removeJob(j)"
              >
                {{ $t('admin.compendium.delete') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const jobs = ref<any[]>([])
const showUpload = ref(false)
const file = ref<File | null>(null)
const pageStart = ref(1)
const pageEnd = ref(2)
const uploading = ref(false)
const uploadError = ref('')
const deletingId = ref('')
const deleteError = ref('')

function statusVariant (status: string) {
  switch (status) {
    case 'completed': return 'success'
    case 'failed': return 'danger'
    case 'extracting': case 'committing': return 'warning'
    default: return 'neutral'
  }
}
function statusLabel (status: string) {
  return t(`admin.compendium.status.${status}`, status)
}
function onFile (e: Event) {
  const input = e.target as HTMLInputElement
  file.value = input.files?.[0] ?? null
}

async function load () {
  const res: any = await $fetch('/api/admin/compendium-imports')
  jobs.value = res?.data?.items ?? res?.items ?? []
}

async function removeJob (j: { id: string; filename?: string; status?: string }) {
  if (j.status === 'extracting' || j.status === 'committing') return
  if (!confirm(t('admin.compendium.deleteConfirm', { file: j.filename || j.id }))) return
  deletingId.value = j.id
  deleteError.value = ''
  try {
    await $fetch(`/api/admin/compendium-imports/${j.id}`, { method: 'DELETE' })
    jobs.value = jobs.value.filter(row => row.id !== j.id)
  } catch (e: any) {
    deleteError.value = e?.data?.error?.message ?? e?.statusMessage ?? t('admin.compendium.deleteFailed')
  } finally {
    deletingId.value = ''
  }
}

async function upload () {
  if (!file.value) return
  uploading.value = true
  uploadError.value = ''
  try {
    const fd = new FormData()
    fd.append('file', file.value)
    fd.append('pageStart', String(pageStart.value))
    fd.append('pageEnd', String(pageEnd.value))
    const res: any = await $fetch('/api/admin/compendium-imports', { method: 'POST', body: fd })
    const id = res?.data?.id ?? res?.id
    await navigateTo(`/admin/compendium-imports/${id}`)
  } catch (e: any) {
    uploadError.value = e?.data?.error?.message ?? e?.statusMessage ?? t('admin.compendium.uploadFailed')
  } finally {
    uploading.value = false
  }
}

onMounted(() => { void load() })
</script>
