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
        <ProButton
          variant="secondary"
          data-testid="admin-pacs-prune"
          :disabled="!pacsOn || pruning || metrics?.status?.state !== 'ready'"
          :loading="pruning"
          @click="pruneOrphans"
        >
          {{ $t('pacs.admin.pruneOrphans') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p v-if="!pacsOn" class="pro-hint" data-testid="admin-pacs-disabled">
      PACS_ENABLED / NUXT_PUBLIC_PACS_ENABLED off
    </p>

    <template v-else>
      <p v-if="loadError" class="pro-error" role="alert">{{ loadError }}</p>
      <p v-if="pruneMsg" class="pro-hint" data-testid="admin-pacs-prune-result" role="status">{{ pruneMsg }}</p>

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

        <ProCard :title="$t('pacs.admin.playgroundTitle')" data-testid="admin-pacs-playground">
          <p class="pro-hint admin-pacs-playground-hint">{{ $t('pacs.admin.playgroundHint') }}</p>
          <div class="admin-pacs-playground-bar">
            <label class="admin-pacs-pet-label">
              <span>{{ $t('pacs.admin.petSelect') }}</span>
              <select
                v-model="selectedPetId"
                data-testid="admin-pacs-pet-select"
                :disabled="!playgroundPets.length"
              >
                <option disabled value="">{{ $t('pacs.admin.petPlaceholder') }}</option>
                <option v-for="p in playgroundPets" :key="p.id" :value="p.id">
                  {{ p.name }} · {{ p.species }} · {{ p.id.slice(0, 8) }}
                </option>
              </select>
            </label>
            <label class="admin-pacs-pet-label">
              <span>{{ $t('pacs.admin.petIdManual') }}</span>
              <input
                v-model="manualPetId"
                type="text"
                data-testid="admin-pacs-pet-id-input"
                :placeholder="$t('pacs.admin.petIdPlaceholder')"
                autocomplete="off"
                spellcheck="false"
              >
            </label>
            <ProButton
              variant="secondary"
              data-testid="admin-pacs-apply-pet"
              :disabled="!manualPetId.trim()"
              @click="applyManualPet"
            >
              {{ $t('pacs.admin.applyPet') }}
            </ProButton>
            <ProButton
              variant="secondary"
              data-testid="admin-pacs-reload-pets"
              :disabled="petsLoading"
              @click="loadPlaygroundPets"
            >
              {{ $t('pacs.admin.reloadPets') }}
            </ProButton>
          </div>
          <p v-if="petsError" class="pro-error" role="alert">{{ petsError }}</p>
          <p v-if="!activePetId" class="pro-hint" data-testid="admin-pacs-no-pet">
            {{ $t('pacs.admin.noPetSelected') }}
          </p>
          <div v-else data-testid="admin-pacs-viewer-mount" class="admin-pacs-viewer-mount">
            <ClientOnly>
              <PacsViewerContainer
                :key="activePetId"
                :pet-id="activePetId"
                debug
                auto-open-first-study
                @debug="onViewerDebug"
              />
              <template #fallback>
                <p class="pro-hint">{{ $t('pacs.admin.viewerLoading') }}</p>
              </template>
            </ClientOnly>
          </div>
        </ProCard>

        <ProCard :title="$t('pacs.admin.debugTitle')" data-testid="admin-pacs-debug">
          <div class="admin-pacs-debug-actions">
            <ProButton variant="secondary" data-testid="admin-pacs-debug-clear" @click="clearDebug">
              {{ $t('pacs.admin.debugClear') }}
            </ProButton>
          </div>
          <ProTable :empty="!debugEntries.length" :empty-title="$t('pacs.admin.debugEmpty')">
            <thead>
              <tr>
                <th>{{ $t('pacs.admin.colAt') }}</th>
                <th>{{ $t('pacs.admin.debugColOk') }}</th>
                <th>{{ $t('pacs.admin.debugColMethod') }}</th>
                <th>{{ $t('pacs.admin.debugColUrl') }}</th>
                <th>{{ $t('pacs.admin.debugColMs') }}</th>
                <th>{{ $t('pacs.admin.colMessage') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, i) in debugEntries" :key="i" data-testid="admin-pacs-debug-row">
                <td>{{ row.at }}</td>
                <td>{{ row.ok ? 'ok' : (row.status || 'err') }}</td>
                <td>{{ row.method }}</td>
                <td class="admin-pacs-debug-url">{{ row.url }}</td>
                <td>{{ row.ms }}</td>
                <td>{{ row.message || '—' }}</td>
              </tr>
            </tbody>
          </ProTable>
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
import type { PacsDebugEntry } from '~/composables/usePacsDebugLog'

definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const runtimeConfig = useRuntimeConfig()
const pacsOn = computed(() => isPublicFlagOn(runtimeConfig.public.pacsEnabled))
const { entries: debugEntries, push: pushDebug, clear: clearDebug } = usePacsDebugLog()

const loading = ref(false)
const waking = ref(false)
const pruning = ref(false)
const loadError = ref('')
const pruneMsg = ref('')
const metrics = ref<any>(null)
const logs = ref<any[]>([])

const playgroundPets = ref<{ id: string; name: string; species: string; practiceId: string }[]>([])
const selectedPetId = ref('')
const manualPetId = ref('')
const petsLoading = ref(false)
const petsError = ref('')

const activePetId = computed(() => selectedPetId.value || manualPetId.value.trim())

watch(selectedPetId, (id) => {
  if (id) manualPetId.value = id
})

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

function onViewerDebug(entry: PacsDebugEntry) {
  pushDebug(entry)
}

function applyManualPet() {
  const id = manualPetId.value.trim()
  if (!id) return
  selectedPetId.value = id
  if (!playgroundPets.value.some(p => p.id === id)) {
    // Keep select in sync when the UUID is not in the seed list.
    manualPetId.value = id
  }
}

async function loadPlaygroundPets() {
  if (!pacsOn.value) return
  petsLoading.value = true
  petsError.value = ''
  try {
    const res: any = await $fetch('/api/admin/pacs/playground-pets')
    const data = res.data ?? res
    playgroundPets.value = data.pets || []
    if (!selectedPetId.value || !playgroundPets.value.some(p => p.id === selectedPetId.value)) {
      selectedPetId.value = data.preferredPetId || playgroundPets.value[0]?.id || ''
    }
    if (selectedPetId.value) manualPetId.value = selectedPetId.value
  } catch (e: any) {
    petsError.value = e?.data?.message || e?.message || 'pets_failed'
  } finally {
    petsLoading.value = false
  }
}

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
    const st = metrics.value?.status?.state
    if (st === 'ready' || st === 'offline') waking.value = false
  } catch (e: any) {
    waking.value = false
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

async function pruneOrphans() {
  if (!pacsOn.value || pruning.value) return
  pruning.value = true
  pruneMsg.value = ''
  loadError.value = ''
  try {
    const dryRes: any = await $fetch('/api/admin/pacs/prune-orphans', {
      method: 'POST',
      body: {},
      query: { dryRun: '1' },
    })
    const dry = dryRes.data ?? dryRes
    const n = Array.isArray(dry.pruned) ? dry.pruned.length : 0
    if (n === 0) {
      pruneMsg.value = t('pacs.admin.pruneNone', { kept: dry.kept ?? 0 })
      return
    }
    if (!confirm(t('pacs.admin.pruneConfirm', { n }))) return
    const res: any = await $fetch('/api/admin/pacs/prune-orphans', { method: 'POST', body: {} })
    const data = res.data ?? res
    const pruned = Array.isArray(data.pruned) ? data.pruned.length : 0
    pruneMsg.value = t('pacs.admin.pruneResult', {
      pruned,
      kept: data.kept ?? 0,
      skipped: data.skipped ?? 0,
    })
    if (data.truncated) {
      pruneMsg.value += ` ${t('pacs.admin.pruneTruncated')}`
    }
    await load()
  } catch (e: any) {
    loadError.value = e?.data?.message || e?.message || 'prune_failed'
  } finally {
    pruning.value = false
  }
}

let poll: ReturnType<typeof setInterval> | null = null
onMounted(() => {
  void load()
  void loadPlaygroundPets()
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
.admin-pacs-playground-hint {
  margin: 0 0 0.75rem;
}
.admin-pacs-playground-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-end;
  margin-bottom: 1rem;
}
.admin-pacs-pet-label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: min(100%, 280px);
  flex: 1;
}
.admin-pacs-pet-label select,
.admin-pacs-pet-label input {
  font: inherit;
  padding: 0.45rem 0.6rem;
}
.admin-pacs-viewer-mount {
  margin-top: 0.5rem;
  min-height: 360px;
}
.admin-pacs-debug-actions {
  margin-bottom: 0.75rem;
}
.admin-pacs-debug-url {
  font-family: ui-monospace, monospace;
  font-size: 0.8rem;
  word-break: break-all;
}
</style>
