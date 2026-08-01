<template>
  <div data-testid="admin-afmps-imports-page">
    <ProPageHeader
      :title="$t('admin.afmps.title')"
      :subtitle="$t('admin.afmps.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="afmps-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton test-id="admin-afmps-new" @click="showUpload = true">
          {{ $t('admin.afmps.newImport') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <ProCard v-if="showUpload" class="pro-mb-lg" data-testid="admin-afmps-upload-card">
      <h3 class="pro-mb-md">{{ $t('admin.afmps.uploadTitle') }}</h3>
      <form class="pro-form" @submit.prevent="upload">
        <div class="pro-field">
          <label class="pro-label" for="afmps-file">{{ $t('admin.afmps.file') }}</label>
          <input
            id="afmps-file"
            type="file"
            accept=".csv,text/csv"
            class="pro-input"
            data-testid="admin-afmps-file"
            required
            @change="onFile"
          >
        </div>
        <p v-if="uploadError" class="pro-hint pro-hint--error" data-testid="admin-afmps-upload-error">{{ uploadError }}</p>
        <div class="pro-flex-gap">
          <ProButton type="submit" test-id="admin-afmps-upload" :disabled="uploading || !file">
            {{ $t('admin.afmps.uploadSubmit') }}
          </ProButton>
          <ProButton variant="ghost" type="button" @click="showUpload = false">{{ $t('common.cancel') }}</ProButton>
        </div>
      </form>
    </ProCard>

    <p v-if="deleteError" class="pro-hint pro-hint--error pro-mb-md">{{ deleteError }}</p>
    <ProCard>
      <ProTable :empty="!jobs.length" :empty-title="$t('admin.afmps.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.afmps.colDate') }}</th>
            <th>{{ $t('admin.afmps.colFile') }}</th>
            <th>{{ $t('admin.afmps.colStatus') }}</th>
            <th>{{ $t('admin.afmps.colReady') }}</th>
            <th>{{ $t('admin.afmps.colPreview') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="j in jobs" :key="j.id" data-testid="admin-afmps-row">
            <td>{{ j.createdAt?.substring(0, 16)?.replace('T', ' ') }}</td>
            <td>{{ j.filename }}</td>
            <td><ProBadge :variant="statusVariant(j.status)">{{ statusLabel(j.status) }}</ProBadge></td>
            <td>{{ j.readyCount }}</td>
            <td>{{ j.insertCount }} / {{ j.updateCount }}</td>
            <td class="pro-flex-gap">
              <NuxtLink :to="`/admin/afmps-imports/${j.id}`" class="pro-link">
                {{ $t('admin.afmps.open') }}
              </NuxtLink>
              <ProButton
                variant="ghost"
                test-id="admin-afmps-delete"
                :disabled="deletingId === j.id || j.status === 'committing'"
                @click="removeJob(j)"
              >
                {{ $t('admin.afmps.delete') }}
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
const uploading = ref(false)
const uploadError = ref('')
const deletingId = ref('')
const deleteError = ref('')

function statusVariant (status: string) {
  switch (status) {
    case 'completed': return 'success'
    case 'failed': case 'blocked': return 'danger'
    case 'committing': case 'reviewed': return 'warning'
    default: return 'neutral'
  }
}
function statusLabel (status: string) {
  return t(`admin.afmps.status.${status}`, status)
}
function onFile (e: Event) {
  const input = e.target as HTMLInputElement
  file.value = input.files?.[0] ?? null
}

async function load () {
  const res: any = await $fetch('/api/admin/afmps-imports')
  jobs.value = res?.data?.items ?? res?.items ?? []
}

async function removeJob (j: { id: string; filename?: string; status?: string }) {
  if (j.status === 'committing') return
  if (!confirm(t('admin.afmps.deleteConfirm', { file: j.filename || j.id }))) return
  deletingId.value = j.id
  deleteError.value = ''
  try {
    await $fetch(`/api/admin/afmps-imports/${j.id}`, { method: 'DELETE' })
    jobs.value = jobs.value.filter(row => row.id !== j.id)
  } catch (e: any) {
    deleteError.value = e?.data?.error?.message ?? e?.statusMessage ?? t('admin.afmps.deleteFailed')
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
    const res: any = await $fetch('/api/admin/afmps-imports', { method: 'POST', body: fd })
    const id = res?.data?.job?.id ?? res?.job?.id ?? res?.data?.id
    if (id) {
      await navigateTo(`/admin/afmps-imports/${id}`)
      return
    }
    uploadError.value = t('admin.afmps.uploadFailed')
  } catch (e: any) {
    // 422 still returns job id in body for blocked gate1
    const id = e?.data?.data?.job?.id ?? e?.data?.job?.id
    if (id) {
      await navigateTo(`/admin/afmps-imports/${id}`)
      return
    }
    uploadError.value = e?.data?.error?.message ?? e?.statusMessage ?? t('admin.afmps.uploadFailed')
  } finally {
    uploading.value = false
  }
}

onMounted(() => { void load() })
</script>
