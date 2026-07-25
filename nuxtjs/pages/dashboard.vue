<template>
  <div data-testid="vet-dashboard-page">
    <ProPageHeader
      :title="welcomeTitle"
      :subtitle="$t('dashboard.subtitle')"
    />
    <div class="pro-grid-kpi">
      <ProKpi
        icon="group"
        :value="clientCount"
        :label="$t('dashboard.activeClients')"
        to="/clients"
      />
      <ProKpi
        icon="chat"
        :value="unreadCount"
        :label="$t('dashboard.unreadMessages')"
        to="/messages"
        :variant="hasUnread ? 'alert' : 'default'"
      />
      <ProKpi
        icon="favorite"
        :value="recentSessions"
        :label="$t('dashboard.recentSessions')"
      />
      <ProKpi
        icon="inbox"
        :value="pendingLinks"
        :label="$t('dashboard.pendingLinks')"
        to="/clients?invitations=1"
        :variant="pendingLinksRaw > 0 ? 'alert' : 'default'"
      />
      <ProKpi
        icon="event"
        :value="pendingVisits"
        :label="$t('dashboard.pendingVisits')"
        to="/calendar"
        :variant="pendingVisitsRaw > 0 ? 'alert' : 'default'"
      />
      <ProKpi
        icon="medical_services"
        :value="overdueCare"
        :label="$t('dashboard.overdueCare')"
        to="/clients"
        :variant="overdueCareRaw > 0 ? 'alert' : 'default'"
      />
    </div>
    <div class="pro-grid-2 pro-mt-lg">
      <ProCard :title="$t('dashboard.quickActions')">
        <div class="pro-flex-gap">
          <ProButton @click="navigateTo('/clients')">{{ $t('dashboard.viewClients') }}</ProButton>
          <ProButton variant="secondary" @click="navigateTo('/calendar')">{{ $t('dashboard.viewCalendar') }}</ProButton>
          <ProButton variant="secondary" @click="navigateTo('/messages')">{{ $t('dashboard.messaging') }}</ProButton>
          <ProButton variant="ghost" @click="navigateTo('/settings')">{{ $t('nav.settings') }}</ProButton>
        </div>
      </ProCard>
      <ProCard :title="$t('dashboard.aiModule.title')" data-testid="dashboard-ai-module">
        <template v-if="aiLoading">
          <p class="text-muted">…</p>
        </template>
        <template v-else-if="aiModule?.status === 'none' || !aiModule">
          <p class="text-muted">{{ $t('dashboard.aiModule.inactive') }}</p>
        </template>
        <template v-else>
          <p v-if="aiModule.status === 'trial'" class="pro-hint">
            {{ $t('dashboard.aiModule.trial', { days: aiModule.daysRemainingTrial }) }}
          </p>
          <p v-else-if="aiModule.status === 'active'" class="pro-hint">
            {{ $t('dashboard.aiModule.active') }}
          </p>
          <p v-else-if="aiModule.status === 'disabled'" class="pro-hint pro-hint--error">
            {{ $t('dashboard.aiModule.disabled') }}
          </p>
          <p v-else class="pro-hint pro-hint--error">
            {{ $t('dashboard.aiModule.expired') }}
          </p>
          <div v-if="aiRoi?.unlocked" class="ai-roi pro-mt-md" data-testid="dashboard-ai-roi">
            <p class="ai-roi__hero">
              <strong>{{ aiRoi.minutesSaved }}</strong>
              {{ $t('dashboard.aiModule.minutesSaved') }}
            </p>
            <p class="text-muted">
              ≈ {{ (aiRoi.netEuroCents / 100).toFixed(0) }} €
              {{ $t('dashboard.aiModule.euroNet') }}
              <span class="pro-hint">({{ $t('dashboard.aiModule.disclaimer') }})</span>
            </p>
          </div>
          <p v-else-if="aiModule.roiUnlocked && !aiRoi" class="pro-hint pro-mt-md">
            {{ $t('dashboard.aiModule.roiUnavailable') }}
          </p>
          <p v-else-if="aiModule.allowed && !aiModule.roiUnlocked" class="pro-hint pro-mt-md">
            {{ $t('dashboard.aiModule.roiLocked', { day: Math.max(0, 60 - (aiModule.daysSinceActivation || 0)) }) }}
          </p>
          <div class="pro-flex-gap pro-mt-md">
            <ProButton
              v-if="aiModule.status === 'trial' || aiModule.status === 'expired'"
              :disabled="aiBusy"
              @click="requestPaid"
            >
              {{ $t('dashboard.aiModule.requestPaid') }}
            </ProButton>
            <ProButton variant="secondary" @click="navigateTo('/calendar')">
              {{ $t('dashboard.aiModule.openCalendar') }}
            </ProButton>
          </div>
          <form v-if="aiModule.allowed" class="pro-form pro-mt-md" @submit.prevent="sendFeedback">
            <label class="pro-label" for="ai-nps">{{ $t('dashboard.aiModule.npsLabel') }}</label>
            <input id="ai-nps" v-model.number="nps" class="pro-input" type="number" min="0" max="10" required>
            <fieldset class="pro-mt-sm">
              <legend class="pro-label">{{ $t('dashboard.aiModule.frictionTagsLabel') }}</legend>
              <div class="ai-friction-tags">
                <label
                  v-for="tag in frictionTagOptions"
                  :key="tag"
                  class="ai-friction-tags__item"
                >
                  <input v-model="frictionTags" type="checkbox" :value="tag">
                  {{ $t(`dashboard.aiModule.frictionTag.${tag}`) }}
                </label>
              </div>
            </fieldset>
            <ProButton type="submit" variant="ghost" :disabled="aiBusy">
              {{ $t('dashboard.aiModule.sendFeedback') }}
            </ProButton>
          </form>
          <p v-if="aiMsg" class="pro-hint">{{ aiMsg }}</p>
        </template>
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const { t } = useI18n()
const welcomeTitle = ref(t('dashboard.title'))
const clientCount = ref('—')
const unreadCount = ref('—')
const recentSessions = ref('—')
const pendingLinks = ref('—')
const pendingVisits = ref('—')
const overdueCare = ref('—')
const unreadRaw = ref(0)
const pendingLinksRaw = ref(0)
const pendingVisitsRaw = ref(0)
const overdueCareRaw = ref(0)
const { fetchUser } = useProUser()
const { mapError } = useApiError()

