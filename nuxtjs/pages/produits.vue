<template>
  <div class="pf-products" data-testid="products-page">
    <header class="pf-products__header">
      <NuxtLink to="/" class="pf-products__brand">
        <PetsFollowLogo variant="default" />
      </NuxtLink>
      <nav class="pf-products__nav">
        <ProLocaleSelect />
        <NuxtLink to="/" class="pf-products__nav-link">{{ $t('products.backHome') }}</NuxtLink>
        <ProButton variant="secondary" class="pf-products__print no-print" @click="printPage">
          {{ $t('products.print') }}
        </ProButton>
      </nav>
    </header>

    <main class="pf-products__main">
      <section class="pf-products__intro">
        <span class="pf-products__badge">{{ $t('products.badge') }}</span>
        <h1>{{ $t('products.title') }}</h1>
        <p class="pf-products__lead">{{ $t('products.lead') }}</p>
        <p class="pf-products__positioning">{{ $t('products.positioning') }}</p>
      </section>

      <section class="pf-products__block">
        <h2>{{ $t('products.includedTitle') }}</h2>
        <p class="pf-products__block-lead">{{ $t('products.includedLead') }}</p>
        <ul class="pf-products__checklist">
          <li v-for="item in includedItems" :key="item">{{ item }}</li>
        </ul>
      </section>

      <section class="pf-products__block">
        <h2>{{ $t('products.plansTitle') }}</h2>
        <p class="pf-products__block-lead">{{ $t('products.plansLead') }}</p>
        <div class="pf-products__grid pf-products__grid--3">
          <article
            v-for="plan in plans"
            :key="plan.key"
            class="pf-products__card"
            :class="{ 'pf-products__card--featured': plan.featured }"
          >
            <div class="pf-products__card-top">
              <h3>{{ plan.name }}</h3>
              <ProBadge v-if="plan.featured" variant="success">{{ $t('products.recommended') }}</ProBadge>
            </div>
            <p class="pf-products__price">{{ plan.price }}</p>
            <p class="pf-products__price-sub">{{ plan.period }} · {{ plan.monthly }}</p>
            <ul class="pf-products__benefits">
              <li v-for="b in plan.benefits" :key="b">{{ b }}</li>
            </ul>
          </article>
        </div>
        <p class="pf-products__tip">{{ $t('products.plansTip') }}</p>
      </section>

      <section class="pf-products__block">
        <h2>{{ $t('products.addonsTitle') }}</h2>
        <p class="pf-products__block-lead">{{ $t('products.addonsLead') }}</p>
        <div class="pf-products__grid pf-products__grid--3">
          <article v-for="addon in addons" :key="addon.key" class="pf-products__card">
            <div class="pf-products__card-top">
              <h3>{{ addon.name }}</h3>
              <ProBadge variant="neutral">{{ addon.scope }}</ProBadge>
            </div>
            <p class="pf-products__price">{{ addon.price }}</p>
            <p class="pf-products__price-sub">{{ addon.tagline }}</p>
            <ul class="pf-products__benefits">
              <li v-for="b in addon.benefits" :key="b">{{ b }}</li>
            </ul>
          </article>
        </div>
        <p class="pf-products__tip">{{ $t('products.addonsTip') }}</p>
      </section>

      <section class="pf-products__block pf-products__block--summary">
        <h2>{{ $t('products.summaryTitle') }}</h2>
        <div class="pf-products__summary">
          <div v-for="row in summaryRows" :key="row.code" class="pf-products__summary-row">
            <strong>{{ row.name }}</strong>
            <span>{{ row.price }}</span>
            <span class="pf-products__summary-role">{{ row.role }}</span>
          </div>
        </div>
      </section>
    </main>

    <footer class="pf-products__footer">
      <PetsFollowLogo variant="compact" />
      <p>{{ $t('products.footer') }}</p>
      <ProLegalFooter />
    </footer>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

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
  (['annual', 'triennial', 'quinquennial'] as const).map((key) => ({
    key,
    name: t(`products.plans.${key}.name`),
    price: t(`products.plans.${key}.price`),
    period: t(`products.plans.${key}.period`),
    monthly: t(`products.plans.${key}.monthly`),
    featured: key === 'triennial',
    benefits: listFrom(`products.plans.${key}.benefits`),
  })),
)

const addons = computed(() =>
  (['family', 'carePlus', 'horse'] as const).map((key) => ({
    key,
    name: t(`products.addons.${key}.name`),
    price: t(`products.addons.${key}.price`),
    tagline: t(`products.addons.${key}.tagline`),
    scope: t(`products.addons.${key}.scope`),
    benefits: listFrom(`products.addons.${key}.benefits`),
  })),
)

