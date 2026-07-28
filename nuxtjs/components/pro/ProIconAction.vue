<template>
  <NuxtLink
    v-if="to"
    :to="to"
    class="pro-icon-action"
    :class="[`pro-icon-action--${variant}`, { 'pro-icon-action--disabled': disabled }]"
    :aria-label="label"
    :title="label"
    :aria-disabled="disabled || undefined"
    :tabindex="disabled ? -1 : undefined"
    :data-testid="testId"
    @click="onLinkClick"
  >
    <ProIcon :name="icon" :size="size" />
  </NuxtLink>
  <a
    v-else-if="href"
    :href="href"
    class="pro-icon-action"
    :class="[`pro-icon-action--${variant}`]"
    :aria-label="label"
    :title="label"
    :target="hrefTarget"
    :rel="hrefTarget === '_blank' ? 'noopener noreferrer' : undefined"
    :data-testid="testId"
    @click="onClick"
  >
    <ProIcon :name="icon" :size="size" />
  </a>
  <button
    v-else
    type="button"
    class="pro-icon-action"
    :class="[`pro-icon-action--${variant}`]"
    :aria-label="label"
    :title="label"
    :disabled="disabled"
    :data-testid="testId"
    @click="onClick"
  >
    <ProIcon :name="icon" :size="size" />
  </button>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    /** Material Symbols ligature */
    icon: string
    /** Accessible label + native tooltip */
    label: string
    to?: string
    href?: string
    hrefTarget?: '_blank' | '_self'
    disabled?: boolean
    testId?: string
    variant?: 'ghost' | 'danger'
    size?: number
  }>(),
  {
    disabled: false,
    variant: 'ghost',
    size: 20,
    hrefTarget: '_blank',
  },
)

const emit = defineEmits<{ click: [MouseEvent] }>()

function onClick(event: MouseEvent) {
  event.stopPropagation()
  emit('click', event)
}

function onLinkClick(event: MouseEvent) {
  event.stopPropagation()
  if ((event.currentTarget as HTMLElement).getAttribute('aria-disabled') === 'true') {
    event.preventDefault()
    return
  }
  emit('click', event)
}
</script>
