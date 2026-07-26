<template>
  <div data-testid="manager-ai-modules-page">
    <ProPageHeader
      :title="$t('manager.aiModules.title')"
      :subtitle="$t('manager.aiModules.subtitle')"
    >
      <template #actions>
        <ProButton variant="secondary" @click="navigateTo('/commercial/ai-cr-playbook')">
          {{ $t('manager.aiModules.openPlaybook') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p v-if="loadError" class="pro-field-error pro-mb-lg" role="alert">{{ loadError }}</p>

    <ProCard class="pro-mb-lg" :title="$t('manager.aiModules.frictionTitle')" data-testid="manager-ai-friction">
      <ProEmptyState
        v-if="!alerts.length"
        :title="$t('manager.aiModules.frictionEmpty')"
        :description="$t('manager.aiModules.frictionEmptyDesc')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('manager.aiModules.colCommercial') }}</th>
            <th>{{ $t('manager.aiModules.colPractice') }}</th>
            <th>{{ $t('manager.aiModules.colSignal') }}</th>
            <th>{{ $t('manager.aiModules.colDetail') }}</th>
            <th>{{ $t('manager.aiModules.colWhen') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in alerts" :key="a.id">
            <td>{{ a.commercialName || a.commercialEmail || '—' }}</td>
            <td>{{ a.practiceName }}</td>
            <td>{{ a.signal }}</td>
            <td>{{ a.detail }}</td>
            <td>{{ formatDate(a.createdAt) }}</td>
            <td class="pf-cta-cell">
              <a
                v-if="a.practicePhone"
                class="pro-link"
                :href="`tel:${a.practicePhone}`"
              >{{ $t('manager.aiModules.call') }}</a>
              <a
                v-if="a.practiceContactEmail"
                class="pro-link"
                :href="`mailto:${a.practiceContactEmail}`"
              >{{ $t('manager.aiModules.email') }}</a>
              <NuxtLink
                v-if="a.commercialUserId"
                :to="`/commercial-manager/member/${a.commercialUserId}`"
                class="pro-link"
              >
                {{ $t('manager.aiModules.openMember') }}
              </NuxtLink>
              <span v-if="!a.practicePhone && !a.practiceContactEmail && !a.commercialUserId">—</span>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('manager.aiModules.listTitle')">
      <ProEmptyState
        v-if="!modules.length"
        :title="$t('manager.aiModules.emptyTitle')"
        :description="$t('manager.aiModules.emptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('manager.aiModules.colPractice') }}</th>
            <th>{{ $t('manager.aiModules.colStatus') }}</th>
            <th>{{ $t('manager.aiModules.colTrial') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in modules" :key="m.practiceId">
            <td>{{ m.practiceName || m.practiceId }}</td>
            <td>{{ m.status }} · {{ $t('manager.aiModules.daysSince', { n: m.daysSinceActivation }) }}</td>
            <td>{{ formatDate(m.trialEndsAt) }} ({{ $t('manager.aiModules.daysRemaining', { n: m.daysRemainingTrial }) }})</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const { t } = useI18n()
const modules = ref<any[]>([])
const alerts = ref<any[]>([])
const loadError = ref('')

function formatDate(iso?: string) {
  if (!iso) return '—'
  return String(iso).slice(0, 10)
}

onMounted(async () => {
  try {
    const [mRes, aRes]: any[] = await Promise.all([
      $fetch('/api/commercial-manager/ai-modules'),
      $fetch('/api/commercial-manager/ai-modules/friction-alerts'),
    ])
    modules.value = mRes.data ?? mRes ?? []
    alerts.value = aRes.data ?? aRes ?? []
  } catch {
    loadError.value = t('manager.aiModules.loadError')
  }
})
</script>

<style scoped>
.pro-mb-lg { margin-bottom: 1.25rem; }
.pro-link { color: var(--pf-vet-accent); margin-right: 0.5rem; }
.pf-cta-cell { white-space: nowrap; }
</style>
