<template>
  <div data-testid="admin-pacs-page">
    <ProPageHeader :title="$t('pacs.admin.title')" :subtitle="$t('pacs.admin.subtitle')">
      <template #actions>
        <ProBadge variant="warning" data-testid="admin-pacs-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton
          data-testid="admin-pacs-wake"
          :disabled="!pacsOn || waking || metrics?.status?.state === 'starting'"
          :loading="waking"
          @click="wakeAdmin"
        >
          {{ waking || metrics?.status?.state === 'starting' ? $t('pacs.launch.waking') : $t('pacs.wake') }}
        </ProButton>
        <ProButton data-testid="admin-pacs-refresh" :disabled="loading" @click="load">
          {{ $t('pacs.admin.refresh') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p v-if="!pacsOn" class="pro-hint" data-testid="admin-pacs-disabled">
      PACS_ENABLED / NUXT_PUBLIC_PACS_ENABLED off
    </p>

    <template v-else>
      <p v-if="loadError" class="pro-error" role="alert">{{ loadError }}</p>

      <div class="admin-pacs-grid">
        <ProCard :title="$t('pacs.admin.metricsTitle')" data-testid="admin-pacs-metrics">
          <dl class="admin-pacs-metrics">
            <div>
              <dt>{{ $t('pacs.admin.state') }}</dt>
              <dd>
                <ProBadge :variant="badgeVariant">{{ stateLabel }}</ProBadge>
              </dd>
            </div>
            <div>
              <dt>{{ $t('pacs.admin.latency') }}</dt>
              <dd>{{ metrics?.status?.latencyMs != null ? `${metrics.status.latencyMs} ms` : '—' }}</dd>
            </div>
            <div>
              <dt>{{ $t('pacs.admin.redis') }}</dt>
              <dd>{{ metrics?.redisConnected ? 'ok' : 'off' }}</dd>
            </div>
            <div>
              <dt>{{ $t('pacs.admin.orthancUrl') }}</dt>
              <dd>{{ metrics?.orthancUrlSet ? 'yes' : 'no' }}</dd>
            </div>
          </dl>
        </ProCard>

        <ProCard :title="$t('pacs.admin.logsTitle')" data-testid="admin-pacs-logs">
          <ProTable :empty="!logs.length" :empty-title="$t('pacs.admin.emptyLogs')">
            <thead>
              <tr>
                <th>{{ $t('pacs.admin.colAt') }}</th>
                <th>{{ $t('pacs.admin.colLevel') }}</th>
                <th>{{ $t('pacs.admin.colEvent') }}</th>
                <th>{{ $t('pacs.admin.colMessage') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, i) in logs" :key="i" data-testid="admin-pacs-log-row">
                <td>{{ row.at }}</td>
                <td>{{ row.level }}</td>
                <td>{{ row.event }}</td>
                <td>{{ row.message }}{{ row.detail ? ` — ${row.detail}` : '' }}</td>
              </tr>
            </tbody>
          </ProTable>
        </ProCard>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { isPublicFlagOn } from '~/utils/public-feature-flag'

definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const runtimeConfig = useRuntimeConfig()
const pacsOn = computed(() => isPublicFlagOn(runtimeConfig.public.pacsEnabled))

const loading = ref(false)
const waking = ref(false)
const loadError = ref('')
const metrics = ref<any>(null)
const logs = ref<any[]>([])

const badgeVariant = computed(() => {
  const s = metrics.value?.status?.state
  if (s === 'ready') return 'success'
  if (s === 'starting') return 'warning'
  return 'neutral'
})

const stateLabel = computed(() => {
  const s = metrics.value?.status?.state || 'offline'
  return t(`pacs.state.${s}`)
})

async function load() {
  if (!pacsOn.value) return
  loading.value = true
  loadError.value = ''
  try {
    const [m, l]: any[] = await Promise.all([
      $fetch('/api/admin/pacs/metrics'),
      $fetch('/api/admin/pacs/logs?limit=100'),
    ])
    metrics.value = m.data ?? m
    logs.value = (l.data ?? l) || []
    if (metrics.value?.status?.state === 'ready') waking.value = false
  } catch (e: any) {
    loadError.value = e?.data?.message || e?.message || 'load_failed'
  } finally {
    loading.value = false
  }
}

async function wakeAdmin() {
  if (!pacsOn.value || waking.value) return
  waking.value = true
  loadError.value = ''
  try {
    const res: any = await $fetch('/api/admin/pacs/wake', { method: 'POST', body: {} })
    const data = res.data ?? res
    if (metrics.value) {
      metrics.value = { ...metrics.value, status: data }
    }
    await load()
  } catch (e: any) {
    waking.value = false
    loadError.value = e?.data?.message || e?.message || 'wake_failed'
  }
}

let poll: ReturnType<typeof setInterval> | null = null
onMounted(() => {
  void load()
  poll = setInterval(() => { void load() }, 5000)
})
onBeforeUnmount(() => {
  if (poll) clearInterval(poll)
})
</script>

<style scoped>
.admin-pacs-grid {
  display: grid;
  gap: 1.25rem;
}
.admin-pacs-metrics {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 1rem;
  margin: 0;
}
.admin-pacs-metrics dt {
  font-size: 0.8rem;
  opacity: 0.7;
}
.admin-pacs-metrics dd {
  margin: 0.25rem 0 0;
  font-weight: 600;
}
</style>
