<template>
  <div class="pro-login-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>Vérifiez votre boîte mail</h2>
      <p>
        Nous avons envoyé un lien de confirmation à votre adresse email.
        Cliquez dessus pour activer votre compte cabinet.
      </p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-login-form">
        <PetsFollowLogo variant="default" />
        <h1>Email envoyé</h1>
        <p class="pro-page-header__subtitle">
          Un message de confirmation a été envoyé à
          <strong>{{ email }}</strong>.
        </p>
        <p class="text-muted">
          Ouvrez le lien dans l'email pour confirmer votre inscription, puis connectez-vous
          pour configurer votre fiche cabinet.
        </p>

        <div v-if="devLink" class="pro-dev-link">
          <p class="pro-dev-link__label">Lien de confirmation (dev)</p>
          <NuxtLink :to="devLink" class="pro-dev-link__url">{{ devLink }}</NuxtLink>
        </div>

        <ProButton block class="pro-mt-lg" @click="navigateTo('/login')">
          Aller à la connexion
        </ProButton>
        <ProButton variant="ghost" block @click="navigateTo('/')">
          Retour à l'accueil
        </ProButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const route = useRoute()
const email = computed(() => String(route.query.email || ''))
const devLink = computed(() => route.query.devLink ? String(route.query.devLink) : '')
</script>

<style scoped>
.pro-dev-link {
  margin-top: 1.5rem;
  padding: 1rem;
  background: var(--pf-vet-bg);
  border-radius: var(--pf-vet-radius);
  border: 1px dashed var(--pf-vet-border);
}

.pro-dev-link__label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--pf-vet-text-muted);
  margin-bottom: 0.5rem;
}

.pro-dev-link__url {
  font-family: var(--pf-font-mono, monospace);
  font-size: 0.8rem;
  word-break: break-all;
  color: var(--pf-vet-accent);
}

.pro-mt-lg {
  margin-top: 1.25rem;
}
</style>
