<template>
  <ProEmptyState
    v-if="empty"
    :title="emptyTitle"
    :description="emptyDescription"
  />
  <div v-else class="pro-table-wrap">
    <table class="pro-table">
      <slot />
    </table>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    empty: false,
    emptyTitle: 'Aucune donnée',
    emptyDescription: undefined,
  },
)
</script>
