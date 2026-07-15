<template>
  <div>
    <nav class="pro-breadcrumb" aria-label="Fil d'Ariane">
      <NuxtLink to="/clients">Clients</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <span>{{ client?.fullName || 'Fiche client' }}</span>
    </nav>
    <ProPageHeader
      :title="client?.fullName || 'Fiche client'"
      :subtitle="clientSubtitle"
    />
    <ProCard v-if="client" title="Identité">
      <p><strong>{{ client.fullName }}</strong></p>
      <p class="text-muted">{{ client.email }}</p>
      <ProBadge variant="neutral">{{ client.petCount }} {{ client.petCount > 1 ? 'animaux' : 'animal' }}</ProBadge>
    </ProCard>
    <ProCard title="Animaux">
      <ProTable
        :empty="!pets.length"
        empty-title="Aucun animal"
        empty-description="Ce client n'a pas encore enregistré d'animal."
      >
        <thead>
          <tr>
            <th>Nom</th>
            <th>Espèce</th>
            <th>Race</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in pets" :key="p.id">
            <td>{{ p.name }}</td>
            <td>{{ p.species }}</td>
            <td>{{ p.breed || '—' }}</td>
            <td>
              <NuxtLink :to="`/clients/${clientId}/pets/${p.id}`">Détail</NuxtLink>
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
          <p class="text-muted">{{ p.species }} · {{ p.breed || 'Race inconnue' }}</p>
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

const route = useRoute()
const clientId = route.params.clientId as string
const client = ref<ClientRow | null>(null)
const pets = ref<any[]>([])

const clientSubtitle = computed(() => client.value?.email || 'Animaux et dossier du client.')

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
