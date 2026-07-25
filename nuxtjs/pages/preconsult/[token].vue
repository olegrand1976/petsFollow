<template>
  <div class="pro-login-page preconsult-public" data-testid="preconsult-public-page">
    <aside class="pro-login-brand">
      <PetsFollowLogo variant="hero" animated />
      <h2>{{ $t('preconsultPublic.brandTitle') }}</h2>
      <p>{{ $t('preconsultPublic.brandText') }}</p>
    </aside>
    <div class="pro-login-form-panel">
      <div class="pro-auth-locale">
        <ProLocaleSelect />
      </div>
      <div class="pro-login-form">
        <PetsFollowLogo variant="default" />
        <div v-if="loading" class="text-muted">{{ $t('preconsultPublic.loading') }}</div>
        <template v-else-if="done || ctx?.status === 'submitted'">
          <h1 data-testid="preconsult-thanks-title">{{ $t('preconsultPublic.thanksTitle') }}</h1>
          <p class="pro-page-header__subtitle">{{ $t('preconsultPublic.thanksIntro', { petName: ctx?.petName || '' }) }}</p>
          <ul class="preconsult-benefits">
            <li>{{ $t('preconsultPublic.benefitFollow') }}</li>
            <li>{{ $t('preconsultPublic.benefitMessages') }}</li>
            <li>{{ $t('preconsultPublic.benefitReminders') }}</li>
          </ul>
          <div class="preconsult-store-qrs" data-testid="preconsult-store-qrs">
            <a
              v-if="qrAndroidUrl"
              class="preconsult-qr"
              :href="storeAndroidUrl || downloadUrl || '#'"
              target="_blank"
              rel="noopener noreferrer"
            >
              <img :src="qrAndroidUrl" :alt="$t('preconsultPublic.qrAndroidAlt')" width="140" height="140">
              <span>{{ $t('preconsultPublic.android') }}</span>
            </a>
            <a
              v-if="qrIosUrl"
              class="preconsult-qr"
              :href="storeIosUrl || downloadUrl || '#'"
              target="_blank"
              rel="noopener noreferrer"
            >
              <img :src="qrIosUrl" :alt="$t('preconsultPublic.qrIosAlt')" width="140" height="140">
              <span>{{ $t('preconsultPublic.ios') }}</span>
            </a>
          </div>
          <a
            v-if="inviteUrl || downloadUrl"
            class="pro-btn pro-btn--primary pro-btn--block"
            data-testid="preconsult-invite-cta"
            :href="inviteUrl || downloadUrl"
            target="_blank"
            rel="noopener noreferrer"
          >
            {{ $t('preconsultPublic.downloadCta') }}
          </a>
        </template>
        <template v-else-if="ctx">
          <h1 data-testid="preconsult-form-title">{{ $t('preconsultPublic.title') }}</h1>
          <p class="pro-page-header__subtitle">
            {{ $t('preconsultPublic.subtitle', { petName: ctx.petName, practiceName: ctx.practiceName || 'petsFollow' }) }}
          </p>
          <p v-if="ctx.scheduledAt" class="pro-hint">{{ $t('preconsultPublic.when', { when: formatWhen(ctx.scheduledAt) }) }}</p>
          <form class="pro-form" data-testid="preconsult-form" @submit.prevent="submit">
            <div class="pro-field">
              <label class="pro-label" for="pc-complaint">{{ $t('preconsultPublic.complaint') }}</label>
              <textarea
                id="pc-complaint"
                v-model="answers.chiefComplaint"
                class="pro-input"
                rows="3"
                maxlength="500"
                required
                data-testid="preconsult-complaint"
              />
            </div>
            <div class="pro-field">
              <label class="pro-label" for="pc-duration">{{ $t('preconsultPublic.duration') }}</label>
              <select id="pc-duration" v-model="answers.duration" class="pro-input" required>
                <option v-for="o in durationOpts" :key="o" :value="o">{{ $t(`calendar.preconsultEnums.duration.${o}`) }}</option>
              </select>
            </div>
            <div class="pro-field">
              <label class="pro-label" for="pc-behavior">{{ $t('preconsultPublic.behavior') }}</label>
              <select id="pc-behavior" v-model="answers.behavior" class="pro-input" required>
                <option v-for="o in behaviorOpts" :key="o" :value="o">{{ $t(`calendar.preconsultEnums.behavior.${o}`) }}</option>
              </select>
            </div>
            <div class="pro-field">
              <label class="pro-label" for="pc-appetite">{{ $t('preconsultPublic.appetite') }}</label>
              <select id="pc-appetite" v-model="answers.appetite" class="pro-input" required>
                <option v-for="o in scaleOpts" :key="o" :value="o">{{ $t(`calendar.preconsultEnums.scale.${o}`) }}</option>
              </select>
            </div>
            <div class="pro-field">
              <label class="pro-label" for="pc-thirst">{{ $t('preconsultPublic.thirst') }}</label>
              <select id="pc-thirst" v-model="answers.thirst" class="pro-input" required>
                <option v-for="o in scaleOpts" :key="o" :value="o">{{ $t(`calendar.preconsultEnums.scale.${o}`) }}</option>
              </select>
            </div>
            <div class="pro-field">
              <label class="pro-label" for="pc-elim">{{ $t('preconsultPublic.elimination') }}</label>
              <select id="pc-elim" v-model="answers.elimination" class="pro-input" required>
                <option v-for="o in scaleOpts" :key="o" :value="o">{{ $t(`calendar.preconsultEnums.scale.${o}`) }}</option>
              </select>
            </div>
            <div class="pro-field">
              <label class="pro-label" for="pc-urgency">{{ $t('preconsultPublic.urgency') }}</label>
              <select id="pc-urgency" v-model="answers.urgency" class="pro-input" required>
                <option v-for="o in urgencyOpts" :key="o" :value="o">{{ $t(`calendar.preconsultEnums.urgency.${o}`) }}</option>
              </select>
            </div>
            <div class="pro-field">
              <label class="pro-label" for="pc-comment">{{ $t('preconsultPublic.comment') }}</label>
              <textarea id="pc-comment" v-model="answers.comment" class="pro-input" rows="2" maxlength="2000" />
            </div>
            <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>
            <ProButton type="submit" block :disabled="submitting" test-id="preconsult-submit">
              {{ $t('preconsultPublic.submit') }}
            </ProButton>
          </form>
        </template>
        <template v-else>
          <h1 data-testid="preconsult-invalid">{{ $t('preconsultPublic.invalidTitle') }}</h1>
          <p class="pro-field-error" role="alert">{{ error || $t('preconsultPublic.invalid') }}</p>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const { t } = useI18n()
