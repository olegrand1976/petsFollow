<template>
  <header class="pro-page-header">
    <div>
      <h1>{{ title }}</h1>
      <p v-if="subtitle" class="pro-page-header__subtitle">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="pro-page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>()
</script>
