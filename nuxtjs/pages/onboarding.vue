<template>
  <div data-testid="onboarding-page" class="pro-onboarding">
    <ProPageHeader
      :title="$t('onboarding.title')"
      :subtitle="$t('onboarding.subtitle')"
    />

    <nav class="pro-onboarding__steps" :aria-label="$t('onboarding.stepsNav')">
      <button
        v-for="(s, i) in stepKeys"
        :key="s"
        type="button"
        class="pro-onboarding__step"
        :class="{
          'pro-onboarding__step--active': i === step,
          'pro-onboarding__step--done': i < step,
        }"
        :disabled="i > step"
        @click="goToStep(i)"
      >
        <span class="pro-onboarding__step-num">{{ i + 1 }}</span>
        <span class="pro-onboarding__step-label">{{ $t(`onboarding.steps.${s}.label`) }}</span>
      </button>
    </nav>

    <ProCard :title="$t(`onboarding.steps.${currentKey}.title`)">
      <p class="pro-onboarding__hint">{{ $t(`onboarding.steps.${currentKey}.text`) }}</p>

      <form class="pro-profile-form" @submit.prevent="onSubmit">
        <template v-if="currentKey === 'identity'">
          <ProInput
            v-model="profile.vetFullName"
            :label="$t('components.profileForm.vetName')"
            name="vetFullName"
            required
            test-id="onboarding-vet-name"
          />
          <ProInput
            v-model="profile.practiceName"
            :label="$t('components.profileForm.practiceName')"
            name="practiceName"
            required
            test-id="onboarding-practice-name"
          />
        </template>

        <template v-else-if="currentKey === 'contact'">
          <ProInput
            v-model="profile.contactEmail"
            :label="$t('components.profileForm.contactEmail')"
            type="email"
            name="contactEmail"
            required
            test-id="onboarding-contact-email"
          />
          <ProInput
            v-model="profile.phone"
            :label="$t('components.profileForm.phone')"
            type="tel"
            name="phone"
            required
            test-id="onboarding-phone"
          />
        </template>

        <template v-else-if="currentKey === 'address'">
          <ProInput
            v-model="profile.addressLine1"
            :label="$t('components.profileForm.address')"
            name="addressLine1"
            required
            test-id="onboarding-address"
          />
          <ProInput
            v-model="profile.addressLine2"
            :label="$t('components.profileForm.addressLine2')"
            name="addressLine2"
            test-id="onboarding-address2"
          />
          <div class="pro-profile-form__row">
            <ProInput
              v-model="profile.postalCode"
              :label="$t('components.profileForm.postalCode')"
              name="postalCode"
              required
              test-id="onboarding-postal"
            />
            <ProInput
              v-model="profile.city"
              :label="$t('components.profileForm.city')"
              name="city"
              required
              test-id="onboarding-city"
            />
          </div>
        </template>

        <template v-else>
          <ProInput
            v-model="profile.website"
            :label="$t('components.profileForm.website')"
            type="url"
            name="website"
            :placeholder="$t('components.profileForm.websitePlaceholder')"
            test-id="onboarding-website"
          />
          <fieldset class="pro-profile-form__heartrate">
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
        </template>

        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

        <div class="pro-onboarding__actions">
          <ProButton
            v-if="step > 0"
            type="button"
            variant="ghost"
            test-id="onboarding-back"
            @click="prev"
          >
            {{ $t('common.previous') }}
          </ProButton>
          <ProButton
            type="submit"
            :loading="saving"
            class="pro-onboarding__primary"
            :test-id="isLast ? 'onboarding-submit' : 'onboarding-next'"
          >
            {{ isLast ? $t('onboarding.submit') : $t('common.next') }}
          </ProButton>
        </div>
      </form>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import type { PracticeProfileForm } from '~/components/pro/PracticeProfileForm.vue'

definePageMeta({ middleware: 'vet-only' })

const { t } = useI18n()
const { mapError } = useApiError()
const { fetchUser } = useProUser()

const stepKeys = ['identity', 'contact', 'address', 'options'] as const
type StepKey = (typeof stepKeys)[number]

