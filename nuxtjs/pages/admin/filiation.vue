<template>
  <div data-testid="admin-filiation-page">
    <ProPageHeader
      :title="$t('filiation.title')"
      :subtitle="$t('filiation.subtitle.admin')"
    />
    <ProCard>
      <ProFiliationTable
        api-path="/api/admin/filiation"
        :show-org-cols="true"
        :show-branch-filter="true"
        :show-commercial-filter="true"
      />
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })
</script>
