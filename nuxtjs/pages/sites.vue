<template>
  <div
    v-if="sitesUiEnabled"
    data-testid="sites-page"
  >
    <ProPageHeader
      :title="$t('sites.pageTitle')"
      :subtitle="$t('sites.pageSubtitle')"
    />
    <ProSitesManager />
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['vet-only', 'practice-perm'],
  practicePerm: 'calendar.manage',
})

const { sitesUiEnabled } = usePracticeSites()

// Soft-GA off: no dedicated Sites surface (nav already hidden).
if (!sitesUiEnabled.value) {
  await navigateTo('/settings')
}
</script>