const step = ref(0)
const saving = ref(false)
const error = ref('')
const heartrateDurationsSec = ref<number[]>([60])
const durationOptions = [15, 30, 60] as const

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

const currentKey = computed(() => stepKeys[step.value] as StepKey)
const isLast = computed(() => step.value === stepKeys.length - 1)

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

function goToStep(i: number) {
  if (i <= step.value) step.value = i
}

function prev() {
  if (step.value > 0) {
    error.value = ''
    step.value -= 1
  }
}

function validateCurrent(): boolean {
  error.value = ''
  const p = profile.value
  switch (currentKey.value) {
    case 'identity':
      if (!p.vetFullName.trim() || !p.practiceName.trim()) {
        error.value = t('onboarding.errors.required')
        return false
      }
      return true
    case 'contact':
      if (!p.contactEmail.trim() || !p.phone.trim()) {
        error.value = t('onboarding.errors.required')
        return false
      }
      return true
    case 'address':
      if (!p.addressLine1.trim() || !p.postalCode.trim() || !p.city.trim()) {
        error.value = t('onboarding.errors.required')
        return false
      }
      return true
    case 'options':
      if (heartrateDurationsSec.value.length === 0) {
        heartrateDurationsSec.value = [60]
      }
      return true
    default: {
      const _exhaustive: never = currentKey.value
      return _exhaustive
    }
  }
}

async function onSubmit() {
  if (!validateCurrent()) return
  if (!isLast.value) {
    step.value += 1
    return
  }
  saving.value = true
  error.value = ''
  try {
    await $fetch('/api/vet/profile?complete=true', {
      method: 'PUT',
      body: {
        ...profile.value,
        heartrateDurationsSec: heartrateDurationsSec.value,
      },
    })
    await fetchUser(true)
    await navigateTo('/dashboard')
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  const me = await fetchUser()
  try {
    const res: any = await $fetch('/api/vet/profile')
    const data = res.data ?? res
    profile.value = mapFromApi(data)
    heartrateDurationsSec.value = data.heartrateDurationsSec?.length
      ? data.heartrateDurationsSec
      : [60]
  } catch {
    profile.value.vetFullName = me?.fullName || ''
    profile.value.contactEmail = me?.email || ''
    profile.value.practiceName = (me as any)?.practiceName || ''
  }
})
</script>

<style scoped>
.pro-onboarding {
  max-width: 44rem;
}

.pro-onboarding__steps {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.5rem;
  margin-bottom: 1.25rem;
}

.pro-onboarding__step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  padding: 0.65rem 0.4rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  background: var(--pf-vet-surface);
  color: var(--pf-vet-text-muted);
  cursor: pointer;
  font: inherit;
}

.pro-onboarding__step:disabled {
  cursor: default;
  opacity: 0.65;
}

.pro-onboarding__step--active {
  border-color: var(--pf-vet-accent);
  color: var(--pf-vet-primary);
  box-shadow: var(--pf-vet-shadow-sm);
}

.pro-onboarding__step--done {
  border-color: var(--pf-vet-accent);
  color: var(--pf-vet-accent);
}

.pro-onboarding__step-num {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 999px;
  display: grid;
  place-items: center;
  font-size: 0.8rem;
  font-weight: 700;
  background: var(--pf-vet-bg);
}

.pro-onboarding__step--active .pro-onboarding__step-num,
.pro-onboarding__step--done .pro-onboarding__step-num {
  background: var(--pf-vet-accent);
  color: #fff;
}

.pro-onboarding__step-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-align: center;
  line-height: 1.2;
}

.pro-onboarding__hint {
  color: var(--pf-vet-text-muted);
  margin: 0 0 1rem;
  font-size: 0.9rem;
}

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

.pro-onboarding__actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 0.75rem;
  margin-top: 1.25rem;
}

.pro-onboarding__primary {
  margin-left: auto;
}

@media (max-width: 640px) {
  .pro-onboarding__steps {
    grid-template-columns: 1fr 1fr;
  }

  .pro-profile-form__row {
    grid-template-columns: 1fr;
  }
}
</style>
