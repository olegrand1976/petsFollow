<template>
  <div data-testid="admin-training-page">
    <ProPageHeader
      :title="$t('training.adminTitle')"
      :subtitle="$t('training.adminSubtitle')"
    />

    <div class="pf-tabs pro-mb-lg">
      <ProButton v-for="tkey in tabs" :key="tkey" :variant="tab === tkey ? 'primary' : 'ghost'" @click="tab = tkey; load()">
        {{ $t(`training.adminTab.${tkey}`) }}
      </ProButton>
    </div>

    <ProCard v-if="tab === 'scripts'" :title="$t('training.adminTab.scripts')">
      <div v-for="s in scripts" :key="s.id" class="pf-row">
        <div>
          <strong>{{ s.title }}</strong>
          <span class="text-muted"> — {{ s.slug }}</span>
          <ProBadge :variant="s.isActive ? 'success' : 'default'">{{ s.isActive ? 'ON' : 'OFF' }}</ProBadge>
        </div>
      </div>
    </ProCard>

    <ProCard v-else-if="tab === 'vet' || tab === 'coach'" :title="$t(`training.adminTab.${tab}`)">
      <div v-for="v in versions" :key="v.id" class="pf-row">
        <div>
          <strong>v{{ v.version }}</strong>
          <ProBadge v-if="v.isCurrent" variant="success">current</ProBadge>
          <span class="text-muted"> — {{ v.source }} — {{ new Date(v.createdAt).toLocaleString() }}</span>
          <p class="pro-hint">{{ v.changelog }}</p>
        </div>
        <ProButton v-if="!v.isCurrent" variant="secondary" @click="restore(v.id)">
          {{ $t('training.restore') }}
        </ProButton>
      </div>
    </ProCard>

    <ProCard v-else-if="tab === 'analyzer'" :title="$t('training.adminTab.analyzer')">
      <h3>{{ $t('training.runs') }}</h3>
      <div v-for="r in runs" :key="r.id" class="pf-row">
        <div>
          <strong>{{ r.status }}</strong>
          <span class="text-muted"> — {{ r.feedbackCount }} retours — {{ new Date(r.startedAt).toLocaleString() }}</span>
        </div>
      </div>
      <h3 class="pro-mt-md">{{ $t('training.feedbacks') }}</h3>
      <div v-for="f in feedbacks" :key="f.id" class="pf-row">
        <div>
          véto {{ f.vetRealism }}/5 · coach {{ f.coachUsefulness }}/5 · {{ f.difficultyFelt }}
          <p class="pro-hint">{{ f.comment || '—' }}</p>
        </div>
      </div>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const tab = ref<'scripts' | 'vet' | 'coach' | 'analyzer'>('scripts')
const tabs = ['scripts', 'vet', 'coach', 'analyzer'] as const
const scripts = ref<any[]>([])
const versions = ref<any[]>([])
const runs = ref<any[]>([])
const feedbacks = ref<any[]>([])

async function load() {
  if (tab.value === 'scripts') {
    const res: any = await $fetch('/api/admin/pitch-scripts')
    scripts.value = res.data ?? res
  } else if (tab.value === 'vet' || tab.value === 'coach') {
    const kind = tab.value === 'vet' ? 'vet_live' : 'coach'
    const res: any = await $fetch(`/api/admin/agent-prompts/${kind}/versions`)
    versions.value = res.data ?? res
  } else {
    const [r1, r2]: any[] = await Promise.all([
      $fetch('/api/admin/pitch-analyzer/runs'),
      $fetch('/api/admin/pitch-feedback'),
    ])
    runs.value = r1.data ?? r1
    feedbacks.value = r2.data ?? r2
  }
}

async function restore(id: string) {
  await $fetch(`/api/admin/agent-prompts/versions/${id}/restore`, { method: 'POST' })
  await load()
}

onMounted(load)
</script>

<style scoped>
.pf-tabs { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.pf-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.pro-mb-lg { margin-bottom: 1.25rem; }
.pro-mt-md { margin-top: 1rem; }
.pro-hint { color: var(--pf-vet-muted, #64748b); font-size: 0.9rem; margin: 0.25rem 0 0; }
</style>
