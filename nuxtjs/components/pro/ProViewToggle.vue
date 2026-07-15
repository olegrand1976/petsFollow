<template>
  <div class="pro-view-toggle" role="group" :aria-label="ariaLabel ?? $t('components.viewToggle.ariaLabel')">
    <button
      type="button"
      class="pro-view-toggle__btn"
      :class="{ 'pro-view-toggle__btn--active': modelValue === 'table' }"
      @click="emit('update:modelValue', 'table')"
    >
      {{ $t('components.viewToggle.table') }}
    </button>
    <button
      type="button"
      class="pro-view-toggle__btn"
      :class="{ 'pro-view-toggle__btn--active': modelValue === 'kanban' }"
      @click="emit('update:modelValue', 'kanban')"
    >
      {{ $t('components.viewToggle.kanban') }}
    </button>
  </div>
</template>

<script setup lang="ts">
export type ProViewMode = 'table' | 'kanban'

defineProps<{
  modelValue: ProViewMode
  ariaLabel?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: ProViewMode] }>()
</script>
