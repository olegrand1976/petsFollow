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

    <fieldset v-if="showPayoutSection" class="pro-profile-form__payout" data-testid="profile-payout-section">
      <legend class="pro-profile-form__section-title">{{ $t('components.profileForm.payoutSection') }}</legend>
      <p class="pro-profile-form__hint">{{ $t('components.profileForm.payoutHint') }}</p>
      <ProInput
        v-model="model.companyLegalName"
        :label="$t('components.profileForm.companyLegalName')"
        name="companyLegalName"
      />
      <div class="pro-profile-form__row">
        <ProInput
          v-model="model.vatNumber"
          :label="$t('components.profileForm.vatNumber')"
          name="vatNumber"
        />
        <ProInput
          v-model="model.companyNumber"
          :label="$t('components.profileForm.companyNumber')"
          name="companyNumber"
        />
      </div>
      <div class="pro-field">
        <label class="pro-label" for="legal-form">{{ $t('components.profileForm.legalForm') }}</label>
        <select id="legal-form" v-model="model.legalForm" class="pro-select" name="legalForm">
          <option value="">{{ $t('components.profileForm.legalFormPlaceholder') }}</option>
          <option value="srl">{{ $t('components.profileForm.legalForms.srl') }}</option>
          <option value="sa">{{ $t('components.profileForm.legalForms.sa') }}</option>
          <option value="asbl">{{ $t('components.profileForm.legalForms.asbl') }}</option>
          <option value="independent">{{ $t('components.profileForm.legalForms.independent') }}</option>
          <option value="other">{{ $t('components.profileForm.legalForms.other') }}</option>
        </select>
      </div>
      <label class="pro-checkbox-row">
        <input v-model="model.billingSameAsPractice" type="checkbox" name="billingSameAsPractice">
        <span>{{ $t('components.profileForm.billingSameAsPractice') }}</span>
      </label>
      <template v-if="!model.billingSameAsPractice">
        <ProInput
          v-model="model.billingAddressLine1"
          :label="$t('components.profileForm.billingAddress')"
          name="billingAddressLine1"
        />
        <ProInput
          v-model="model.billingAddressLine2"
          :label="$t('components.profileForm.billingAddressLine2')"
          name="billingAddressLine2"
        />
        <div class="pro-profile-form__row">
          <ProInput
            v-model="model.billingPostalCode"
            :label="$t('components.profileForm.postalCode')"
            name="billingPostalCode"
          />
          <ProInput
            v-model="model.billingCity"
            :label="$t('components.profileForm.city')"
            name="billingCity"
          />
        </div>
      </template>
      <ProInput
        v-model="model.payoutIban"
        :label="$t('components.profileForm.payoutIban')"
        name="payoutIban"
        autocomplete="off"
      />
      <ProInput
        v-model="model.payoutBic"
        :label="$t('components.profileForm.payoutBic')"
        name="payoutBic"
        autocomplete="off"
      />
      <ProInput
        v-model="model.payoutAccountHolder"
        :label="$t('components.profileForm.payoutAccountHolder')"
        name="payoutAccountHolder"
      />
    </fieldset>

    <slot name="actions" />
  </form>
</template>

<script lang="ts">
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
  companyLegalName: string
  vatNumber: string
  companyNumber: string
  legalForm: string
  billingSameAsPractice: boolean
  billingAddressLine1: string
  billingAddressLine2: string
  billingPostalCode: string
  billingCity: string
  payoutIban: string
  payoutBic: string
  payoutAccountHolder: string
}

export function emptyPracticeProfileForm(): PracticeProfileForm {
  return {
    vetFullName: '',
    practiceName: '',
    contactEmail: '',
    phone: '',
    addressLine1: '',
    addressLine2: '',
    city: '',
    postalCode: '',
    website: '',
    companyLegalName: '',
    vatNumber: '',
    companyNumber: '',
    legalForm: '',
    billingSameAsPractice: true,
    billingAddressLine1: '',
    billingAddressLine2: '',
    billingPostalCode: '',
    billingCity: '',
    payoutIban: '',
    payoutBic: '',
    payoutAccountHolder: '',
  }
}

export function mapPracticeProfileFromApi(data: any): PracticeProfileForm {
  return {
    ...emptyPracticeProfileForm(),
    vetFullName: data?.vetFullName || '',
    practiceName: data?.practiceName || '',
    contactEmail: data?.contactEmail || '',
    phone: data?.phone || '',
    addressLine1: data?.addressLine1 || '',
    addressLine2: data?.addressLine2 || '',
    city: data?.city || '',
    postalCode: data?.postalCode || '',
    website: data?.website || '',
    companyLegalName: data?.companyLegalName || '',
    vatNumber: data?.vatNumber || '',
    companyNumber: data?.companyNumber || '',
    legalForm: data?.legalForm || '',
    billingSameAsPractice: data?.billingSameAsPractice !== false,
    billingAddressLine1: data?.billingAddressLine1 || '',
    billingAddressLine2: data?.billingAddressLine2 || '',
    billingPostalCode: data?.billingPostalCode || '',
    billingCity: data?.billingCity || '',
    payoutIban: data?.payoutIban || '',
    payoutBic: data?.payoutBic || '',
    payoutAccountHolder: data?.payoutAccountHolder || '',
  }
}
</script>

<script setup lang="ts">
const model = defineModel<PracticeProfileForm>({ required: true })

withDefaults(
  defineProps<{
    showHeartrateDurations?: boolean
    showPayoutSection?: boolean
  }>(),
  { showHeartrateDurations: false, showPayoutSection: true },
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

.pro-profile-form__heartrate,
.pro-profile-form__payout {
  border: none;
  margin: 1rem 0 0;
  padding: 0;
}

.pro-profile-form__section-title {
  font-weight: 600;
  color: var(--pf-vet-primary);
  margin: 0 0 0.25rem;
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

.pro-checkbox-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.75rem 0;
  cursor: pointer;
}

.pro-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  margin-bottom: 0.5rem;
}

@media (max-width: 640px) {
  .pro-profile-form__row {
    grid-template-columns: 1fr;
  }
}
</style>
