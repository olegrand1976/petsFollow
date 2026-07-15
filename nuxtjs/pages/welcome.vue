<template>
  <div class="pro-welcome">
    <header class="pro-welcome__header">
      <PetsFollowLogo variant="default" />
      <ProButton @click="navigateTo('/login')">Se connecter</ProButton>
    </header>

    <section class="pro-welcome__hero">
      <span class="pro-landing__badge">Bienvenue sur petsFollow Pro</span>
      <h1>Votre cabinet cardiaque, prêt à démarrer</h1>
      <p>
        petsFollow Pro vous accompagne au quotidien : relevés cardiaques, alertes, messagerie
        et dossiers patients — le tout <strong>gratuit pour les vétérinaires</strong>.
      </p>
    </section>

    <section class="pro-welcome__steps">
      <h2>Comment ça marche ?</h2>
      <ol class="pro-welcome__step-list">
        <li v-for="(step, i) in steps" :key="step.title" class="pro-welcome__step">
          <span class="pro-welcome__step-num">{{ i + 1 }}</span>
          <div>
            <h3>{{ step.title }}</h3>
            <p>{{ step.text }}</p>
          </div>
        </li>
      </ol>
    </section>

    <section class="pro-welcome__features">
      <div v-for="item in highlights" :key="item.title" class="pro-welcome__highlight">
        <span aria-hidden="true">{{ item.icon }}</span>
        <h3>{{ item.title }}</h3>
        <p>{{ item.text }}</p>
      </div>
    </section>

    <section class="pro-welcome__cta">
      <h2>Prochaine étape : configurez votre fiche cabinet</h2>
      <p>
        À la première connexion, renseignez les coordonnées de votre cabinet pour que vos clients
        puissent vous contacter facilement.
      </p>
      <div class="pro-welcome__cta-actions">
        <ProButton @click="navigateTo('/login')">Me connecter et configurer</ProButton>
        <ProButton variant="ghost" @click="navigateTo('/')">Retour à l'accueil</ProButton>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const steps = [
  {
    title: 'Configurez votre fiche cabinet',
    text: 'Renseignez téléphone, adresse et email de contact lors de votre première connexion.',
  },
  {
    title: 'Invitez vos clients',
    text: 'Vos clients s\'inscrivent via l\'app mobile petsFollow et vous rejoignent automatiquement.',
  },
  {
    title: 'Suivez les relevés cardiaques',
    text: 'Recevez les BPM, validez les sessions et intervenez en cas d\'alerte.',
  },
  {
    title: 'Communiquez en toute simplicité',
    text: 'Messagerie intégrée avec statut de disponibilité configurable.',
  },
]

const highlights = [
  { icon: '❤️', title: 'Relevés BPM', text: 'Suivi cardiaque guidé pour chiens et chats.' },
  { icon: '🔔', title: 'Alertes', text: 'Notification immédiate en cas d\'anomalie.' },
  { icon: '🆓', title: 'Gratuit véto', text: 'Aucun coût pour votre cabinet.' },
]
</script>
