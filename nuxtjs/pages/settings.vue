<template>
  <div>
    <ProPageHeader
      :title="$t('settings.title')"
      :subtitle="$t('settings.subtitle')"
    />

    <ProSectionTabs
      v-model="activeTab"
      :tabs="settingsTabs"
      :aria-label="$t('settings.tabsLabel')"
    />

    <div
      v-show="activeTab === 'profile'"
      role="tabpanel"
      aria-labelledby="tab-profile"
      data-testid="settings-tab-profile"
    >
      <ProCard :title="$t('settings.avatar.title')" class="pro-settings-card">
        <ProAvatarUpload
          v-model="avatarUrl"
          :name="user?.fullName || ''"
          upload-url="/api/me/avatar"
          :label="$t('settings.avatar.change')"
          :hint="$t('settings.avatar.hint')"
          @uploaded="onAvatarUploaded"
        />
        <p v-if="avatarSaved" class="text-muted" role="status">{{ $t('settings.avatar.saved') }}</p>
      </ProCard>

      <ProCard
        v-if="canManagePractice"
        :title="$t('settings.profileCard')"
        class="pro-settings-card"
        data-testid="settings-practice-profile"
      >
        <ProPracticeProfileForm
          v-model="profile"
          v-model:heartrate-durations-sec="selectedDurations"
          v-model:desk-idle-minutes="deskIdleMinutes"
          show-heartrate-durations
          show-desk-idle
          @submit="saveProfile"
        >
          <template #actions>
            <p v-if="profileSaved" class="text-muted" role="status">{{ $t('settings.profileSaved') }}</p>
            <p v-if="profileError" class="pro-field-error" role="alert">{{ profileError }}</p>
            <ProButton type="submit" :loading="profileSaving" :disabled="!profileLoaded" class="pro-save-btn">
              {{ $t('settings.saveProfile') }}
            </ProButton>
          </template>
        </ProPracticeProfileForm>
      </ProCard>
      <p
        v-else
        class="pro-hint pro-mb-md"
        data-testid="settings-practice-profile-denied"
      >
        {{ $t('settings.aclDenied') }}
      </p>

      <ProCard
        v-if="canManagePractice && researchFlagOn"
        :title="$t('research.settings.title')"
        class="pro-settings-card"
        data-testid="settings-research-opt-in"
      >
        <div class="pro-flex-row pro-mb-sm">
          <ProBadge variant="warning">{{ $t('research.settings.devBadge') }}</ProBadge>
        </div>
        <p class="pro-hint">{{ $t('research.settings.hint') }}</p>
        <p class="pro-mb-sm" data-testid="settings-research-status">
          <template v-if="researchStatusKnown">
            {{ researchOptedIn ? $t('research.settings.optedIn') : $t('research.settings.optedOut') }}
          </template>
          <template v-else>{{ $t('research.settings.statusUnknown') }}</template>
        </p>
        <p v-if="researchSaved" class="text-muted" role="status">{{ $t('research.settings.saved') }}</p>
        <p v-if="researchError" class="pro-field-error" role="alert">{{ researchError }}</p>
        <div class="pro-flex-row">
          <ProButton
            v-if="researchStatusKnown && !researchOptedIn"
            :loading="researchBusy"
            data-testid="settings-research-opt-in-btn"
            @click="setResearchOptIn(true)"
          >
            {{ $t('research.settings.optIn') }}
          </ProButton>
          <ProButton
            v-else-if="researchStatusKnown && researchOptedIn"
            variant="secondary"
            :loading="researchBusy"
            data-testid="settings-research-opt-out-btn"
            @click="confirmResearchOptOut"
          >
            {{ $t('research.settings.optOut') }}
          </ProButton>
        </div>
      </ProCard>
    </div>

    <div
      v-show="activeTab === 'calendar'"
      role="tabpanel"
      aria-labelledby="tab-calendar"
      data-testid="settings-tab-calendar"
    >
      <ProCard
        v-if="canMessage"
        :title="$t('settings.availability')"
        class="pro-settings-card"
        data-testid="settings-availability"
      >
        <div class="pro-toggle" role="group" :aria-label="$t('settings.availability')">
          <button
            type="button"
            class="pro-toggle-btn"
            :class="{ 'pro-toggle-btn--active': status === 'available' }"
            @click="status = 'available'"
          >
            {{ $t('settings.available') }}
          </button>
          <button
            type="button"
            class="pro-toggle-btn"
            :class="{ 'pro-toggle-btn--active': status === 'unavailable' }"
            @click="status = 'unavailable'"
          >
            {{ $t('settings.unavailable') }}
          </button>
        </div>
        <div class="pro-field pro-field-spaced">
          <label class="pro-label" for="auto-reply">{{ $t('settings.autoReply') }}</label>
          <textarea
            id="auto-reply"
            v-model="autoReply"
            class="pro-textarea"
            rows="4"
            :placeholder="$t('settings.autoReplyPlaceholder')"
          />
        </div>
        <p v-if="saved" class="text-muted" role="status">{{ $t('settings.availabilitySaved') }}</p>
        <p v-if="availabilityError" class="pro-field-error" role="alert">{{ availabilityError }}</p>
        <ProButton class="pro-save-btn" :loading="saving" @click="save">
          {{ $t('settings.saveAvailability') }}
        </ProButton>
      </ProCard>

      <ProCard
        v-if="canManageCalendar"
        :title="$t('settings.calendar.title')"
        class="pro-settings-card"
        data-testid="settings-calendar"
      >        <p v-if="!vacationsConfigured" class="pro-inline-feedback" role="status">
          {{ $t('settings.calendar.vacationsReminder') }}
        </p>
        <p class="pro-settings-hint">{{ $t('settings.calendar.slotsHint') }}</p>
        <div class="pro-field pro-field-spaced">
          <label class="pro-label" for="slot-duration">{{ $t('settings.calendar.duration') }}</label>
          <select id="slot-duration" v-model.number="slotDuration" class="pro-select">
            <option :value="15">15</option>
            <option :value="30">30</option>
            <option :value="60">60</option>
          </select>
        </div>
        <div v-for="(slot, idx) in scheduleSlots" :key="idx" class="calendar-slot-row">
          <select v-model.number="slot.weekday" class="pro-select">
            <option v-for="d in weekdayOptions" :key="d.value" :value="d.value">{{ d.label }}</option>
          </select>
          <input v-model="slot.startTime" type="time" class="pro-input" required>
          <input v-model="slot.endTime" type="time" class="pro-input" required>
          <ProButton variant="ghost" type="button" @click="scheduleSlots.splice(idx, 1)">×</ProButton>
        </div>
        <ProButton variant="secondary" type="button" class="pro-mb-md" @click="addSlot">
          {{ $t('settings.calendar.addSlot') }}
        </ProButton>
        <label class="pro-checkbox-row">
          <input v-model="clientBookingEnabled" type="checkbox" :disabled="!scheduleSlots.length" data-testid="calendar-booking-toggle">
          <span>{{ $t('settings.calendar.enableBooking') }}</span>
        </label>
        <p class="pro-settings-hint">{{ $t('settings.calendar.enableHelp') }}</p>
        <p v-if="scheduleError" class="pro-field-error" role="alert">{{ scheduleError }}</p>
        <p v-if="scheduleSaved" class="text-muted" role="status">{{ $t('settings.calendar.saved') }}</p>
        <ProButton class="pro-save-btn" :loading="scheduleSaving" data-testid="calendar-schedule-save" @click="saveSchedule">
          {{ $t('settings.calendar.save') }}
        </ProButton>

        <hr class="pro-settings-hr">
        <h3 class="pro-settings-subtitle">{{ $t('settings.calendar.vacationsTitle') }}</h3>
        <label class="pro-checkbox-row">
          <input v-model="noVacationsThisYear" type="checkbox">
          <span>{{ $t('settings.calendar.noVacations') }}</span>
        </label>
        <div class="calendar-slot-row">
          <input v-model="vacStart" type="date" class="pro-input">
          <input v-model="vacEnd" type="date" class="pro-input">
          <ProButton variant="secondary" type="button" :disabled="!vacStart || !vacEnd" @click="addVacation">
            {{ $t('settings.calendar.addVacation') }}
          </ProButton>
        </div>
        <ul class="calendar-vac-list">
          <li v-for="v in vacations" :key="v.id">
            <span>{{ v.startsOn }} → {{ v.endsOn }} {{ v.label ? `— ${v.label}` : '' }}</span>
            <ProButton variant="ghost" type="button" @click="removeVacation(v.id)">×</ProButton>
          </li>
        </ul>

        <hr class="pro-settings-hr">
        <h3 class="pro-settings-subtitle">{{ $t('settings.calendar.visitTypesTitle') }}</h3>
        <p class="pro-settings-hint">{{ $t('settings.calendar.visitTypesHint') }}</p>
        <div
          v-for="(vt, idx) in visitTypes"
          :key="vt.id || `new-${idx}`"
          class="visit-type-row"
          data-testid="settings-visit-type-row"
        >
          <span class="visit-type-row__swatch" :style="{ background: vt.color }" aria-hidden="true" />
          <input
            v-model="vt.name"
            type="text"
            class="pro-input"
            :placeholder="$t('settings.calendar.visitTypeName')"
            required
          >
          <input
            v-model.number="vt.durationMinutes"
            type="number"
            class="pro-input visit-type-row__duration"
            min="5"
            max="480"
            step="5"
            :aria-label="$t('settings.calendar.visitTypeDuration')"
          >
          <input
            v-model="vt.color"
            type="color"
            class="visit-type-row__color"
            :aria-label="$t('settings.calendar.visitTypeColor')"
          >
          <label class="pro-checkbox-row visit-type-row__active">
            <input v-model="vt.isActive" type="checkbox">
            <span>{{ $t('settings.calendar.visitTypeActive') }}</span>
          </label>
          <ProButton variant="ghost" type="button" @click="visitTypes.splice(idx, 1)">×</ProButton>
        </div>
        <ProButton variant="secondary" type="button" class="pro-mb-md" data-testid="settings-visit-type-add" @click="addVisitType">
          {{ $t('settings.calendar.addVisitType') }}
        </ProButton>
        <p v-if="visitTypesError" class="pro-field-error" role="alert">{{ visitTypesError }}</p>
        <p v-if="visitTypesSaved" class="text-muted" role="status">{{ $t('settings.calendar.visitTypesSaved') }}</p>
        <ProButton
          class="pro-save-btn"
          :loading="visitTypesSaving"
          data-testid="settings-visit-types-save"
          @click="saveVisitTypes"
        >
          {{ $t('settings.calendar.saveVisitTypes') }}
        </ProButton>
      </ProCard>
    </div>

    <div
      v-show="activeTab === 'notifications'"
      role="tabpanel"
      aria-labelledby="tab-notifications"
      data-testid="settings-tab-notifications"
    >
      <ProCard
        v-if="canMessage"
        :title="$t('settings.notifications.title')"
        class="pro-settings-card"
      >        <p class="pro-settings-hint">{{ $t('settings.notifications.subtitle') }}</p>
        <label class="pro-checkbox-row">
          <input v-model="emailOnMessage" type="checkbox">
          <span>{{ $t('settings.notifications.onMessage') }}</span>
        </label>
        <label class="pro-checkbox-row">
          <input v-model="emailOnHeartrate" type="checkbox">
          <span>{{ $t('settings.notifications.onHeartrate') }}</span>
        </label>
        <label class="pro-checkbox-row">
          <input v-model="emailOnVisitRequest" type="checkbox">
          <span>{{ $t('settings.notifications.onVisitRequest') }}</span>
        </label>
        <p v-if="notifSaved" class="text-muted" role="status">{{ $t('settings.notifications.saved') }}</p>
        <p v-if="notifError" class="pro-field-error" role="alert">{{ notifError }}</p>
        <ProButton class="pro-save-btn" :loading="notifSaving" @click="saveNotifications">
          {{ $t('common.save') }}
        </ProButton>
      </ProCard>
    </div>

    <div
      v-show="activeTab === 'account'"
      role="tabpanel"
      aria-labelledby="tab-account"
      data-testid="settings-tab-account"
    >
      <ProCard :title="$t('settings.password.title')" class="pro-settings-card">
        <p class="pro-settings-hint">{{ $t('settings.password.subtitle') }}</p>
        <ProInput
          v-model="currentPassword"
          :label="$t('settings.password.current')"
          type="password"
          autocomplete="current-password"
          test-id="settings-current-password"
        />
        <ProInput
          v-model="newPassword"
          :label="$t('settings.password.new')"
          type="password"
          autocomplete="new-password"
          test-id="settings-new-password"
        />
        <ProInput
          v-model="confirmPassword"
          :label="$t('settings.password.confirm')"
          type="password"
          autocomplete="new-password"
          test-id="settings-confirm-password"
        />
        <p v-if="passwordSaved" class="text-muted" role="status">{{ $t('settings.password.saved') }}</p>
        <p v-if="passwordError" class="pro-field-error" role="alert">{{ passwordError }}</p>
        <ProButton class="pro-save-btn" :loading="passwordSaving" test-id="settings-password-save" @click="savePassword">
          {{ $t('settings.password.save') }}
        </ProButton>
      </ProCard>

      <ProCard :title="$t('settings.language.title')" class="pro-settings-card">
        <p class="pro-settings-hint">{{ $t('settings.language.subtitle') }}</p>
        <div class="pro-toggle" role="group" :aria-label="$t('settings.language.title')">
          <button
            v-for="loc in supportedLocales"
            :key="loc"
            type="button"
            class="pro-toggle-btn"
            :class="{ 'pro-toggle-btn--active': selectedLocale === loc }"
            :data-testid="`settings-locale-${loc}`"
            @click="selectLocale(loc)"
          >
            {{ $t(`settings.language.${loc}`) }}
          </button>
        </div>
        <p v-if="localeSaved" class="text-muted" role="status">{{ $t('settings.language.saved') }}</p>
        <p v-if="localeError" class="pro-field-error" role="alert">{{ localeError }}</p>
        <ProButton class="pro-save-btn" :loading="localeSaving" test-id="settings-locale-save" @click="saveLocale">
          {{ $t('common.save') }}
        </ProButton>
      </ProCard>

      <ProCard :title="$t('settings.twoFa.title')" class="pro-settings-card">
        <p class="pro-settings-hint">{{ $t('settings.twoFa.hint') }}</p>

        <p v-if="twoFactorEnabled" class="pro-2fa-status pro-2fa-status--on" role="status">
          {{ $t('settings.twoFa.enabled') }}
        </p>
        <p v-else class="pro-2fa-status" role="status">{{ $t('settings.twoFa.disabled') }}</p>

        <div v-if="!twoFactorEnabled && setupData" class="pro-2fa-setup">
          <img :src="setupData.qrCodeDataUrl" :alt="$t('settings.twoFa.qrAlt')" width="200" height="200" class="pro-2fa-qr">
          <p class="pro-settings-hint">
            {{ $t('settings.twoFa.scanHint') }}
            <code>{{ setupData.secret }}</code>
          </p>
          <ProInput
            v-model="setupCode"
            :label="$t('settings.twoFa.verifyCode')"
            type="text"
            inputmode="numeric"
            maxlength="6"
            test-id="2fa-setup-code"
          />
          <p v-if="twoFactorError" class="pro-field-error" role="alert">{{ twoFactorError }}</p>
          <ProButton :loading="twoFactorLoading" @click="confirm2FA">
            {{ $t('settings.twoFa.activate') }}
          </ProButton>
        </div>

        <div v-else-if="!twoFactorEnabled" class="pro-2fa-actions">
          <ProButton :loading="twoFactorLoading" @click="start2FASetup">
            {{ $t('settings.twoFa.setup') }}
          </ProButton>
        </div>

        <div v-else class="pro-2fa-disable">
          <ProInput
            v-model="disableCode"
            :label="$t('auth.twoFa.codeLabel')"
            type="text"
            inputmode="numeric"
            maxlength="6"
          />
          <ProInput
            v-model="disablePassword"
            :label="$t('settings.twoFa.passwordOptional')"
            type="password"
            autocomplete="current-password"
          />
          <p v-if="twoFactorError" class="pro-field-error" role="alert">{{ twoFactorError }}</p>
          <ProButton variant="secondary" :loading="twoFactorLoading" @click="disable2FA">
            {{ $t('settings.twoFa.disable') }}
          </ProButton>
        </div>
      </ProCard>

      <ProPrivacyAccountCard />

      <ProCard :title="$t('settings.legalLinks')" class="pro-settings-card">
        <ProLegalFooter />
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { PracticeProfileForm } from '~/components/pro/PracticeProfileForm.vue'
import { emptyPracticeProfileForm, mapPracticeProfileFromApi } from '~/components/pro/PracticeProfileForm.vue'
import type { AppLocale } from '~/composables/useLocaleSync'
import {
  DEFAULT_DESK_IDLE_MINUTES,
  normalizeDeskIdleMinutes,
} from '~/utils/deskIdleMinutes'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

