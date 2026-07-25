<template>
  <div data-testid="commercial-competition-page">
    <ProPageHeader :title="$t('competition.title')" :subtitle="$t('competition.subtitle')" />

    <ProCard class="pro-mb-lg competition-banner">
      <h3>{{ banner.title }}</h3>
      <p>{{ banner.body }}</p>
    </ProCard>

    <div class="competition-tabs pro-mb-lg" role="tablist">
      <button
        v-for="m in marketKeys"
        :key="m"
        type="button"
        class="competition-tab"
        :class="{ active: market === m }"
        :data-testid="`competition-tab-${m}`"
        @click="market = m"
      >
        {{ $t(`competition.market.${m}`) }}
      </button>
    </div>

    <ProCard class="pro-mb-lg" data-testid="competition-synthesis">
      <h3>{{ loc(current.synthesis.headline) }}</h3>
      <p><strong>{{ $t('competition.opportunities') }}</strong> — {{ loc(current.synthesis.opportunities) }}</p>
      <p><strong>{{ $t('competition.threats') }}</strong> — {{ loc(current.synthesis.threats) }}</p>
      <p><strong>{{ $t('competition.angle') }}</strong> — {{ loc(current.synthesis.petsFollowAngle) }}</p>
    </ProCard>

    <div class="competition-grid">
      <ProCard v-for="p in current.products" :key="p.id" class="competition-card">
        <div class="competition-card-head">
          <h3>{{ p.name }}</h3>
          <span class="pro-hint">{{ p.vendor }} · {{ p.category }}</span>
        </div>
        <div class="competition-badges">
          <ProBadge>{{ presenceLabel(p.presenceLevel) }}</ProBadge>
          <ProBadge>{{ $t('competition.since', { year: p.foundedYear, years: yearsActive(p.foundedYear) }) }}</ProBadge>
        </div>
        <p class="pro-hint">{{ loc(p.presenceNote) }}</p>
        <h4>{{ $t('competition.features') }}</h4>
        <ul>
          <li v-for="(f, i) in p.features" :key="i">{{ loc(f) }}</li>
        </ul>
        <h4>{{ $t('competition.pricing') }}</h4>
        <p>{{ p.pricing.model }} — {{ p.pricing.rangeHint }} ({{ p.pricing.currency }})</p>
        <p class="pro-hint">{{ loc(p.pricing.notes) }}</p>
        <h4>{{ $t('competition.positioning') }}</h4>
        <p>{{ loc(p.positioning.functional) }}</p>
        <p class="pro-hint">{{ loc(p.positioning.pricing) }}</p>
        <div class="competition-pf">
          <strong>petsFollow</strong>
          <p>{{ loc(p.petsFollowCompare.wins) }}</p>
          <p class="pro-hint">{{ loc(p.petsFollowCompare.watchouts) }}</p>
          <p><em>{{ loc(p.petsFollowCompare.talkTrack) }}</em></p>
        </div>
      </ProCard>
    </div>

    <p class="pro-hint pro-mt-lg">
      {{ $t('competition.reviewed', { date: data.lastReviewedAt }) }}
      — {{ loc(data.analystNote) }}
    </p>
  </div>
</template>

<script setup lang="ts">
import marketsData from '~/data/competition/markets.json'

definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

type Loc = { fr: string; en: string; es?: string; nl?: string; et?: string }

const { locale, t } = useI18n()
const data = marketsData as typeof marketsData
const marketKeys = ['fr', 'be', 'es'] as const
const market = ref<(typeof marketKeys)[number]>('fr')

const current = computed(() => data.markets[market.value])

const banner = computed(() => {
  const b = data.petsFollowBanner as Record<string, { title: string; body: string }>
  return b[locale.value] ?? b.fr ?? b.en
})

function loc(v: Loc | string | undefined): string {
  if (!v) return ''
  if (typeof v === 'string') return v
  const l = locale.value as keyof Loc
  return v[l] ?? v.fr ?? v.en ?? ''
}

function yearsActive(year: number) {
  return Math.max(0, new Date().getFullYear() - year)
}

function presenceLabel(level: string) {
  return t(`competition.presence.${level}`)
}
</script>

<style scoped>
.competition-tabs {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.competition-tab {
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface);
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  cursor: pointer;
}
.competition-tab.active {
  border-color: var(--pf-vet-accent);
  color: var(--pf-vet-accent);
  font-weight: 600;
}
.competition-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}
.competition-card-head h3 {
  margin: 0;
}
.competition-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin: 0.5rem 0;
}
.competition-pf {
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--pf-vet-border);
}
.competition-banner h3 {
  margin-top: 0;
}
</style>
