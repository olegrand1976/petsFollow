<template>
  <div class="pro-landing">
    <header class="pro-landing__header">
      <PetsFollowLogo variant="default" />
      <nav class="pro-landing__nav">
        <ProLocaleSelect />
        <a href="#produits" class="pro-landing__nav-link">{{ $t('index.productsLink') }}</a>
        <NuxtLink to="/login" class="pro-landing__nav-link">{{ $t('index.login') }}</NuxtLink>
        <NuxtLink to="/register" class="pro-btn pro-btn--primary" data-testid="landing-cta">
          {{ $t('index.registerCta') }}
        </NuxtLink>
      </nav>
    </header>

    <section class="pro-landing__hero">
      <div class="pro-landing__hero-content">
        <span class="pro-landing__badge">{{ $t('index.badge') }}</span>
        <h1>
          <span class="pro-landing__hero-line">{{ $t('index.heroTitleLine1') }}</span>
          <span class="pro-landing__hero-line">{{ $t('index.heroTitleLine2') }}</span>
        </h1>
        <p class="pro-landing__lead">{{ $t('index.heroLead') }}</p>
        <div class="pro-landing__hero-actions">
          <NuxtLink to="/register" class="pro-btn pro-btn--primary" data-testid="landing-hero-cta">
            {{ $t('index.heroCta') }}
          </NuxtLink>
          <NuxtLink to="/login" class="pro-landing__secondary-link">{{ $t('index.heroLogin') }}</NuxtLink>
        </div>
      </div>

      <div class="pro-landing__triangle" aria-hidden="true">
        <div class="pro-landing__actor pro-landing__actor--cabinet">
          <ProIcon name="local_hospital" :size="22" />
          <strong>{{ $t('index.actors.cabinet.title') }}</strong>
          <span>{{ $t('index.actors.cabinet.text') }}</span>
        </div>
        <div class="pro-landing__triangle-mid">
          <div class="pro-landing__actor pro-landing__actor--care">
            <ProIcon name="handshake" :size="22" />
            <strong>{{ $t('index.actors.care.title') }}</strong>
            <span>{{ $t('index.actors.care.text') }}</span>
          </div>
          <div class="pro-landing__passport-node">
            <ProIcon name="folder_shared" :size="26" />
            <strong>{{ $t('index.actors.passport') }}</strong>
          </div>
          <div class="pro-landing__actor pro-landing__actor--owner">
            <ProIcon name="person" :size="22" />
            <strong>{{ $t('index.actors.owner.title') }}</strong>
            <span>{{ $t('index.actors.owner.text') }}</span>
          </div>
        </div>
      </div>
    </section>

    <section class="pro-landing__problem">
      <div class="pro-landing__problem-inner">
        <h2>{{ $t('index.problemTitle') }}</h2>
        <p>{{ $t('index.problemLead') }}</p>
      </div>
    </section>

    <section class="pro-landing__passport">
      <h2>{{ $t('index.passportTitle') }}</h2>
      <p class="pro-landing__section-lead">{{ $t('index.passportLead') }}</p>
      <div class="pro-landing__faces">
        <article v-for="face in faces" :key="face.key" class="pro-landing__face">
          <span class="pro-landing__face-role">{{ $t(`index.faces.${face.key}.role`) }}</span>
          <ProIcon :name="face.icon" class="pro-landing__face-icon" :size="28" />
          <h3>{{ $t(`index.faces.${face.key}.title`) }}</h3>
          <p>{{ $t(`index.faces.${face.key}.text`) }}</p>
        </article>
      </div>
    </section>

    <section class="pro-landing__continuity">
      <h2>{{ $t('index.continuityTitle') }}</h2>
      <p class="pro-landing__section-lead">{{ $t('index.continuityLead') }}</p>
      <ol class="pro-landing__steps">
        <li v-for="(step, i) in steps" :key="step.key" class="pro-landing__step">
          <span class="pro-landing__step-num">{{ i + 1 }}</span>
          <div>
            <h3>{{ $t(`index.steps.${step.key}.title`) }}</h3>
            <p>{{ $t(`index.steps.${step.key}.text`) }}</p>
          </div>
        </li>
      </ol>
    </section>

    <section id="produits" class="pro-landing__products">
      <h2>{{ $t('index.productsTitle') }}</h2>
      <p class="pro-landing__products-lead">{{ $t('index.productsLead') }}</p>

      <div class="pro-landing__solutions-grid">
        <article
          v-for="sol in solutions"
          :key="sol.key"
          class="pro-landing__solution"
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
        <article v-for="item in productHighlights" :key="item.key" class="pro-landing__plan">
          <strong class="pro-landing__products-price">{{ $t(`index.productHighlights.${item.key}.price`) }}</strong>
          <h3>{{ $t(`index.productHighlights.${item.key}.title`) }}</h3>
          <p>{{ $t(`index.productHighlights.${item.key}.text`) }}</p>
        </article>
      </div>

      <div class="pro-landing__products-actions">
        <NuxtLink to="/register" class="pro-btn pro-btn--primary" data-testid="landing-products-cta">
          {{ $t('index.productsCta') }}
        </NuxtLink>
      </div>
    </section>

    <section class="pro-landing__cta">
      <div class="pro-landing__cta-inner">
        <h2>{{ $t('index.ctaTitle') }}</h2>
        <p>{{ $t('index.ctaText') }}</p>
        <NuxtLink to="/register" class="pro-btn pro-btn--primary" data-testid="landing-footer-cta">
          {{ $t('index.ctaButton') }}
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

const faces = [
  { key: 'pro', icon: 'local_hospital' },
  { key: 'proLight', icon: 'phone_android' },
  { key: 'owner', icon: 'pets' },
] as const

const steps = [
  { key: 'prescribe' },
  { key: 'observe' },
  { key: 'document' },
  { key: 'retrieve' },
] as const

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
