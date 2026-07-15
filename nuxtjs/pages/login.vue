<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>Suivi cardiaque et dossier client en un seul outil</h2>
      <p>
        Bienvenue sur petsFollow Pro — l’espace pensé pour les pros du soin animal,
        avec une touche de bonne humeur.
      </p>
    </aside>
    <div class="pro-login-form-panel">
      <form
        class="pro-login-form"
        data-testid="login-form"
        @submit.prevent="submit"
      >
        <PetsFollowLogo variant="default" />
        <h1>Connexion Pro</h1>
        <p class="pro-page-header__subtitle">Accès réservé aux vétérinaires et administrateurs.</p>
        <ProInput
          v-model="email"
          label="Email"
          type="email"
          name="email"
          autocomplete="email"
          required
          test-id="login-email"
        />
        <ProInput
          v-model="password"
          label="Mot de passe"
          type="password"
          name="password"
          autocomplete="current-password"
          required
          test-id="login-password"
        />
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <ProButton type="submit" block :loading="loading" test-id="login-submit">
          Se connecter
        </ProButton>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const email = ref(import.meta.dev ? 'vet.demo@petsfollow.test' : '')
const password = ref(import.meta.dev ? 'VetDemo123!' : '')
const error = ref('')
const loading = ref(false)
const token = useCookie('pf_token')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const res: any = await $fetch('/api/auth/login', {
      method: 'POST',
      body: { email: email.value, password: password.value },
    })
    token.value = res.data?.accessToken || res.accessToken
    const me: any = await $fetch('/api/me')
    const role = me.data?.role || me.role
    if (role === 'admin') await navigateTo('/admin')
    else if (role === 'vet') await navigateTo('/clients')
    else error.value = 'Accès réservé aux profils Pro (véto / admin)'
  } catch {
    error.value = 'Identifiants invalides'
  } finally {
    loading.value = false
  }
}
</script>
