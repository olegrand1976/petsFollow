<template>
  <div data-testid="admin-brand-assets-page">
    <ProPageHeader
      :title="$t('admin.brandAssets.title')"
      :subtitle="$t('admin.brandAssets.subtitle')"
    />
    <p class="pro-hint pro-mb-md">{{ $t('admin.brandAssets.hint') }}</p>

    <div class="pro-grid-2">
      <ProCard
        v-for="slot in slots"
        :key="slot.key"
        :title="$t(`admin.brandAssets.${slot.labelKey}`)"
        :data-testid="`admin-brand-${slot.key}`"
      >
        <div v-if="assetMap[slot.key]?.publicUrl" class="brand-preview pro-mb-md">
          <img :src="assetMap[slot.key].publicUrl" :alt="slot.key" width="160" height="160">
        </div>
        <p v-else class="pro-hint pro-mb-md">{{ $t('admin.brandAssets.missing') }}</p>
        <form class="pro-form" @submit.prevent="upload(slot.key)">
          <div class="pro-field">
            <label class="pro-label" :for="`${slot.key}-file`">{{ $t('admin.brandAssets.file') }}</label>
            <input
              :id="`${slot.key}-file`"
              type="file"
              accept="image/png,image/webp,image/jpeg"
              class="pro-input"
              @change="onFile(slot.key, $event)"
            >
          </div>
          <div class="pro-field">
            <label class="pro-label" :for="`${slot.key}-url`">{{ $t('admin.brandAssets.storeUrl') }}</label>
            <input
              :id="`${slot.key}-url`"
              v-model="storeUrls[slot.key]"
              class="pro-input"
              type="url"
              :placeholder="slot.placeholder"
            >
          </div>
          <div class="pro-flex-gap">
            <ProButton type="submit" :disabled="busy || !files[slot.key]">
              {{ $t('admin.brandAssets.upload') }}
            </ProButton>
            <ProButton type="button" variant="secondary" :disabled="busy" @click="patchUrl(slot.key)">
              {{ $t('admin.brandAssets.saveUrl') }}
            </ProButton>
          </div>
        </form>
      </ProCard>
    </div>
    <p v-if="msg" class="pro-hint pro-mt-md" data-testid="admin-brand-msg">{{ msg }}</p>
    <p v-if="err" class="pro-hint pro-hint--error pro-mt-md" role="alert">{{ err }}</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

const { t } = useI18n()
const { mapError } = useApiError()

type Asset = { key: string, publicUrl: string, storeUrl: string }

const slots = [
  { key: 'qr_android', labelKey: 'android', placeholder: 'https://play.google.com/...' },
  { key: 'qr_ios', labelKey: 'ios', placeholder: 'https://apps.apple.com/...' },
] as const

const assets = ref<Asset[]>([])
const assetMap = computed(() => Object.fromEntries(assets.value.map(a => [a.key, a])))
const files = reactive<Record<string, File | null>>({ qr_android: null, qr_ios: null })
const storeUrls = reactive<Record<string, string>>({ qr_android: '', qr_ios: '' })
const busy = ref(false)
const msg = ref('')
const err = ref('')

function onFile(key: string, e: Event) {
  const input = e.target as HTMLInputElement
  files[key] = input.files?.[0] || null
}

async function load() {
  const res: any = await $fetch('/api/admin/brand-assets')
  assets.value = res.data ?? res ?? []
  for (const a of assets.value) {
    storeUrls[a.key] = a.storeUrl || ''
  }
}

async function upload(key: string) {
  const file = files[key]
  if (!file) return
  busy.value = true
  err.value = ''
  msg.value = ''
  try {
    const fd = new FormData()
    fd.append('file', file)
    if (storeUrls[key]) fd.append('storeUrl', storeUrls[key])
    await $fetch(`/api/admin/brand-assets/${key}`, { method: 'POST', body: fd })
    msg.value = t('admin.brandAssets.uploaded')
    files[key] = null
    await load()
  } catch (e: any) {
    err.value = mapError(e)
  } finally {
    busy.value = false
  }
}

async function patchUrl(key: string) {
  busy.value = true
  err.value = ''
  msg.value = ''
  try {
    await $fetch(`/api/admin/brand-assets/${key}`, {
      method: 'PATCH',
      body: { storeUrl: storeUrls[key] || '' },
    })
    msg.value = t('admin.brandAssets.urlSaved')
    await load()
  } catch (e: any) {
    err.value = mapError(e)
  } finally {
    busy.value = false
  }
}

onMounted(() => { void load() })
</script>

<style scoped>
.brand-preview img {
  border-radius: 0.5rem;
  background: #fff;
  border: 1px solid var(--pf-vet-border, #ddd);
}
</style>
