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
</style>