definePageMeta({ middleware: 'vet-only' })

const { t } = useI18n()
const { mapError } = useApiError()
const { saveLocale: persistLocale, supportedLocales } = useLocaleSync()
const { user, fetchUser } = useProUser()
const { canPractice } = usePracticePerms()
const runtimeConfig = useRuntimeConfig()
const canManagePractice = computed(() => canPractice('practice.settings'))
const canManageCalendar = computed(() => canPractice('calendar.manage'))
const canMessage = computed(() => canPractice('messaging'))
const researchFlagOn = computed(() => isPublicFlagOn(runtimeConfig.public.researchEnabled))
const researchOptedIn = ref(false)
const researchStatusKnown = ref(false)
const researchBusy = ref(false)
const researchSaved = ref(false)
const researchError = ref('')

const activeTab = ref('profile')
const settingsTabs = computed(() => {
  const tabs = [
    { id: 'profile', label: t('settings.tabs.profile') },
  ]
  if (canManageCalendar.value || canMessage.value) {
    tabs.push({ id: 'calendar', label: t('settings.tabs.calendar') })
  }
  if (canMessage.value) {
    tabs.push({ id: 'notifications', label: t('settings.tabs.notifications') })
  }
  tabs.push({ id: 'account', label: t('settings.tabs.account') })
  return tabs
})

