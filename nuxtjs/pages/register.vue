<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('auth.register.brandTitle') }}</h2>
      <p>{{ $t('auth.register.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <form
        class="pro-login-form"
        data-testid="register-form"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.register.title') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.register.subtitle') }}</p>

        <ProInput
          v-model="fullName"
          :label="$t('auth.register.fullName')"
          name="fullName"
          autocomplete="name"
          required
          test-id="register-fullname"
        />
        <ProInput
          v-model="practiceName"
          :label="$t('auth.register.practiceName')"
          name="practiceName"
          required
          test-id="register-practice"
        />
        <ProInput
          v-model="email"
          :label="$t('auth.register.email')"
          type="email"
          name="email"
          autocomplete="email"
          required
          test-id="register-email"
        />
        <ProInput
          v-model="password"
          :label="$t('auth.register.password')"
          type="password"
          name="password"
          autocomplete="new-password"
          required
          test-id="register-password"
        />
        <p class="pro-field-hint">{{ $t('auth.register.passwordHint') }}</p>

        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

        <ProButton type="submit" block :loading="loading" test-id="register-submit">
          {{ $t('auth.register.submit') }}
        </ProButton>

        <p class="pro-login-form__footer">
          {{ $t('auth.register.alreadyRegistered') }}
          <NuxtLink to="/login">{{ $t('auth.register.loginLink') }}</NuxtLink>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { mapError } = useApiError()

const fullName = ref('')
const practiceName = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res: any = await $fetch('/api/auth/register', {
      method: 'POST',
      body: {
        fullName: fullName.value,
        practiceName: practiceName.value,
        email: email.value,
        password: password.value,
      },
    })
    const data = res.data ?? res
    await navigateTo({
      path: '/register/sent',
      query: { email: email.value, devLink: import.meta.dev ? data.confirmPath : undefined },
    })
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.pro-field-hint {
  font-size: 0.8rem;
  color: var(--pf-vet-text-muted);
  margin-top: -0.5rem;
  margin-bottom: 0.75rem;
}

.pro-login-form__footer {
  margin-top: 1.25rem;
  text-align: center;
  font-size: 0.9rem;
  color: var(--pf-vet-text-muted);
}

.pro-login-form__footer a {
  color: var(--pf-vet-accent);
  font-weight: 600;
}
</style>