const { mapError } = useApiError()
const route = useRoute()
const token = computed(() => String(route.params.token || ''))

const loading = ref(true)
const submitting = ref(false)
const done = ref(false)
const error = ref('')
const ctx = ref<{
  petName: string
  practiceName: string
  scheduledAt?: string
  status: string
  inviteUrl?: string
  downloadUrl?: string
  qrAndroid?: { publicUrl?: string, storeUrl?: string } | null
  qrIos?: { publicUrl?: string, storeUrl?: string } | null
} | null>(null)

const answers = reactive({
  chiefComplaint: '',
  duration: 'unknown',
  behavior: 'unknown',
  appetite: 'unknown',
  thirst: 'unknown',
  elimination: 'unknown',
  urgency: 'medium',
  comment: '',
})

const durationOpts = ['today', 'few_days', 'week', 'weeks', 'months', 'unknown'] as const
const behaviorOpts = ['normal', 'lethargic', 'restless', 'aggressive', 'anxious', 'other', 'unknown'] as const
const scaleOpts = ['normal', 'decreased', 'increased', 'unknown'] as const
const urgencyOpts = ['low', 'medium', 'high'] as const

const inviteUrl = computed(() => ctx.value?.inviteUrl || '')
const downloadUrl = computed(() => ctx.value?.downloadUrl || '')
const qrAndroidUrl = computed(() => ctx.value?.qrAndroid?.publicUrl || '')
const qrIosUrl = computed(() => ctx.value?.qrIos?.publicUrl || '')
const storeAndroidUrl = computed(() => ctx.value?.qrAndroid?.storeUrl || '')
const storeIosUrl = computed(() => ctx.value?.qrIos?.storeUrl || '')

function formatWhen(iso: string) {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function applyPayload(data: any) {
  ctx.value = {
    petName: data.petName || '',
    practiceName: data.practiceName || '',
    scheduledAt: data.scheduledAt,
    status: data.status || data.intake?.status || 'pending',
    inviteUrl: data.inviteUrl,
    downloadUrl: data.downloadUrl,
    qrAndroid: data.qrAndroid,
    qrIos: data.qrIos,
  }
}

onMounted(async () => {
  loading.value = true
  error.value = ''
  try {
    const res: any = await $fetch(`/api/public/preconsult/${encodeURIComponent(token.value)}`)
    applyPayload(res.data ?? res)
    if (route.query.done === '1' || ctx.value?.status === 'submitted') {
      done.value = true
    }
  } catch (e: any) {
    ctx.value = null
    error.value = mapError(e) || t('preconsultPublic.invalid')
  } finally {
    loading.value = false
  }
})

async function submit() {
  submitting.value = true
  error.value = ''
  try {
    const res: any = await $fetch(`/api/public/preconsult/${encodeURIComponent(token.value)}`, {
      method: 'POST',
      body: { answers: { ...answers } },
    })
    applyPayload(res.data ?? res)
    done.value = true
    await navigateTo({ path: route.path, query: { done: '1' } }, { replace: true })
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.preconsult-benefits {
  margin: 1rem 0 1.25rem;
  padding-left: 1.25rem;
}
.preconsult-store-qrs {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin: 1rem 0 1.25rem;
  justify-content: center;
}
.preconsult-qr {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  text-decoration: none;
  color: inherit;
}
.preconsult-qr img {
  border-radius: 0.5rem;
  background: #fff;
}
</style>
