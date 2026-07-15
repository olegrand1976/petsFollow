<template>
  <form class="pro-profile-form" @submit.prevent="$emit('submit')">
    <ProInput
      v-model="model.vetFullName"
      label="Nom du vétérinaire"
      name="vetFullName"
      required
    />
    <ProInput
      v-model="model.practiceName"
      label="Nom du cabinet"
      name="practiceName"
      required
    />
    <ProInput
      v-model="model.contactEmail"
      label="Email de contact"
      type="email"
      name="contactEmail"
      required
    />
    <ProInput
      v-model="model.phone"
      label="Téléphone"
      type="tel"
      name="phone"
      required
    />
    <ProInput
      v-model="model.addressLine1"
      label="Adresse"
      name="addressLine1"
      required
    />
    <ProInput
      v-model="model.addressLine2"
      label="Complément d'adresse"
      name="addressLine2"
    />
    <div class="pro-profile-form__row">
      <ProInput
        v-model="model.postalCode"
        label="Code postal"
        name="postalCode"
        required
      />
      <ProInput
        v-model="model.city"
        label="Ville"
        name="city"
        required
      />
    </div>
    <ProInput
      v-model="model.website"
      label="Site web (optionnel)"
      type="url"
      name="website"
      placeholder="https://"
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