const route = useRoute()
const router = useRouter()

const avatarUrl = ref('')
const avatarSaved = ref(false)

const profile = ref<PracticeProfileForm>(emptyPracticeProfileForm())
const profileLoaded = ref(false)
const profileSaving = ref(false)
const profileSaved = ref(false)
const profileError = ref('')

const status = ref('available')
const autoReply = ref('')
const saving = ref(false)
const saved = ref(false)
const availabilityError = ref('')
const notifError = ref('')

const selectedLocale = ref<AppLocale>('fr')
const localeSaving = ref(false)
const localeSaved = ref(false)
const localeError = ref('')

const twoFactorEnabled = ref(false)
const twoFactorLoading = ref(false)
const twoFactorError = ref('')
const setupData = ref<{ secret: string; qrCodeDataUrl: string } | null>(null)
const setupCode = ref('')
const disableCode = ref('')
const disablePassword = ref('')

const selectedDurations = ref<number[]>([60])
const deskIdleMinutes = ref(DEFAULT_DESK_IDLE_MINUTES)

const emailOnMessage = ref(true)
const emailOnHeartrate = ref(true)
const emailOnVisitRequest = ref(true)
const notifSaving = ref(false)
const notifSaved = ref(false)

const scheduleSlots = ref<{ weekday: number; startTime: string; endTime: string }[]>([])
const slotDuration = ref(30)
const clientBookingEnabled = ref(false)
const vacationsConfigured = ref(true)
const noVacationsThisYear = ref(false)
const vacations = ref<any[]>([])
const vacStart = ref('')
const vacEnd = ref('')
const scheduleSaving = ref(false)
const scheduleSaved = ref(false)
const scheduleError = ref('')

