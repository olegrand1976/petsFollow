<template>
  <div>
    <ProPageHeader
      title="Configuration du cabinet"
      subtitle="Complétez votre fiche d'information — visible par vos clients et modifiable à tout moment dans les paramètres."
    />
    <ProCard title="Fiche d'information">
      <PracticeProfileForm v-model="profile" @submit="save">
        <template #actions>
          <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
          <ProButton type="submit" :loading="saving" class="pro-save-btn">
            Enregistrer et accéder au dashboard
          </ProButton>
        </template>
      </PracticeProfileForm>
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
const saving = ref(false)
const error = ref('')
const { fetchUser } = useProUser()

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
  const me = await fetchUser()
  try {
    const res: any = await $fetch('/api/vet/profile')
    const data = res.data ?? res
    profile.value = mapFromApi(data)
  } catch {
    profile.value.vetFullName = me?.fullName || ''
    profile.value.contactEmail = me?.email || ''
    profile.value.practiceName = (me as any)?.practiceName || ''
  }
})

async function save() {
  error.value = ''
  saving.value = true
  try {
    await $fetch('/api/vet/profile?complete=true', {
      method: 'PUT',
      body: profile.value,
    })
    await fetchUser(true)
    await navigateTo('/dashboard')
  } catch (e: any) {
    error.value = e?.data?.message || 'Enregistrement impossible'
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.pro-save-btn {
  margin-top: 1rem;
}
</style>
