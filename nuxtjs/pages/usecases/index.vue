<template>
  <UsecasesIndex base-path="/usecases" />
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['staging-usecases'],
})
</script>