type VisitTypeDraft = {
  id?: string
  name: string
  durationMinutes: number
  color: string
  isActive: boolean
  sortOrder: number
}
const visitTypes = ref<VisitTypeDraft[]>([])
const visitTypesSaving = ref(false)
const visitTypesSaved = ref(false)
const visitTypesError = ref('')

const weekdayOptions = computed(() => [
  { value: 1, label: t('settings.calendar.weekday.1') },
  { value: 2, label: t('settings.calendar.weekday.2') },
  { value: 3, label: t('settings.calendar.weekday.3') },
  { value: 4, label: t('settings.calendar.weekday.4') },
  { value: 5, label: t('settings.calendar.weekday.5') },
  { value: 6, label: t('settings.calendar.weekday.6') },
  { value: 0, label: t('settings.calendar.weekday.0') },
])

function addSlot() {
  scheduleSlots.value.push({ weekday: 1, startTime: '09:00', endTime: '12:00' })
}

function addVisitType() {
  const palette = ['#2A9D8F', '#E9C46A', '#E76F51', '#264653', '#457B9D', '#F4A261']
  visitTypes.value.push({
    name: '',
    durationMinutes: slotDuration.value || 30,
    color: palette[visitTypes.value.length % palette.length],
    isActive: true,
    sortOrder: visitTypes.value.length,
  })
}

