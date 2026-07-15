<template>
  <section class="pro-kanban-column">
    <header class="pro-kanban-column__header">
      <h3 class="pro-kanban-column__title">{{ title }}</h3>
      <span class="pro-kanban-column__count">{{ count }}</span>
    </header>
    <div class="pro-kanban-column__cards">
      <slot />
      <ProEmptyState
        v-if="empty"
        :title="emptyTitle"
        :description="emptyDescription"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    count: number
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    empty: false,
    emptyTitle: 'Aucun élément',
    emptyDescription: undefined,
  },
)
</script>
