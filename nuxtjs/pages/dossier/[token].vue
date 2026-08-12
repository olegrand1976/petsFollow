<template>
  <div class="pro-login-page" data-testid="dossier-public-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('dossierPublic.brandTitle') }}</h2>
      <p>{{ $t('dossierPublic.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <div class="pro-login-form">
        <PetsFollowLogo variant="default" />
        <div v-if="loading" class="text-muted">{{ $t('dossierPublic.loading') }}</div>
        <template v-else-if="expired">
          <h1 data-testid="dossier-expired-title">{{ $t('dossierPublic.expiredTitle') }}</h1>
          <p class="pro-page-header__subtitle">{{ $t('dossierPublic.expiredText') }}</p>
          <a
            v-if="registerUrl"
            class="pro-btn pro-btn--primary pro-btn--block"
            :href="registerUrl"
            data-testid="dossier-register-cta"
          >
            {{ $t('dossierPublic.registerCta') }}
          </a>
        </template>
        <template v-else-if="error">
          <h1>{{ $t('dossierPublic.errorTitle') }}</h1>
          <p class="pro-page-header__subtitle">{{ error }}</p>
        </template>
        <template v-else-if="meta">
          <h1 data-testid="dossier-title">{{ $t('dossierPublic.title', { petName: meta.petName }) }}</h1>
          <p class="pro-page-header__subtitle">{{ $t('dossierPublic.subtitle') }}</p>
          <p v-if="meta.expiresAt" class="pro-hint">{{ $t('dossierPublic.expires', { when: formatWhen(meta.expiresAt) }) }}</p>

          <ProButton
            block
            :loading="downloading"
            test-id="dossier-download"
            @click="download"
          >
            {{ $t('dossierPublic.downloadCta') }}
          </ProButton>
          <p v-if="downloadError" class="pro-field-error" role="alert">{{ downloadError }}</p>

          <div class="pro-mt-lg">
            <h2 class="pro-h3">{{ $t('dossierPublic.marketingTitle') }}</h2>
            <p>{{ $t('dossierPublic.marketingText') }}</p>
            <ul class="preconsult-benefits">
              <li>{{ $t('dossierPublic.benefitFollow') }}</li>
              <li>{{ $t('dossierPublic.benefitMessages') }}</li>
              <li>{{ $t('dossierPublic.benefitContinuity') }}</li>
            </ul>
            <a
              v-if="meta.registerUrl"
              class="pro-btn pro-btn--secondary pro-btn--block"
              :href="meta.registerUrl"
              data-testid="dossier-register-cta"
            >
              {{ $t('dossierPublic.registerCta') }}
            </a>
            <p v-if="meta.productsUrl" class="pro-hint pro-mt-md">
              <a :href="meta.productsUrl">{{ $t('dossierPublic.productsLink') }}</a>
            </p>
            <p v-if="meta.commercialPhone" class="pro-mt-md" data-testid="dossier-commercial-phone">
              <strong>{{ meta.commercialName || 'petsFollow' }}</strong><br>
              {{ $t('dossierPublic.commercialPhone', { phone: meta.commercialPhone }) }}
            </p>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const route = useRoute()
const { t, locale } = useI18n()
const { mapError } = useApiError()

type DossierMeta = {
  petName?: string
  expiresAt?: string
  commercialName?: string
  commercialPhone?: string
  registerUrl?: string
  siteUrl?: string
  productsUrl?: string
}

const loading = ref(true)
const downloading = ref(false)
const expired = ref(false)
const error = ref('')
const downloadError = ref('')
const meta = ref<DossierMeta | null>(null)
const registerUrl = ref('')

function formatWhen(iso: string) {
  try {
    return new Date(iso).toLocaleString(locale.value || 'fr')
  } catch {
    return iso
  }
}

onMounted(async () => {
  const token = String(route.params.token || '')
  try {
    const res: any = await $fetch(`/api/public/pet-dossier/${encodeURIComponent(token)}`)
    meta.value = res.data ?? res
    registerUrl.value = meta.value?.registerUrl || ''
  } catch (e: any) {
    const status = e?.statusCode || e?.status || e?.response?.status
    if (status === 410) {
      expired.value = true
      // Keep a default signup CTA even when the share payload is gone.
      registerUrl.value = '/register'
    } else {
      error.value = mapError(e) || t('dossierPublic.errorTitle')
    }
  } finally {
    loading.value = false
  }
})

async function download() {
  downloadError.value = ''
  downloading.value = true
  const token = String(route.params.token || '')
  try {
    const blob = await ($fetch as any)(`/api/public/pet-dossier/${encodeURIComponent(token)}/download`, {
      responseType: 'blob',
    }) as Blob
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `dossier-${meta.value?.petName || 'animal'}.zip`
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  } catch (e: any) {
    const status = e?.statusCode || e?.status || e?.response?.status
    if (status === 410) {
      expired.value = true
      meta.value = null
    } else {
      downloadError.value = mapError(e) || t('dossierPublic.downloadError')
    }
  } finally {
    downloading.value = false
  }
}
</script>
