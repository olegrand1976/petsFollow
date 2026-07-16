<template>
  <div data-testid="requests-page">
    <ProPageHeader :title="$t('requests.title')" :subtitle="$t('requests.subtitle')" />

    <ProCard :title="$t('requests.linkTitle')" class="pro-mb-lg">
      <ProEmptyState
        v-if="!linkRequests.length"
        :title="$t('requests.linkEmptyTitle')"
        :description="$t('requests.linkEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('requests.columnClient') }}</th>
            <th>{{ $t('requests.columnEmail') }}</th>
            <th>{{ $t('requests.columnDate') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="req in linkRequests" :key="req.id" :data-testid="`link-request-${req.id}`">
            <td>{{ req.clientName }}</td>
            <td>{{ req.clientEmail }}</td>
            <td>{{ formatDate(req.createdAt) }}</td>
            <td>
              <div class="pro-flex-gap">
                <ProButton
                  :disabled="busyId === req.id"
                  @click="acceptLink(req.id)"
                >
                  {{ $t('requests.accept') }}
                </ProButton>
                <ProButton
                  variant="ghost"
                  :disabled="busyId === req.id"
                  @click="rejectLink(req.id)"
                >
                  {{ $t('requests.reject') }}
                </ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('requests.visitsTitle')">
      <ProEmptyState
        v-if="!visitRequests.length"
        :title="$t('requests.visitsEmptyTitle')"
        :description="$t('requests.visitsEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('requests.columnClient') }}</th>
            <th>{{ $t('requests.columnPet') }}</th>
            <th>{{ $t('requests.columnNotes') }}</th>
            <th>{{ $t('requests.columnDate') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="v in visitRequests" :key="v.id" :data-testid="`visit-request-${v.id}`">
            <td>
              <NuxtLink v-if="v.clientId" :to="`/clients/${v.clientId}`">{{ v.clientName }}</NuxtLink>
              <span v-else>{{ v.clientName }}</span>
            </td>
            <td>
              <NuxtLink
                v-if="v.clientId && v.petId"
                :to="`/clients/${v.clientId}/pets/${v.petId}`"
              >
                {{ v.petName }}
              </NuxtLink>
              <span v-else>{{ v.petName }}</span>
            </td>
            <td>{{ v.notes || '—' }}</td>
            <td>{{ formatDate(v.scheduledAt || v.createdAt) }}</td>
            <td>
              <div class="pro-flex-gap">
                <ProButton :disabled="busyId === v.id" @click="setVisitStatus(v.id, 'confirmed')">
                  {{ $t('requests.confirm') }}
                </ProButton>
                <ProButton variant="ghost" :disabled="busyId === v.id" @click="setVisitStatus(v.id, 'cancelled')">
                  {{ $t('requests.cancel') }}
                </ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const { formatDate } = useFormatters()

const linkRequests = ref<any[]>([])
const visitRequests = ref<any[]>([])
const busyId = ref('')

async function load() {
  const [linksRes, visitsRes]: any[] = await Promise.all([
    $fetch('/api/vet/link-requests'),
    $fetch('/api/vet/visits?status=requested'),
  ])
  linkRequests.value = linksRes.data ?? linksRes ?? []
  visitRequests.value = visitsRes.data ?? visitsRes ?? []
}

async function acceptLink(id: string) {
  busyId.value = id
  try {
    await $fetch(`/api/vet/link-requests/${id}/accept`, { method: 'POST' })
    await load()
  } finally {
    busyId.value = ''
  }
}

async function rejectLink(id: string) {
  busyId.value = id
  try {
    await $fetch(`/api/vet/link-requests/${id}/reject`, { method: 'POST' })
    await load()
  } finally {
    busyId.value = ''
  }
}

async function setVisitStatus(id: string, status: string) {
  busyId.value = id
  try {
    await $fetch(`/api/visits/${id}`, { method: 'PATCH', body: { status } })
    await load()
  } finally {
    busyId.value = ''
  }
}

onMounted(load)
</script>
