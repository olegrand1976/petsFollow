<template>
  <div class="pro-landing">
    <header class="pro-landing__header">
      <PetsFollowLogo variant="default" />
      <nav class="pro-landing__nav">
        <NuxtLink to="/login" class="pro-landing__nav-link">Connexion</NuxtLink>
        <NuxtLink to="/register">
          <ProButton test-id="landing-cta">S'inscrire gratuitement</ProButton>
        </NuxtLink>
      </nav>
    </header>

    <section class="pro-landing__hero">
      <div class="pro-landing__hero-content">
        <span class="pro-landing__badge">100 % gratuit pour les vétérinaires</span>
        <h1>Le suivi cardiaque de vos patients, simplifié</h1>
        <p class="pro-landing__lead">
          petsFollow Pro centralise relevés cardiaques, dossiers clients et messagerie sécurisée
          dans un seul outil pensé pour les cabinets vétérinaires.
        </p>
        <div class="pro-landing__hero-actions">
          <NuxtLink to="/register">
            <ProButton test-id="landing-hero-cta">Créer mon cabinet — c'est gratuit</ProButton>
          </NuxtLink>
          <NuxtLink to="/login" class="pro-landing__secondary-link">J'ai déjà un compte</NuxtLink>
        </div>
      </div>
      <div class="pro-landing__hero-visual" aria-hidden="true">
        <div class="pro-landing__card pro-landing__card--float">
          <span class="pro-landing__card-icon">❤️</span>
          <strong>Relevés cardiaques</strong>
          <p>Suivi BPM et alertes en temps réel</p>
        </div>
        <div class="pro-landing__card pro-landing__card--float pro-landing__card--delay">
          <span class="pro-landing__card-icon">💬</span>
          <strong>Messagerie sécurisée</strong>
          <p>Échanges direct avec vos clients</p>
        </div>
        <div class="pro-landing__card pro-landing__card--float pro-landing__card--delay2">
          <span class="pro-landing__card-icon">📋</span>
          <strong>Dossiers patients</strong>
          <p>Historique complet par animal</p>
        </div>
      </div>
    </section>

    <section class="pro-landing__features">
      <h2>Pourquoi choisir petsFollow Pro ?</h2>
      <div class="pro-landing__feature-grid">
        <article v-for="feature in features" :key="feature.title" class="pro-landing__feature">
          <span class="pro-landing__feature-icon" aria-hidden="true">{{ feature.icon }}</span>
          <h3>{{ feature.title }}</h3>
          <p>{{ feature.text }}</p>
        </article>
      </div>
    </section>

    <section class="pro-landing__cta">
      <div class="pro-landing__cta-inner">
        <h2>Rejoignez petsFollow Pro dès aujourd'hui</h2>
        <p>
          L'application est <strong>entièrement gratuite pour les vétérinaires</strong>.
          Vos clients paient uniquement l'abonnement de suivi cardiaque de leurs animaux.
        </p>
        <NuxtLink to="/register">
          <ProButton variant="secondary">Commencer l'inscription</ProButton>
        </NuxtLink>
      </div>
    </section>

    <footer class="pro-landing__footer">
      <PetsFollowLogo variant="compact" />
      <p>© {{ year }} petsFollow — Suivi cardiaque vétérinaire</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const year = new Date().getFullYear()

const features = [
  {
    icon: '❤️',
    title: 'Relevés cardiaques guidés',
    text: 'Vos clients réalisent des relevés BPM depuis l\'app mobile. Vous validez et suivez l\'évolution.',
  },
  {
    icon: '🔔',
    title: 'Alertes intelligentes',
    text: 'Soyez notifié immédiatement en cas de tachycardie ou d\'anomalie détectée.',
  },
  {
    icon: '💬',
    title: 'Messagerie intégrée',
    text: 'Communiquez avec vos clients sans quitter la plateforme. Disponibilité configurable.',
  },
  {
    icon: '🆓',
    title: 'Gratuit pour les vétos',
    text: 'Aucun abonnement cabinet. petsFollow Pro est offert aux professionnels vétérinaires.',
  },
  {
    icon: '🔒',
    title: 'Données sécurisées',
    text: 'Hébergement conforme, accès restreint par rôle, messagerie chiffrée.',
  },
  {
    icon: '⚡',
    title: 'Prise en main rapide',
    text: 'Inscription en 2 minutes, configuration de votre fiche cabinet au premier accès.',
  },
]
</script>
