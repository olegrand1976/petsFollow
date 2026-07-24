<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ done ? $t('auth.reset.brandDone') : $t('auth.reset.brandTitle') }}</h2>
      <p>{{ done ? $t('auth.reset.brandDoneText') : $t('auth.reset.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <form
        v-if="!done && token"
        class="pro-login-form"
        data-testid="reset-form"
        method="post"
        action="#"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.reset.title') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.reset.subtitle') }}</p>
        <ProInput
          v-model="password"
          :label="$t('auth.reset.password')"
          type="password"
          name="password"
          autocomplete="new-password"
          required
          test-id="reset-password"
        />
        <ProInput
          v-model="confirmPassword"
          :label="$t('auth.reset.passwordConfirm')"
          type="password"
          name="password-confirm"
          autocomplete="new-password"
          required
          test-id="reset-password-confirm"
        />
        <p class="pro-field-hint">{{ $t('auth.reset.passwordHint') }}</p>
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="reset-submit">
          {{ $t('auth.reset.submit') }}
        </ProButton>
        <p class="pro-login-form__footer">
          <NuxtLink to="/login">{{ $t('auth.reset.backToLogin') }}</NuxtLink>
        </p>
      </form>

      <div v-else-if="!done && !token" class="pro-login-form" data-testid="reset-invalid">
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.reset.title') }}</h1>
        <p class="pro-field-error" role="alert">{{ $t('auth.reset.invalidLink') }}</p>
        <p class="pro-login-form__footer">
          <NuxtLink to="/login">{{ $t('auth.reset.backToLogin') }}</NuxtLink>
        </p>
      </div>

      <div v-else class="pro-login-form" data-testid="reset-done">
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.reset.doneTitle') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.reset.doneSubtitle') }}</p>
        <ProButton block test-id="reset-go-login" @click="navigateTo('/login')">
          {{ $t('auth.reset.backToLogin') }}
        </ProButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()
const route = useRoute()

const token = computed(() => String(route.query.token || ''))
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)
const done = ref(false)

onMounted(() => {
  if (!token.value) {
    error.value = t('auth.reset.invalidLink')
  }
})

async function submit() {
  error.value = ''
  if (!token.value) {
    error.value = t('auth.reset.invalidLink')
    return
  }
  if (password.value !== confirmPassword.value) {
    error.value = t('auth.reset.passwordMismatch')
    return
  }
  if (password.value.length < 8) {
    error.value = t('auth.reset.passwordHint')
    return
  }
  loading.value = true
  try {
    await $fetch('/api/auth/reset-password', {
      method: 'POST',
      body: { token: token.value, password: password.value },
    })
    done.value = true
  } catch (e: any) {
    error.value = mapError(e) || t('auth.reset.failed')
  } finally {
    loading.value = false
  }
}
</script>
