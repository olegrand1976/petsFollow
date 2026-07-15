<template>
  <span
    class="pro-avatar"
    :class="[
      `pro-avatar--${size}`,
      { 'pro-avatar--image': !!src },
    ]"
    :aria-hidden="alt ? undefined : true"
    role="img"
    :aria-label="alt || undefined"
  >
    <img v-if="src" :src="src" :alt="alt || ''" class="pro-avatar__img">
    <span v-else>{{ initialsText }}</span>
  </span>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    src?: string | null
    name?: string
    alt?: string
    size?: 'sm' | 'md' | 'lg'
  }>(),
  {
    src: null,
    name: '',
    alt: '',
    size: 'md',
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
