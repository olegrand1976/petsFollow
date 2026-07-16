<template>
  <div data-testid="admin-prospects-page">
    <ProPageHeader :title="$t('admin.prospects.title')" :subtitle="$t('admin.prospects.subtitle')" />
    <ProCard>
      <ProTable :empty="!rows.length" :empty-title="$t('admin.prospects.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.prospects.columnPractice') }}</th>
            <th>{{ $t('admin.prospects.columnContact') }}</th>
            <th>{{ $t('admin.prospects.columnStatus') }}</th>
            <th>{{ $t('admin.prospects.columnCommercial') }}</th>
            <th>{{ $t('admin.prospects.columnDays') }}</th>
            <th>{{ $t('admin.prospects.columnCity') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in rows" :key="p.id">
            <td>{{ p.practiceName }}</td>
            <td>{{ p.contactName }}</td>
            <td><ProBadge variant="neutral">{{ p.status }}</ProBadge></td>
            <td>{{ p.commercialName || p.commercialEmail }}</td>
            <td>{{ p.daysInStatus }}</td>
            <td>{{ p.city }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const rows = ref<any[]>([])

onMounted(async () => {
  const res: any = await $fetch('/api/admin/prospects')
  rows.value = res.data ?? res ?? []
})
</script>
