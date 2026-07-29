<template>
  <div data-testid="admin-client-imports-page">
    <ProPageHeader
      :title="$t('admin.clientImport.title')"
      :subtitle="$t('admin.clientImport.subtitle')"
    >
      <template #actions>
        <ProButton test-id="admin-import-new" @click="showUpload = true">
          {{ $t('admin.clientImport.newImport') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <ProCard v-if="showUpload" class="pro-mb-lg" data-testid="admin-import-upload-card">
      <h3 class="pro-mb-md">{{ $t('admin.clientImport.uploadTitle') }}</h3>
      <form class="pro-form" @submit.prevent="upload">
        <div class="pro-field">
          <label class="pro-label" for="import-vet">{{ $t('admin.clientImport.vet') }}</label>
          <select id="import-vet" v-model="vetUserId" class="pro-select" required data-testid="admin-import-vet">
            <option value="">{{ $t('admin.clientImport.vetPlaceholder') }}</option>
            <option v-for="v in vetOptions" :key="v.userId" :value="v.userId">
              {{ v.fullName }} — {{ v.practiceName }}
            </option>
          </select>
        </div>
        <div class="pro-field">
          <label class="pro-label" for="import-file">{{ $t('admin.clientImport.file') }}</label>
          <input
            id="import-file"
            type="file"
            accept=".csv,.xls,.xlsx,text/csv,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
            class="pro-input"
            data-testid="admin-import-file"
            required
            @change="onFile"
          >
        </div>
        <p v-if="uploadError" class="pro-hint pro-hint--error" data-testid="admin-import-upload-error">{{ uploadError }}</p>
        <div class="pro-flex-gap">
          <ProButton type="submit" test-id="admin-import-upload" :disabled="uploading || !file || !vetUserId">
            {{ $t('admin.clientImport.uploadSubmit') }}
          </ProButton>
          <ProButton variant="ghost" type="button" @click="showUpload = false">{{ $t('common.cancel') }}</ProButton>
        </div>
      </form>
    </ProCard>

    <ProCard>
      <ProTable :empty="!jobs.length" :empty-title="$t('admin.clientImport.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.clientImport.colDate') }}</th>
            <th>{{ $t('admin.clientImport.colFile') }}</th>
            <th>{{ $t('admin.clientImport.colVet') }}</th>
            <th>{{ $t('admin.clientImport.colStatus') }}</th>
            <th>{{ $t('admin.clientImport.colRows') }}</th>
            <th>{{ $t('admin.clientImport.colCreated') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="j in jobs" :key="j.id" data-testid="admin-import-row">
            <td>{{ j.createdAt?.substring(0, 16)?.replace('T', ' ') }}</td>
            <td>{{ j.filename }}</td>
            <td>{{ j.vetFullName }} — {{ j.practiceName }}</td>
            <td><ProBadge :variant="statusVariant(j.status)">{{ statusLabel(j.status) }}</ProBadge></td>
            <td>{{ j.rowCount }}</td>
            <td>{{ j.createdCount }}</td>
            <td>
              <NuxtLink :to="`/admin/client-imports/${j.id}`" class="pro-link">
                {{ $t('admin.clientImport.open') }}
              </NuxtLink>
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

type VetOption = { userId: string, fullName: string, email: string, practiceName: string }
type ImportJob = {
  id: string
  filename: string
  status: string
  rowCount: number
  createdCount: number
  createdAt: string
  vetFullName?: string
  practiceName?: string
}

const jobs = ref<ImportJob[]>([])
const vetOptions = ref<VetOption[]>([])
const showUpload = ref(false)
const vetUserId = ref('')
const file = ref<File | null>(null)
const uploading = ref(false)
const uploadError = ref('')

function onFile(e: Event) {
  const input = e.target as HTMLInputElement
  file.value = input.files?.[0] ?? null
}

function statusLabel(status: string) {
  return t(`admin.clientImport.status.${status}`, status)
}

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'completed':
      return 'success'
    case 'failed':
      return 'danger'
    case 'importing':
    case 'preview_ready':
    case 'mapping_ready':
    case 'uploaded':
      return 'warning'
    default:
      return 'neutral'
  }
}

async function load() {
  const [listRes, vetsRes] = await Promise.all([
    $fetch<any>('/api/admin/client-imports'),
    $fetch<any>('/api/admin/vets'),
  ])
  jobs.value = listRes?.data?.items ?? listRes?.items ?? []
  vetOptions.value = vetsRes?.data ?? vetsRes ?? []
}

async function upload() {
  if (!file.value || !vetUserId.value) return
  uploading.value = true
  uploadError.value = ''
  try {
    const fd = new FormData()
    fd.append('file', file.value)
    fd.append('vetUserId', vetUserId.value)
    const res: any = await $fetch('/api/admin/client-imports', { method: 'POST', body: fd })
    const detail = res?.data ?? res
    const id = detail?.job?.id ?? detail?.id
    if (!id) throw new Error('invalid_response')
    await navigateTo(`/admin/client-imports/${id}`)
  } catch (e: any) {
    uploadError.value = e?.data?.error?.message ?? e?.statusMessage ?? t('admin.clientImport.uploadFailed')
  } finally {
    uploading.value = false
  }
}

onMounted(() => load())
</script>
