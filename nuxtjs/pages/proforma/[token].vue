<template>
  <div class="pro-login-page proforma-public" data-testid="proforma-public-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('proformaPublic.brandTitle') }}</h2>
      <p>{{ $t('proformaPublic.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <div class="pro-login-form">
        <PetsFollowLogo variant="default" />
        <div v-if="loading" class="text-muted">{{ $t('proformaPublic.loading') }}</div>
        <template v-else-if="done || ctx?.status === 'accepted'">
          <h1 data-testid="proforma-thanks-title">{{ $t('proformaPublic.thanksTitle') }}</h1>
          <p class="pro-page-header__subtitle">{{ $t('proformaPublic.thanksIntro') }}</p>
        </template>
        <template v-else-if="ctx">
          <h1 data-testid="proforma-form-title">{{ $t('proformaPublic.title') }}</h1>
          <p class="pro-page-header__subtitle">
            {{ $t('proformaPublic.subtitle', { practiceName: ctx.practiceName || 'petsFollow' }) }}
          </p>
          <p class="proforma-total" data-testid="proforma-total">
            {{ $t('proformaPublic.total') }} :
            <strong>{{ formatMoney(ctx.totalInclCents) }}</strong>
          </p>
          <ul v-if="ctx.lines?.length" class="proforma-lines">
            <li v-for="(line, i) in ctx.lines" :key="i">
              {{ line.description }} — {{ line.quantity }} × {{ formatMoney(line.unitPriceExclCents) }}
            </li>
          </ul>
          <p v-if="error" class="pro-alert pro-alert--danger" data-testid="proforma-error">{{ error }}</p>
          <ProButton
            v-if="ctx.canAccept"
            test-id="proforma-accept"
            :disabled="busy"
            @click="accept"
          >
            {{ $t('proformaPublic.accept') }}
          </ProButton>
          <p v-else class="pro-hint">{{ $t('proformaPublic.expired') }}</p>
        </template>
        <p v-else class="pro-alert pro-alert--danger" data-testid="proforma-invalid">
          {{ error || $t('proformaPublic.invalid') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false, auth: false })

const { t } = useI18n()
const route = useRoute()
const token = computed(() => String(route.params.token || ''))

type Line = { description: string, quantity: number, unitPriceExclCents: number }
type Ctx = {
  status: string
  practiceName?: string
  totalInclCents: number
  lines?: Line[]
  canAccept: boolean
}

const loading = ref(true)
const busy = ref(false)
const done = ref(false)
const error = ref('')
const ctx = ref<Ctx | null>(null)

function unwrap<T>(res: any): T {
  return (res?.data ?? res) as T
}

function formatMoney(cents: number) {
  return `${(Number(cents) / 100).toFixed(2)} €`
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await $fetch(`/api/public/proforma/${encodeURIComponent(token.value)}`)
    ctx.value = unwrap<Ctx>(res)
  } catch (e: any) {
    ctx.value = null
    const msgKey = e?.data?.error?.msgKey || e?.data?.data?.error?.msgKey
    if (msgKey === 'proforma_expired') error.value = t('proformaPublic.expired')
    else if (msgKey === 'proforma_token_invalid') error.value = t('proformaPublic.invalid')
    else error.value = t('proformaPublic.errorGeneric')
  } finally {
    loading.value = false
  }
}

async function accept() {
  busy.value = true
  error.value = ''
  try {
    await $fetch(`/api/public/proforma/${encodeURIComponent(token.value)}/accept`, { method: 'POST' })
    done.value = true
  } catch (e: any) {
    const msgKey = e?.data?.error?.msgKey || e?.data?.data?.error?.msgKey
    if (msgKey === 'proforma_expired') error.value = t('proformaPublic.expired')
    else if (msgKey === 'proforma_already_accepted') {
      done.value = true
    } else {
      error.value = t('proformaPublic.errorGeneric')
    }
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.proforma-total {
  margin: 1rem 0 0.5rem;
  font-size: 1.1rem;
}
.proforma-lines {
  margin: 0 0 1.25rem;
  padding-left: 1.1rem;
  color: var(--pf-vet-muted, #64748b);
}
</style>
