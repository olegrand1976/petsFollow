<template>
  <PresentationWizard />
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['presentation'],
})

useHead({
  title: 'petsFollow — Présentation cabinet',
})
</script>
