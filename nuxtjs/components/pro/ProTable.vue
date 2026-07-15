<template>
  <ProEmptyState
    v-if="empty"
    :title="resolvedEmptyTitle"
    :description="emptyDescription"
  />
  <div v-else class="pro-table-wrap">
    <table class="pro-table">
      <slot />
    </table>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    empty: false,
    emptyTitle: undefined,
    emptyDescription: undefined,
  },
)

const { t } = useI18n()
const resolvedEmptyTitle = computed(() => props.emptyTitle ?? t('components.table.emptyTitle'))
</script>
