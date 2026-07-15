<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>Rejoignez petsFollow Pro — gratuit pour les vétérinaires</h2>
      <p>
        Créez votre cabinet en quelques minutes et commencez à suivre vos patients cardiaques
        dès aujourd'hui.
      </p>
    </aside>
    <div class="pro-login-form-panel">
      <form
        class="pro-login-form"
        data-testid="register-form"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>Inscription cabinet</h1>
        <p class="pro-page-header__subtitle">100 % gratuit pour les professionnels vétérinaires.</p>

        <ProInput
          v-model="fullName"
          label="Votre nom"
          name="fullName"
          autocomplete="name"
          required
          test-id="register-fullname"
        />
        <ProInput
          v-model="practiceName"
          label="Nom du cabinet"
          name="practiceName"
          required
          test-id="register-practice"
        />
        <ProInput
          v-model="email"
          label="Email professionnel"
          type="email"
          name="email"
          autocomplete="email"
          required
          test-id="register-email"
        />
        <ProInput
          v-model="password"
          label="Mot de passe"
          type="password"
          name="password"
          autocomplete="new-password"
          required
          test-id="register-password"
        />
        <p class="pro-field-hint">Minimum 8 caractères</p>

        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

        <ProButton type="submit" block :loading="loading" test-id="register-submit">
          Créer mon compte
        </ProButton>

        <p class="pro-login-form__footer">
          Déjà inscrit ?
          <NuxtLink to="/login">Se connecter</NuxtLink>
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

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
    const msg = e?.data?.message || e?.data?.error || 'Inscription impossible'
    error.value = typeof msg === 'string' ? msg : 'Inscription impossible'
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
