<template>
  <div data-testid="commercial-network-page">
    <ProPageHeader
      :title="$t('commercial.network.title')"
      :subtitle="$t('commercial.network.subtitle')"
    />

    <ProEmptyState v-if="loadError" :title="$t('commercial.network.loadError')" />
    <p v-else-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>

    <template v-else-if="network">
      <div class="pro-grid-kpi pro-mb-lg">
        <ProKpi
          :value="network.branch?.name || '—'"
          :label="$t('commercial.network.branch')"
        />
        <ProKpi
          :value="personLabel(network.sponsor)"
          :label="$t('commercial.network.sponsor')"
        />
        <ProKpi
          :value="personLabel(network.manager)"
          :label="$t('commercial.network.manager')"
        />
        <ProKpi
          :value="network.salesRank ?? 0"
          :label="$t('commercial.network.rank')"
        />
      </div>

      <ProCard>
        <h3 class="pro-mb-md">{{ $t('commercial.network.downlineTitle') }}</h3>
        <ProTable
          v-if="network.downline?.length"
          :empty="false"
        >
          <thead>
            <tr>
              <th>{{ $t('commercial.network.colName') }}</th>
              <th>{{ $t('commercial.network.colEmail') }}</th>
              <th>{{ $t('commercial.network.colRole') }}</th>
              <th>{{ $t('commercial.network.colDepth') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="n in network.downline" :key="n.userId">
              <td>{{ n.fullName }}</td>
              <td>{{ n.email }}</td>
              <td>{{ n.role }}</td>
              <td>{{ n.depth }}</td>
            </tr>
          </tbody>
        </ProTable>
        <ProEmptyState
          v-else
          :title="network.mlmOrgEnabled ? $t('commercial.network.downlineEmpty') : $t('commercial.network.mlmPending')"
          :description="network.mlmOrgEnabled ? $t('commercial.network.downlineEmptyDesc') : $t('commercial.network.mlmPendingDesc')"
        />
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const network = ref<any>(null)
const loading = ref(true)
const loadError = ref(false)

function personLabel(p?: { fullName?: string; email?: string } | null) {
  if (!p?.fullName) return '—'
  return p.email ? `${p.fullName}` : p.fullName
}

onMounted(async () => {
  loading.value = true
  loadError.value = false
  try {
    const res: any = await $fetch('/api/commercial/network')
    network.value = res.data ?? res
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
})
</script>
