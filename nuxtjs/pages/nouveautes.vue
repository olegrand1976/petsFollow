<template>
  <div data-testid="nouveautes-page">
    <ProPageHeader
      :title="$t('nouveautes.title')"
      :subtitle="$t('nouveautes.subtitle')"
    />

    <p v-if="pending" class="pro-hint" data-testid="nouveautes-loading">{{ $t('nouveautes.loading') }}</p>

    <ProEmptyState
      v-else-if="!items.length && !error"
      data-testid="nouveautes-empty"
      :title="$t('nouveautes.emptyTitle')"
      :description="$t('nouveautes.emptyBody')"
    />

    <div v-else-if="items.length" class="nouveautes-list" data-testid="nouveautes-list">
      <ProCard
        v-for="item in items"
        :key="item.date"
        class="pro-mb-md nouveautes-card"
        :data-testid="`nouveautes-item-${item.date}`"
      >
        <div class="nouveautes-card__meta">
          <time class="nouveautes-card__date" :datetime="item.date">{{ formatDate(item.date) }}</time>
          <ProBadge v-if="item.branch" variant="neutral">{{ item.branch }}</ProBadge>
        </div>
        <h2 v-if="item.headline" class="nouveautes-card__headline">{{ item.headline }}</h2>
        <pre v-if="item.body" class="nouveautes-card__body">{{ item.body }}</pre>
      </ProCard>
    </div>

    <p v-if="error" class="pro-error" data-testid="nouveautes-error">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
// Page partagée entre les 4 shells (vet / admin / commercial / manager) :
// le middleware shared-pro-layout pose le layout selon le rôle, sinon retombe sur `default`.
definePageMeta({
  middleware: ['shared-pro-layout'],
})

type DigestItem = {
  date: string
  headline: string
  body: string
  status: string
  branch?: string
}

const { t, locale } = useI18n()

const pending = ref(true)
const error = ref('')
const items = ref<DigestItem[]>([])

function formatDate(iso: string): string {
  const d = new Date(`${iso}T12:00:00`)
  if (Number.isNaN(d.getTime())) return iso
  try {
    return new Intl.DateTimeFormat(locale.value || 'fr', {
      weekday: 'long',
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    }).format(d)
  } catch {
    return iso
  }
}

onMounted(async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await $fetch<{ data?: { items?: DigestItem[] }, items?: DigestItem[] }>('/api/product-digests', {
      query: { limit: 30 },
    })
    const payload = res?.data ?? res
    items.value = (payload as { items?: DigestItem[] })?.items ?? []
  } catch {
    error.value = t('nouveautes.loadError')
  } finally {
    pending.value = false
  }
})
</script>

<style scoped>
.nouveautes-card__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
  margin-bottom: 0.5rem;
}
.nouveautes-card__date {
  font-size: 0.875rem;
  color: var(--pf-vet-muted, #5b6b7c);
  text-transform: capitalize;
}
.nouveautes-card__headline {
  margin: 0 0 0.75rem;
  font-size: 1.125rem;
  font-weight: 650;
  color: var(--pf-vet-primary);
}
.nouveautes-card__body {
  margin: 0;
  white-space: pre-wrap;
  font-family: inherit;
  font-size: 0.95rem;
  line-height: 1.55;
  color: var(--pf-vet-text, #1a2332);
}
</style>
