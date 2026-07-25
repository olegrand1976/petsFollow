<template>
  <div class="ai-playbook" data-testid="commercial-ai-cr-playbook">
    <div class="ai-playbook__toolbar no-print">
      <ProPageHeader
        :title="$t('commercial.aiPlaybook.title')"
        :subtitle="$t('commercial.aiPlaybook.subtitle')"
      >
        <template #actions>
          <ProButton variant="secondary" @click="copyLink">{{ $t('commercial.aiPlaybook.copyLink') }}</ProButton>
          <ProButton variant="secondary" @click="mailTo">{{ $t('commercial.aiPlaybook.mailto') }}</ProButton>
          <ProButton @click="printDoc">{{ $t('commercial.aiPlaybook.print') }}</ProButton>
        </template>
      </ProPageHeader>
      <p v-if="toast" class="pro-hint">{{ toast }}</p>
    </div>

    <article class="ai-playbook__doc">
      <header class="ai-playbook__hero">
        <p class="ai-playbook__badge">{{ $t('commercial.aiPlaybook.badge') }}</p>
        <h1>{{ $t('commercial.aiPlaybook.docTitle') }}</h1>
        <p class="ai-playbook__lead">{{ $t('commercial.aiPlaybook.promise') }}</p>
        <p class="ai-playbook__price">
          <strong>{{ $t('commercial.aiPlaybook.priceMonthly') }}</strong>
          · {{ $t('commercial.aiPlaybook.priceAnnual') }}
        </p>
      </header>

      <section>
        <h2>{{ $t('commercial.aiPlaybook.whoTitle') }}</h2>
        <p>{{ $t('commercial.aiPlaybook.whoBody') }}</p>
      </section>

      <section>
        <h2>{{ $t('commercial.aiPlaybook.changeTitle') }}</h2>
        <ul>
          <li>{{ $t('commercial.aiPlaybook.change1') }}</li>
          <li>{{ $t('commercial.aiPlaybook.change2') }}</li>
          <li>{{ $t('commercial.aiPlaybook.change3') }}</li>
        </ul>
      </section>

      <section>
        <h2>{{ $t('commercial.aiPlaybook.howtoTitle') }}</h2>
        <ol class="ai-playbook__steps">
          <li>{{ $t('commercial.aiPlaybook.step1') }}</li>
          <li>{{ $t('commercial.aiPlaybook.step2') }}</li>
          <li>{{ $t('commercial.aiPlaybook.step3') }}</li>
          <li>{{ $t('commercial.aiPlaybook.step4') }}</li>
          <li>{{ $t('commercial.aiPlaybook.step5') }}</li>
          <li>{{ $t('commercial.aiPlaybook.step6') }}</li>
        </ol>
      </section>

      <section>
        <h2>{{ $t('commercial.aiPlaybook.pitchTitle') }}</h2>
        <blockquote>{{ $t('commercial.aiPlaybook.pitchBody') }}</blockquote>
        <h3>{{ $t('commercial.aiPlaybook.objectionsTitle') }}</h3>
        <dl class="ai-playbook__objections">
          <div>
            <dt>{{ $t('commercial.aiPlaybook.obj1q') }}</dt>
            <dd>{{ $t('commercial.aiPlaybook.obj1a') }}</dd>
          </div>
          <div>
            <dt>{{ $t('commercial.aiPlaybook.obj2q') }}</dt>
            <dd>{{ $t('commercial.aiPlaybook.obj2a') }}</dd>
          </div>
          <div>
            <dt>{{ $t('commercial.aiPlaybook.obj3q') }}</dt>
            <dd>{{ $t('commercial.aiPlaybook.obj3a') }}</dd>
          </div>
        </dl>
      </section>

      <section>
        <h2>{{ $t('commercial.aiPlaybook.roiTitle') }}</h2>
        <p>{{ $t('commercial.aiPlaybook.roiBody') }}</p>
      </section>

      <section>
        <h2>{{ $t('commercial.aiPlaybook.calendarTitle') }}</h2>
        <ul>
          <li>{{ $t('commercial.aiPlaybook.calJ0') }}</li>
          <li>{{ $t('commercial.aiPlaybook.calJ14') }}</li>
          <li>{{ $t('commercial.aiPlaybook.calJ45') }}</li>
          <li>{{ $t('commercial.aiPlaybook.calJ60') }}</li>
          <li>{{ $t('commercial.aiPlaybook.calJ75') }}</li>
          <li>{{ $t('commercial.aiPlaybook.calJ90') }}</li>
        </ul>
      </section>

      <section>
        <h2>{{ $t('commercial.aiPlaybook.frictionTitle') }}</h2>
        <p>{{ $t('commercial.aiPlaybook.frictionBody') }}</p>
      </section>

      <footer class="ai-playbook__footer">
        <p>{{ $t('commercial.aiPlaybook.footer') }}</p>
      </footer>
    </article>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { t } = useI18n()
