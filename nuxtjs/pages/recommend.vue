<template>
  <div data-testid="recommend-page">
    <ProPageHeader :title="$t('recommend.title')" :subtitle="$t('recommend.subtitle')" />

    <ProCard v-if="thanks" class="pro-mb-lg" data-testid="recommend-thanks">
      <div class="recommend-thanks">
        <ProIcon name="favorite" :size="36" class="recommend-thanks__icon" />
        <h3 class="pro-mb-sm">{{ $t('recommend.thanksTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('recommend.thanksBody') }}</p>
        <ProButton test-id="recommend-another" @click="resetForm">
          {{ $t('recommend.another') }}
        </ProButton>
      </div>
    </ProCard>

    <ProCard v-else data-testid="vet-referral-form">
      <p class="pro-hint pro-mb-md">{{ $t('recommend.hint') }}</p>
      <form class="pro-form" @submit.prevent="submitReferral">
        <ProInput
          v-model="referral.practiceName"
          test-id="referral-practice"
          :label="$t('recommend.practiceName')"
          required
        />
        <ProInput
          v-model="referral.contactName"
          test-id="referral-contact"
          :label="$t('recommend.contactName')"
        />
        <ProInput
          v-model="referral.contactEmail"
          test-id="referral-email"
          type="email"
          :label="$t('recommend.contactEmail')"
        />
        <ProInput
          v-model="referral.city"
          test-id="referral-city"
          :label="$t('recommend.city')"
        />
        <ProInput
          v-model="referral.notes"
          test-id="referral-notes"
          :label="$t('recommend.notes')"
        />
        <p v-if="referralError" class="pro-error" data-testid="referral-error">{{ referralError }}</p>
        <ProButton type="submit" test-id="referral-submit" :disabled="referralSaving">
          {{ $t('recommend.submit') }}
        </ProButton>
      </form>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const { t } = useI18n()

const referral = reactive({
  practiceName: '',
  contactName: '',
  contactEmail: '',
  city: '',
  notes: '',
})
const referralSaving = ref(false)
const referralError = ref('')
const thanks = ref(false)

function resetForm() {
  thanks.value = false
  referralError.value = ''
  Object.assign(referral, {
    practiceName: '',
    contactName: '',
    contactEmail: '',
    city: '',
    notes: '',
  })
}

async function submitReferral() {
  referralSaving.value = true
  referralError.value = ''
  try {
    await $fetch('/api/vet/prospects', { method: 'POST', body: { ...referral } })
    thanks.value = true
    Object.assign(referral, {
      practiceName: '',
      contactName: '',
      contactEmail: '',
      city: '',
      notes: '',
    })
  } catch {
    referralError.value = t('recommend.error')
  } finally {
    referralSaving.value = false
  }
}
</script>

<style scoped>
.recommend-thanks {
  text-align: center;
  padding: 1.25rem 0.5rem;
}

.recommend-thanks__icon {
  color: var(--pf-vet-accent);
  margin-bottom: 0.75rem;
}

.recommend-thanks h3 {
  margin: 0;
  font-size: 1.25rem;
  color: var(--pf-vet-primary);
}
</style>
