<template>
  <ProCard>
    <div class="pro-kpi">
      <span class="pro-kpi__value">{{ value }}</span>
      <span class="pro-kpi__label">{{ label }}</span>
      <span v-if="trend" class="pro-kpi__label">{{ trend }}</span>
    </div>
  </ProCard>
</template>

<script setup lang="ts">
defineProps<{ value: string | number; label: string; trend?: string }>()
</script>