const toast = ref('')

function printDoc() {
  window.print()
}

async function copyLink() {
  const url = `${window.location.origin}/commercial/ai-cr-playbook`
  try {
    await navigator.clipboard.writeText(url)
    toast.value = t('commercial.aiPlaybook.linkCopied')
  } catch {
    toast.value = url
  }
}

function mailTo() {
  const subject = encodeURIComponent(t('commercial.aiPlaybook.mailSubject'))
  const body = encodeURIComponent(t('commercial.aiPlaybook.mailBodyPdf'))
  window.location.href = `mailto:?subject=${subject}&body=${body}`
}
</script>

<style scoped>
.ai-playbook__doc {
  max-width: 42rem;
  margin: 0 auto;
  padding: 1.5rem 0 3rem;
  line-height: 1.55;
}
.ai-playbook__hero {
  margin-bottom: 2rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--pf-vet-border, #d8e0ea);
}
.ai-playbook__badge {
  font-size: 0.8rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--pf-vet-muted, #5a6b7d);
  margin: 0 0 0.5rem;
}
.ai-playbook__hero h1 {
  font-size: 1.75rem;
  margin: 0 0 0.75rem;
}
.ai-playbook__lead {
  font-size: 1.05rem;
  margin: 0 0 0.75rem;
}
.ai-playbook__price {
  margin: 0;
  font-size: 1rem;
}
.ai-playbook__doc section {
  margin-bottom: 1.75rem;
}
.ai-playbook__doc h2 {
  font-size: 1.15rem;
  margin: 0 0 0.5rem;
}
.ai-playbook__doc h3 {
  font-size: 1rem;
  margin: 1rem 0 0.4rem;
}
.ai-playbook__steps,
.ai-playbook__doc ul {
  margin: 0.25rem 0 0;
  padding-left: 1.25rem;
}
.ai-playbook__steps li,
.ai-playbook__doc li {
  margin-bottom: 0.35rem;
}
.ai-playbook__doc blockquote {
  margin: 0.5rem 0 1rem;
  padding: 0.75rem 1rem;
  border-left: 3px solid var(--pf-vet-primary, #1f6b5a);
  background: var(--pf-vet-surface-2, #f4f7fa);
}
.ai-playbook__objections dt {
  font-weight: 600;
  margin-top: 0.65rem;
}
.ai-playbook__objections dd {
  margin: 0.2rem 0 0;
}
.ai-playbook__footer {
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid var(--pf-vet-border, #d8e0ea);
  font-size: 0.9rem;
  color: var(--pf-vet-muted, #5a6b7d);
}

@media print {
  .no-print {
    display: none !important;
  }
  .ai-playbook__doc {
    max-width: none;
    padding: 0;
  }
}
</style>

<style>
/* Unscoped: hide commercial shell when printing the playbook for email/PDF. */
@media print {
  .pro-topbar,
  .pro-sidebar,
  .pro-app-shell > .pro-sidebar,
  aside.pro-sidebar {
    display: none !important;
  }
  .pro-app-shell {
    display: block !important;
  }
  .pro-app-body,
  .pro-main,
  .pro-main-inner {
    margin: 0 !important;
    padding: 0 !important;
    max-width: none !important;
  }
}
</style>
