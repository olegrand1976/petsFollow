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

    <ProCard :title="$t('products.includedTitle')" class="pro-mb-lg">
      <p class="pro-hint">{{ $t('products.includedLead') }}</p>
      <ul class="pro-feature-list">
        <li v-for="item in includedItems" :key="item">{{ item }}</li>
      </ul>
    </ProCard>

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

    <ProCard :title="$t('products.summaryTitle')">
      <div class="pf-products-pro__summary">
        <div v-for="row in summaryRows" :key="row.code" class="pf-products-pro__summary-row">
          <strong>{{ row.name }}</strong>
          <span>{{ row.price }}</span>
          <span class="text-muted">{{ row.role }}</span>
        </div>
      </div>
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

const includedItems = computed(() => listFrom('products.included'))

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

const summaryRows = computed(() =>
  (['monthly', 'annual', 'triennial'] as const).map((key) => ({
    code: key,
    name: t(`products.summary.${key}.name`),
    price: t(`products.summary.${key}.price`),
    role: t(`products.summary.${key}.role`),
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
