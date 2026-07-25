<template>
  <span
    class="pro-avatar"
    :class="[
      `pro-avatar--${size}`,
      `pro-avatar--fit-${fit}`,
      { 'pro-avatar--image': !!src },
    ]"
    :aria-hidden="alt ? undefined : true"
    role="img"
    :aria-label="alt || undefined"
  >
    <img
      v-if="src"
      :src="src"
      :alt="alt || ''"
      class="pro-avatar__img"
      :style="{ objectFit: fit }"
    >
    <span v-else class="pro-avatar__initials">{{ initialsText }}</span>
  </span>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    src?: string | null
    name?: string
    alt?: string
    size?: 'sm' | 'md' | 'lg' | 'xl'
    fit?: 'cover' | 'contain'
  }>(),
  {
    src: null,
    name: '',
    alt: '',
    size: 'md',
    fit: 'cover',
  },
)

const initialsText = computed(() => {
  const name = props.name?.trim() || '?'
  return name
    .split(/\s+/)
    .map((p) => p[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
})
</script>
