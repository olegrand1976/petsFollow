<template>
  <div class="pro-landing">
    <header class="pro-landing__header">
      <PetsFollowLogo variant="default" />
      <nav class="pro-landing__nav">
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
          <span class="pro-landing__card-icon">❤️</span>
          <strong>{{ $t('index.cards.heartrate.title') }}</strong>
          <p>{{ $t('index.cards.heartrate.text') }}</p>
        </div>
        <div class="pro-landing__card pro-landing__card--float pro-landing__card--delay">
          <span class="pro-landing__card-icon">💬</span>
          <strong>{{ $t('index.cards.messaging.title') }}</strong>
          <p>{{ $t('index.cards.messaging.text') }}</p>
        </div>
        <div class="pro-landing__card pro-landing__card--float pro-landing__card--delay2">
          <span class="pro-landing__card-icon">📋</span>
          <strong>{{ $t('index.cards.records.title') }}</strong>
          <p>{{ $t('index.cards.records.text') }}</p>
        </div>
      </div>
    </section>

    <section class="pro-landing__features">
      <h2>{{ $t('index.featuresTitle') }}</h2>
      <div class="pro-landing__feature-grid">
        <article v-for="feature in features" :key="feature.key" class="pro-landing__feature">
          <span class="pro-landing__feature-icon" aria-hidden="true">{{ feature.icon }}</span>
          <h3>{{ $t(`index.features.${feature.key}.title`) }}</h3>
          <p>{{ $t(`index.features.${feature.key}.text`) }}</p>
        </article>
      </div>
    </section>

    <section class="pro-landing__cta">
      <div class="pro-landing__cta-inner">
        <h2>{{ $t('index.ctaTitle') }}</h2>
        <p>
          <i18n-t keypath="index.ctaText" tag="span">
            <template #free>
              <strong>{{ $t('index.ctaFree') }}</strong>
            </template>
          </i18n-t>
        </p>
        <NuxtLink to="/register">
          <ProButton variant="secondary">{{ $t('index.ctaButton') }}</ProButton>
        </NuxtLink>
      </div>
    </section>

    <footer class="pro-landing__footer">
      <PetsFollowLogo variant="compact" />
      <p>{{ $t('index.footer', { year }) }}</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const year = new Date().getFullYear()

const features = [
  { key: 'heartrate', icon: '❤️' },
  { key: 'alerts', icon: '🔔' },
  { key: 'messaging', icon: '💬' },
  { key: 'free', icon: '🆓' },
  { key: 'security', icon: '🔒' },
  { key: 'onboarding', icon: '⚡' },
]
</script>
