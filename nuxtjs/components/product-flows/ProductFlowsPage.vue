<template>
  <div class="pf-flows" data-testid="product-flows-page">
    <ProPageHeader :title="$t('productFlows.title')" :subtitle="$t('productFlows.subtitle')" />

    <ProCard class="pf-flows__howto">
      <h2 class="pf-flows__h2">{{ $t('productFlows.howtoTitle') }}</h2>
      <ol class="pf-flows__howto-list">
        <li>{{ $t('productFlows.howto1') }}</li>
        <li>{{ $t('productFlows.howto2') }}</li>
        <li>{{ $t('productFlows.howto3') }}</li>
      </ol>
    </ProCard>

    <h2 id="pf-flows-profiles-label" class="pf-flows__h2">{{ $t('productFlows.selectProfile') }}</h2>
    <div
      class="pf-flows__profiles"
      data-testid="product-flows-profiles"
      role="radiogroup"
      aria-labelledby="pf-flows-profiles-label"
      @keydown="onProfileKeydown"
    >
      <button
        v-for="p in profiles"
        :key="p.id"
        type="button"
        class="pro-card pro-card--interactive pf-flows__profile"
        :class="{ 'pf-flows__profile--active': selectedId === p.id }"
        role="radio"
        :aria-checked="selectedId === p.id"
        :tabindex="selectedId === p.id ? 0 : -1"
        :data-testid="`flow-profile-${p.id}`"
        @click="selectProfile(p.id)"
      >
        <ProIcon :name="p.icon" :size="28" />
        <strong>{{ $t(`productFlows.profiles.${p.id}.label`) }}</strong>
        <span class="pf-flows__profile-surface">{{ $t(`productFlows.surface.${p.surface}`) }}</span>
        <ProBadge v-if="p.tagDev" variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
      </button>
    </div>

    <template v-if="selected">
      <ProCard class="pro-mb-lg" data-testid="product-flows-detail">
        <div class="pf-flows__detail-head">
          <div>
            <h2 class="pf-flows__h2 pf-flows__h2--flush">
              {{ $t(`productFlows.profiles.${selected.id}.label`) }}
            </h2>
            <p class="pf-flows__mission">
              {{ $t(`productFlows.profiles.${selected.id}.mission`) }}
            </p>
          </div>
          <ProBadge variant="neutral">{{ $t(`productFlows.surface.${selected.surface}`) }}</ProBadge>
        </div>

        <h3 class="pf-flows__h3">{{ $t('productFlows.flowTitle') }}</h3>
        <ol class="pf-flows__steps" data-testid="product-flows-steps">
          <li v-for="(stepId, idx) in selected.stepIds" :key="stepId" class="pf-flows__step">
            <span class="pf-flows__step-n" aria-hidden="true">{{ idx + 1 }}</span>
            <span>{{ $t(`productFlows.steps.${stepId}`) }}</span>
            <ProIcon
              v-if="idx < selected.stepIds.length - 1"
              class="pf-flows__step-arrow"
              name="arrow_forward"
              :size="18"
              aria-hidden="true"
            />
          </li>
        </ol>

        <h3 class="pf-flows__h3">{{ $t('productFlows.featuresTitle') }}</h3>
        <ul class="pf-flows__features" data-testid="product-flows-features">
          <li v-for="fid in selected.featureIds" :key="fid" class="pf-flows__feature">
            <strong>{{ $t(`productFlows.features.${fid}.title`) }}</strong>
            <span>{{ $t(`productFlows.features.${fid}.desc`) }}</span>
          </li>
        </ul>
      </ProCard>

      <ProCard class="pro-mb-lg" data-testid="product-flows-links">
        <h2 class="pf-flows__h2">{{ $t('productFlows.linksTitle') }}</h2>
        <p class="pf-flows__blurb">{{ $t('productFlows.linksBlurb') }}</p>
        <ul v-if="profileLinks.length" class="pf-flows__links">
          <li v-for="link in profileLinks" :key="link.id" class="pf-flows__link">
            <button
              type="button"
              class="pf-flows__link-btn"
              :data-testid="`flow-link-${link.id}`"
              @click="focusOtherProfile(link)"
            >
              <span class="pf-flows__link-profiles">
                <span>{{ $t(`productFlows.profiles.${link.fromProfile}.label`) }}</span>
                <ProIcon name="sync_alt" :size="16" aria-hidden="true" />
                <span>{{ $t(`productFlows.profiles.${link.toProfile}.label`) }}</span>
              </span>
              <strong>{{ $t(`productFlows.links.${link.id}.label`) }}</strong>
              <span class="pf-flows__link-desc">{{ $t(`productFlows.links.${link.id}.desc`) }}</span>
              <ProBadge variant="neutral">{{ $t(`productFlows.features.${link.featureId}.title`) }}</ProBadge>
            </button>
          </li>
        </ul>
        <p v-else class="pro-hint">{{ $t('productFlows.linksEmpty') }}</p>
      </ProCard>
    </template>

    <ProCard v-if="isStagingLike" data-testid="product-flows-usecases-cta">
      <p class="pro-hint">{{ $t('productFlows.usecasesCtaHint') }}</p>
      <ProButton test-id="product-flows-open-usecases" @click="navigateTo('/usecases')">
        {{ $t('productFlows.usecasesCta') }}
      </ProButton>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import {
  linksForProfile,
  parseFlowProfileId,
  productFlowsCatalog,
} from '~/data/product-flows/catalog'
import type { FlowLink, FlowProfileId } from '~/data/product-flows/types'

