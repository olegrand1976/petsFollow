<template>
  <ProAccordionSection
    :title="$t('settings.privacy.title')"
    :description="$t('settings.sections.privacy.description')"
    data-testid="privacy-account-card"
  >
    <p class="pro-settings-hint">{{ $t('settings.privacy.exportHint') }}</p>
    <ProButton variant="secondary" :loading="exporting" test-id="settings-export-data" @click="exportData">
      {{ $t('settings.privacy.exportButton') }}
    </ProButton>
    <p v-if="exportError" class="pro-field-error" role="alert">{{ exportError }}</p>

    <hr class="pro-settings-hr">
    <h3 class="pro-settings-subtitle pro-danger-title">{{ $t('settings.deleteAccount.title') }}</h3>
    <p class="pro-settings-hint">{{ $t('settings.deleteAccount.hint') }}</p>
    <ProButton
      v-if="!deleteConfirming"
      variant="secondary"
      test-id="settings-delete-account"
      @click="deleteConfirming = true"
    >
      {{ $t('settings.deleteAccount.button') }}
    </ProButton>
    <template v-else>
      <p class="pro-field-error" role="alert">{{ $t('settings.deleteAccount.confirmText') }}</p>
      <div class="pro-danger-actions">
        <ProButton
          variant="secondary"
          :loading="deleting"
          test-id="settings-delete-account-confirm"
          @click="deleteAccount"
        >
          {{ $t('settings.deleteAccount.confirm') }}
        </ProButton>
        <ProButton variant="ghost" @click="deleteConfirming = false">
          {{ $t('common.cancel') }}
        </ProButton>
      </div>
      <p v-if="deleteError" class="pro-field-error" role="alert">{{ deleteError }}</p>
    </template>
  </ProAccordionSection>
</template>

<script setup lang="ts">
import { clearAuthTokens } from '~/composables/useAuth'

const { mapError } = useApiError()
const localePath = useLocalePath()

const exporting = ref(false)
const exportError = ref('')
const deleteConfirming = ref(false)
const deleting = ref(false)
const deleteError = ref('')

async function exportData() {
  exporting.value = true
  exportError.value = ''
  try {
    const res = await ($fetch as any)('/api/me/export', { responseType: 'blob' }) as Blob
    const url = URL.createObjectURL(res)
    const a = document.createElement('a')
    a.href = url
    a.download = `petsfollow-export-${new Date().toISOString().slice(0, 10)}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e: any) {
    exportError.value = mapError(e)
  } finally {
    exporting.value = false
  }
}

async function deleteAccount() {
  deleting.value = true
  deleteError.value = ''
  try {
    await $fetch('/api/me', { method: 'DELETE' })
    await clearAuthTokens()
    await navigateTo(localePath('/login'))
  } catch (e: any) {
    deleteError.value = mapError(e)
  } finally {
    deleting.value = false
  }
}
</script>

<style scoped>
.pro-settings-hint {
  color: var(--pf-vet-text-muted);
  font-size: 0.9rem;
  margin-bottom: 1rem;
}
.pro-settings-hr {
  border: 0;
  border-top: 1px solid var(--pf-vet-border);
  margin: 1.25rem 0;
}
.pro-settings-subtitle {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}
.pro-danger-title {
  color: var(--pf-vet-alert);
}
.pro-danger-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
</style>
