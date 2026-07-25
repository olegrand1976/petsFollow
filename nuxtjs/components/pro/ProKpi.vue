<template>
  <component
    :is="to ? NuxtLink : 'div'"
    v-bind="rootBind"
    class="pro-card pro-kpi-card"
    :class="{
      'pro-card--interactive': Boolean(to),
      'pro-kpi-card--alert': variant === 'alert',
    }"
  >
    <div class="pro-kpi" :class="{ 'pro-kpi--with-icon': Boolean(icon) }">
      <span v-if="icon" class="pro-kpi__icon-wrap" aria-hidden="true">
        <ProIcon :name="icon" :size="22" class="pro-kpi__icon" />
      </span>
      <div class="pro-kpi__body">
        <span class="pro-kpi__value">{{ value }}</span>
        <span class="pro-kpi__label">{{ label }}</span>
        <span v-if="trend" class="pro-kpi__trend">{{ trend }}</span>
      </div>
    </div>
  </component>
</template>

<script setup lang="ts">
import { NuxtLink } from '#components'

const props = withDefaults(
  defineProps<{
    value: string | number
    label: string
    trend?: string
    icon?: string
    to?: string
    variant?: 'default' | 'alert'
  }>(),
  { variant: 'default' },
)

const rootBind = computed(() => (props.to ? { to: props.to } : {}))
</script>
