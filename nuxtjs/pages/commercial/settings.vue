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
const saving = ref(false)
const saved = ref(false)
const error = ref('')

onMounted(async () => {
  const res: any = await $fetch('/api/commercial/me/payout-profile')
  const data = res.data ?? res
  iban.value = data.iban || ''
  bic.value = data.bic || ''
  accountHolder.value = data.accountHolder || ''
})

async function save() {
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    await $fetch('/api/commercial/me/payout-profile', {
      method: 'PATCH',
      body: {
        iban: iban.value,
        bic: bic.value,
        accountHolder: accountHolder.value,
      },
    })
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
.pro-success {
  color: var(--pf-vet-accent);
  margin: 0;
}
</style>