async function saveVisitTypes() {
  visitTypesSaving.value = true
  visitTypesSaved.value = false
  visitTypesError.value = ''
  try {
    for (const vt of visitTypes.value) {
      if (!String(vt.name || '').trim()) {
        visitTypesError.value = t('settings.calendar.visitTypeNameRequired')
        return
      }
      const dur = Number(vt.durationMinutes)
      if (!Number.isFinite(dur) || dur < 5 || dur > 480) {
        visitTypesError.value = t('settings.calendar.visitTypeDurationInvalid')
        return
      }
    }
    const res: any = await $fetch('/api/vet/visit-types', {
      method: 'PUT',
      body: {
        items: visitTypes.value.map((vt, i) => ({
          id: vt.id || undefined,
          name: vt.name.trim(),
          durationMinutes: Number(vt.durationMinutes),
          color: vt.color,
          isActive: vt.isActive !== false,
          sortOrder: i,
        })),
      },
    })
    const items = res.data ?? res ?? []
    visitTypes.value = (Array.isArray(items) ? items : []).map((vt: any, i: number) => ({
      id: vt.id,
      name: vt.name || '',
      durationMinutes: vt.durationMinutes || 30,
      color: vt.color || '#2A9D8F',
      isActive: vt.isActive !== false,
      sortOrder: vt.sortOrder ?? i,
    }))
    visitTypesSaved.value = true
  } catch (e: any) {
    visitTypesError.value = mapError(e) || t('settings.calendar.visitTypesSaveFailed')
  } finally {
    visitTypesSaving.value = false
  }
}

