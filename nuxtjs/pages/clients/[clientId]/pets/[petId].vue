<template>
  <div>
    <nav class="pro-breadcrumb" aria-label="Fil d'Ariane">
      <NuxtLink to="/clients">Clients</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <NuxtLink :to="`/clients/${clientId}`">Fiche</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <span>{{ pet?.name || 'Animal' }}</span>
    </nav>
    <ProPageHeader
      :title="pet?.name || 'Animal'"
      :subtitle="petSubtitle"
    />
    <ProCard title="Relevés cardiaques validés">
      <ProTable
        :empty="!sessions.length"
        empty-title="Aucun relevé validé"
        empty-description="Les sessions BPM validées apparaîtront ici."
      >
        <thead>
          <tr>
            <th>Date</th>
            <th>BPM</th>
            <th>Statut</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in sessions" :key="s.id">
            <td>{{ formatDate(s.startedAt) }}</td>
            <td><code>{{ s.bpm }}</code></td>
            <td>
              <ProBadge :variant="s.isAlert ? 'danger' : 'success'">
                {{ s.isAlert ? 'Alerte' : 'OK' }}
              </ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
    <ProCard title="Historique suivi">
      <ul v-if="timeline.length" class="pro-timeline">
        <li v-for="item in timeline" :key="item.id" class="pro-timeline__item">
          <div class="pro-timeline__dot" aria-hidden="true" />
          <div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.body }}</p>
            <small class="text-muted">{{ formatDate(item.createdAt) }}</small>
          </div>
        </li>
      </ul>
      <ProEmptyState
        v-else
        title="Historique vide"
        description="Aucun événement de suivi pour cet animal."
      />
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const route = useRoute()
const clientId = route.params.clientId as string
const petId = route.params.petId as string
const pet = ref<any>(null)
const sessions = ref<any[]>([])
const timeline = ref<any[]>([])

const petSubtitle = computed(() => {
  if (!pet.value) return ''
  return [pet.value.species, pet.value.breed].filter(Boolean).join(' · ')
})

function formatDate(value: string) {
  return new Date(value).toLocaleString('fr-FR')
}

onMounted(async () => {
  const petRes: any = await $fetch(`/api/pets/${petId}`)
  pet.value = petRes.data ?? petRes

  const sessionsRes: any = await $fetch(`/api/pets/${petId}/heartrate`)
  sessions.value = sessionsRes.data ?? sessionsRes ?? []

  const timelineRes: any = await $fetch(`/api/pets/${petId}/timeline`)
  timeline.value = timelineRes.data ?? timelineRes ?? []
})
</script>

<style scoped>
.pro-timeline {
  list-style: none;
  margin: 0;
  padding: 0;
}

.pro-timeline__item {
  display: grid;
  grid-template-columns: 1rem 1fr;
  gap: 0.75rem 1rem;
  padding-bottom: 1.25rem;
  border-left: 2px solid var(--pf-vet-border);
  margin-left: 0.35rem;
  padding-left: 1.25rem;
  position: relative;
}

.pro-timeline__item:last-child {
  border-left-color: transparent;
  padding-bottom: 0;
}

.pro-timeline__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--pf-vet-accent);
  position: absolute;
  left: -6px;
  top: 0.35rem;
}

.pro-timeline__item p {
  margin: 0.25rem 0;
}
</style>
