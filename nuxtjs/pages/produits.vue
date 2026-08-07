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

    <!-- 1. Accroche -->
    <p class="pf-products-pro__hook">{{ $t('products.hook') }}</p>

    <!-- 2. Offre Pro = Plateforme Web cabinet -->
    <article
      class="pf-products-pro__hero pro-mb-lg"
      data-testid="products-solution-proComplete"
    >
      <div class="pf-products-pro__hero-top">
        <div>
          <h2 class="pf-products-pro__hero-title">{{ $t('products.pro.name') }}</h2>
          <p class="pf-products-pro__hero-badge-line">
            <ProBadge variant="success">{{ $t('products.pro.badge') }}</ProBadge>
          </p>
        </div>
        <div class="pf-products-pro__hero-price-block">
          <p class="pf-products-pro__price">{{ $t('products.pro.price') }}</p>
          <p class="pf-products-pro__price-sub">{{ $t('products.pro.setup') }}</p>
        </div>
      </div>
      <p class="pf-products-pro__tagline">{{ $t('products.pro.tagline') }}</p>
      <ul class="pro-feature-list">
        <li v-for="f in proFeatures" :key="f">{{ f }}</li>
      </ul>
      <p class="pf-products-pro__tip">{{ $t('products.pro.invoiceNote') }}</p>
    </article>

    <!-- 3. Pro Light -->
    <article
      class="pf-products-pro__light pro-mb-lg"
      data-testid="products-solution-proLight"
    >
      <div class="pf-products-pro__light-top">
        <h2 class="pf-products-pro__section-title">{{ $t('products.proLight.name') }}</h2>
        <ProBadge variant="neutral">{{ $t('products.proLight.badge') }}</ProBadge>
        <p class="pf-products-pro__price pf-products-pro__price--sm">{{ $t('products.proLight.price') }}</p>
      </div>
      <p class="pf-products-pro__tagline">{{ $t('products.proLight.tagline') }}</p>
      <ul class="pro-feature-list">
        <li v-for="f in proLightFeatures" :key="f">{{ f }}</li>
      </ul>
    </article>

    <!-- 4. Clients : app + plans + inclus -->
    <ProCard :title="$t('products.clientsTitle')" class="pro-mb-lg">
      <p class="pro-hint">{{ $t('products.clientsLead') }}</p>
      <ul class="pro-feature-list pf-products-pro__client-features">
        <li v-for="item in clientFeatures" :key="item">{{ item }}</li>
      </ul>
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
      <p class="pf-products-pro__included-line">{{ $t('products.includedLine') }}</p>
      <p class="pf-products-pro__tip">{{ $t('products.plansTip') }}</p>
    </ProCard>

    <!-- 5. Autofinancement + engagements SaaS -->
    <ProCard :title="$t('products.economyTitle')" class="pro-mb-lg">
      <p class="pro-hint">{{ $t('products.partnerLead') }}</p>
      <ol class="pf-products-pro__steps">
        <li v-for="step in partnerSteps" :key="step">{{ step }}</li>
      </ol>
      <p class="pf-products-pro__tip">{{ $t('products.partnerTip') }}</p>
      <h3 class="pf-products-pro__subsection">{{ $t('products.saasTitle') }}</h3>
      <p class="pro-hint">{{ $t('products.saasLead') }}</p>
      <div class="pf-products-pro__summary">
        <div
          v-for="row in saasRows"
          :key="row.key"
          class="pf-products-pro__summary-row"
          :class="{ 'pf-products-pro__summary-row--featured': row.featured }"
        >
          <strong>
            {{ row.name }}
            <ProBadge v-if="row.featured" variant="success">{{ $t('products.recommended') }}</ProBadge>
          </strong>
          <span>{{ row.price }}</span>
          <span class="text-muted">{{ row.description }}</span>
        </div>
      </div>
      <p class="pf-products-pro__migration">{{ $t('products.saasMigration') }}</p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['shared-pro-layout'],
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

const proFeatures = computed(() => listFrom('products.pro.features'))
const proLightFeatures = computed(() => listFrom('products.proLight.features'))
const clientFeatures = computed(() => listFrom('products.clientFeatures'))
const partnerSteps = computed(() => listFrom('products.partnerSteps'))

const saasRows = computed(() =>
  (['setup', 'annual', 'longTerm'] as const).map((key) => ({
    key,
    name: t(`products.saas.${key}.name`),
    price: t(`products.saas.${key}.price`),
    description: t(`products.saas.${key}.description`),
    featured: key === 'longTerm',
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
.pf-products-pro__hook {
  margin: 0 0 1.5rem;
  color: var(--pf-vet-text-muted);
  line-height: 1.55;
  max-width: 42rem;
  font-size: 1.05rem;
}

.pf-products-pro__section-title {
  margin: 0;
  font-size: 1.15rem;
  color: var(--pf-vet-primary);
}

.pf-products-pro__hero {
  background: var(--pf-vet-bg);
  border: 1px solid var(--pf-vet-primary);
  box-shadow: 0 0 0 1px var(--pf-vet-primary);
  border-radius: 12px;
  padding: 1.35rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.pf-products-pro__hero-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}

.pf-products-pro__hero-title {
  margin: 0 0 0.4rem;
  font-size: 1.35rem;
  color: var(--pf-vet-primary);
}

.pf-products-pro__hero-badge-line {
  margin: 0;
}

.pf-products-pro__hero-price-block {
  text-align: right;
}

.pf-products-pro__light {
  background: var(--pf-vet-bg);
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  padding: 1.1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.pf-products-pro__light-top {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.pf-products-pro__tagline {
  margin: 0.1rem 0 0.25rem;
  color: var(--pf-vet-text-muted);
  line-height: 1.45;
  font-size: 0.95rem;
}

.pf-products-pro__client-features {
  margin-bottom: 1.25rem;
}

.pf-products-pro__included-line {
  margin: 1rem 0 0;
  color: var(--pf-vet-text);
  line-height: 1.5;
  font-size: 0.95rem;
}

.pf-products-pro__subsection {
  margin: 1.25rem 0 0.5rem;
  font-size: 1.05rem;
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

.pf-products-pro__price--sm {
  font-size: 1.25rem;
  margin-left: auto;
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

.pf-products-pro__summary-row--featured {
  padding: 0.85rem 0.75rem;
  border-radius: 8px;
  background: rgba(42, 157, 143, 0.06);
  border-bottom-color: transparent;
}

.pf-products-pro__summary-row strong {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pf-products-pro__summary-row:last-child {
  border-bottom: none;
}

@media (max-width: 900px) {
  .pf-products-pro__grid {
    grid-template-columns: 1fr;
  }

  .pf-products-pro__hero-price-block {
    text-align: left;
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
