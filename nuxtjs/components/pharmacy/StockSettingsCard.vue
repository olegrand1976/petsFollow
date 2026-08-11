<template>
  <ProCard class="pro-mb-lg" data-testid="stock-settings">
    <h3 class="pro-mb-sm">{{ $t('pharmacy.stock.settingsTitle') }}</h3>
    <div class="stock-settings__form">
      <label class="stock-settings__check">
        <input
          type="checkbox"
          data-testid="stock-settings-auto-quarantine"
          :checked="settings.autoQuarantineExpired"
          @change="patch('autoQuarantineExpired', ($event.target as HTMLInputElement).checked)"
        >
        {{ $t('pharmacy.stock.settingsAutoQuarantine') }}
      </label>
      <label class="stock-settings__check">
        <input
          type="checkbox"
          data-testid="stock-settings-digest"
          :checked="settings.expiryDigestEnabled"
          @change="patch('expiryDigestEnabled', ($event.target as HTMLInputElement).checked)"
        >
        {{ $t('pharmacy.stock.settingsDigest') }}
      </label>
      <label class="stock-settings__check">
        <input
          type="checkbox"
          data-testid="stock-settings-notify"
          :checked="settings.notifyOnAutoQuarantine"
          @change="patch('notifyOnAutoQuarantine', ($event.target as HTMLInputElement).checked)"
        >
        {{ $t('pharmacy.stock.settingsNotifyQuarantine') }}
      </label>
      <ProButton
        variant="primary"
        test-id="stock-settings-save"
        :disabled="busy"
        @click="$emit('save')"
      >
        {{ $t('pharmacy.stock.settingsSave') }}
      </ProButton>
      <p v-if="msg" class="pro-hint" data-testid="stock-settings-msg">{{ msg }}</p>
    </div>
  </ProCard>
</template>

<script setup lang="ts">
import type { PharmacySettingsForm } from '~/composables/usePharmacyStockPage'

const props = defineProps<{
  settings: PharmacySettingsForm
  busy: boolean
  msg: string
}>()
const emit = defineEmits<{
  'update:settings': [value: PharmacySettingsForm]
  save: []
}>()

function patch<K extends keyof PharmacySettingsForm>(key: K, value: PharmacySettingsForm[K]) {
  emit('update:settings', { ...props.settings, [key]: value })
}
</script>

<style scoped>
.stock-settings__form {
  display: grid;
  gap: 0.65rem;
  max-width: 28rem;
}
.stock-settings__check {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.95rem;
}
</style>
