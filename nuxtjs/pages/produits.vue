<template>
  <div data-testid="products-page" class="pf-products-pro">
    <ProPageHeader
      :title="$t('products.title')"
      :subtitle="$t('products.lead')"
    >
      <template #actions>
        <ProButton variant="secondary" class="no-print" test-id="products-print" @click="printPage">
          {{ $t('products.print') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p class="pf-products-pro__positioning">{{ $t('products.positioning') }}</p>

    <!-- Deux solutions Pro -->
    <section class="pf-products-pro__solutions pro-mb-lg" aria-labelledby="products-solutions-title">
      <h2 id="products-solutions-title" class="pf-products-pro__section-title">
        {{ $t('products.solutionsTitle') }}
      </h2>
      <p class="pro-hint">{{ $t('products.solutionsLead') }}</p>
      <div class="pf-products-pro__solutions-grid">
        <article
          v-for="sol in solutions"
          :key="sol.key"
          class="pf-products-pro__solution"
          :class="{ 'pf-products-pro__solution--featured': sol.featured }"
          :data-testid="`products-solution-${sol.key}`"
        >
          <div class="pf-products-pro__solution-top">
            <h3>{{ sol.name }}</h3>
            <ProBadge :variant="sol.featured ? 'success' : 'neutral'">{{ sol.badge }}</ProBadge>
          </div>
          <p class="pf-products-pro__price">{{ sol.price }}</p>
          <p v-if="sol.setup" class="pf-products-pro__price-sub">{{ sol.setup }}</p>
          <p class="pf-products-pro__solution-tagline">{{ sol.tagline }}</p>
          <ul class="pro-feature-list">
            <li v-for="f in sol.features" :key="f">{{ f }}</li>
          </ul>
        </article>
      </div>
      <p class="pf-products-pro__tip">{{ $t('products.saasInvoiceNote') }}</p>
    </section>

    <!-- Web + Mobile -->
    <div class="pf-products-pro__duo pro-mb-lg">
      <ProCard :title="$t('products.webTitle')">
        <p class="pro-hint">{{ $t('products.webLead') }}</p>
        <ul class="pro-feature-list">
          <li v-for="item in webFeatures" :key="item">{{ item }}</li>
        </ul>
      </ProCard>
      <ProCard :title="$t('products.mobileTitle')">
        <p class="pro-hint">{{ $t('products.mobileLead') }}</p>
        <ul class="pro-feature-list">
          <li v-for="item in mobileFeatures" :key="item">{{ item }}</li>
        </ul>
      </ProCard>
    </div>

    <!-- Bénéfices -->
    <ProCard :title="$t('products.benefitsTitle')" class="pro-mb-lg">
      <div class="pf-products-pro__benefits">
        <div>
          <h3 class="pf-products-pro__benefits-h">{{ $t('products.benefitsCabinetTitle') }}</h3>
          <ul class="pro-feature-list">
            <li v-for="item in benefitsCabinet" :key="item">{{ item }}</li>
          </ul>
        </div>
        <div>
          <h3 class="pf-products-pro__benefits-h">{{ $t('products.benefitsClientsTitle') }}</h3>
          <ul class="pro-feature-list">
            <li v-for="item in benefitsClients" :key="item">{{ item }}</li>
          </ul>
        </div>
      </div>
    </ProCard>

    <!-- Modèle partenaire -->
    <ProCard :title="$t('products.partnerTitle')" class="pro-mb-lg">
      <p class="pro-hint">{{ $t('products.partnerLead') }}</p>
      <ol class="pf-products-pro__steps">
        <li v-for="step in partnerSteps" :key="step">{{ step }}</li>
      </ol>
      <p class="pf-products-pro__tip">{{ $t('products.partnerTip') }}</p>
    </ProCard>

    <!-- Tarifs SaaS Pro -->
    <ProCard :title="$t('products.saasTitle')" class="pro-mb-lg">
      <p class="pro-hint">{{ $t('products.saasLead') }}</p>
      <div class="pf-products-pro__summary">
        <div v-for="row in saasRows" :key="row.key" class="pf-products-pro__summary-row">
          <strong>{{ row.name }}</strong>
          <span>{{ row.price }}</span>
          <span class="text-muted">{{ row.description }}</span>
        </div>
      </div>
      <p class="pf-products-pro__migration">{{ $t('products.saasMigration') }}</p>
    </ProCard>

    <!-- Plans clients (app) -->
    <ProCard :title="$t('products.plansTitle')" class="pro-mb-lg">
      <p class="pro-hint">{{ $t('products.plansLead') }}</p>
      <div class="pf-products-pro__grid">
        <article
          v-for="plan in plans"
          :key="plan.key"
          class="pf-products-pro__card"
          :class="{ 'pf-products-pro__card--featured': plan.featured }"
        >
          <div class="pf-products-pro__card-top">
            <h3>{{ plan.name }}</h3>
            <ProBadge v-if="plan.featured" variant="success">{{ $t('products.recommended') }}</ProBadge>
          </div>
          <p class="pf-products-pro__price">{{ plan.price }}</p>
          <p class="pf-products-pro__price-sub">{{ plan.period }} · {{ plan.monthly }}</p>
          <ul class="pro-feature-list">
            <li v-for="b in plan.benefits" :key="b">{{ b }}</li>
          </ul>
        </article>
      </div>
      <p class="pf-products-pro__tip">{{ $t('products.plansTip') }}</p>
    </ProCard>

    <ProCard :title="$t('products.includedTitle')">
      <p class="pro-hint">{{ $t('products.includedLead') }}</p>
      <ul class="pro-feature-list">
        <li v-for="item in includedItems" :key="item">{{ item }}</li>
      </ul>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['products-layout'],
})

