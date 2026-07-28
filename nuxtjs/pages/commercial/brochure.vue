<template>
  <div class="brochure" data-testid="commercial-brochure">
    <div class="brochure__toolbar no-print">
      <ProPageHeader
        :title="$t('commercial.brochure.title')"
        :subtitle="$t('commercial.brochure.subtitle')"
        title-tag="p"
      >
        <template #actions>
          <ProButton variant="secondary" @click="copyLink">{{ $t('commercial.brochure.copyLink') }}</ProButton>
          <ProButton variant="secondary" @click="mailTo">{{ $t('commercial.brochure.mailto') }}</ProButton>
          <ProButton test-id="brochure-print" @click="printDoc">{{ $t('commercial.brochure.print') }}</ProButton>
        </template>
      </ProPageHeader>
      <p class="pro-hint">{{ $t('commercial.brochure.printHint') }}</p>
      <p v-if="toast" class="pro-hint">{{ toast }}</p>
    </div>

    <article class="brochure__doc">
      <header class="brochure__hero">
        <div class="brochure__brand">
          <PetsFollowLogo variant="default" />
          <p class="brochure__badge">{{ $t('commercial.brochure.badge') }}</p>
        </div>
        <h1>{{ $t('commercial.brochure.docTitle') }}</h1>
        <p class="brochure__lead">{{ $t('commercial.brochure.lead') }}</p>
        <p class="brochure__tagline">{{ $t('commercial.brochure.tagline') }}</p>
      </header>

      <section class="brochure__section">
        <h2>{{ $t('commercial.brochure.problemTitle') }}</h2>
        <ul>
          <li>{{ $t('commercial.brochure.problem1') }}</li>
          <li>{{ $t('commercial.brochure.problem2') }}</li>
          <li>{{ $t('commercial.brochure.problem3') }}</li>
        </ul>
      </section>

      <section class="brochure__section">
        <h2>{{ $t('commercial.brochure.passportTitle') }}</h2>
        <p>{{ $t('commercial.brochure.passportLead') }}</p>
        <div class="brochure__actors" :aria-label="$t('commercial.brochure.actorsLabel')">
          <div class="brochure__actor">
            <strong>{{ $t('commercial.brochure.actorPro') }}</strong>
            <span>{{ $t('commercial.brochure.actorProRole') }}</span>
          </div>
          <div class="brochure__actor-arrow" aria-hidden="true">↔</div>
          <div class="brochure__actor">
            <strong>{{ $t('commercial.brochure.actorOwner') }}</strong>
            <span>{{ $t('commercial.brochure.actorOwnerRole') }}</span>
          </div>
          <div class="brochure__actor-arrow" aria-hidden="true">↔</div>
          <div class="brochure__actor">
            <strong>{{ $t('commercial.brochure.actorCare') }}</strong>
            <span>{{ $t('commercial.brochure.actorCareRole') }}</span>
          </div>
        </div>
      </section>

      <section class="brochure__section">
        <h2>{{ $t('commercial.brochure.stepsTitle') }}</h2>
        <ol class="brochure__steps">
          <li><strong>{{ $t('commercial.brochure.step1Label') }}</strong> — {{ $t('commercial.brochure.step1') }}</li>
          <li><strong>{{ $t('commercial.brochure.step2Label') }}</strong> — {{ $t('commercial.brochure.step2') }}</li>
          <li><strong>{{ $t('commercial.brochure.step3Label') }}</strong> — {{ $t('commercial.brochure.step3') }}</li>
          <li><strong>{{ $t('commercial.brochure.step4Label') }}</strong> — {{ $t('commercial.brochure.step4') }}</li>
        </ol>
      </section>

      <section class="brochure__section">
        <h2>{{ $t('commercial.brochure.featuresTitle') }}</h2>
        <ul class="brochure__features">
          <li>{{ $t('commercial.brochure.feature1') }}</li>
          <li>{{ $t('commercial.brochure.feature2') }}</li>
          <li>{{ $t('commercial.brochure.feature3') }}</li>
          <li>{{ $t('commercial.brochure.feature4') }}</li>
          <li>{{ $t('commercial.brochure.feature5') }}</li>
          <li>{{ $t('commercial.brochure.feature6') }}</li>
        </ul>
      </section>

      <div class="brochure__page-break" aria-hidden="true" />

      <section class="brochure__section brochure__section--grid">
        <div>
          <h2>{{ $t('commercial.brochure.terrainTitle') }}</h2>
          <ul>
            <li>{{ $t('commercial.brochure.terrain1') }}</li>
            <li>{{ $t('commercial.brochure.terrain2') }}</li>
            <li>{{ $t('commercial.brochure.terrain3') }}</li>
          </ul>
        </div>
        <div>
          <h2>{{ $t('commercial.brochure.clientTitle') }}</h2>
          <ul>
            <li>{{ $t('commercial.brochure.client1') }}</li>
            <li>{{ $t('commercial.brochure.client2') }}</li>
            <li>{{ $t('commercial.brochure.client3') }}</li>
            <li>{{ $t('commercial.brochure.client4') }}</li>
          </ul>
        </div>
      </section>

      <section class="brochure__section brochure__offer">
        <h2>{{ $t('commercial.brochure.offerTitle') }}</h2>
        <div class="brochure__offer-grid">
          <div>
            <strong>{{ $t('commercial.brochure.offerProLabel') }}</strong>
            <p>{{ $t('commercial.brochure.offerProAnnual') }}</p>
            <p>{{ $t('commercial.brochure.offerProTriennial') }}</p>
            <p>{{ $t('commercial.brochure.offerSetup') }}</p>
          </div>
          <div>
            <strong>{{ $t('commercial.brochure.offerClientLabel') }}</strong>
            <p>{{ $t('commercial.brochure.offerClientPlans') }}</p>
            <p class="brochure__steer">{{ $t('commercial.brochure.offerSteer') }}</p>
          </div>
        </div>
        <p class="brochure__autofinance">{{ $t('commercial.brochure.offerAutofinance') }}</p>
      </section>

      <section class="brochure__section brochure__asv">
        <h2>{{ $t('commercial.brochure.asvTitle') }}</h2>
        <ul>
          <li>{{ $t('commercial.brochure.asv1') }}</li>
          <li>{{ $t('commercial.brochure.asv2') }}</li>
          <li>{{ $t('commercial.brochure.asv3') }}</li>
        </ul>
        <p class="brochure__asv-link">{{ $t('commercial.brochure.asvLink') }}</p>
      </section>

      <section class="brochure__section">
        <h2>{{ $t('commercial.brochure.notTitle') }}</h2>
        <p>{{ $t('commercial.brochure.notBody') }}</p>
      </section>

      <footer class="brochure__footer">
        <p class="brochure__disclaimer">{{ $t('commercial.brochure.disclaimer') }}</p>
        <p class="brochure__cta">{{ $t('commercial.brochure.cta') }}</p>
        <p v-if="contactLine" class="brochure__contact" data-testid="brochure-contact">
          {{ contactLine }}
        </p>
        <p class="brochure__brand-line">{{ $t('commercial.brochure.footerBrand') }}</p>
      </footer>
    </article>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { t } = useI18n()
