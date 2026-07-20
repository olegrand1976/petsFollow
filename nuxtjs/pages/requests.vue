<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const route = useRoute()
const tab = typeof route.query.tab === 'string' ? route.query.tab : ''
if (tab === 'links' || tab === 'invitations') {
  await navigateTo('/clients?invitations=1')
} else {
  await navigateTo('/calendar')
}
</script>

<template>
  <div />
</template>
