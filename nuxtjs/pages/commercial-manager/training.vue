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

    <div v-if="memberScores.length" class="pro-grid-kpi pro-mb-lg" data-testid="manager-training-kpis">
      <ProKpi
        v-for="m in memberScores"
        :key="m.userId"
        :value="m.avgLabel"
        :label="m.name"
      />
    </div>

    <ProCard :title="$t('training.teamHistory')">
      <table v-if="list.length" class="pro-table">
        <thead>
          <tr>
            <th>{{ $t('training.colName') }}</th>
            <th>{{ $t('training.colEmail') }}</th>
            <th>{{ $t('training.colDate') }}</th>
            <th>{{ $t('training.colDifficulty') }}</th>
            <th>{{ $t('training.colOutcome') }}</th>
            <th>{{ $t('training.colScore') }}</th>
            <th>{{ $t('training.colNote') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="h in list" :key="h.id">
            <td>
              <NuxtLink
                v-if="h.userId"
                :to="`/commercial-manager/member/${h.userId}`"
              >
                {{ h.userFullName || h.userId?.slice(0, 8) }}
              </NuxtLink>
              <span v-else>—</span>
            </td>
            <td>{{ h.userEmail || '—' }}</td>
            <td>{{ formatDate(h.createdAt) }}</td>
            <td>{{ $t(`training.difficulty.${h.interestLevel}.label`) }}</td>
            <td>{{ $t(`training.outcome.${h.outcome}`) }}</td>
            <td>{{ h.userScore ?? h.aiScore ?? '—' }}</td>
            <td class="pf-note-cell">
              <textarea
                class="pro-input"
                rows="2"
                :value="noteDrafts[h.id] ?? h.managerNote ?? ''"
                :data-testid="`manager-sim-note-${h.id}`"
                @input="(e) => noteDrafts[h.id] = (e.target as HTMLTextAreaElement).value"
              />
              <ProButton
                variant="secondary"
                class="pro-mt-sm"
                :disabled="savingId === h.id"
                :test-id="`manager-sim-note-save-${h.id}`"
                @click="saveNote(h.id)"
              >
                {{ $t('training.saveNote') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </table>
      <ProEmptyState v-else :title="$t('training.historyEmpty')" />
      <p v-if="noteMsg" class="pro-hint pro-mt-md">{{ noteMsg }}</p>
      <p v-if="noteError" class="pro-field-error pro-mt-md" role="alert">{{ noteError }}</p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const { formatDate } = useFormatters()
const { t } = useI18n()
const list = ref<any[]>([])
const noteDrafts = reactive<Record<string, string>>({})
const savingId = ref('')
const noteMsg = ref('')
const noteError = ref('')

const memberScores = computed(() => {
  const byUser = new Map<string, { name: string; scores: number[] }>()
  for (const h of list.value) {
    const score = Number(h.userScore ?? h.aiScore)
    if (!h.userId || Number.isNaN(score)) continue
    const cur = byUser.get(h.userId) ?? { name: h.userFullName || h.userEmail || h.userId.slice(0, 8), scores: [] }
    cur.scores.push(score)
    byUser.set(h.userId, cur)
  }
  return [...byUser.entries()].map(([userId, v]) => {
    const avg = v.scores.reduce((a, b) => a + b, 0) / v.scores.length
    return { userId, name: `${v.name} (${t('training.avgScore')})`, avgLabel: avg.toFixed(1) }
  })
})

async function load() {
  try {
    const res: any = await $fetch('/api/commercial-manager/pitch-sims')
    list.value = res.data ?? res ?? []
  } catch {
    list.value = []
    noteError.value = t('training.noteFailed')
  }
}

async function saveNote(simId: string) {
  savingId.value = simId
  noteMsg.value = ''
  noteError.value = ''
  const existing = list.value.find((h) => h.id === simId)?.managerNote ?? ''
  const note = noteDrafts[simId] ?? existing
  try {
    await $fetch(`/api/commercial-manager/pitch-sims/${simId}/note`, {
      method: 'PATCH',
      body: { note },
    })
    noteMsg.value = t('training.noteSaved')
    delete noteDrafts[simId]
    await load()
  } catch {
    noteError.value = t('training.noteFailed')
  } finally {
    savingId.value = ''
  }
}

onMounted(load)
</script>

<style scoped>
.pro-mb-lg { margin-bottom: 1.25rem; }
.pro-mt-md { margin-top: 0.75rem; }
.pro-mt-sm { margin-top: 0.35rem; }
.pf-note-cell { min-width: 12rem; }
</style>