const { user, fetchUser } = useProUser()
const toast = ref('')

if (!user.value) {
  await fetchUser().catch(() => null)
}

const contactLine = computed(() => {
  const name = user.value?.fullName?.trim()
  const phone = user.value?.contactPhone?.trim()
  if (name && phone) return `${name} · ${phone}`
  if (name) return name
  if (phone) return phone
  return ''
})

function printDoc() {
  window.print()
}

async function copyLink() {
  const url = `${window.location.origin}/commercial/brochure`
  try {
    await navigator.clipboard.writeText(url)
    toast.value = t('commercial.brochure.linkCopied')
  } catch {
    toast.value = url
  }
}

function mailTo() {
  const subject = encodeURIComponent(t('commercial.brochure.mailSubject'))
  const body = encodeURIComponent(t('commercial.brochure.mailBodyPdf'))
  window.location.href = `mailto:?subject=${subject}&body=${body}`
}
</script>

<style scoped>
.brochure__doc {
  max-width: 48rem;
  margin: 0 auto;
  padding: 1.25rem 0 3rem;
  line-height: 1.5;
  color: var(--pf-vet-text, #1a2b3c);
}
.brochure__hero {
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid var(--pf-vet-primary, #1a3a5c);
}
.brochure__brand {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 0.75rem;
}
.brochure__brand :deep(.logo--default) {
  margin-bottom: 0;
}
.brochure__badge {
  font-size: 0.75rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--pf-vet-muted, #5a6b7d);
  margin: 0;
}
.brochure__hero h1 {
  font-size: 1.65rem;
  line-height: 1.25;
  margin: 0 0 0.65rem;
  color: var(--pf-vet-primary, #1a3a5c);
}
.brochure__lead {
  font-size: 1.05rem;
  margin: 0 0 0.5rem;
  font-weight: 600;
}
.brochure__tagline {
  margin: 0;
  color: var(--pf-vet-muted, #5a6b7d);
  font-size: 0.95rem;
}
.brochure__section {
  margin-bottom: 1.25rem;
}
.brochure__section h2 {
  font-size: 1.05rem;
  margin: 0 0 0.4rem;
  color: var(--pf-vet-primary, #1a3a5c);
}
.brochure__section ul,
.brochure__steps {
  margin: 0.2rem 0 0;
  padding-left: 1.2rem;
}
.brochure__section li,
.brochure__steps li {
  margin-bottom: 0.28rem;
}
.brochure__actors {
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  gap: 0.5rem;
  margin-top: 0.75rem;
}
.brochure__actor {
  flex: 1 1 8rem;
  padding: 0.65rem 0.75rem;
  background: var(--pf-vet-surface-2, #f4f7fa);
  border: 1px solid var(--pf-vet-border, #d8e0ea);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  font-size: 0.9rem;
}
.brochure__actor-arrow {
  align-self: center;
  color: var(--pf-vet-accent, #0d9488);
  font-weight: 700;
}
.brochure__section--grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
}
.brochure__offer {
  padding: 0.85rem 1rem;
  background: var(--pf-vet-surface-2, #f4f7fa);
  border: 1px solid var(--pf-vet-border, #d8e0ea);
  border-radius: 8px;
}
.brochure__offer-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-top: 0.35rem;
}
.brochure__offer-grid p {
  margin: 0.25rem 0 0;
  font-size: 0.92rem;
}
.brochure__steer {
  font-weight: 600;
  color: var(--pf-vet-accent, #0d9488);
}
.brochure__autofinance {
  margin: 0.75rem 0 0;
  font-size: 0.9rem;
  color: var(--pf-vet-muted, #5a6b7d);
}
.brochure__asv {
  padding: 0.75rem 1rem;
  border-left: 3px solid var(--pf-vet-accent, #0d9488);
  background: color-mix(in srgb, var(--pf-vet-accent, #0d9488) 6%, transparent);
}
.brochure__asv-link {
  margin: 0.5rem 0 0;
  font-size: 0.88rem;
  color: var(--pf-vet-muted, #5a6b7d);
}
.brochure__footer {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--pf-vet-border, #d8e0ea);
}
.brochure__disclaimer {
  margin: 0 0 0.5rem;
  font-size: 0.9rem;
}
.brochure__cta {
  margin: 0 0 0.35rem;
  font-weight: 600;
}
.brochure__contact {
  margin: 0 0 0.35rem;
  font-size: 1rem;
  color: var(--pf-vet-primary, #1a3a5c);
}
.brochure__brand-line {
  margin: 0.5rem 0 0;
  font-size: 0.85rem;
  color: var(--pf-vet-muted, #5a6b7d);
}
.brochure__page-break {
  display: none;
}

@media (max-width: 640px) {
  .brochure__section--grid,
  .brochure__offer-grid {
    grid-template-columns: 1fr;
  }
  .brochure__actor-arrow {
    display: none;
  }
}

@media print {
  .brochure__doc {
    max-width: none;
    padding: 0;
    font-size: 10.5pt;
  }
  .brochure__hero h1 {
    font-size: 16pt;
  }
  .brochure__page-break {
    display: block;
    break-before: auto;
    page-break-before: auto;
    height: 0.75rem;
    margin: 0.5rem 0;
    border-top: 1px solid var(--pf-vet-border, #d8e0ea);
  }
  .brochure__actors {
    break-inside: avoid;
  }
  .brochure__offer,
  .brochure__asv,
  .brochure__footer {
    break-inside: avoid;
  }
}
</style>
