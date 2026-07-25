<template>
  <nav class="pro-legal-footer" :aria-label="$t('legal.navAria')">
    <NuxtLink to="/legal/terms">{{ $t('legal.terms.link') }}</NuxtLink>
    <span aria-hidden="true">·</span>
    <NuxtLink to="/legal/privacy">{{ $t('legal.privacy.link') }}</NuxtLink>
    <span aria-hidden="true">·</span>
    <NuxtLink to="/legal/mentions">{{ $t('legal.mentions.link') }}</NuxtLink>
  </nav>
</template>

<style scoped>
.pro-legal-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 0.75rem;
  justify-content: center;
  font-size: 0.875rem;
  margin-top: 1rem;
}

.pro-legal-footer a {
  color: var(--pf-vet-text-muted);
  text-decoration: none;
}

.pro-legal-footer a:hover {
  color: var(--pf-vet-accent);
  text-decoration: underline;
}
</style>