const route = useRoute()
const router = useRouter()
const { isStagingLike } = useAppEnv()
const profiles = productFlowsCatalog.profiles
const profileIds = profiles.map((p) => p.id)

/** Source de vérité = query `?profile=` (évite race remount middleware). */
const selectedId = computed<FlowProfileId>(
  () => parseFlowProfileId(route.query.profile) ?? 'vet',
)

const selected = computed(() => profiles.find((p) => p.id === selectedId.value))

const profileLinks = computed(() => linksForProfile(selectedId.value))

function selectProfile(id: FlowProfileId) {
  if (selectedId.value === id && route.query.profile === id) return
  void router.replace({ path: '/flux', query: { profile: id } })
}

function focusProfileButton(id: FlowProfileId) {
  void nextTick(() => {
    const el = document.querySelector<HTMLElement>(`[data-testid="flow-profile-${id}"]`)
    el?.focus()
  })
}

function onProfileKeydown(e: KeyboardEvent) {
  const idx = profileIds.indexOf(selectedId.value)
  if (idx < 0) return
  let next: FlowProfileId | null = null
  switch (e.key) {
    case 'ArrowRight':
    case 'ArrowDown':
      next = profileIds[(idx + 1) % profileIds.length] ?? null
      break
    case 'ArrowLeft':
    case 'ArrowUp':
      next = profileIds[(idx - 1 + profileIds.length) % profileIds.length] ?? null
      break
    case 'Home':
      next = profileIds[0] ?? null
      break
    case 'End':
      next = profileIds[profileIds.length - 1] ?? null
      break
    default:
      return
  }
  if (!next) return
  e.preventDefault()
  selectProfile(next)
  focusProfileButton(next)
}

function focusOtherProfile(link: FlowLink) {
  const next =
    link.fromProfile === selectedId.value ? link.toProfile : link.fromProfile
  selectProfile(next)
  focusProfileButton(next)
}
</script>

<style scoped>
.pf-flows__howto {
  margin-bottom: 1.25rem;
}
.pf-flows__h2 {
  margin: 0 0 0.75rem;
  font-size: 1.125rem;
  color: var(--pf-vet-primary);
}
.pf-flows__h2--flush {
  margin-bottom: 0.35rem;
}
.pf-flows__h3 {
  margin: 1.25rem 0 0.65rem;
  font-size: 1rem;
  color: var(--pf-vet-primary);
}
.pf-flows__howto-list {
  margin: 0;
  padding-left: 1.25rem;
  line-height: 1.55;
}
.pf-flows__profiles {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(10.5rem, 1fr));
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}
.pf-flows__profile {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
  text-align: left;
  cursor: pointer;
  border: 1px solid var(--pf-vet-border);
  padding: 0.85rem;
}
.pf-flows__profile--active {
  border-color: var(--pf-vet-accent);
  box-shadow: var(--pf-vet-shadow-sm);
}
.pf-flows__profile-surface {
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
}
.pf-flows__detail-head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
}
.pf-flows__mission {
  margin: 0;
  color: var(--pf-vet-muted, #64748b);
  line-height: 1.45;
  font-size: 0.95rem;
}
.pf-flows__steps {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 0.35rem;
  align-items: center;
}
.pf-flows__step {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  border-radius: 0.5rem;
  padding: 0.45rem 0.65rem;
  font-size: 0.9rem;
}
.pf-flows__step-n {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.35rem;
  height: 1.35rem;
  border-radius: 999px;
  background: var(--pf-vet-primary);
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
  flex-shrink: 0;
}
.pf-flows__step-arrow {
  color: var(--pf-vet-muted, #64748b);
  margin-left: 0.15rem;
}
.pf-flows__features {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.65rem;
}
.pf-flows__feature {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  padding: 0.55rem 0.7rem;
  border-left: 3px solid var(--pf-vet-accent);
  background: var(--pf-vet-bg, #f8fafc);
}
.pf-flows__feature span {
  font-size: 0.875rem;
  color: var(--pf-vet-muted, #64748b);
  line-height: 1.4;
}
.pf-flows__blurb {
  margin: 0 0 0.85rem;
  font-size: 0.9rem;
  color: var(--pf-vet-muted, #64748b);
}
.pf-flows__links {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.65rem;
}
.pf-flows__link-btn {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.3rem;
  text-align: left;
  padding: 0.75rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 0.5rem;
  background: var(--pf-vet-surface);
  cursor: pointer;
  color: inherit;
  font: inherit;
}
.pf-flows__link-btn:hover {
  border-color: var(--pf-vet-accent);
}
.pf-flows__link-profiles {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
}
.pf-flows__link-desc {
  font-size: 0.875rem;
  color: var(--pf-vet-muted, #64748b);
  line-height: 1.4;
}
@media (max-width: 640px) {
  .pf-flows__detail-head {
    flex-direction: column;
  }
  .pf-flows__step-arrow {
    display: none;
  }
}
</style>