watch(scheduleSlots, (slots) => {
  if (!slots.length && clientBookingEnabled.value) {
    clientBookingEnabled.value = false
  }
}, { deep: true })

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordSaving = ref(false)
const passwordSaved = ref(false)
const passwordError = ref('')

function onAvatarUploaded(data: any) {
  avatarUrl.value = data?.avatarUrl || avatarUrl.value
  avatarSaved.value = true
  fetchUser(true).catch(() => {})
}

function mapFromApi(data: any): PracticeProfileForm {
  return mapPracticeProfileFromApi(data)
}

async function setResearchOptIn(optIn: boolean) {
  researchBusy.value = true
  researchError.value = ''
  researchSaved.value = false
  try {
    const res: any = await $fetch('/api/vet/practice/research-opt-in', {
      method: optIn ? 'POST' : 'DELETE',
    })
    const data = res.data ?? res
    researchOptedIn.value = !!data.optedIn
    researchStatusKnown.value = true
    researchSaved.value = true
  } catch (e: any) {
    researchError.value = mapError(e) || t('research.settings.error')
  } finally {
    researchBusy.value = false
  }
}

function confirmResearchOptOut() {
  if (!window.confirm(t('research.settings.optOutConfirm'))) return
  void setResearchOptIn(false)
}

onMounted(async () => {
  if (route.hash === '#calendar' && (canManageCalendar.value || canMessage.value)) {
    activeTab.value = 'calendar'
    await router.replace({ query: { ...route.query, tab: 'calendar' }, hash: '' })
  }

  autoReply.value = t('settings.autoReplyDefault')
  avatarUrl.value = user.value?.avatarUrl || ''
  if (!avatarUrl.value) {
    try {
      const me = await fetchUser(true)
      avatarUrl.value = me?.avatarUrl || ''
    } catch { /* ignore */ }
  }

  if (canManagePractice.value) {
    try {
      const res: any = await $fetch('/api/vet/profile')
      const data = res.data ?? res
      profile.value = mapFromApi(data)
      const durations = (data.heartrateDurationsSec ?? []).map(Number).filter((n: number) => [15, 30, 60].includes(n))
      selectedDurations.value = durations.length ? durations : [60]
      const idle = Number(data.deskIdleMinutes)
      deskIdleMinutes.value = normalizeDeskIdleMinutes(idle)
      profileLoaded.value = true
    } catch (e: any) {
      profileError.value = mapError(e) || t('settings.profileLoadFailed')
    }
    if (researchFlagOn.value) {
      try {
        const ro: any = await $fetch('/api/vet/practice/research-opt-in')
        const data = ro.data ?? ro
        researchOptedIn.value = !!data.optedIn
        researchStatusKnown.value = true
      } catch (e: any) {
        researchStatusKnown.value = false
        const code = e?.data?.error?.code
        if (code !== 'research_disabled') {
          researchError.value = mapError(e) || t('research.settings.error')
        }
      }
    }
  }

  if (canMessage.value) {
    try {
      const avail: any = await $fetch('/api/vet/availability')
      const data = avail.data ?? avail
      status.value = data.status ?? status.value
      autoReply.value = data.autoReply || autoReply.value
    } catch { /* ignore */ }
  }

  const preferred = user.value?.preferredLocale as AppLocale | undefined
  if (preferred && supportedLocales.includes(preferred)) {
    selectedLocale.value = preferred
  }

  try {
    const tfa: any = await $fetch('/api/auth/2fa/status')
    twoFactorEnabled.value = (tfa.data ?? tfa).enabled === true
  } catch { /* ignore */ }

  if (canMessage.value) {
    try {
      const prefs: any = await $fetch('/api/vet/notification-preferences')
      const data = prefs.data ?? prefs
      emailOnMessage.value = data.emailOnMessage !== false
      emailOnHeartrate.value = data.emailOnHeartrate !== false
      emailOnVisitRequest.value = data.emailOnVisitRequest !== false
    } catch { /* ignore */ }
  }

  if (canManageCalendar.value) {
    try {
      const [schedRes, vacRes, typesRes]: any[] = await Promise.all([
        $fetch('/api/vet/schedule'),
        $fetch('/api/vet/vacations'),
        $fetch('/api/vet/visit-types'),
      ])
      const sched = schedRes.data ?? schedRes
      scheduleSlots.value = (sched.slots ?? []).map((s: any) => ({
        weekday: s.weekday,
        startTime: s.startTime,
        endTime: s.endTime,
      }))
      slotDuration.value = sched.slotDurationMinutes || 30
      clientBookingEnabled.value = !!sched.clientBookingEnabled
      vacationsConfigured.value = !!sched.vacationsConfiguredForYear
      noVacationsThisYear.value = !!sched.vacationsDeclaredYear && sched.vacationsDeclaredYear >= new Date().getFullYear()
      vacations.value = vacRes.data ?? vacRes ?? []
      const types = typesRes.data ?? typesRes ?? []
      visitTypes.value = (Array.isArray(types) ? types : []).map((vt: any, i: number) => ({
        id: vt.id,
        name: vt.name || '',
        durationMinutes: vt.durationMinutes || 30,
        color: vt.color || '#2A9D8F',
        isActive: vt.isActive !== false,
        sortOrder: vt.sortOrder ?? i,
      }))
    } catch (e: any) {
      scheduleError.value = mapError(e) || t('settings.calendar.loadFailed')
    }
  }
})

