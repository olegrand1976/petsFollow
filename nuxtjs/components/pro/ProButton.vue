<template>
  <button
    :type="type"
    class="pro-btn"
    :class="[`pro-btn--${variant}`, { 'pro-btn--block': block }]"
    :disabled="disabled || loading"
    :data-testid="testId"
  >
    <span v-if="loading" aria-hidden="true">…</span>
    <slot />
  </button>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost'
    type?: 'button' | 'submit' | 'reset'
    block?: boolean
    disabled?: boolean
    loading?: boolean
    testId?: string
  }>(),
  {
    variant: 'primary',
    type: 'button',
    block: false,
    disabled: false,
    loading: false,
  },
)
</script>
