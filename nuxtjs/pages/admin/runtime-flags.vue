<template>
  <div data-testid="admin-runtime-flags-page">
    <ProPageHeader
      :title="$t('admin.runtimeFlags.title')"
      :subtitle="$t('admin.runtimeFlags.subtitle')"
    />
    <ProCard>
      <p v-if="loadError" class="pro-error" role="alert">{{ loadError }}</p>
      <ProTable v-else :empty="!rows.length" :empty-title="$t('admin.runtimeFlags.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.runtimeFlags.colFlag') }}</th>
            <th>{{ $t('admin.runtimeFlags.colValue') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in rows" :key="row.key" data-testid="admin-runtime-flag-row">
            <td><code>{{ row.key }}</code></td>
            <td>
              <ProBadge :variant="badgeVariant(row.value)">{{ formatValue(row.value) }}</ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

const { t } = useI18n()
const flags = ref<Record<string, unknown>>({})
const loadError = ref('')

const rows = computed(() =>
  Object.entries(flags.value).map(([key, value]) => ({ key, value })),
)

function formatValue(v: unknown) {
  if (typeof v === 'boolean') return v ? 'true' : 'false'
  if (v == null || v === '') return '—'
  return String(v)
}

function badgeVariant(v: unknown): 'success' | 'neutral' | 'warning' {
  if (v === true) return 'success'
  if (v === false) return 'neutral'
  return 'warning'
}

onMounted(async () => {
  try {
    const res: any = await $fetch('/api/admin/runtime-flags')
    flags.value = (res?.data ?? res) as Record<string, unknown>
  } catch {
    loadError.value = t('admin.runtimeFlags.loadError')
  }
})
</script>
