<template>
  <component :is="linkTo ? 'NuxtLink' : 'div'" :to="linkTo" class="logo" :class="`logo--${variant}`">
    <img
      src="/brand/emblem.svg"
      alt="petsFollow"
      class="logo__emblem"
      :class="{ 'logo__emblem--animated': animated }"
      :width="emblemSize"
      :height="emblemSize"
    />
    <span v-if="showText" class="logo__text">
      <strong>petsFollow</strong> <em>Pro</em>
    </span>
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

const showText = computed(() => props.variant !== 'compact')

const emblemSize = computed(() => {
  switch (props.variant) {
    case 'hero':
      return 72
    case 'compact':
      return 28
    default:
      return 36
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

.logo--hero .logo__text {
  font-size: 1.35rem;
  color: white;
}

.logo--hero .logo__text em {
  color: var(--pf-brand-gold, #E9C46A);
}

.logo__text em {
  color: var(--pf-brand-teal, #2A9D8F);
  font-style: normal;
  font-weight: 600;
}

.logo__emblem--animated {
  animation: pro-emblem-float 3s ease-in-out infinite;
}

@keyframes pro-emblem-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}
</style>
