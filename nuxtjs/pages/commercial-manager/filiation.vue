<template>
  <div data-testid="manager-filiation-page">
    <ProPageHeader
      :title="$t('filiation.title')"
      :subtitle="$t('filiation.subtitle.manager')"
    />
    <ProCard>
      <ProFiliationTable api-path="/api/commercial-manager/filiation" :show-org-cols="true" />
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })
</script>
