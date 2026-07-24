<template>
  <div class="pro-welcome">
    <header class="pro-welcome__header">
      <PetsFollowLogo variant="default" />
      <div class="pro-welcome__header-actions">
        <ProLocaleSelect />
        <ProButton @click="goPrimary">{{ primaryLabel }}</ProButton>
      </div>
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
        <ProIcon :name="item.icon" class="pro-welcome__highlight-icon" :size="32" />
        <h3>{{ $t(`welcome.highlights.${item.key}.title`) }}</h3>
        <p>{{ $t(`welcome.highlights.${item.key}.text`) }}</p>
      </div>
    </section>

    <section class="pro-welcome__cta">
      <h2>{{ ctaTitle }}</h2>
      <p>{{ ctaText }}</p>
      <div class="pro-welcome__cta-actions">
        <ProButton @click="goPrimary">{{ primaryLabel }}</ProButton>
        <ProButton variant="ghost" @click="navigateTo('/')">{{ $t('common.backHome') }}</ProButton>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const token = useCookie('pf_token')
const session = useCookie('pf_session')
const isAuthenticated = computed(() => !!(token.value || session.value))

const steps = [
  { key: 'profile' },
  { key: 'invite' },
  { key: 'monitor' },
  { key: 'communicate' },
]

const highlights = [
  { key: 'bpm', icon: 'favorite' },
  { key: 'alerts', icon: 'notifications' },
  { key: 'free', icon: 'card_giftcard' },
]

const primaryLabel = computed(() =>
  isAuthenticated.value ? t('welcome.ctaContinue') : t('welcome.ctaLogin'),
)
const ctaTitle = computed(() =>
  isAuthenticated.value ? t('welcome.ctaTitleAuthed') : t('welcome.ctaTitle'),
)
const ctaText = computed(() =>
  isAuthenticated.value ? t('welcome.ctaTextAuthed') : t('welcome.ctaText'),
)

async function goPrimary() {
  if (!isAuthenticated.value) {
    await navigateTo('/login')
    return
  }
  try {
    const me: any = await $fetch('/api/me')
    const data = me.data ?? me
    if (data.role === 'vet' && data.profileComplete !== true) {
      await navigateTo('/onboarding')
      return
    }
    if (isProRole(data.role)) {
      await navigateTo(homePathForRole(data.role, { profileComplete: data.profileComplete }))
      return
    }
    await clearAuthTokens()
    await navigateTo('/login')
  } catch {
    await clearAuthTokens()
    await navigateTo('/login')
  }
}
</script>