const hasUnread = computed(() => unreadRaw.value > 0)

const aiLoading = ref(true)
const aiBusy = ref(false)
const aiModule = ref<any>(null)
const aiRoi = ref<any>(null)
const nps = ref(9)
const frictionTags = ref<string[]>([])
const frictionTagOptions = ['audio', 'quality', 'time', 'ux', 'other'] as const
const aiMsg = ref('')

async function loadAi() {
  aiLoading.value = true
  try {
    const res: any = await $fetch('/api/me/ai-module')
    aiModule.value = res.data ?? res
  } catch {
    aiModule.value = { status: 'none' }
    aiRoi.value = null
    aiLoading.value = false
    return
  }
  if (aiModule.value?.roiUnlocked) {
    try {
      const roiRes: any = await $fetch('/api/me/ai-module/roi')
      aiRoi.value = roiRes.data ?? roiRes
    } catch {
      aiRoi.value = null
    }
  } else {
    aiRoi.value = null
  }
  aiLoading.value = false
}

async function requestPaid() {
  aiBusy.value = true
  aiMsg.value = ''
  try {
    await $fetch('/api/me/ai-module/request-paid', { method: 'POST' })
    aiMsg.value = t('dashboard.aiModule.requestSent')
  } catch (e: any) {
    aiMsg.value = mapError(e)
  } finally {
    aiBusy.value = false
  }
}

async function sendFeedback() {
  aiBusy.value = true
  aiMsg.value = ''
  try {
    await $fetch('/api/me/ai-module/feedback', {
      method: 'POST',
      body: { nps: nps.value, source: 'in_app', comment: '', frictionTags: frictionTags.value },
    })
    aiMsg.value = t('dashboard.aiModule.feedbackOk')
    frictionTags.value = []
  } catch (e: any) {
    aiMsg.value = mapError(e)
  } finally {
    aiBusy.value = false
  }
}

onMounted(async () => {
  try {
    const me = await fetchUser()
    const name = me?.fullName
    if (name) welcomeTitle.value = t('dashboard.welcome', { name: name.split(' ')[0] })
  } catch { /* ignore */ }
  try {
    const res: any = await $fetch('/api/vet/overview')
    const data = res.data ?? res
    clientCount.value = String(data.clientCount ?? 0)
    unreadRaw.value = Number(data.unreadMessages ?? 0)
    unreadCount.value = String(unreadRaw.value)
    recentSessions.value = String(data.recentSessions7d ?? 0)
    pendingLinksRaw.value = Number(data.pendingLinkRequests ?? 0)
    pendingLinks.value = String(pendingLinksRaw.value)
    pendingVisitsRaw.value = Number(data.pendingVisits ?? 0)
    pendingVisits.value = String(pendingVisitsRaw.value)
    overdueCareRaw.value = Number(data.overdueCareCount ?? 0)
    overdueCare.value = String(overdueCareRaw.value)
  } catch { /* ignore */ }
  void loadAi()
})
</script>

<style scoped>
.ai-roi__hero {
  font-size: 1.25rem;
  margin: 0 0 0.35rem;
}
.ai-roi__hero strong {
  font-size: 1.75rem;
}
.ai-friction-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  margin: 0.35rem 0 0.75rem;
}
.ai-friction-tags__item {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.9rem;
}
</style>
