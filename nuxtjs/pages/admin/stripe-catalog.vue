<template>
  <div data-testid="admin-stripe-catalog-page">
    <ProPageHeader
      :title="$t('admin.stripeCatalog.title')"
      :subtitle="$t('admin.stripeCatalog.subtitle')"
    />

    <p class="pro-hint pro-mb-md">{{ $t('admin.stripeCatalog.importHint') }}</p>

    <div class="pro-flex-gap pro-mb-lg" style="align-items: stretch">
      <ProCard style="flex: 1; min-width: 16rem" data-testid="admin-stripe-import-products">
        <h3 class="pro-mb-md">{{ $t('admin.stripeCatalog.importProducts') }}</h3>
        <form class="pro-form" @submit.prevent="upload('products')">
          <div class="pro-field">
            <label class="pro-label" for="products-file">{{ $t('admin.stripeCatalog.file') }}</label>
            <input
              id="products-file"
              type="file"
              accept=".csv,text/csv"
              class="pro-input"
              data-testid="admin-stripe-products-file"
              required
              @change="onProductsFile"
            >
          </div>
          <ProButton type="submit" test-id="admin-stripe-products-submit" :disabled="uploading || !productsFile">
            {{ $t('admin.stripeCatalog.uploadSubmit') }}
          </ProButton>
        </form>
      </ProCard>

      <ProCard style="flex: 1; min-width: 16rem" data-testid="admin-stripe-import-prices">
        <h3 class="pro-mb-md">{{ $t('admin.stripeCatalog.importPrices') }}</h3>
        <form class="pro-form" @submit.prevent="upload('prices')">
          <div class="pro-field">
            <label class="pro-label" for="prices-file">{{ $t('admin.stripeCatalog.file') }}</label>
            <input
              id="prices-file"
              type="file"
              accept=".csv,text/csv"
              class="pro-input"
              data-testid="admin-stripe-prices-file"
              required
              @change="onPricesFile"
            >
          </div>
          <ProButton type="submit" test-id="admin-stripe-prices-submit" :disabled="uploading || !pricesFile">
            {{ $t('admin.stripeCatalog.uploadSubmit') }}
          </ProButton>
        </form>
      </ProCard>
    </div>

    <p v-if="uploadError" class="pro-hint pro-hint--error pro-mb-md" data-testid="admin-stripe-upload-error">
      {{ uploadError }}
    </p>
    <p v-if="uploadOk" class="pro-hint pro-mb-md" data-testid="admin-stripe-upload-ok">
      {{ uploadOk }}
    </p>

    <ProCard class="pro-mb-lg">
      <h3 class="pro-mb-md">{{ $t('admin.stripeCatalog.productsTitle') }}</h3>
      <ProTable :empty="!catalog.products.length" :empty-title="$t('admin.stripeCatalog.emptyProducts')">
        <thead>
          <tr>
            <th>{{ $t('admin.stripeCatalog.colProductId') }}</th>
            <th>{{ $t('admin.stripeCatalog.colName') }}</th>
            <th>{{ $t('admin.stripeCatalog.colTaxCode') }}</th>
            <th>{{ $t('admin.stripeCatalog.colPlanSlug') }}</th>
            <th>{{ $t('admin.stripeCatalog.colActive') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in catalog.products" :key="p.stripeProductId">
            <td><code>{{ p.stripeProductId }}</code></td>
            <td>{{ p.name }}</td>
            <td>{{ p.taxCode || '—' }}</td>
            <td>{{ p.metadataPlanSlug || '—' }}</td>
            <td>
              <ProBadge :variant="p.active ? 'success' : 'neutral'">
                {{ p.active ? $t('admin.stripeCatalog.active') : $t('admin.stripeCatalog.inactive') }}
              </ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard>
      <h3 class="pro-mb-md">{{ $t('admin.stripeCatalog.pricesTitle') }}</h3>
      <ProTable :empty="!catalog.prices.length" :empty-title="$t('admin.stripeCatalog.emptyPrices')">
        <thead>
          <tr>
            <th>{{ $t('admin.stripeCatalog.colPriceId') }}</th>
            <th>{{ $t('admin.stripeCatalog.colProduct') }}</th>
            <th>{{ $t('admin.stripeCatalog.colPlan') }}</th>
            <th>{{ $t('admin.stripeCatalog.colMode') }}</th>
            <th>{{ $t('admin.stripeCatalog.colAmount') }}</th>
            <th>{{ $t('admin.stripeCatalog.colInterval') }}</th>
            <th>{{ $t('admin.stripeCatalog.colActive') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="pr in catalog.prices" :key="pr.stripePriceId">
            <td><code>{{ pr.stripePriceId }}</code></td>
            <td>{{ pr.productName || pr.stripeProductId }}</td>
            <td>{{ pr.planCode || '—' }}</td>
            <td>{{ pr.billingMode || '—' }}</td>
            <td>{{ formatAmount(pr.amountCents, pr.currency) }}</td>
            <td>{{ formatInterval(pr.interval, pr.intervalCount) }}</td>
            <td>
              <ProBadge :variant="pr.active ? 'success' : 'neutral'">
                {{ pr.active ? $t('admin.stripeCatalog.active') : $t('admin.stripeCatalog.inactive') }}
              </ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()

type StripeProduct = {
  stripeProductId: string
  name: string
  description?: string
  taxCode?: string
  metadataPlanSlug?: string
  active: boolean
}

type StripePrice = {
  stripePriceId: string
  stripeProductId: string
  productName?: string
  amountCents: number
  currency: string
  interval: string
  intervalCount: number
  planCode?: string | null
  billingMode?: string | null
  active: boolean
}

const catalog = ref<{ products: StripeProduct[], prices: StripePrice[] }>({
  products: [],
  prices: [],
})
const productsFile = ref<File | null>(null)
const pricesFile = ref<File | null>(null)
const uploading = ref(false)
const uploadError = ref('')
const uploadOk = ref('')

function onProductsFile(e: Event) {
  productsFile.value = (e.target as HTMLInputElement).files?.[0] ?? null
}

function onPricesFile(e: Event) {
  pricesFile.value = (e.target as HTMLInputElement).files?.[0] ?? null
}

function formatAmount(cents: number, currency: string) {
  const cur = (currency || 'eur').toUpperCase()
  return `${(cents / 100).toFixed(2)} ${cur}`
}

function formatInterval(interval: string, count: number) {
  if (!interval) return t('admin.stripeCatalog.oneTime')
  return count > 1 ? `${interval}×${count}` : interval
}

async function load() {
  const res: any = await $fetch('/api/admin/stripe-catalog')
  const data = res?.data ?? res
  catalog.value = {
    products: data?.products ?? [],
    prices: data?.prices ?? [],
  }
}

async function upload(kind: 'products' | 'prices') {
  const file = kind === 'products' ? productsFile.value : pricesFile.value
  if (!file) return
  uploading.value = true
  uploadError.value = ''
  uploadOk.value = ''
  try {
    const fd = new FormData()
    fd.append('file', file)
    fd.append('kind', kind)
    const res: any = await $fetch(`/api/admin/stripe-catalog/import?kind=${kind}`, {
      method: 'POST',
      body: fd,
    })
    const result = res?.data ?? res
    uploadOk.value = t('admin.stripeCatalog.importResult', {
      inserted: result?.inserted ?? 0,
      updated: result?.updated ?? 0,
      skipped: result?.skipped ?? 0,
    })
    if (Array.isArray(result?.errors) && result.errors.length) {
      uploadError.value = result.errors.slice(0, 5).join(' · ')
    }
    await load()
  } catch (e: any) {
    uploadError.value = e?.data?.error?.message ?? e?.statusMessage ?? t('admin.stripeCatalog.uploadFailed')
  } finally {
    uploading.value = false
  }
}

onMounted(() => { void load() })
</script>
