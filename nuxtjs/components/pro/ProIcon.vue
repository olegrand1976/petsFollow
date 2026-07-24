<template>
  <span
    class="material-symbols-outlined pro-icon"
    :style="sizeStyle"
    aria-hidden="true"
  >{{ name }}</span>
</template>

<script setup lang="ts">
const props = defineProps<{
  /** Material Symbols ligature name, e.g. `light_mode`, `dashboard` */
  name: string
  /** Pixel size (number) or CSS length (string). Defaults to inherited / 20px base. */
  size?: number | string
}>()

const sizeStyle = computed(() => {
  if (props.size == null) return undefined
  const value = typeof props.size === 'number' ? `${props.size}px` : props.size
  return { fontSize: value }
})
</script>
