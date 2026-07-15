<template>
  <div>
    <nav class="pro-breadcrumb" :aria-label="$t('common.breadcrumb')">
      <NuxtLink to="/clients">{{ $t('nav.clients') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <span>{{ client?.fullName || $t('clients.detail.title') }}</span>
    </nav>
    <ProPageHeader
      :title="client?.fullName || $t('clients.detail.title')"
      :subtitle="clientSubtitle"
    />
    <ProCard v-if="client" :title="$t('clients.detail.identity')">
      <p><strong>{{ client.fullName }}</strong></p>
      <p class="text-muted">{{ client.email }}</p>
      <ProBadge variant="neutral">{{ client.petCount }} {{ petLabel(client.petCount) }}</ProBadge>
    </ProCard>
    <ProCard :title="$t('clients.detail.petsTitle')">
      <ProTable
        :empty="!pets.length"
        :empty-title="$t('clients.detail.petsEmptyTitle')"
        :empty-description="$t('clients.detail.petsEmptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('clients.detail.columnName') }}</th>
            <th>{{ $t('clients.detail.columnSpecies') }}</th>
            <th>{{ $t('clients.detail.columnBreed') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in pets" :key="p.id">
            <td>{{ p.name }}</td>
            <td>{{ p.species }}</td>
            <td>{{ p.breed || $t('common.dash') }}</td>
            <td>
              <NuxtLink :to="`/clients/${clientId}/pets/${p.id}`">{{ $t('common.detail') }}</NuxtLink>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
    <div v-if="pets.length" class="pro-grid-2">
      <NuxtLink
        v-for="p in pets"
        :key="`card-${p.id}`"
        :to="`/clients/${clientId}/pets/${p.id}`"
        class="pro-pet-card-link"
      >
        <ProCard>
          <strong>{{ p.name }}</strong>
          <p class="text-muted">{{ p.species }} · {{ p.breed || $t('common.unknownBreed') }}</p>
        </ProCard>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

type ClientRow = {
  userId: string
  email: string
  fullName: string
  petCount: number
}

const { t } = useI18n()

const route = useRoute()
const clientId = route.params.clientId as string
const client = ref<ClientRow | null>(null)
const pets = ref<any[]>([])

function petLabel(count: number) {
  return count > 1 ? t('common.pets') : t('common.pet')
}

const clientSubtitle = computed(() => client.value?.email || t('clients.detail.subtitle'))

onMounted(async () => {
  try {
    const clientRes: any = await $fetch(`/api/clients/${clientId}`)
    client.value = clientRes.data ?? clientRes
  } catch {
    client.value = null
  }
  const petsRes: any = await $fetch(`/api/clients/${clientId}/pets`)
  pets.value = petsRes.data ?? petsRes ?? []
})
</script>

<style scoped>
.pro-pet-card-link {
  text-decoration: none;
  color: inherit;
}
.pro-pet-card-link:hover .pro-card {
  border-color: var(--pf-vet-accent);
}
</style>
