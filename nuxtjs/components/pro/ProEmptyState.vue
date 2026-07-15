<template>
  <div class="pro-empty">
    <p class="pro-empty__title">{{ title }}</p>
    <p v-if="description">{{ description }}</p>
    <div v-if="$slots.default" class="pro-empty__actions">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>
