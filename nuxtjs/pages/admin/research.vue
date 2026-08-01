<template>
  <div data-testid="admin-research-page">
    <ProPageHeader
      :title="$t('admin.research.title')"
      :subtitle="$t('admin.research.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="admin-research-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <ProEmptyState
      v-if="disabled"
      data-testid="research-disabled"
      :title="$t('research.disabledTitle')"
      :description="$t('research.disabledHint')"
    />
    <template v-else>
      <ProCard :title="$t('admin.research.optInsTitle')" class="pro-mb-lg">
        <p v-if="loadingOptIns" class="pro-hint">{{ $t('common.loading') }}</p>
        <p v-else-if="optInError" class="pro-error" role="alert">{{ optInError }}</p>
        <ProTable
          v-else
          :empty="!optIns.length"
          :empty-title="$t('admin.research.empty')"
          data-testid="admin-research-optins-table"
        >
          <thead>
            <tr>
              <th>{{ $t('admin.research.colName') }}</th>
              <th>{{ $t('admin.research.colPostal') }}</th>
              <th>{{ $t('admin.research.colCity') }}</th>
              <th>{{ $t('admin.research.colCountry') }}</th>
              <th>{{ $t('admin.research.colOptedInAt') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in optIns" :key="row.practiceId" data-testid="admin-research-optin-row">
              <td>{{ row.name || '—' }}</td>
              <td>{{ row.postalCode || '—' }}</td>
              <td>{{ row.city || '—' }}</td>
              <td>{{ row.countryCode }}</td>
              <td>{{ formatDate(row.optedInAt) }}</td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>

      <ProCard :title="$t('admin.research.groupsTitle')" data-testid="admin-research-groups-card">
        <p class="pro-hint pro-mb-sm">{{ $t('admin.research.groupsHint') }}</p>
        <p v-if="loadingGroups" class="pro-hint">{{ $t('common.loading') }}</p>
        <p v-else-if="groupsError" class="pro-error" role="alert">{{ groupsError }}</p>
        <ProTable
          v-else
          :empty="!groups.length"
          :empty-title="$t('admin.research.groupsEmpty')"
          data-testid="admin-research-groups-table"
        >
          <thead>
            <tr>
              <th>{{ $t('research.groups.name') }}</th>
              <th>{{ $t('research.groups.members') }}</th>
              <th>{{ $t('research.groups.dataroom') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr v-for="g in groups" :key="g.id" data-testid="admin-research-group-row">
              <td>{{ g.name }}</td>
              <td>{{ g.memberCount }}</td>
              <td>
                <ProBadge :variant="g.dataroomEnabled ? 'success' : 'neutral'">
                  {{ g.dataroomEnabled ? $t('research.groups.dataroomOn') : $t('research.groups.dataroomOff') }}
                </ProBadge>
              </td>
              <td>
                <ProButton
                  variant="secondary"
                  :loading="togglingId === g.id"
                  test-id="admin-research-dataroom-toggle"
                  @click="toggleDataroom(g)"
                >
                  {{ g.dataroomEnabled ? $t('admin.research.disableDataroom') : $t('admin.research.enableDataroom') }}
                </ProButton>
              </td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import { isPublicFlagOn } from '~/utils/public-feature-flag'

definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

const { t, locale } = useI18n()
const runtimeConfig = useRuntimeConfig()
const disabled = computed(() => !isPublicFlagOn(runtimeConfig.public.researchEnabled))

type OptIn = {
  practiceId: string
  name: string
  postalCode: string
  city: string
  countryCode: string
  optedInAt: string
}
type Group = {
  id: string
  name: string
  memberCount: number
  dataroomEnabled: boolean
}

const optIns = ref<OptIn[]>([])
const groups = ref<Group[]>([])
const loadingOptIns = ref(false)
const loadingGroups = ref(false)
const optInError = ref('')
const groupsError = ref('')
const togglingId = ref('')

function formatDate(iso: string) {
  if (!iso) return '—'
  try {
    return new Intl.DateTimeFormat(locale.value || 'fr', {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(new Date(iso))
  } catch {
    return iso
  }
}

async function loadOptIns() {
  loadingOptIns.value = true
  optInError.value = ''
  try {
    const res: any = await $fetch('/api/admin/research/opt-ins')
    const data = res?.data ?? res
    optIns.value = Array.isArray(data?.items) ? data.items : []
  } catch (e: any) {
    const code = e?.data?.error?.code
    optInError.value = code === 'research_disabled' || e?.statusCode === 404
      ? t('research.disabledTitle')
      : t('admin.research.loadError')
    optIns.value = []
  } finally {
    loadingOptIns.value = false
  }
}

async function loadGroups() {
  loadingGroups.value = true
  groupsError.value = ''
  try {
    const res: any = await $fetch('/api/admin/research/groups')
    const data = res?.data ?? res
    groups.value = Array.isArray(data?.items) ? data.items : []
  } catch {
    groupsError.value = t('admin.research.groupsLoadError')
    groups.value = []
  } finally {
    loadingGroups.value = false
  }
}

async function toggleDataroom(g: Group) {
  togglingId.value = g.id
  try {
    await $fetch(`/api/admin/research/groups/${g.id}/dataroom`, {
      method: 'PATCH',
      body: { enabled: !g.dataroomEnabled },
    })
    await loadGroups()
  } catch {
    groupsError.value = t('admin.research.groupsLoadError')
  } finally {
    togglingId.value = ''
  }
}

onMounted(() => {
  if (disabled.value) return
  void loadOptIns()
  void loadGroups()
})
</script>