const summaryRows = computed(() =>
  (['annual', 'triennial', 'quinquennial', 'family', 'carePlus', 'horse'] as const).map((key) => ({
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
.pf-products {
  min-height: 100vh;
  background: var(--pf-vet-bg);
  color: var(--pf-vet-text);
}

.pf-products__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem 2rem;
  background: var(--pf-vet-surface);
  border-bottom: 1px solid var(--pf-vet-border);
  position: sticky;
  top: 0;
  z-index: 10;
}

.pf-products__brand {
  text-decoration: none;
}

.pf-products__nav {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.pf-products__nav-link {
  color: var(--pf-vet-primary);
  font-weight: 600;
  text-decoration: none;
}

.pf-products__main {
  max-width: 1100px;
  margin: 0 auto;
  padding: 2.5rem 1.5rem 4rem;
}

.pf-products__intro {
  margin-bottom: 2.5rem;
}

.pf-products__badge {
  display: inline-block;
  background: rgba(42, 157, 143, 0.12);
  color: var(--pf-vet-primary);
  font-weight: 600;
  font-size: 0.85rem;
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
  margin-bottom: 1rem;
}

.pf-products__intro h1 {
  font-size: clamp(1.75rem, 3vw, 2.4rem);
  color: var(--pf-vet-primary);
  margin: 0 0 0.75rem;
  line-height: 1.2;
}

.pf-products__lead {
  font-size: 1.125rem;
  line-height: 1.55;
  margin: 0 0 0.75rem;
  max-width: 46rem;
}

.pf-products__positioning {
  margin: 0;
  color: var(--pf-vet-text-muted);
  line-height: 1.55;
  max-width: 46rem;
}

.pf-products__block {
  margin-bottom: 2.75rem;
}

.pf-products__block h2 {
  color: var(--pf-vet-primary);
  margin: 0 0 0.5rem;
  font-size: 1.35rem;
}

.pf-products__block-lead {
  margin: 0 0 1.25rem;
  color: var(--pf-vet-text-muted);
  line-height: 1.5;
  max-width: 42rem;
}

.pf-products__checklist,
.pf-products__benefits {
  margin: 0;
  padding-left: 1.2rem;
  display: grid;
  gap: 0.45rem;
  line-height: 1.45;
}

.pf-products__checklist {
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  padding: 1.25rem 1.25rem 1.25rem 2.25rem;
}

.pf-products__grid {
  display: grid;
  gap: 1rem;
}

.pf-products__grid--3 {
  grid-template-columns: repeat(3, 1fr);
}

.pf-products__card {
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.pf-products__card--featured {
  border-color: var(--pf-vet-primary);
  box-shadow: 0 0 0 1px var(--pf-vet-primary);
}

.pf-products__card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pf-products__card h3 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--pf-vet-primary);
}

.pf-products__price {
  margin: 0.25rem 0 0;
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--pf-vet-text);
}

.pf-products__price-sub {
  margin: 0 0 0.5rem;
  color: var(--pf-vet-text-muted);
  font-size: 0.9rem;
}

.pf-products__tip {
  margin: 1rem 0 0;
  padding: 0.85rem 1rem;
  background: rgba(42, 157, 143, 0.08);
  border-radius: 10px;
  color: var(--pf-vet-text);
  line-height: 1.5;
  font-size: 0.95rem;
}

.pf-products__summary {
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  overflow: hidden;
}

.pf-products__summary-row {
  display: grid;
  grid-template-columns: 1.2fr 0.7fr 1.4fr;
  gap: 0.75rem;
  padding: 0.85rem 1.1rem;
  border-bottom: 1px solid var(--pf-vet-border);
  align-items: baseline;
}

.pf-products__summary-row:last-child {
  border-bottom: none;
}

.pf-products__summary-role {
  color: var(--pf-vet-text-muted);
  font-size: 0.9rem;
}

.pf-products__footer {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2rem;
  color: var(--pf-vet-text-muted);
  font-size: 0.875rem;
  border-top: 1px solid var(--pf-vet-border);
}

@media (max-width: 900px) {
  .pf-products__grid--3 {
    grid-template-columns: 1fr;
  }

  .pf-products__summary-row {
    grid-template-columns: 1fr;
    gap: 0.25rem;
  }
}

@media print {
  .pf-products__header,
  .no-print {
    display: none !important;
  }

  .pf-products {
    background: white;
  }

  .pf-products__main {
    padding-top: 0;
  }

  .pf-products__card,
  .pf-products__checklist,
  .pf-products__summary {
    break-inside: avoid;
  }
}
</style>
