<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('auth.contactPhone.brandTitle') }}</h2>
      <p>{{ $t('auth.contactPhone.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <form
        class="pro-login-form"
        data-testid="complete-contact-phone-form"
        method="post"
        action="#"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>{{ $t('auth.contactPhone.title') }}</h1>
        <p class="pro-page-header__subtitle">{{ $t('auth.contactPhone.subtitle') }}</p>
        <ProInput
          v-model="phone"
          :label="$t('auth.contactPhone.phone')"
          type="tel"
          name="phone"
          autocomplete="tel"
          required
          test-id="complete-contact-phone"
        />
        <p class="pro-field-hint">{{ $t('auth.contactPhone.hint') }}</p>
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="complete-contact-phone-submit">
          {{ $t('auth.contactPhone.submit') }}
        </ProButton>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()
const { fetchUser } = useProUser()

const phone = ref('')
const error = ref('')
const loading = ref(false)

async function redirectAfterSave() {
  try {
    const me = await fetchUser(true)
    await navigateTo(homePathForRole(me?.role, { profileComplete: me?.profileComplete }))
  } catch {
    await navigateTo('/login')
  }
}

async function submit() {
  error.value = ''
  const value = phone.value.trim()
  if (value.length < 6) {
    error.value = t('auth.contactPhone.hint')
    return
  }
  loading.value = true
  try {
    await $fetch('/api/me', {
      method: 'PATCH',
      body: { contactPhone: value },
    })
    await redirectAfterSave()
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    loading.value = false
  }
}
</script>
