<template>
  <div data-testid="commercial-prospects-page">
    <ProPageHeader
      :title="$t('commercial.prospects.title')"
      :subtitle="$t('commercial.prospects.subtitle')"
    >
      <template #actions>
        <ProButton data-testid="commercial-prospect-new-btn" @click="showForm = !showForm">
          {{ showForm ? $t('common.cancel') : $t('commercial.prospects.create') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <ProCard v-if="showForm" class="pro-mb-lg" data-testid="commercial-prospect-form">
      <form class="pro-form" @submit.prevent="createProspect">
        <ProInput v-model="form.practiceName" data-testid="prospect-practice" :label="$t('commercial.prospects.practiceName')" required />
        <ProInput v-model="form.contactName" data-testid="prospect-contact" :label="$t('commercial.prospects.contactName')" />
        <ProInput v-model="form.contactEmail" data-testid="prospect-email" type="email" :label="$t('commercial.prospects.contactEmail')" />
        <ProInput v-model="form.contactPhone" data-testid="prospect-phone" :label="$t('commercial.prospects.contactPhone')" />
        <ProInput v-model="form.city" data-testid="prospect-city" :label="$t('commercial.prospects.city')" />
        <ProInput v-model="form.notes" data-testid="prospect-notes" :label="$t('commercial.prospects.notes')" />
        <ProButton type="submit" data-testid="prospect-submit">{{ $t('commercial.prospects.save') }}</ProButton>
      </form>
    </ProCard>

    <ProCard>
      <ProListToolbar>
        <template #filters>
          <select v-model="statusFilter" class="pro-select" data-testid="prospect-status-filter">
            <option value="">{{ $t('commercial.prospects.statusAll') }}</option>
            <option v-for="s in statuses" :key="s" :value="s">{{ $t(`commercial.prospects.status.${s}`) }}</option>
          </select>
        </template>
      </ProListToolbar>
      <ProTable :empty="!filtered.length" :empty-title="$t('commercial.prospects.empty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('commercial.prospects.contactName') }}</th>
            <th>{{ $t('commercial.prospects.statusLabel') }}</th>
            <th>{{ $t('commercial.prospects.daysInStatus') }}</th>
            <th>{{ $t('commercial.prospects.city') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in filtered" :key="p.id" :data-testid="`prospect-row-${p.id}`">
            <td>{{ p.practiceName }}</td>
            <td>{{ p.contactName }}</td>
            <td>
              <select
                class="pro-select"
                :value="p.status"
                :data-testid="`prospect-status-${p.id}`"
                @change="(e) => updateStatus(p.id, (e.target as HTMLSelectElement).value)"
              >
                <option v-for="s in statuses" :key="s" :value="s">{{ $t(`commercial.prospects.status.${s}`) }}</option>
              </select>
            </td>
            <td>{{ p.daysInStatus }}</td>
            <td>{{ p.city }}</td>
            <td>
              <ProButton variant="ghost" :data-testid="`prospect-delete-${p.id}`" @click="remove(p.id)">{{ $t('common.delete') }}</ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const statuses = ['new', 'contacted', 'qualified', 'converted', 'lost'] as const
const prospects = ref<any[]>([])
const showForm = ref(false)
const statusFilter = ref('')
const form = reactive({
  practiceName: '',
  contactName: '',
  contactEmail: '',
  contactPhone: '',
  city: '',
  notes: '',
})

const filtered = computed(() =>
  statusFilter.value ? prospects.value.filter((p) => p.status === statusFilter.value) : prospects.value,
)

async function load() {
  const res: any = await $fetch('/api/commercial/prospects')
  prospects.value = res.data ?? res ?? []
}

async function createProspect() {
  await $fetch('/api/commercial/prospects', { method: 'POST', body: { ...form } })
  showForm.value = false
  Object.assign(form, { practiceName: '', contactName: '', contactEmail: '', contactPhone: '', city: '', notes: '' })
  await load()
}

async function updateStatus(id: string, status: string) {
  await $fetch(`/api/commercial/prospects/${id}`, { method: 'PATCH', body: { status } })
  await load()
}

async function remove(id: string) {
  await $fetch(`/api/commercial/prospects/${id}`, { method: 'DELETE' })
  await load()
}

onMounted(load)
</script>
