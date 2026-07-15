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
    <fieldset v-if="showHeartrateDurations" class="pro-profile-form__heartrate">
      <legend class="pro-label">{{ $t('settings.heartrate.title') }}</legend>
      <p class="pro-profile-form__hint">{{ $t('settings.heartrate.subtitle') }}</p>
      <div class="pro-heartrate-durations" role="group" :aria-label="$t('settings.heartrate.title')">
        <label v-for="opt in durationOptions" :key="opt" class="pro-heartrate-durations__item">
          <input v-model="heartrateDurationsSec" type="checkbox" :value="opt">
          <span>{{ $t(`settings.heartrate.duration${opt}`) }}</span>
        </label>
      </div>
      <p class="pro-profile-form__hint">{{ $t('settings.heartrate.hint') }}</p>
    </fieldset>
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

withDefaults(
  defineProps<{
    showHeartrateDurations?: boolean
  }>(),
  { showHeartrateDurations: false },
)

const heartrateDurationsSec = defineModel<number[]>('heartrateDurationsSec', { default: () => [60] })

defineEmits<{ submit: [] }>()

const durationOptions = [15, 30, 60] as const
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

.pro-profile-form__heartrate {
  border: none;
  margin: 1rem 0 0;
  padding: 0;
}

.pro-profile-form__hint {
  color: var(--pf-vet-text-muted);
  font-size: 0.875rem;
  margin: 0.35rem 0 0.75rem;
}

.pro-heartrate-durations {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

.pro-heartrate-durations__item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

@media (max-width: 640px) {
  .pro-profile-form__row {
    grid-template-columns: 1fr;
  }
}
</style>
