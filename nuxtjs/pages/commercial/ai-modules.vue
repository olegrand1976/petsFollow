<template>
  <div data-testid="commercial-ai-modules-page">
    <ProPageHeader
      :title="$t('commercial.aiModules.title')"
      :subtitle="$t('commercial.aiModules.subtitle')"
    >
      <template #actions>
        <ProButton variant="secondary" @click="navigateTo('/commercial/ai-cr-playbook')">
          {{ $t('commercial.aiModules.openPlaybook') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <ProCard class="pro-mb-lg" :title="$t('commercial.aiModules.activateTitle')">
      <form class="pro-form" @submit.prevent="activate">
        <div class="pro-field">
          <label class="pro-label" for="c-ai-practice">{{ $t('commercial.aiModules.practice') }}</label>
          <select id="c-ai-practice" v-model="selectedPracticeId" class="pro-input" required data-testid="commercial-ai-practice">
            <option value="">{{ $t('commercial.aiModules.practicePlaceholder') }}</option>
            <option v-for="v in vetsWithPractice" :key="v.practiceId" :value="v.practiceId">
              {{ v.practiceName }} — {{ v.fullName }}
            </option>
          </select>
        </div>
        <ProButton type="submit" :disabled="busy || !selectedPracticeId" test-id="commercial-ai-activate">
          {{ $t('commercial.aiModules.activate') }}
        </ProButton>
      </form>
      <p v-if="msg" class="pro-hint pro-mt-md">{{ msg }}</p>
      <p v-if="err" class="pro-hint pro-hint--error pro-mt-md" role="alert">{{ err }}</p>
    </ProCard>

    <ProCard class="pro-mb-lg" :title="$t('commercial.aiModules.frictionTitle')" data-testid="commercial-ai-friction">
      <ProEmptyState
        v-if="!alerts.length"
        :title="$t('commercial.aiModules.frictionEmpty')"
        :description="$t('commercial.aiModules.frictionEmptyDesc')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('commercial.aiModules.colPractice') }}</th>
            <th>{{ $t('commercial.aiModules.colSignal') }}</th>
            <th>{{ $t('commercial.aiModules.colDetail') }}</th>
            <th>{{ $t('commercial.aiModules.colWhen') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in alerts" :key="a.id">
            <td>{{ a.practiceName }}</td>
            <td>{{ a.signal }}</td>
            <td>{{ a.detail }}</td>
            <td>{{ formatDate(a.createdAt) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('commercial.aiModules.listTitle')">
      <ProEmptyState
        v-if="!modules.length"
        :title="$t('commercial.aiModules.emptyTitle')"
        :description="$t('commercial.aiModules.emptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('commercial.aiModules.colPractice') }}</th>
            <th>{{ $t('commercial.aiModules.colStatus') }}</th>
            <th>{{ $t('commercial.aiModules.colTrial') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in modules" :key="m.practiceId">
            <td>{{ m.practiceName || m.practiceId }}</td>
            <td>{{ m.status }} · {{ $t('commercial.aiModules.daysSince', { n: m.daysSinceActivation }) }}</td>
            <td>{{ formatDate(m.trialEndsAt) }} ({{ $t('commercial.aiModules.daysRemaining', { n: m.daysRemainingTrial }) }})</td>
            <td class="pro-flex-gap">
              <ProButton
                v-if="m.status === 'trial' || m.status === 'expired'"
                :disabled="busy"
                @click="convert(m.practiceId, 'monthly_39')"
              >
                39 €/mois
              </ProButton>
              <ProButton
                v-if="m.status === 'trial' || m.status === 'expired'"
                variant="secondary"
                :disabled="busy"
                @click="convert(m.practiceId, 'annual_390')"
              >
                390 €/an
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

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
const alerts = ref<Array<{ id: string, practiceName?: string, signal: string, detail: string, createdAt?: string }>>([])
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
  const [vRes, mRes, aRes]: any[] = await Promise.all([
    $fetch('/api/commercial/vets'),
    $fetch('/api/commercial/ai-modules'),
    $fetch('/api/commercial/ai-modules/friction-alerts'),
  ])
  vets.value = vRes.data ?? vRes ?? []
  modules.value = mRes.data ?? mRes ?? []
  alerts.value = aRes.data ?? aRes ?? []
}

async function activate() {
  busy.value = true
  err.value = ''
  msg.value = ''
  try {
    await $fetch(`/api/commercial/ai-modules/${selectedPracticeId.value}/activate`, { method: 'POST' })
    msg.value = t('commercial.aiModules.activatedOk')
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
    await $fetch(`/api/commercial/ai-modules/${practiceId}/convert`, {
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

onMounted(() => { void load().catch((e) => { err.value = mapError(e) }) })
</script>
