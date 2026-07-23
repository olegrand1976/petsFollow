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
    >
      <template #actions>
        <ProButton
          v-if="client"
          :disabled="sendingAppLink"
          :loading="sendingAppLink"
          data-testid="send-app-link"
          @click="sendAppLink"
        >
          <ProIcon name="smartphone" />
          {{ $t('clients.detail.sendAppLink') }}
        </ProButton>
      </template>
    </ProPageHeader>
    <p v-if="appLinkFeedback" class="pro-inline-feedback" role="status">{{ appLinkFeedback }}</p>
    <ProCard v-if="client" :title="$t('clients.detail.identity')">
      <p><strong>{{ client.fullName }}</strong></p>
      <p class="text-muted">{{ client.email }}</p>
      <ProBadge variant="neutral">{{ client.petCount }} {{ petLabel(client.petCount) }}</ProBadge>
      <div class="pro-mt-md">
        <ProButton
          variant="secondary"
          :disabled="sendingAppLink"
          :loading="sendingAppLink"
          @click="sendAppLink"
        >
          <ProIcon name="mail" />
          {{ $t('clients.detail.sendAppLink') }}
        </ProButton>
        <p class="text-muted pro-hint">{{ $t('clients.detail.sendAppLinkHint') }}</p>
      </div>
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
const sendingAppLink = ref(false)
const appLinkFeedback = ref('')

function petLabel(count: number) {
  return count > 1 ? t('common.pets') : t('common.pet')
}

const clientSubtitle = computed(() => client.value?.email || t('clients.detail.subtitle'))

async function sendAppLink() {
  if (!client.value || sendingAppLink.value) return
  sendingAppLink.value = true
  appLinkFeedback.value = ''
  try {
    const res: any = await $fetch(`/api/clients/${clientId}/send-app-link`, { method: 'POST' })
    const data = res.data ?? res
    appLinkFeedback.value = data.message || t('clients.detail.sendAppLinkSuccess', { email: client.value.email })
  } catch {
    appLinkFeedback.value = t('clients.detail.sendAppLinkError')
  } finally {
    sendingAppLink.value = false
  }
}

onMounted(async () => {
  try {
    const clientRes: any = await $fetch(`/api/clients/${clientId}`)
    client.value = clientRes.data ?? clientRes
  } catch {
    client.value = null
  }
  try {
    const petsRes: any = await $fetch(`/api/clients/${clientId}/pets`)
    pets.value = petsRes.data ?? petsRes ?? []
  } catch {
    pets.value = []
  }
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
.pro-hint {
  margin-top: 0.5rem;
  font-size: 0.875rem;
}
.pro-inline-feedback {
  margin: 0 0 1rem;
  padding: 0.75rem 1rem;
  border-radius: var(--pf-vet-radius);
  background: color-mix(in srgb, var(--pf-vet-accent) 10%, var(--pf-vet-surface));
  border: 1px solid color-mix(in srgb, var(--pf-vet-accent) 30%, transparent);
}
.pro-mt-md {
  margin-top: 1rem;
}
</style>
