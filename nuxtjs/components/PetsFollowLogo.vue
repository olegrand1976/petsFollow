<template>
  <component :is="linkTo ? 'NuxtLink' : 'div'" :to="linkTo" class="logo" :class="`logo--${variant}`">
    <img
      src="/brand/logo-mark.png"
      :alt="markAlt"
      class="logo__mark"
      :class="{ 'logo__mark--animated': animated }"
      :width="markSize"
      :height="markSize"
    />
    <span v-if="showPro" class="logo__pro">Pro</span>
  </component>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    variant?: 'default' | 'compact' | 'hero'
    animated?: boolean
    linkTo?: string
  }>(),
  {
    variant: 'default',
    animated: false,
    linkTo: undefined,
  },
)

/** Le mark inclut déjà « Petsfollow » — n’afficher que le suffixe Pro hors compact/hero. */
const showPro = computed(() => props.variant === 'default')

const markAlt = computed(() => (showPro.value ? 'petsFollow' : 'petsFollow'))

const markSize = computed(() => {
  switch (props.variant) {
    case 'hero':
      return 120
    case 'compact':
      return 36
    default:
      return 48
  }
})
</script>

<style scoped>
.logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  text-decoration: none;
  color: inherit;
}

.logo--default {
  margin-bottom: 1.5rem;
}

.logo--compact {
  gap: 0.5rem;
}

.logo--hero {
  gap: 1rem;
  margin-bottom: 0;
}

.logo__mark {
  display: block;
  border-radius: 50%;
  flex-shrink: 0;
}

.logo__pro {
  font-weight: 600;
  font-style: normal;
  color: var(--pf-brand-teal, #2A9D8F);
  font-size: 1.15rem;
}

.logo--hero .logo__pro {
  color: var(--pf-brand-gold, #E9C46A);
}

.logo__mark--animated {
  animation: pro-emblem-float 3s ease-in-out infinite;
}

@keyframes pro-emblem-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}
</style>
