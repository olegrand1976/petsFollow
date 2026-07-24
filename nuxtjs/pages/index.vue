<template>
  <div class="pro-landing">
    <header class="pro-landing__header">
      <PetsFollowLogo variant="default" />
      <nav class="pro-landing__nav">
        <ProLocaleSelect />
        <a href="#produits" class="pro-landing__nav-link">{{ $t('index.productsLink') }}</a>
        <NuxtLink to="/login" class="pro-landing__nav-link">{{ $t('index.login') }}</NuxtLink>
        <NuxtLink to="/register">
          <ProButton test-id="landing-cta">{{ $t('index.registerCta') }}</ProButton>
        </NuxtLink>
      </nav>
    </header>

    <section class="pro-landing__hero">
      <div class="pro-landing__hero-content">
        <span class="pro-landing__badge">{{ $t('index.badge') }}</span>
        <h1>{{ $t('index.heroTitle') }}</h1>
        <p class="pro-landing__lead">{{ $t('index.heroLead') }}</p>
        <div class="pro-landing__hero-actions">
          <NuxtLink to="/register">
            <ProButton test-id="landing-hero-cta">{{ $t('index.heroCta') }}</ProButton>
          </NuxtLink>
          <NuxtLink to="/login" class="pro-landing__secondary-link">{{ $t('index.heroLogin') }}</NuxtLink>
        </div>
      </div>
      <div class="pro-landing__hero-visual" aria-hidden="true">
        <div class="pro-landing__card pro-landing__card--float">
          <ProIcon name="favorite" class="pro-landing__card-icon" :size="24" />
          <strong>{{ $t('index.cards.heartrate.title') }}</strong>
          <p>{{ $t('index.cards.heartrate.text') }}</p>
        </div>
        <div class="pro-landing__card pro-landing__card--float pro-landing__card--delay">
          <ProIcon name="chat" class="pro-landing__card-icon" :size="24" />
          <strong>{{ $t('index.cards.messaging.title') }}</strong>
          <p>{{ $t('index.cards.messaging.text') }}</p>
        </div>
        <div class="pro-landing__card pro-landing__card--float pro-landing__card--delay2">
          <ProIcon name="description" class="pro-landing__card-icon" :size="24" />
          <strong>{{ $t('index.cards.records.title') }}</strong>
          <p>{{ $t('index.cards.records.text') }}</p>
        </div>
      </div>
    </section>

    <section class="pro-landing__features">
      <h2>{{ $t('index.featuresTitle') }}</h2>
      <div class="pro-landing__feature-grid">
        <article v-for="feature in features" :key="feature.key" class="pro-landing__feature">
          <ProIcon :name="feature.icon" class="pro-landing__feature-icon" :size="28" />
          <h3>{{ $t(`index.features.${feature.key}.title`) }}</h3>
          <p>{{ $t(`index.features.${feature.key}.text`) }}</p>
        </article>
      </div>
    </section>

    <section id="produits" class="pro-landing__products">
      <h2>{{ $t('index.productsTitle') }}</h2>
      <p class="pro-landing__products-lead">{{ $t('index.productsLead') }}</p>

      <div class="pro-landing__solutions-grid">
        <article
          v-for="sol in solutions"
          :key="sol.key"
          class="pro-landing__feature pro-landing__solution"
          :class="{ 'pro-landing__solution--featured': sol.featured }"
          :data-testid="`landing-solution-${sol.key}`"
        >
          <strong class="pro-landing__products-price">{{ $t(`index.solutions.${sol.key}.price`) }}</strong>
          <h3>{{ $t(`index.solutions.${sol.key}.title`) }}</h3>
          <p>{{ $t(`index.solutions.${sol.key}.text`) }}</p>
          <ul class="pro-landing__solution-list">
            <li v-for="item in sol.features" :key="item">{{ item }}</li>
          </ul>
        </article>
      </div>

      <h3 class="pro-landing__products-sub">{{ $t('index.clientPlansTitle') }}</h3>
      <p class="pro-landing__products-lead pro-landing__products-lead--tight">{{ $t('index.clientPlansLead') }}</p>
      <div class="pro-landing__products-grid">
        <article v-for="item in productHighlights" :key="item.key" class="pro-landing__feature">
          <strong class="pro-landing__products-price">{{ $t(`index.productHighlights.${item.key}.price`) }}</strong>
          <h3>{{ $t(`index.productHighlights.${item.key}.title`) }}</h3>
          <p>{{ $t(`index.productHighlights.${item.key}.text`) }}</p>
        </article>
      </div>

      <div class="pro-landing__products-actions">
        <NuxtLink to="/register">
          <ProButton test-id="landing-products-cta">{{ $t('index.productsCta') }}</ProButton>
        </NuxtLink>
      </div>
    </section>

    <section class="pro-landing__cta">
      <div class="pro-landing__cta-inner">
        <h2>{{ $t('index.ctaTitle') }}</h2>
        <p>{{ $t('index.ctaText') }}</p>
        <NuxtLink to="/register">
          <ProButton variant="secondary">{{ $t('index.ctaButton') }}</ProButton>
        </NuxtLink>
      </div>
    </section>

    <footer class="pro-landing__footer">
      <PetsFollowLogo variant="compact" />
      <p>{{ $t('index.footer', { year }) }}</p>
      <ProLegalFooter />
    </footer>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const year = new Date().getFullYear()

const { tm, rt } = useI18n()

function listFrom(key: string): string[] {
  const raw = tm(key) as unknown
  if (!Array.isArray(raw)) return []
  return raw.map((x) => (typeof x === 'string' ? x : rt(x as any)))
}

const features = [
  { key: 'heartrate', icon: 'favorite' },
  { key: 'alerts', icon: 'notifications' },
  { key: 'messaging', icon: 'chat' },
  { key: 'partner', icon: 'handshake' },
  { key: 'security', icon: 'lock' },
  { key: 'onboarding', icon: 'bolt' },
]

const solutions = computed(() =>
  (['proComplete', 'proLight'] as const).map((key) => ({
    key,
    featured: key === 'proComplete',
    features: listFrom(`index.solutions.${key}.features`),
  })),
)

const productHighlights = [
  { key: 'monthly' },
  { key: 'annual' },
  { key: 'triennial' },
]
</script>
