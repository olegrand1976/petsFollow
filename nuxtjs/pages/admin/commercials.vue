<template>
  <div data-testid="admin-commercials-page">
    <ProPageHeader :title="$t('admin.commercials.title')" :subtitle="$t('admin.commercials.subtitle')" />
    <ProCard>
      <ProTable :empty="!rows.length" :empty-title="$t('admin.commercials.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commercials.columnName') }}</th>
            <th>{{ $t('admin.commercials.columnEmail') }}</th>
            <th>{{ $t('admin.commercials.columnVets') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in rows" :key="r.userId">
            <td>{{ r.fullName }}</td>
            <td>{{ r.email }}</td>
            <td>{{ r.clientCount }}</td>
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
  const res: any = await $fetch('/api/admin/commercials')
  rows.value = res.data ?? res ?? []
})
</script>
