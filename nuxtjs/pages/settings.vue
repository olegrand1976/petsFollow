<template>
  <div>
    <ProPageHeader
      title="Paramètres"
      subtitle="Fiche d'information, disponibilité et messages automatiques."
    />

    <ProCard title="Fiche d'information du cabinet" class="pro-settings-card">
      <PracticeProfileForm v-model="profile" @submit="saveProfile">
        <template #actions>
          <p v-if="profileSaved" class="text-muted" role="status">Fiche enregistrée.</p>
          <p v-if="profileError" class="pro-field-error" role="alert">{{ profileError }}</p>
          <ProButton type="submit" :loading="profileSaving" class="pro-save-btn">
            Enregistrer la fiche
          </ProButton>
        </template>
      </PracticeProfileForm>
    </ProCard>

    <ProCard title="Disponibilité" class="pro-settings-card">
      <div class="pro-toggle" role="group" aria-label="Statut de disponibilité">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': status === 'available' }"
          @click="status = 'available'"
        >
          Disponible
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': status === 'unavailable' }"
          @click="status = 'unavailable'"
        >
          Indisponible
        </button>
      </div>
      <div class="pro-field pro-field-spaced">
        <label class="pro-label" for="auto-reply">Message auto-réponse</label>
        <textarea
          id="auto-reply"
          v-model="autoReply"
          class="pro-textarea"
          rows="4"
          placeholder="Message envoyé automatiquement lorsque vous êtes indisponible."
        />
      </div>
      <p v-if="saved" class="text-muted" role="status">Paramètres enregistrés.</p>
      <ProButton class="pro-save-btn" :loading="saving" @click="save">
        Enregistrer la disponibilité
      </ProButton>
    </ProCard>

    <ProCard title="Authentification à deux facteurs (2FA)" class="pro-settings-card">
      <p class="pro-settings-hint">
        Optionnel — ajoute une couche de sécurité via une application authenticator (Google Authenticator, Authy…).
      </p>

      <p v-if="twoFactorEnabled" class="pro-2fa-status pro-2fa-status--on" role="status">
        2FA activée sur votre compte.
      </p>
      <p v-else class="pro-2fa-status" role="status">2FA désactivée.</p>

      <div v-if="!twoFactorEnabled && setupData" class="pro-2fa-setup">
        <img :src="setupData.qrCodeDataUrl" alt="QR code 2FA" width="200" height="200" class="pro-2fa-qr">
        <p class="pro-settings-hint">
          Scannez le QR code ou saisissez le secret :
          <code>{{ setupData.secret }}</code>
        </p>
        <ProInput
          v-model="setupCode"
          label="Code de vérification"
          type="text"
          inputmode="numeric"
          maxlength="6"
          test-id="2fa-setup-code"
        />
        <p v-if="twoFactorError" class="pro-field-error" role="alert">{{ twoFactorError }}</p>
        <ProButton :loading="twoFactorLoading" @click="confirm2FA">
          Activer la 2FA
        </ProButton>
      </div>

      <div v-else-if="!twoFactorEnabled" class="pro-2fa-actions">
        <ProButton :loading="twoFactorLoading" @click="start2FASetup">
          Configurer la 2FA
        </ProButton>
      </div>

      <div v-else class="pro-2fa-disable">
        <ProInput
          v-model="disableCode"
          label="Code authenticator"
          type="text"
          inputmode="numeric"
          maxlength="6"
        />
        <ProInput
          v-model="disablePassword"
          label="Mot de passe (si défini)"
          type="password"
          autocomplete="current-password"
        />
        <p v-if="twoFactorError" class="pro-field-error" role="alert">{{ twoFactorError }}</p>
        <ProButton variant="secondary" :loading="twoFactorLoading" @click="disable2FA">
          Désactiver la 2FA
        </ProButton>
      </div>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import type { PracticeProfileForm } from '~/components/pro/PracticeProfileForm.vue'

definePageMeta({ middleware: 'vet-only' })

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
const autoReply = ref('Je suis indisponible, je reviens vers vous rapidement.')
const saving = ref(false)
const saved = ref(false)

const twoFactorEnabled = ref(false)
const twoFactorLoading = ref(false)
const twoFactorError = ref('')
const setupData = ref<{ secret: string; qrCodeDataUrl: string } | null>(null)
const setupCode = ref('')
const disableCode = ref('')
const disablePassword = ref('')

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
  try {
    const res: any = await $fetch('/api/vet/profile')
    profile.value = mapFromApi(res.data ?? res)
  } catch { /* ignore */ }

  const avail: any = await $fetch('/api/vet/availability')
  const data = avail.data ?? avail
  status.value = data.status ?? status.value
  autoReply.value = data.autoReply || autoReply.value

  try {
    const tfa: any = await $fetch('/api/auth/2fa/status')
    twoFactorEnabled.value = (tfa.data ?? tfa).enabled === true
  } catch { /* ignore */ }
})

async function saveProfile() {
  profileSaving.value = true
  profileSaved.value = false
  profileError.value = ''
  try {
    await $fetch('/api/vet/profile', { method: 'PUT', body: profile.value })
    profileSaved.value = true
    await useProUser().fetchUser(true)
  } catch (e: any) {
    profileError.value = e?.data?.message || 'Enregistrement impossible'
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

async function start2FASetup() {
  twoFactorLoading.value = true
  twoFactorError.value = ''
  try {
    const res: any = await $fetch('/api/auth/2fa/setup', { method: 'POST' })
    setupData.value = res.data ?? res
    setupCode.value = ''
  } catch (e: any) {
    twoFactorError.value = e?.data?.message || 'Configuration 2FA impossible'
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
    twoFactorError.value = 'Code invalide'
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
    twoFactorError.value = 'Désactivation impossible — vérifiez le code et le mot de passe'
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
</style>
