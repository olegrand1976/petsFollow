<template>
  <div class="pro-list-toolbar">
    <div class="pro-list-toolbar__filters">
      <slot name="filters" />
    </div>
    <div class="pro-list-toolbar__right">
      <slot name="actions" />
      <ProViewToggle v-if="showViewToggle" :model-value="viewMode" @update:model-value="emit('update:viewMode', $event)" />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProViewMode } from '~/components/pro/ProViewToggle.vue'

withDefaults(
  defineProps<{
    viewMode?: ProViewMode
    showViewToggle?: boolean
  }>(),
  {
    viewMode: 'table',
    showViewToggle: true,
  },
)

const emit = defineEmits<{ 'update:viewMode': [value: ProViewMode] }>()
</script>
