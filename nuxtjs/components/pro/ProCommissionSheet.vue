<template>
  <div class="pf-commission-sheet" :data-audience="audience" data-testid="commission-sheet">
    <p v-if="leadText" class="pf-commission-sheet__lead">{{ leadText }}</p>
    <p class="pf-commission-sheet__meta text-muted">{{ $t('commissionSheet.htvaNote') }}</p>

    <div v-if="showVetGrid" class="pf-commission-sheet__block">
      <h3 v-if="audience !== 'vet'" class="pf-commission-sheet__h">{{ $t('commissionSheet.vetTitle') }}</h3>
      <p class="pf-commission-sheet__explain">{{ $t('commissionSheet.vetBaseExplain') }}</p>
      <ul class="pf-commission-sheet__tiers">
        <li v-for="t in vetTierLabels" :key="t">{{ t }}</li>
      </ul>
      <p class="pf-commission-sheet__explain">{{ $t('commissionSheet.vetMaxExplain') }}</p>
      <ProTable>
        <thead>
          <tr>
            <th>{{ $t('commissionSheet.colPlan') }}</th>
            <th>{{ $t('commissionSheet.colVetRate') }}</th>
            <th>{{ $t('commissionSheet.colVetEuros') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in planRates"
            :key="'vet-' + row.code"
            :class="{ 'pf-commission-sheet__row--rec': row.recommended }"
          >
            <td>
              {{ planLabel(row.code) }}
              <ProBadge v-if="row.recommended" variant="success">{{ $t('commissionSheet.recommended') }}</ProBadge>
            </td>
            <td>{{ formatPct(row.vetRateBpsMax) }}</td>
            <td>{{ formatCurrency(row.vetCentsMax) }}</td>
          </tr>
          <tr v-for="row in vetAddonRows" :key="'vet-addon-' + row.code">
            <td>{{ planLabel(row.code) }}</td>
            <td>{{ formatPct(row.vetRateBpsMax) }}</td>
            <td>{{ formatCurrency(row.vetCentsMax) }}</td>
          </tr>
        </tbody>
      </ProTable>
      <p
        v-if="(audience === 'vet' || audience === 'admin') && vetAddonRows.length"
        class="pf-commission-sheet__addon-note"
        role="note"
        data-testid="vet-addon-commission-note"
      >
        <ProIcon name="info" :size="18" />
        <span>{{ $t('commissionSheet.vetAddonNote') }}</span>
      </p>
      <p class="pro-hint">{{ $t('commissionSheet.discountNote') }}</p>
    </div>

    <div v-if="showCommercialGrid" class="pf-commission-sheet__block">
      <h3 class="pf-commission-sheet__h">{{ $t('commissionSheet.commercialTitle') }}</h3>
      <ProTable>
        <thead>
          <tr>
            <th>{{ $t('commissionSheet.colPlan') }}</th>
            <th>{{ $t('commissionSheet.colCommRate') }}</th>
            <th>{{ $t('commissionSheet.colCommEuros') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in planRates"
            :key="'comm-' + row.code"
            :class="{ 'pf-commission-sheet__row--rec': row.recommended }"
          >
            <td>{{ planLabel(row.code) }}</td>
            <td>{{ formatPct(row.commercialRateBps) }}</td>
            <td>{{ formatCurrency(row.commercialCents) }}</td>
          </tr>
          <tr v-for="row in addonRates" :key="'addon-' + row.code">
            <td>{{ planLabel(row.code) }}</td>
            <td>{{ formatPct(row.commercialRateBps) }}</td>
            <td>{{ formatCurrency(row.commercialCents) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </div>

    <div v-if="visibleBonuses.length" class="pf-commission-sheet__block">
      <h3 class="pf-commission-sheet__h">{{ $t('commissionSheet.bonusesTitle') }}</h3>
      <div class="pf-commission-sheet__bonuses">
        <ProCard v-for="b in visibleBonuses" :key="b.code" class="pf-commission-sheet__bonus">
          <strong>{{ bonusTitle(b) }}</strong>
          <p>{{ formatCurrency(b.amountCents) }}</p>
          <p class="text-muted">{{ bonusHint(b) }}</p>
          <ProBadge v-if="b.status" :variant="bonusVariant(b.status)">{{ $t(`commissionSheet.status.${b.status}`) }}</ProBadge>
        </ProCard>
      </div>
    </div>

    <p v-if="audience === 'commercial'" class="pro-hint">{{ $t('commissionSheet.coSellNote') }}</p>
    <p v-if="audience === 'admin'" class="pro-hint">{{ $t('commissionSheet.adminGuardrails') }}</p>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  audience: 'vet' | 'commercial' | 'admin'
  planRates?: any[]
  addonRates?: any[]
  bonuses?: any[]
}>(), {
  planRates: () => [],
  addonRates: () => [],
  bonuses: () => [],
})

const { t } = useI18n()
const { formatCurrency } = useFormatters()

const showVetGrid = computed(() => props.audience === 'vet' || props.audience === 'commercial' || props.audience === 'admin')
const showCommercialGrid = computed(() => props.audience === 'commercial' || props.audience === 'admin')

const vetAddonRows = computed(() =>
  (props.addonRates || []).filter((r: any) => (r.vetRateBpsMax || 0) > 0),
)

const leadText = computed(() => {
  switch (props.audience) {
    case 'vet':
      // Page véto : pas de lead redondant — grille + tableau suffisent.
      return ''
    case 'commercial':
      return t('commissionSheet.leadCommercial')
    case 'admin':
      return t('commissionSheet.leadAdmin')
    default: {
      const _exhaustive: never = props.audience
      return _exhaustive
    }
  }
})

const vetTierLabels = computed(() => [
  t('commissionSheet.tier', { min: 1, max: 10, pct: 7 }),
  t('commissionSheet.tier', { min: 11, max: 30, pct: 9 }),
  t('commissionSheet.tier', { min: 31, max: 60, pct: 11 }),
  t('commissionSheet.tierOpen', { min: 61, pct: 12 }),
])

const visibleBonuses = computed(() => {
  const list = props.bonuses?.length ? props.bonuses : []
  if (props.audience === 'admin') return list
  if (props.audience === 'vet') return list.filter((b: any) => b.audience === 'vet')
  // commercial: own bonuses; vet bonus as summary if present
  return list.filter((b: any) => b.audience === 'commercial' || b.audience === 'vet')
})

function formatPct(bps: number) {
  return `${((bps || 0) / 100).toFixed(0)} %`
}

function planLabel(code: string) {
  const key = `commissionSheet.plans.${code}`
  const translated = t(key)
  return translated === key ? code : translated
}

function bonusTitle(b: any) {
  return t(`commissionSheet.bonusTitles.${b.code}`, b.code)
}

function bonusHint(b: any) {
  return t(`commissionSheet.bonusHints.${b.code}`, '')
}

function bonusVariant(status: string): 'success' | 'warning' | 'neutral' {
  if (status === 'earned' || status === 'paid') return 'success'
  if (status === 'in_progress') return 'warning'
  return 'neutral'
}
</script>

<style scoped>
.pf-commission-sheet__lead {
  font-weight: 600;
  margin: 0 0 0.35rem;
}
.pf-commission-sheet__meta {
  margin: 0 0 1rem;
  font-size: 0.9rem;
}
.pf-commission-sheet__block {
  margin-bottom: 1.25rem;
}
.pf-commission-sheet__h {
  margin: 0 0 0.5rem;
  font-size: 1rem;
}
.pf-commission-sheet__tiers {
  margin: 0 0 0.75rem;
  padding-left: 1.2rem;
}
.pf-commission-sheet__explain {
  margin: 0 0 0.5rem;
  font-size: 0.9rem;
  color: var(--pf-vet-text-muted);
  line-height: 1.45;
}
.pf-commission-sheet__addon-note {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin: 0.85rem 0 0;
  padding: 0.75rem 0.9rem;
  border-radius: var(--pf-vet-radius);
  background: color-mix(in srgb, var(--pf-vet-accent) 8%, var(--pf-vet-surface));
  border: 1px solid color-mix(in srgb, var(--pf-vet-accent) 28%, transparent);
  font-size: 0.9rem;
  line-height: 1.4;
  color: var(--pf-vet-primary);
}
.pf-commission-sheet__addon-note .pro-icon {
  flex-shrink: 0;
  margin-top: 0.1rem;
  color: var(--pf-vet-accent);
}
.pf-commission-sheet__row--rec td {
  font-weight: 600;
}
.pf-commission-sheet__bonuses {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}
.pf-commission-sheet__bonus p {
  margin: 0.25rem 0;
}
@media print {
  .pf-commission-sheet {
    break-inside: avoid;
  }
}
</style>
