<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ confirmed ? 'Compte activé !' : 'Confirmation en cours…' }}</h2>
      <p v-if="confirmed">
        Votre inscription est confirmée. Découvrez petsFollow Pro et configurez votre cabinet.
      </p>
      <p v-else-if="error">
        Ce lien de confirmation n'est plus valide.
      </p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-login-form">
        <PetsFollowLogo variant="default" />
        <div v-if="loading" class="text-muted">Confirmation en cours…</div>
        <template v-else-if="confirmed">
          <h1>Inscription confirmée</h1>
          <p class="pro-page-header__subtitle">
            Bienvenue sur petsFollow Pro{{ confirmedEmail ? `, ${confirmedEmail}` : '' }} !
          </p>
          <ProButton block @click="navigateTo('/welcome')">
            Découvrir l'application
          </ProButton>
          <ProButton variant="secondary" block class="pro-mt-sm" @click="navigateTo('/login')">
            Se connecter
          </ProButton>
        </template>
        <template v-else>
          <h1>Confirmation impossible</h1>
          <p class="pro-field-error" role="alert">{{ error }}</p>
          <ProButton block @click="navigateTo('/register')">Réessayer l'inscription</ProButton>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const route = useRoute()
const loading = ref(true)
const confirmed = ref(false)
const confirmedEmail = ref('')
const error = ref('')

onMounted(async () => {
  const token = String(route.query.token || '')
  if (!token) {
    error.value = 'Lien de confirmation invalide.'
    loading.value = false
    return
  }
  try {
    const res: any = await $fetch('/api/auth/confirm-email', {
      method: 'POST',
      body: { token },
    })
    const data = res.data ?? res
    confirmed.value = true
    confirmedEmail.value = data.email || ''
  } catch (e: any) {
    error.value = e?.data?.message || 'Ce lien a expiré ou a déjà été utilisé.'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.pro-mt-sm {
  margin-top: 0.75rem;
}
</style>
