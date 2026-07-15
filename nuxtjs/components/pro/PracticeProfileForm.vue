<template>
  <form class="pro-profile-form" @submit.prevent="$emit('submit')">
    <ProInput
      v-model="model.vetFullName"
      :label="$t('components.profileForm.vetName')"
      name="vetFullName"
      required
    />
    <ProInput
      v-model="model.practiceName"
      :label="$t('components.profileForm.practiceName')"
      name="practiceName"
      required
    />
    <ProInput
      v-model="model.contactEmail"
      :label="$t('components.profileForm.contactEmail')"
      type="email"
      name="contactEmail"
      required
    />
    <ProInput
      v-model="model.phone"
      :label="$t('components.profileForm.phone')"
      type="tel"
      name="phone"
      required
    />
    <ProInput
      v-model="model.addressLine1"
      :label="$t('components.profileForm.address')"
      name="addressLine1"
      required
    />
    <ProInput
      v-model="model.addressLine2"
      :label="$t('components.profileForm.addressLine2')"
      name="addressLine2"
    />
    <div class="pro-profile-form__row">
      <ProInput
        v-model="model.postalCode"
        :label="$t('components.profileForm.postalCode')"
        name="postalCode"
        required
      />
      <ProInput
        v-model="model.city"
        :label="$t('components.profileForm.city')"
        name="city"
        required
      />
    </div>
    <ProInput
      v-model="model.website"
      :label="$t('components.profileForm.website')"
      type="url"
      name="website"
      :placeholder="$t('components.profileForm.websitePlaceholder')"
    />
    <slot name="actions" />
  </form>
</template>

<script setup lang="ts">
export type PracticeProfileForm = {
  vetFullName: string
  practiceName: string
  contactEmail: string
  phone: string
  addressLine1: string
  addressLine2: string
  city: string
  postalCode: string
  website: string
}

const model = defineModel<PracticeProfileForm>({ required: true })

defineEmits<{ submit: [] }>()
</script>

<style scoped>
.pro-profile-form {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.pro-profile-form__row {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 1rem;
}

@media (max-width: 640px) {
  .pro-profile-form__row {
    grid-template-columns: 1fr;
  }
}
</style>
