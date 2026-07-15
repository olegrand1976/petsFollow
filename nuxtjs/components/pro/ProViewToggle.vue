<template>
  <div class="pro-view-toggle" role="group" :aria-label="ariaLabel">
    <button
      type="button"
      class="pro-view-toggle__btn"
      :class="{ 'pro-view-toggle__btn--active': modelValue === 'table' }"
      @click="emit('update:modelValue', 'table')"
    >
      Liste
    </button>
    <button
      type="button"
      class="pro-view-toggle__btn"
      :class="{ 'pro-view-toggle__btn--active': modelValue === 'kanban' }"
      @click="emit('update:modelValue', 'kanban')"
    >
      Kanban
    </button>
  </div>
</template>

<script setup lang="ts">
export type ProViewMode = 'table' | 'kanban'

withDefaults(
  defineProps<{
    modelValue: ProViewMode
    ariaLabel?: string
  }>(),
  { ariaLabel: 'Mode affichage' },
)

const emit = defineEmits<{ 'update:modelValue': [value: ProViewMode] }>()
</script>
