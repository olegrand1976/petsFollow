<template>
  <div data-testid="manager-leaderboard-page">
    <ProPageHeader
      :title="$t('manager.leaderboard.title')"
      :subtitle="$t('manager.leaderboard.subtitle')"
    />

    <ProCard class="pro-mb-lg">
      <div class="pro-field pro-field-inline pro-mb-md">
        <label class="pro-label" for="lb-period">{{ $t('manager.leaderboard.period') }}</label>
        <input
          id="lb-period"
          v-model="periodYm"
          class="pro-input"
          type="month"
          data-testid="manager-leaderboard-period"
        >
      </div>

      <ProEmptyState v-if="loadError" :title="$t('manager.leaderboard.loadError')" />
      <p v-else-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>

      <ProTable
        v-else
        :empty="!rows.length"
        :empty-title="$t('manager.leaderboard.empty')"
      >
        <thead>
          <tr>
            <th>{{ $t('manager.leaderboard.colRank') }}</th>
            <th>{{ $t('manager.leaderboard.colCommercial') }}</th>
            <th>{{ $t('manager.leaderboard.colConverted') }}</th>
            <th>{{ $t('manager.leaderboard.colEarned') }}</th>
            <th>{{ $t('manager.leaderboard.colQuotaActivations') }}</th>
            <th>{{ $t('manager.leaderboard.colQuotaEarned') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="(m, idx) in rows" :key="m.userId" :data-testid="`manager-lb-row-${m.userId}`">
            <td>{{ idx + 1 }}</td>
            <td>
              <NuxtLink :to="`/commercial-manager/member/${m.userId}`">{{ m.fullName }}</NuxtLink>
              <br><span class="pro-hint">{{ m.email }}</span>
            </td>
            <td>{{ m.prospectsConverted }}</td>
            <td>{{ formatCurrency(m.monthEarnedCents) }}</td>
            <td>
              <span>{{ quotaFor(m.userId)?.actualActivations ?? 0 }} / </span>
              <input
                class="pro-input pf-quota-input"
                type="number"
                min="0"
                :value="editTargets[m.userId]?.activations ?? quotaFor(m.userId)?.targetActivations ?? 0"
                @input="(e) => setTarget(m.userId, 'activations', (e.target as HTMLInputElement).value)"
              >
            </td>
            <td>
              <span>{{ formatCurrency(quotaFor(m.userId)?.actualEarnedCents ?? 0) }} / </span>
              <input
                class="pro-input pf-quota-input"
                type="number"
                min="0"
                step="0.01"
                :aria-label="$t('manager.leaderboard.quotaEuros')"
                :value="earnedEurosDisplay(m.userId)"
                @input="(e) => setTargetEuros(m.userId, (e.target as HTMLInputElement).value)"
              >
              <span class="pro-hint">€</span>
            </td>
            <td>
              <ProButton
                variant="secondary"
                :test-id="`manager-quota-save-${m.userId}`"
                :disabled="savingId === m.userId"
                @click="saveQuota(m.userId)"
              >
                {{ $t('manager.leaderboard.saveQuota') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
      <p v-if="saveMsg" class="pro-hint pro-mt-md">{{ saveMsg }}</p>
      <p v-if="saveError" class="pro-field-error pro-mt-md" role="alert">{{ saveError }}</p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const { formatCurrency } = useFormatters()
const { t } = useI18n()

function currentPeriodYm() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

const periodYm = ref(currentPeriodYm())
const rows = ref<any[]>([])
const quotas = ref<any[]>([])
const loading = ref(true)
const loadError = ref(false)
const savingId = ref('')
const saveMsg = ref('')
const saveError = ref('')
const editTargets = reactive<Record<string, { activations: number; earnedCents: number }>>({})

function quotaFor(userId: string) {
  return quotas.value.find((q: any) => q.userId === userId)
}

function earnedEurosDisplay(userId: string) {
  const cents = editTargets[userId]?.earnedCents ?? quotaFor(userId)?.targetEarnedCents ?? 0
  return (Number(cents) / 100).toFixed(2)
}

function setTarget(userId: string, field: 'activations', raw: string) {
  const q = quotaFor(userId)
  if (!editTargets[userId]) {
    editTargets[userId] = {
      activations: q?.targetActivations ?? 0,
      earnedCents: q?.targetEarnedCents ?? 0,
    }
  }
  editTargets[userId][field] = Number(raw) || 0
}

function setTargetEuros(userId: string, raw: string) {
  const q = quotaFor(userId)
  if (!editTargets[userId]) {
    editTargets[userId] = {
      activations: q?.targetActivations ?? 0,
      earnedCents: q?.targetEarnedCents ?? 0,
    }
  }
  const euros = Number(String(raw).replace(',', '.')) || 0
  editTargets[userId].earnedCents = Math.round(euros * 100)
}

async function load() {
  loading.value = true
  loadError.value = false
  saveMsg.value = ''
  saveError.value = ''
  try {
    const query = { periodYm: periodYm.value }
    const [lbRes, qRes]: any[] = await Promise.all([
      $fetch('/api/commercial-manager/leaderboard', { query }),
      $fetch('/api/commercial-manager/quotas', { query }),
    ])
    rows.value = lbRes.data ?? lbRes ?? []
    quotas.value = qRes.data ?? qRes ?? []
    for (const key of Object.keys(editTargets)) {
      delete editTargets[key]
    }
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
}

async function saveQuota(userId: string) {
  savingId.value = userId
  saveMsg.value = ''
  saveError.value = ''
  const q = quotaFor(userId)
  const targets = editTargets[userId] ?? {
    activations: q?.targetActivations ?? 0,
    earnedCents: q?.targetEarnedCents ?? 0,
  }
  try {
    await $fetch(`/api/commercial-manager/quotas/${userId}`, {
      method: 'PUT',
      body: {
        periodYm: periodYm.value,
        targetActivations: targets.activations,
        targetEarnedCents: targets.earnedCents,
      },
    })
    saveMsg.value = t('manager.leaderboard.quotaSaved')
    await load()
  } catch {
    saveError.value = t('manager.leaderboard.quotaFailed')
  } finally {
    savingId.value = ''
  }
}

watch(periodYm, load)
onMounted(load)
</script>

<style scoped>
.pf-quota-input {
  display: inline-block;
  width: 5.5rem;
  margin-left: 0.25rem;
}
.pro-mt-md { margin-top: 0.75rem; }
.pro-mb-md { margin-bottom: 0.75rem; }
</style>