async function savePassword() {
  passwordError.value = ''
  passwordSaved.value = false
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = t('settings.password.mismatch')
    return
  }
  if (newPassword.value.length < 8) {
    passwordError.value = t('errors.password_too_short')
    return
  }
  passwordSaving.value = true
  try {
    await $fetch('/api/me/password', {
      method: 'PATCH',
      body: {
        currentPassword: currentPassword.value,
        newPassword: newPassword.value,
      },
    })
    passwordSaved.value = true
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (e: any) {
    passwordError.value = mapError(e)
  } finally {
    passwordSaving.value = false
  }
}

async function saveNotifications() {
  notifSaving.value = true
  notifSaved.value = false
  notifError.value = ''
  try {
    await $fetch('/api/vet/notification-preferences', {
      method: 'PUT',
      body: {
        emailOnMessage: emailOnMessage.value,
        emailOnHeartrate: emailOnHeartrate.value,
        emailOnVisitRequest: emailOnVisitRequest.value,
      },
    })
    notifSaved.value = true
  } catch (e: any) {
    notifError.value = mapError(e)
  } finally {
    notifSaving.value = false
  }
}

async function saveSchedule() {
  scheduleSaving.value = true
  scheduleSaved.value = false
  scheduleError.value = ''
  try {
    const body: any = {
      clientBookingEnabled: clientBookingEnabled.value && scheduleSlots.value.length > 0,
      slotDurationMinutes: slotDuration.value,
      slots: scheduleSlots.value,
    }
    if (noVacationsThisYear.value) {
      body.vacationsDeclaredYear = new Date().getFullYear()
    }
    const res: any = await $fetch('/api/vet/schedule', { method: 'PUT', body })
    const sched = res.data ?? res
    clientBookingEnabled.value = !!sched.clientBookingEnabled
    vacationsConfigured.value = !!sched.vacationsConfiguredForYear
    scheduleSaved.value = true
  } catch (e: any) {
    const raw = [
      e?.data?.error?.message,
      e?.data?.message,
      e?.data?.error?.code,
      e?.statusMessage,
      mapError(e),
    ].filter(Boolean).join(' ')
    scheduleError.value = /schedule_incomplete/i.test(raw)
      ? t('settings.calendar.scheduleIncomplete')
      : mapError(e)
  } finally {
    scheduleSaving.value = false
  }
}

async function addVacation() {
  if (!vacStart.value || !vacEnd.value) return
  try {
    await $fetch('/api/vet/vacations', {
      method: 'POST',
      body: { startsOn: vacStart.value, endsOn: vacEnd.value },
    })
    vacStart.value = ''
    vacEnd.value = ''
    const vacRes: any = await $fetch('/api/vet/vacations')
    vacations.value = vacRes.data ?? vacRes ?? []
    vacationsConfigured.value = true
  } catch (e: any) {
    scheduleError.value = mapError(e)
  }
}

async function removeVacation(id: string) {
  try {
    await $fetch(`/api/vet/vacations/${id}`, { method: 'DELETE' })
    vacations.value = vacations.value.filter((v) => v.id !== id)
  } catch (e: any) {
    scheduleError.value = mapError(e)
  }
}

