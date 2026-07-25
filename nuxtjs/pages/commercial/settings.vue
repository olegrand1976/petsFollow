<template>
  <div data-testid="commercial-settings-page">
    <ProPageHeader
      :title="$t('commercial.settings.title')"
      :subtitle="$t('commercial.settings.subtitle')"
    />
    <ProCard class="pro-mb-lg">
      <h3 class="pro-mb-md">{{ $t('settings.language.title') }}</h3>
      <p class="pro-hint pro-mb-md">{{ $t('settings.language.subtitle') }}</p>
      <ProLocaleSelect persist />
    </ProCard>
    <ProCard>
      <form class="pro-form" @submit.prevent="save">
        <h3 class="pro-mb-md">{{ $t('commercial.settings.baseLocationTitle') }}</h3>
        <p class="pro-hint pro-mb-md">{{ $t('commercial.settings.baseLocationHint') }}</p>
        <div class="pro-field">
          <label class="pro-label" for="base-city">{{ $t('commercial.settings.baseCity') }}</label>
          <input id="base-city" v-model="baseCity" class="pro-input" data-testid="commercial-base-city">
        </div>
        <div class="pro-field">
          <label class="pro-label" for="base-postal">{{ $t('commercial.settings.basePostalCode') }}</label>
          <input id="base-postal" v-model="basePostalCode" class="pro-input" data-testid="commercial-base-postal">
        </div>
        <div class="pro-field-row">
          <div class="pro-field">
            <label class="pro-label" for="base-lat">{{ $t('commercial.settings.baseLat') }}</label>
            <input id="base-lat" v-model="baseLat" class="pro-input" inputmode="decimal" data-testid="commercial-base-lat">
          </div>
          <div class="pro-field">
            <label class="pro-label" for="base-lng">{{ $t('commercial.settings.baseLng') }}</label>
            <input id="base-lng" v-model="baseLng" class="pro-input" inputmode="decimal" data-testid="commercial-base-lng">
          </div>
        </div>

        <h3 class="pro-mb-md pro-mt-lg">{{ $t('commercial.settings.title') }}</h3>
        <div class="pro-field">
          <label class="pro-label" for="iban">{{ $t('commercial.settings.iban') }}</label>
          <input id="iban" v-model="iban" class="pro-input" autocomplete="off" required>
        </div>
        <div class="pro-field">
          <label class="pro-label" for="bic">{{ $t('commercial.settings.bic') }}</label>
          <input id="bic" v-model="bic" class="pro-input" autocomplete="off">
        </div>
        <div class="pro-field">
          <label class="pro-label" for="holder">{{ $t('commercial.settings.accountHolder') }}</label>
          <input id="holder" v-model="accountHolder" class="pro-input" required>
        </div>
        <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
        <p v-if="saved" class="pro-success">{{ $t('commercial.settings.saved') }}</p>
        <ProButton type="submit" :loading="saving">{{ $t('commercial.settings.save') }}</ProButton>
      </form>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { mapError } = useApiError()
const iban = ref('')
const bic = ref('')
const accountHolder = ref('')
const baseCity = ref('')
const basePostalCode = ref('')
const baseLat = ref('')
const baseLng = ref('')
const saving = ref(false)
const saved = ref(false)
const error = ref('')

onMounted(async () => {
  const res: any = await $fetch('/api/commercial/me/payout-profile')
  const data = res.data ?? res
  iban.value = data.iban || ''
  bic.value = data.bic || ''
  accountHolder.value = data.accountHolder || ''
  baseCity.value = data.baseCity || ''
  basePostalCode.value = data.basePostalCode || ''
  baseLat.value = data.baseLat != null ? String(data.baseLat) : ''
  baseLng.value = data.baseLng != null ? String(data.baseLng) : ''
})

function parseCoord(raw: string): number | null {
  const t = raw.trim()
  if (!t) return null
  const n = Number(t)
  return Number.isFinite(n) ? n : NaN
}

async function save() {
  saving.value = true
  error.value = ''
  saved.value = false
  const lat = parseCoord(baseLat.value)
  const lng = parseCoord(baseLng.value)
  if (Number.isNaN(lat as number) || Number.isNaN(lng as number)) {
    error.value = mapError({ data: { error: { message: 'invalid_base_location' } } })
    saving.value = false
    return
  }
  try {
    await Promise.all([
      $fetch('/api/commercial/me/payout-profile', {
        method: 'PATCH',
        body: {
          iban: iban.value,
          bic: bic.value,
          accountHolder: accountHolder.value,
        },
      }),
      $fetch('/api/commercial/me/base-location', {
        method: 'PATCH',
        body: {
          baseCity: baseCity.value,
          basePostalCode: basePostalCode.value,
          baseLat: lat,
          baseLng: lng,
        },
      }),
    ])
    saved.value = true
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.pro-form {
  display: grid;
  gap: 1rem;
  max-width: 32rem;
}
.pro-field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
.pro-mt-lg {
  margin-top: 0.5rem;
}
.pro-success {
  color: var(--pf-vet-accent);
  margin: 0;
}
</style>
