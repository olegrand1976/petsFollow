<template>
  <div data-testid="commercial-pitch-page">
    <ProPageHeader
      :title="$t('commercial.pitch.title')"
      :subtitle="$t('commercial.pitch.subtitle')"
    />

    <div v-if="!audience" class="pf-audience-grid" data-testid="commercial-pitch-cards">
      <button
        type="button"
        class="pro-card pro-card--interactive pf-audience-card"
        data-testid="pitch-card-vet"
        @click="audience = 'vet'"
      >
        <ProIcon name="medical_services" :size="36" />
        <strong>{{ $t('commercial.pitch.cardVetTitle') }}</strong>
        <span>{{ $t('commercial.pitch.cardVetDesc') }}</span>
      </button>
      <button
        type="button"
        class="pro-card pro-card--interactive pf-audience-card"
        data-testid="pitch-card-client"
        @click="audience = 'client'"
      >
        <ProIcon name="person" :size="36" />
        <strong>{{ $t('commercial.pitch.cardClientTitle') }}</strong>
        <span>{{ $t('commercial.pitch.cardClientDesc') }}</span>
      </button>
    </div>

    <template v-else>
      <div class="pf-form-toolbar pro-mb-lg">
        <h2 class="pf-pitch-heading">
          {{ audience === 'vet' ? $t('commercial.pitch.cardVetTitle') : $t('commercial.pitch.cardClientTitle') }}
        </h2>
        <ProButton variant="ghost" test-id="pitch-back-cards" @click="audience = null">
          {{ $t('commercial.pitch.backToCards') }}
        </ProButton>
      </div>

      <ProCard :title="$t('commercial.pitch.benefitsTitle')" class="pro-mb-lg">
        <p class="pro-hint">{{ $t('products.includedLead') }}</p>
        <ul class="pro-feature-list">
          <li v-for="item in includedBenefits" :key="item">{{ item }}</li>
        </ul>
        <ul v-if="audienceFeatures.length" class="pro-feature-list pro-mt-md">
          <li v-for="item in audienceFeatures" :key="item">{{ item }}</li>
        </ul>
      </ProCard>

      <ProCard :title="$t('products.plansTitle')" class="pro-mb-lg">
        <p class="pro-hint">{{ $t('products.plansLead') }}</p>
        <div class="pf-plan-grid">
          <div
            v-for="code in planCodes"
            :key="code"
            class="pf-plan-item"
            :class="{ 'pf-plan-item--rec': code === 'triennial' }"
          >
            <strong>{{ $t(`products.plans.${code}.name`) }}</strong>
            <span class="pf-plan-price">{{ $t(`products.plans.${code}.price`) }}</span>
            <span class="text-muted">{{ $t(`products.plans.${code}.period`) }}</span>
            <ProBadge v-if="code === 'triennial'" variant="success">{{ $t('products.recommended') }}</ProBadge>
            <ul>
              <li v-for="b in planBenefits(code)" :key="b">{{ b }}</li>
            </ul>
          </div>
        </div>
        <p class="pro-hint">{{ $t('products.plansTip') }}</p>
      </ProCard>

      <ProCard :title="$t('products.addonsTitle')" class="pro-mb-lg">
        <p class="pro-hint">{{ $t('products.addonsLead') }}</p>
        <div class="pf-plan-grid">
          <div v-for="code in addonCodes" :key="code" class="pf-plan-item">
            <strong>{{ $t(`products.addons.${code}.name`) }}</strong>
            <span class="pf-plan-price">{{ $t(`products.addons.${code}.price`) }}</span>
            <span class="text-muted">{{ $t(`products.addons.${code}.tagline`) }}</span>
          </div>
        </div>
      </ProCard>

      <ProCard
        v-if="audience === 'vet'"
        :title="$t('commercial.pitch.vetRetributionTitle')"
        class="pro-mb-lg"
        data-testid="pitch-vet-retribution"
      >
        <ProCommissionSheet
          audience="vet"
          :plan-rates="planRates"
          :addon-rates="addonRates"
          :bonuses="vetBonuses"
        />
      </ProCard>

      <NuxtLink to="/produits">
        <ProButton variant="secondary">{{ $t('commercial.pitch.productsLink') }}</ProButton>
      </NuxtLink>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { tm, rt } = useI18n()

const audience = ref<null | 'vet' | 'client'>(null)
const planRates = ref<any[]>([])
const addonRates = ref<any[]>([])
const bonuses = ref<any[]>([])

const planCodes = ['annual', 'triennial', 'quinquennial'] as const
const addonCodes = ['family', 'carePlus', 'kennel', 'horse'] as const

function mapList(key: string): string[] {
  const raw = tm(key) as string[]
  return Array.isArray(raw) ? raw.map((x) => (typeof x === 'string' ? x : rt(x as any))) : []
}

const includedBenefits = computed(() => mapList('products.included'))
const audienceFeatures = computed(() =>
  audience.value === 'vet' ? mapList('commercial.pitch.webFeatures') : mapList('commercial.pitch.mobileFeatures'),
)
const vetBonuses = computed(() => (bonuses.value || []).filter((b: any) => b.audience === 'vet'))

function planBenefits(code: string): string[] {
  return mapList(`products.plans.${code}.benefits`)
}

onMounted(async () => {
  const res: any = await $fetch('/api/commercial/commissions')
  const data = res.data ?? res
  planRates.value = data.planRates ?? []
  addonRates.value = data.addonRates ?? []
  bonuses.value = data.bonuses ?? []
})
</script>

<style scoped>
.pf-audience-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.pf-audience-card {
  display: grid;
  gap: 0.5rem;
  text-align: left;
  cursor: pointer;
  width: 100%;
  font: inherit;
  color: inherit;
}
.pf-audience-card strong { font-size: 1.05rem; }
.pf-audience-card span {
  color: var(--pf-vet-muted, #64748b);
  font-size: 0.9rem;
}
.pf-form-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}
.pf-pitch-heading {
  margin: 0;
  font-size: 1.15rem;
}
.pro-feature-list {
  margin: 0.5rem 0 0;
  padding-left: 1.25rem;
  display: grid;
  gap: 0.45rem;
}
.pro-mb-lg { margin-bottom: 1.25rem; }
.pro-mt-md { margin-top: 1rem; }
.pf-plan-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.75rem;
  margin: 0.75rem 0;
}
.pf-plan-item {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  padding: 0.75rem;
  display: grid;
  gap: 0.35rem;
}
.pf-plan-item--rec { border-color: var(--pf-vet-accent); }
.pf-plan-price {
  font-size: 1.25rem;
  font-weight: 700;
}
.pf-plan-item ul {
  margin: 0.25rem 0 0;
  padding-left: 1.1rem;
  display: grid;
  gap: 0.25rem;
  font-size: 0.875rem;
}
@media (max-width: 700px) {
  .pf-audience-grid { grid-template-columns: 1fr; }
}
</style>