async function saveProfile() {
  profileSaving.value = true
  profileSaved.value = false
  profileError.value = ''
  if (!profileLoaded.value) {
    profileError.value = t('settings.profileLoadFailed')
    profileSaving.value = false
    return
  }
  const durations = selectedDurations.value.map(Number).filter((n) => [15, 30, 60].includes(n))
  if (!durations.length) {
    profileError.value = t('settings.heartrate.hint')
    profileSaving.value = false
    return
  }
  selectedDurations.value = durations
  const idle = normalizeDeskIdleMinutes(deskIdleMinutes.value)
  deskIdleMinutes.value = idle
  try {
    await $fetch('/api/vet/profile', {
      method: 'PUT',
      body: { ...profile.value, heartrateDurationsSec: durations, deskIdleMinutes: idle },
    })
    profileSaved.value = true
    useDeskSession().setDeskIdleMinutes(idle)
    await useProUser().fetchUser(true)
  } catch (e: any) {
    profileError.value = mapError(e)
  } finally {
    profileSaving.value = false
  }
}

async function save() {
  saving.value = true
  saved.value = false
  availabilityError.value = ''
  try {
    await $fetch('/api/vet/availability', {
      method: 'PUT',
      body: { status: status.value, autoReply: autoReply.value },
    })
    saved.value = true
  } catch (e: any) {
    availabilityError.value = mapError(e)
  } finally {
    saving.value = false
  }
}

async function saveLocale() {
  localeSaving.value = true
  localeSaved.value = false
  localeError.value = ''
  try {
    await persistLocale(selectedLocale.value)
    localeSaved.value = true
  } catch (e: any) {
    localeError.value = mapError(e)
  } finally {
    localeSaving.value = false
  }
}

function selectLocale(loc: AppLocale) {
  selectedLocale.value = loc
  localeSaved.value = false
  localeError.value = ''
}

async function start2FASetup() {
  twoFactorLoading.value = true
  twoFactorError.value = ''
  try {
    const res: any = await $fetch('/api/auth/2fa/setup', { method: 'POST' })
    setupData.value = res.data ?? res
    setupCode.value = ''
  } catch (e: any) {
    twoFactorError.value = mapError(e)
  } finally {
    twoFactorLoading.value = false
  }
}

async function confirm2FA() {
  twoFactorLoading.value = true
  twoFactorError.value = ''
  try {
    await $fetch('/api/auth/2fa/confirm', { method: 'POST', body: { code: setupCode.value } })
    twoFactorEnabled.value = true
    setupData.value = null
    setupCode.value = ''
  } catch {
    twoFactorError.value = t('settings.twoFa.invalidCode')
  } finally {
    twoFactorLoading.value = false
  }
}

async function disable2FA() {
  twoFactorLoading.value = true
  twoFactorError.value = ''
  try {
    await $fetch('/api/auth/2fa/disable', {
      method: 'POST',
      body: { code: disableCode.value, password: disablePassword.value || undefined },
    })
    twoFactorEnabled.value = false
    disableCode.value = ''
    disablePassword.value = ''
  } catch {
    twoFactorError.value = t('settings.twoFa.disableFailed')
  } finally {
    twoFactorLoading.value = false
  }
}
</script>

<style scoped>
.calendar-slot-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;
}
.calendar-vac-list {
  list-style: none;
  padding: 0;
  margin: 0.75rem 0 0;
}
.calendar-vac-list li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.35rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.visit-type-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;
}
.visit-type-row__swatch {
  width: 1rem;
  height: 1rem;
  border-radius: 3px;
  border: 1px solid var(--pf-vet-border);
  flex-shrink: 0;
}
.visit-type-row__duration {
  max-width: 5.5rem;
}
.visit-type-row__color {
  width: 2.5rem;
  height: 2rem;
  padding: 0;
  border: 1px solid var(--pf-vet-border);
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
}
.visit-type-row__active {
  margin: 0;
}
.pro-settings-hr {
  border: 0;
  border-top: 1px solid var(--pf-vet-border);
  margin: 1.25rem 0;
}
.pro-settings-subtitle {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}
.pro-settings-card {
  margin-bottom: 1.5rem;
}

.pro-field-spaced {
  margin-top: 1.25rem;
}

.pro-save-btn {
  margin-top: 1rem;
}

.pro-settings-hint {
  color: var(--pf-vet-text-muted);
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.pro-2fa-status {
  font-weight: 600;
  margin-bottom: 1rem;
}

.pro-2fa-status--on {
  color: var(--pf-vet-accent);
}

.pro-2fa-setup,
.pro-2fa-disable {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-width: 24rem;
}

.pro-2fa-qr {
  border-radius: 8px;
  border: 1px solid var(--pf-vet-border);
}

.pro-checkbox-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
  cursor: pointer;
}

.pro-danger-title {
  color: var(--pf-vet-alert);
}

.pro-danger-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
</style>
