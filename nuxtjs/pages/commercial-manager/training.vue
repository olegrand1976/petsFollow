<template>
  <div data-testid="manager-training-page">
    <ProPageHeader
      :title="$t('training.managerTitle')"
      :subtitle="$t('training.managerSubtitle')"
    />
    <p class="pro-mb-lg">
      <NuxtLink to="/commercial/training">
        <ProButton>{{ $t('training.goTrain') }}</ProButton>
      </NuxtLink>
    </p>
    <ProCard :title="$t('training.teamHistory')">
      <table v-if="list.length" class="pro-table">
        <thead>
          <tr>
            <th>{{ $t('training.colUser') }}</th>
            <th>{{ $t('training.colDate') }}</th>
            <th>{{ $t('training.colDifficulty') }}</th>
            <th>{{ $t('training.colOutcome') }}</th>
            <th>{{ $t('training.colScore') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="h in list" :key="h.id">
            <td>{{ h.userId?.slice(0, 8) || '—' }}</td>
            <td>{{ new Date(h.createdAt).toLocaleString() }}</td>
            <td>{{ $t(`training.difficulty.${h.interestLevel}.label`) }}</td>
            <td>{{ $t(`training.outcome.${h.outcome}`) }}</td>
            <td>{{ h.userScore ?? h.aiScore ?? '—' }}</td>
          </tr>
        </tbody>
      </table>
      <ProEmptyState v-else :title="$t('training.historyEmpty')" />
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const list = ref<any[]>([])

onMounted(async () => {
  const res: any = await $fetch('/api/commercial-manager/pitch-sims')
  list.value = res.data ?? res
})
</script>

<style scoped>
.pro-mb-lg { margin-bottom: 1.25rem; }
</style>
