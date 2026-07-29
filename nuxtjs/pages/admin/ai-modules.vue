<template>
  <div data-testid="admin-ai-modules-page">
    <ProPageHeader
      :title="$t('admin.aiModules.title')"
      :subtitle="$t('admin.aiModules.subtitle')"
    />

    <ProCard class="pro-mb-lg" :title="$t('admin.aiModules.activateTitle')">
      <form class="pro-form" @submit.prevent="activate">
        <div class="pro-field">
          <label class="pro-label" for="ai-practice">{{ $t('admin.aiModules.practice') }}</label>
          <select id="ai-practice" v-model="selectedPracticeId" class="pro-input" required data-testid="admin-ai-practice">
            <option value="">{{ $t('admin.aiModules.practicePlaceholder') }}</option>
            <option v-for="v in vetsWithPractice" :key="v.practiceId" :value="v.practiceId">
              {{ v.practiceName }} — {{ v.fullName }}
            </option>
          </select>
        </div>
        <ProButton type="submit" :disabled="busy || !selectedPracticeId" test-id="admin-ai-activate">
          {{ $t('admin.aiModules.activate') }}
        </ProButton>
      </form>
      <p v-if="msg" class="pro-hint pro-mt-md">{{ msg }}</p>
      <p v-if="err" class="pro-hint pro-hint--error pro-mt-md" role="alert">{{ err }}</p>
    </ProCard>

    <ProCard :title="$t('admin.aiModules.listTitle')">
      <ProEmptyState
        v-if="!modules.length"
        :title="$t('admin.aiModules.emptyTitle')"
        :description="$t('admin.aiModules.emptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('admin.aiModules.colPractice') }}</th>
            <th>{{ $t('admin.aiModules.colStatus') }}</th>
            <th>{{ $t('admin.aiModules.colTrial') }}</th>
            <th>{{ $t('admin.aiModules.colDays') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in modules" :key="m.practiceId">
            <td>{{ m.practiceName || m.practiceId }}</td>
            <td>{{ m.status }}</td>
            <td>{{ formatDate(m.trialEndsAt) }}</td>
            <td>{{ m.daysSinceActivation }} / −{{ m.daysRemainingTrial }}</td>
            <td class="pro-flex-gap">
              <ProButton
                v-if="m.status === 'trial' || m.status === 'expired'"
                :disabled="busy"
                @click="convert(m.practiceId, 'monthly_39')"
              >
                {{ $t('admin.aiModules.convertMonthly') }}
              </ProButton>
              <ProButton
                v-if="m.status === 'trial' || m.status === 'expired'"
                variant="secondary"
                :disabled="busy"
                @click="convert(m.practiceId, 'annual_390')"
              >
                {{ $t('admin.aiModules.convertAnnual') }}
              </ProButton>
              <ProButton
                v-if="m.status !== 'disabled'"
                variant="ghost"
                :disabled="busy"
                @click="setStatus(m.practiceId, 'disabled')"
              >
                {{ $t('admin.aiModules.disable') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const { mapError } = useApiError()

type VetOpt = { userId: string, fullName: string, practiceId?: string, practiceName?: string }
type Mod = {
  practiceId: string
  practiceName?: string
  status: string
  trialEndsAt?: string
  daysSinceActivation?: number
  daysRemainingTrial?: number
}

const vets = ref<VetOpt[]>([])
const modules = ref<Mod[]>([])
const selectedPracticeId = ref('')
const busy = ref(false)
const msg = ref('')
const err = ref('')

const vetsWithPractice = computed(() => {
  const seen = new Set<string>()
  return vets.value.filter((v) => {
    if (!v.practiceId || seen.has(v.practiceId)) return false
    seen.add(v.practiceId)
    return true
  })
})

function formatDate(iso?: string) {
  if (!iso) return '—'
  return iso.slice(0, 10)
}

async function load() {
  const [vRes, mRes]: any[] = await Promise.all([
    $fetch('/api/admin/vets'),
    $fetch('/api/admin/ai-modules'),
  ])
  vets.value = vRes.data ?? vRes ?? []
  modules.value = mRes.data ?? mRes ?? []
}

async function activate() {
  busy.value = true
  err.value = ''
  msg.value = ''
  try {
    await $fetch(`/api/admin/ai-modules/${selectedPracticeId.value}/activate`, { method: 'POST' })
    msg.value = t('admin.aiModules.activatedOk')
    await load()
  } catch (e: any) {
    err.value = mapError(e)
  } finally {
    busy.value = false
  }
}

async function convert(practiceId: string, pricePlan: string) {
  busy.value = true
  err.value = ''
  try {
    await $fetch(`/api/admin/ai-modules/${practiceId}/convert`, {
      method: 'POST',
      body: { pricePlan },
    })
    await load()
  } catch (e: any) {
    err.value = mapError(e)
  } finally {
    busy.value = false
  }
}

async function setStatus(practiceId: string, status: string) {
  busy.value = true
  err.value = ''
  try {
    await $fetch(`/api/admin/ai-modules/${practiceId}`, {
      method: 'PATCH',
      body: { status },
    })
    await load()
  } catch (e: any) {
    err.value = mapError(e)
  } finally {
    busy.value = false
  }
}

onMounted(() => { void load().catch((e) => { err.value = mapError(e) }) })
</script>
