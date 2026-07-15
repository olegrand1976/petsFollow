<template>
  <div class="pro-welcome">
    <header class="pro-welcome__header">
      <PetsFollowLogo variant="default" />
      <ProButton @click="navigateTo('/login')">{{ $t('welcome.login') }}</ProButton>
    </header>

    <section class="pro-welcome__hero">
      <span class="pro-landing__badge">{{ $t('welcome.badge') }}</span>
      <h1>{{ $t('welcome.title') }}</h1>
      <p>
        <i18n-t keypath="welcome.lead" tag="span">
          <template #free>
            <strong>{{ $t('welcome.free') }}</strong>
          </template>
        </i18n-t>
      </p>
    </section>

    <section class="pro-welcome__steps">
      <h2>{{ $t('welcome.howItWorks') }}</h2>
      <ol class="pro-welcome__step-list">
        <li v-for="(step, i) in steps" :key="step.key" class="pro-welcome__step">
          <span class="pro-welcome__step-num">{{ i + 1 }}</span>
          <div>
            <h3>{{ $t(`welcome.steps.${step.key}.title`) }}</h3>
            <p>{{ $t(`welcome.steps.${step.key}.text`) }}</p>
          </div>
        </li>
      </ol>
    </section>

    <section class="pro-welcome__features">
      <div v-for="item in highlights" :key="item.key" class="pro-welcome__highlight">
        <span aria-hidden="true">{{ item.icon }}</span>
        <h3>{{ $t(`welcome.highlights.${item.key}.title`) }}</h3>
        <p>{{ $t(`welcome.highlights.${item.key}.text`) }}</p>
      </div>
    </section>

    <section class="pro-welcome__cta">
      <h2>{{ $t('welcome.ctaTitle') }}</h2>
      <p>{{ $t('welcome.ctaText') }}</p>
      <div class="pro-welcome__cta-actions">
        <ProButton @click="navigateTo('/login')">{{ $t('welcome.ctaLogin') }}</ProButton>
        <ProButton variant="ghost" @click="navigateTo('/')">{{ $t('common.backHome') }}</ProButton>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const steps = [
  { key: 'profile' },
  { key: 'invite' },
  { key: 'monitor' },
  { key: 'communicate' },
]

const highlights = [
  { key: 'bpm', icon: '❤️' },
  { key: 'alerts', icon: '🔔' },
  { key: 'free', icon: '🆓' },
]
</script>
