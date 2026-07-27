<template>
  <div data-testid="commercial-filiation-page">
    <ProPageHeader
      :title="$t('filiation.title')"
      :subtitle="$t('filiation.subtitle.commercial')"
    />
    <ProCard>
      <ProFiliationTable api-path="/api/commercial/filiation" />
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })
</script>
