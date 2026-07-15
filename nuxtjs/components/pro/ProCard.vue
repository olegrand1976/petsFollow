<template>
  <section class="pro-card">
    <header v-if="title || $slots.header" class="pro-card__header">
      <h2 v-if="title" class="pro-card__title">{{ title }}</h2>
      <slot name="header" />
    </header>
    <slot />
  </section>
</template>

<script setup lang="ts">
defineProps<{ title?: string }>()
</script>
