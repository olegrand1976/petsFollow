<template>
  <div>
    <ProPageHeader
      :title="$t('settings.title')"
      :subtitle="$t('settings.subtitle')"
    />

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

    <ProCard :title="$t('settings.profileCard')" class="pro-settings-card">
      <PracticeProfileForm
        v-model="profile"
        v-model:heartrate-durations-sec="selectedDurations"
        show-heartrate-durations
        @submit="saveProfile"
      >
        <template #actions>
          <p v-if="profileSaved" class="text-muted" role="status">{{ $t('settings.profileSaved') }}</p>
          <p v-if="profileError" class="pro-field-error" role="alert">{{ profileError }}</p>
          <ProButton type="submit" :loading="profileSaving" class="pro-save-btn">
            {{ $t('settings.saveProfile') }}
          </ProButton>
        </template>
      </PracticeProfileForm>
    </ProCard>

    <ProCard :title="$t('settings.availability')" class="pro-settings-card">
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
      <ProButton class="pro-save-btn" :loading="saving" @click="save">
        {{ $t('settings.saveAvailability') }}
      </ProButton>
    </ProCard>

    <ProCard :title="$t('settings.notifications.title')" class="pro-settings-card">
      <p class="pro-settings-hint">{{ $t('settings.notifications.subtitle') }}</p>
      <label class="pro-checkbox-row">
        <input v-model="emailOnMessage" type="checkbox">
        <span>{{ $t('settings.notifications.onMessage') }}</span>
      </label>
      <label class="pro-checkbox-row">
        <input v-model="emailOnHeartrate" type="checkbox">
        <span>{{ $t('settings.notifications.onHeartrate') }}</span>
      </label>
      <p v-if="notifSaved" class="text-muted" role="status">{{ $t('settings.notifications.saved') }}</p>
      <ProButton class="pro-save-btn" :loading="notifSaving" @click="saveNotifications">
        {{ $t('common.save') }}
      </ProButton>
    </ProCard>

    <ProCard :title="$t('settings.legalLinks')" class="pro-settings-card">
      <ProLegalFooter />
    </ProCard>

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
          @click="selectedLocale = loc"
        >
          {{ $t(`settings.language.${loc}`) }}
        </button>
      </div>
      <p v-if="localeSaved" class="text-muted" role="status">{{ $t('settings.language.saved') }}</p>
      <p v-if="localeError" class="pro-field-error" role="alert">{{ localeError }}</p>
      <ProButton class="pro-save-btn" :loading="localeSaving" @click="saveLocale">
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
  </div>
</template>

<script setup lang="ts">
import type { PracticeProfileForm } from '~/components/pro/PracticeProfileForm.vue'
import type { AppLocale } from '~/composables/useLocaleSync'

definePageMeta({ middleware: 'vet-only' })

const { t } = useI18n()
const { mapError } = useApiError()
const { saveLocale: persistLocale, supportedLocales } = useLocaleSync()
const { user, fetchUser } = useProUser()

const avatarUrl = ref('')
const avatarSaved = ref(false)

const profile = ref<PracticeProfileForm>({
  vetFullName: '',
  practiceName: '',
  contactEmail: '',
  phone: '',
  addressLine1: '',
  addressLine2: '',
  city: '',
  postalCode: '',
  website: '',
})
const profileSaving = ref(false)
const profileSaved = ref(false)
const profileError = ref('')

const status = ref('available')
const autoReply = ref('')
const saving = ref(false)
const saved = ref(false)

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

const emailOnMessage = ref(true)
const emailOnHeartrate = ref(true)
const notifSaving = ref(false)
const notifSaved = ref(false)

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordSaving = ref(false)
const passwordSaved = ref(false)
const passwordError = ref('')

function onAvatarUploaded(data: any) {
  avatarUrl.value = data?.avatarUrl || avatarUrl.value
  avatarSaved.value = true
  fetchUser(true)
}

function mapFromApi(data: any): PracticeProfileForm {
  return {
    vetFullName: data.vetFullName || '',
    practiceName: data.practiceName || '',
    contactEmail: data.contactEmail || '',
    phone: data.phone || '',
    addressLine1: data.addressLine1 || '',
    addressLine2: data.addressLine2 || '',
    city: data.city || '',
    postalCode: data.postalCode || '',
    website: data.website || '',
  }
}

onMounted(async () => {
  autoReply.value = t('settings.autoReplyDefault')
  avatarUrl.value = user.value?.avatarUrl || ''
  if (!avatarUrl.value) {
    try {
      const me = await fetchUser(true)
      avatarUrl.value = me?.avatarUrl || ''
    } catch { /* ignore */ }
  }

  try {
    const res: any = await $fetch('/api/vet/profile')
    profile.value = mapFromApi(res.data ?? res)
    selectedDurations.value = (res.data ?? res).heartrateDurationsSec ?? [60]
  } catch { /* ignore */ }

  const avail: any = await $fetch('/api/vet/availability')
  const data = avail.data ?? avail
  status.value = data.status ?? status.value
  autoReply.value = data.autoReply || autoReply.value

  const preferred = user.value?.preferredLocale as AppLocale | undefined
  if (preferred && supportedLocales.includes(preferred)) {
    selectedLocale.value = preferred
  }

  try {
    const tfa: any = await $fetch('/api/auth/2fa/status')
    twoFactorEnabled.value = (tfa.data ?? tfa).enabled === true
  } catch { /* ignore */ }

  try {
    const prefs: any = await $fetch('/api/vet/notification-preferences')
    const data = prefs.data ?? prefs
    emailOnMessage.value = data.emailOnMessage !== false
    emailOnHeartrate.value = data.emailOnHeartrate !== false
  } catch { /* ignore */ }
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
  try {
    await $fetch('/api/vet/notification-preferences', {
      method: 'PUT',
      body: { emailOnMessage: emailOnMessage.value, emailOnHeartrate: emailOnHeartrate.value },
    })
    notifSaved.value = true
  } finally {
    notifSaving.value = false
  }
}

async function saveProfile() {
  profileSaving.value = true
  profileSaved.value = false
  profileError.value = ''
  if (!selectedDurations.value.length) {
    profileError.value = t('settings.heartrate.hint')
    profileSaving.value = false
    return
  }
  try {
    await $fetch('/api/vet/profile', {
      method: 'PUT',
      body: { ...profile.value, heartrateDurationsSec: selectedDurations.value },
    })
    profileSaved.value = true
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
  try {
    await $fetch('/api/vet/availability', {
      method: 'PUT',
      body: { status: status.value, autoReply: autoReply.value },
    })
    saved.value = true
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
</style>
