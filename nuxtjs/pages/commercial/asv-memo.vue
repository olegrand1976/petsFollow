<template>
  <div class="asv-memo" data-testid="commercial-asv-memo">
    <div class="asv-memo__toolbar no-print">
      <ProPageHeader
        :title="$t('commercial.asvMemo.title')"
        :subtitle="$t('commercial.asvMemo.subtitle')"
        title-tag="p"
      >
        <template #actions>
          <ProButton variant="secondary" @click="copyLink">{{ $t('commercial.asvMemo.copyLink') }}</ProButton>
          <ProButton variant="secondary" @click="mailTo">{{ $t('commercial.asvMemo.mailto') }}</ProButton>
          <ProButton test-id="asv-memo-print" @click="printDoc">{{ $t('commercial.asvMemo.print') }}</ProButton>
        </template>
      </ProPageHeader>
      <p v-if="toast" class="pro-hint">{{ toast }}</p>
    </div>

    <article class="asv-memo__doc">
      <header class="asv-memo__hero">
        <p class="asv-memo__badge">{{ $t('commercial.asvMemo.badge') }}</p>
        <h1>{{ $t('commercial.asvMemo.docTitle') }}</h1>
        <p class="asv-memo__lead">{{ $t('commercial.asvMemo.promise') }}</p>
      </header>

      <section>
        <h2>{{ $t('commercial.asvMemo.painTitle') }}</h2>
        <ul>
          <li>{{ $t('commercial.asvMemo.pain1') }}</li>
          <li>{{ $t('commercial.asvMemo.pain2') }}</li>
          <li>{{ $t('commercial.asvMemo.pain3') }}</li>
          <li>{{ $t('commercial.asvMemo.pain4') }}</li>
        </ul>
      </section>

      <section>
        <h2>{{ $t('commercial.asvMemo.gainTitle') }}</h2>
        <ul>
          <li>{{ $t('commercial.asvMemo.gain1') }}</li>
          <li>{{ $t('commercial.asvMemo.gain2') }}</li>
          <li>{{ $t('commercial.asvMemo.gain3') }}</li>
          <li>{{ $t('commercial.asvMemo.gain4') }}</li>
          <li>{{ $t('commercial.asvMemo.gain5') }}</li>
          <li>{{ $t('commercial.asvMemo.gain6') }}</li>
        </ul>
      </section>

      <section>
        <h2>{{ $t('commercial.asvMemo.notTitle') }}</h2>
        <p>{{ $t('commercial.asvMemo.notBody') }}</p>
      </section>

      <section>
        <h2>{{ $t('commercial.asvMemo.demoTitle') }}</h2>
        <ol class="asv-memo__steps">
          <li>{{ $t('commercial.asvMemo.demo1') }}</li>
          <li>{{ $t('commercial.asvMemo.demo2') }}</li>
          <li>{{ $t('commercial.asvMemo.demo3') }}</li>
        </ol>
      </section>

      <section>
        <h2>{{ $t('commercial.asvMemo.objectionsTitle') }}</h2>
        <dl class="asv-memo__objections">
          <div>
            <dt>{{ $t('commercial.asvMemo.obj1q') }}</dt>
            <dd>{{ $t('commercial.asvMemo.obj1a') }}</dd>
          </div>
          <div>
            <dt>{{ $t('commercial.asvMemo.obj2q') }}</dt>
            <dd>{{ $t('commercial.asvMemo.obj2a') }}</dd>
          </div>
          <div>
            <dt>{{ $t('commercial.asvMemo.obj3q') }}</dt>
            <dd>{{ $t('commercial.asvMemo.obj3a') }}</dd>
          </div>
          <div>
            <dt>{{ $t('commercial.asvMemo.obj4q') }}</dt>
            <dd>{{ $t('commercial.asvMemo.obj4a') }}</dd>
          </div>
        </dl>
      </section>

      <section>
        <h2>{{ $t('commercial.asvMemo.takeawayTitle') }}</h2>
        <blockquote>{{ $t('commercial.asvMemo.takeawayBody') }}</blockquote>
      </section>

      <footer class="asv-memo__footer">
        <p>{{ $t('commercial.asvMemo.footer') }}</p>
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
  const url = `${window.location.origin}/commercial/asv-memo`
  try {
    await navigator.clipboard.writeText(url)
    toast.value = t('commercial.asvMemo.linkCopied')
  } catch {
    toast.value = url
  }
}

function mailTo() {
  const subject = encodeURIComponent(t('commercial.asvMemo.mailSubject'))
  const body = encodeURIComponent(t('commercial.asvMemo.mailBodyPdf'))
  window.location.href = `mailto:?subject=${subject}&body=${body}`
}
</script>

<style scoped>
.asv-memo__doc {
  max-width: 42rem;
  margin: 0 auto;
  padding: 1.5rem 0 3rem;
  line-height: 1.55;
}
.asv-memo__hero {
  margin-bottom: 2rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--pf-vet-border, #d8e0ea);
}
.asv-memo__badge {
  font-size: 0.8rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--pf-vet-muted, #5a6b7d);
  margin: 0 0 0.5rem;
}
.asv-memo__hero h1 {
  font-size: 1.75rem;
  margin: 0 0 0.75rem;
}
.asv-memo__lead {
  font-size: 1.05rem;
  margin: 0;
}
.asv-memo__doc section {
  margin-bottom: 1.75rem;
}
.asv-memo__doc h2 {
  font-size: 1.15rem;
  margin: 0 0 0.5rem;
}
.asv-memo__steps,
.asv-memo__doc ul {
  margin: 0.25rem 0 0;
  padding-left: 1.25rem;
}
.asv-memo__steps li,
.asv-memo__doc li {
  margin-bottom: 0.35rem;
}
.asv-memo__doc blockquote {
  margin: 0.5rem 0 1rem;
  padding: 0.75rem 1rem;
  border-left: 3px solid var(--pf-vet-primary, #1f6b5a);
  background: var(--pf-vet-surface-2, #f4f7fa);
}
.asv-memo__objections dt {
  font-weight: 600;
  margin-top: 0.65rem;
}
.asv-memo__objections dd {
  margin: 0.2rem 0 0;
}
.asv-memo__footer {
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid var(--pf-vet-border, #d8e0ea);
  font-size: 0.9rem;
  color: var(--pf-vet-muted, #5a6b7d);
}

@media print {
  .asv-memo__doc {
    max-width: none;
    padding: 0;
  }
}
</style>
