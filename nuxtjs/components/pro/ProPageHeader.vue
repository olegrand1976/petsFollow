<template>
  <header class="pro-page-header">
    <div>
      <component :is="titleTag" class="pro-page-header__title">{{ title }}</component>
      <p v-if="subtitle" class="pro-page-header__subtitle">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="pro-page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    /** Use `p` when the page already has a document `<h1>` (print leave-behinds). */
    titleTag?: 'h1' | 'p'
  }>(),
  { titleTag: 'h1' },
)
</script>
