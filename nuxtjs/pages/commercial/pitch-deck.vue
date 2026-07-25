<template>
  <CommercialPitchDeck />
</template>

<script setup lang="ts">
definePageMeta({
  layout: false,
  middleware: 'commercial-only',
})

useHead({
  title: 'petsFollow Pro — Pitch deck B2B',
})
</script>