const { t, tm, rt } = useI18n()

useSeoMeta({
  title: () => t('products.seoTitle'),
  description: () => t('products.seoDescription'),
})

function listFrom(key: string): string[] {
  const raw = tm(key) as unknown
  if (!Array.isArray(raw)) return []
  return raw.map((x) => (typeof x === 'string' ? x : rt(x as any)))
}

const solutions = computed(() =>
  (['proComplete', 'proLight'] as const).map((key) => ({
    key,
    name: t(`products.solutions.${key}.name`),
    badge: t(`products.solutions.${key}.badge`),
    price: t(`products.solutions.${key}.price`),
    setup: key === 'proComplete' ? t(`products.solutions.${key}.setup`) : '',
    tagline: t(`products.solutions.${key}.tagline`),
    featured: key === 'proComplete',
    features: listFrom(`products.solutions.${key}.features`),
  })),
)

const webFeatures = computed(() => listFrom('products.webFeatures'))
const mobileFeatures = computed(() => listFrom('products.mobileFeatures'))
const benefitsCabinet = computed(() => listFrom('products.benefitsCabinet'))
const benefitsClients = computed(() => listFrom('products.benefitsClients'))
const partnerSteps = computed(() => listFrom('products.partnerSteps'))
const includedItems = computed(() => listFrom('products.included'))

const saasRows = computed(() =>
  (['setup', 'monthly', 'annual', 'longTerm'] as const).map((key) => ({
    key,
    name: t(`products.saas.${key}.name`),
    price: t(`products.saas.${key}.price`),
    description: t(`products.saas.${key}.description`),
  })),
)

const plans = computed(() =>
  (['monthly', 'annual', 'triennial'] as const).map((key) => ({
    key,
    name: t(`products.plans.${key}.name`),
    price: t(`products.plans.${key}.price`),
    period: t(`products.plans.${key}.period`),
    monthly: t(`products.plans.${key}.monthly`),
    featured: key === 'triennial',
    benefits: listFrom(`products.plans.${key}.benefits`),
  })),
)

function printPage() {
  if (import.meta.client) window.print()
}
</script>

<style scoped>
.pf-products-pro__positioning {
  margin: 0 0 1.5rem;
  color: var(--pf-vet-text-muted);
  line-height: 1.55;
  max-width: 46rem;
}

.pf-products-pro__section-title {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  color: var(--pf-vet-primary);
}

.pf-products-pro__solutions-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, 1fr);
  margin-top: 1rem;
}

.pf-products-pro__solution {
  background: var(--pf-vet-bg);
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.pf-products-pro__solution--featured {
  border-color: var(--pf-vet-primary);
  box-shadow: 0 0 0 1px var(--pf-vet-primary);
}

.pf-products-pro__solution-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pf-products-pro__solution h3 {
  margin: 0;
  font-size: 1.15rem;
  color: var(--pf-vet-primary);
}

.pf-products-pro__solution-tagline {
  margin: 0.15rem 0 0.35rem;
  color: var(--pf-vet-text-muted);
  line-height: 1.45;
  font-size: 0.95rem;
}

.pf-products-pro__duo {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, 1fr);
}

.pf-products-pro__benefits {
  display: grid;
  gap: 1.25rem;
  grid-template-columns: repeat(2, 1fr);
}

.pf-products-pro__benefits-h {
  margin: 0 0 0.5rem;
  font-size: 1rem;
  color: var(--pf-vet-primary);
}

.pf-products-pro__steps {
  margin: 0.75rem 0 1rem;
  padding-left: 1.25rem;
  line-height: 1.55;
  color: var(--pf-vet-text);
}

.pf-products-pro__steps li + li {
  margin-top: 0.4rem;
}

.pf-products-pro__migration {
  margin: 1rem 0 0;
  color: var(--pf-vet-text-muted);
  font-size: 0.9rem;
  line-height: 1.45;
}

.pf-products-pro__grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(3, 1fr);
}

.pf-products-pro__card {
  background: var(--pf-vet-bg);
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.pf-products-pro__card--featured {
  border-color: var(--pf-vet-primary);
  box-shadow: 0 0 0 1px var(--pf-vet-primary);
}

.pf-products-pro__card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pf-products-pro__card h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--pf-vet-primary);
}

.pf-products-pro__price {
  margin: 0.25rem 0 0;
  font-size: 1.75rem;
  font-weight: 700;
}

.pf-products-pro__price-sub {
  margin: 0 0 0.5rem;
  color: var(--pf-vet-text-muted);
  font-size: 0.9rem;
}

.pf-products-pro__tip {
  margin: 1rem 0 0;
  padding: 0.85rem 1rem;
  background: rgba(42, 157, 143, 0.08);
  border-radius: 10px;
  line-height: 1.5;
  font-size: 0.95rem;
}

.pf-products-pro__summary {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.pf-products-pro__summary-row {
  display: grid;
  grid-template-columns: 1.2fr 0.7fr 1.4fr;
  gap: 0.75rem;
  padding: 0.85rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
  align-items: baseline;
}

.pf-products-pro__summary-row:last-child {
  border-bottom: none;
}

@media (max-width: 900px) {
  .pf-products-pro__solutions-grid,
  .pf-products-pro__duo,
  .pf-products-pro__benefits,
  .pf-products-pro__grid {
    grid-template-columns: 1fr;
  }

  .pf-products-pro__summary-row {
    grid-template-columns: 1fr;
    gap: 0.25rem;
  }
}

@media print {
  .no-print {
    display: none !important;
  }
}
</style>
