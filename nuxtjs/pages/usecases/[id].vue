<template>
  <UsecasesDetail base-path="/usecases" :case-id="caseId" />
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['staging-usecases'],
})

const route = useRoute()
const caseId = computed(() => String(route.params.id || ''))
</script>
