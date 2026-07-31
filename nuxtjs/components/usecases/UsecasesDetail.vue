<template>
  <div v-if="item" class="usecase-detail" data-testid="usecase-detail">
    <p class="usecase-detail__back">
      <NuxtLink :to="basePath">← {{ $t('usecases.back') }}</NuxtLink>
    </p>
    <ProPageHeader :title="item.title" :subtitle="item.objective">
      <template #actions>
        <ProBadge v-if="item.destructive" variant="danger">{{ $t('usecases.destructiveShort') }}</ProBadge>
        <ProBadge variant="neutral">{{ priorityLabel }}</ProBadge>
      </template>
    </ProPageHeader>

    <p class="usecase-detail__id">{{ $t('usecases.refId', { id: item.id }) }}</p>

    <div class="usecase-detail__meta">
      <span v-if="item.duration">{{ $t('usecases.durationLabel', { duration: item.duration }) }}</span>
      <span v-if="friendlySurface">{{ friendlySurface }}</span>
    </div>

    <ProCard v-if="item.destructive" class="usecase-detail__warn" data-testid="usecase-destructive-warn">
      <p><strong>{{ $t('usecases.destructive') }}</strong></p>
      <p>{{ $t('usecases.destructiveHint') }}</p>
    </ProCard>

    <ProCard v-if="item.actors.length" class="usecase-detail__block">
      <h2>{{ $t('usecases.actors') }}</h2>
      <p class="usecase-detail__help">{{ $t('usecases.actorsHelp') }}</p>
      <div v-for="(a, i) in item.actors" :key="i" class="usecase-detail__actor">
        <p class="usecase-detail__actor-role">{{ a.role }}</p>
        <dl class="usecase-detail__creds">
          <div>
            <dt>{{ $t('usecases.colAccount') }}</dt>
            <dd>
              <code>{{ a.account }}</code>
              <button type="button" class="usecase-detail__copy" @click="copy(a.account)">
                {{ $t('usecases.copy') }}
              </button>
            </dd>
          </div>
          <div>
            <dt>{{ $t('usecases.colPassword') }}</dt>
            <dd>
              <code>{{ a.password }}</code>
              <button type="button" class="usecase-detail__copy" @click="copy(a.password)">
                {{ $t('usecases.copy') }}
              </button>
            </dd>
          </div>
        </dl>
      </div>
      <p v-if="copied" class="usecase-detail__copied" role="status">{{ $t('usecases.copied') }}</p>
    </ProCard>

    <ProCard v-if="item.prerequisites.length" class="usecase-detail__block">
      <h2>{{ $t('usecases.prerequisites') }}</h2>
      <ul>
        <li v-for="(p, i) in item.prerequisites" :key="i">{{ p }}</li>
      </ul>
    </ProCard>

    <ProCard v-if="item.steps.length" class="usecase-detail__block usecase-detail__block--steps">
      <h2>{{ $t('usecases.steps') }}</h2>
      <p class="usecase-detail__help">{{ $t('usecases.stepsHelp') }}</p>
      <ol class="usecase-detail__steps">
        <li v-for="(s, i) in item.steps" :key="i">{{ s }}</li>
      </ol>
    </ProCard>

    <ProCard v-if="item.expected.length" class="usecase-detail__block">
      <h2>{{ $t('usecases.expected') }}</h2>
      <p class="usecase-detail__help">{{ $t('usecases.expectedHelp') }}</p>
      <ul>
        <li v-for="(e, i) in item.expected" :key="i">{{ e }}</li>
      </ul>
    </ProCard>

    <ProCard v-if="item.checklist.length" class="usecase-detail__block">
      <h2>{{ $t('usecases.checklist') }}</h2>
      <p class="usecase-detail__help">{{ $t('usecases.checklistHelp') }}</p>
      <ul class="usecase-detail__check">
        <li v-for="(c, i) in item.checklist" :key="i">
          <span class="usecase-detail__check-label">{{ c }}</span>
          <span class="usecase-detail__check-opts">
            {{ $t('usecases.checkOk') }} · {{ $t('usecases.checkKo') }} · {{ $t('usecases.checkNa') }}
          </span>
        </li>
      </ul>
    </ProCard>

    <ProCard class="usecase-detail__block">
      <h2>{{ $t('usecases.feedbackTitle') }}</h2>
      <p>{{ $t('usecases.feedbackBody') }}</p>
    </ProCard>
  </div>
  <ProEmptyState
    v-else
    :title="$t('usecases.notFound')"
    :description="$t('usecases.notFoundHint')"
  />
</template>

<script setup lang="ts">
import { usecaseFriendlyPriority, usecaseFriendlySurface } from '~/utils/usecasesPlain'

const props = defineProps<{ basePath: string; caseId: string }>()

const { t } = useI18n()
const { getCase } = useUsecasesCatalog()
const basePath = computed(() => props.basePath.replace(/\/$/, ''))
const item = computed(() => getCase(props.caseId))
const friendlySurface = computed(() => usecaseFriendlySurface(item.value?.surface))
const priorityLabel = computed(() => {
  const key = usecaseFriendlyPriority(item.value?.priority)
  if (key === 'demo') return t('usecases.priorityDemo')
  if (key === 'important') return t('usecases.priorityImportant')
  if (key === 'secondary') return t('usecases.prioritySecondary')
  return item.value?.priority || ''
})

const copied = ref(false)
let copyTimer: ReturnType<typeof setTimeout> | undefined

async function copy(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    copied.value = true
    clearTimeout(copyTimer)
    copyTimer = setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    /* ignore */
  }
}
</script>

<style scoped>
.usecase-detail__back {
  margin: 0 0 0.75rem;
}
.usecase-detail__back a {
  color: var(--pf-vet-accent);
  text-decoration: none;
}
.usecase-detail__id {
  margin: -0.35rem 0 0.75rem;
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
}
.usecase-detail__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin: 0 0 1.25rem;
  font-size: 0.9rem;
  color: var(--pf-vet-muted, #64748b);
}
.usecase-detail__warn {
  margin-bottom: 1rem;
  border-color: var(--pf-vet-alert);
}
.usecase-detail__warn p {
  margin: 0;
}
.usecase-detail__warn p + p {
  margin-top: 0.35rem;
}
.usecase-detail__block {
  margin-bottom: 1rem;
}
.usecase-detail__block h2 {
  margin: 0 0 0.35rem;
  font-size: 1.05rem;
  color: var(--pf-vet-primary);
}
.usecase-detail__help {
  margin: 0 0 0.65rem;
  font-size: 0.875rem;
  color: var(--pf-vet-muted, #64748b);
}
.usecase-detail__block p,
.usecase-detail__block ul,
.usecase-detail__block ol {
  margin: 0;
}
.usecase-detail__block--steps {
  background: color-mix(in srgb, var(--pf-vet-accent) 6%, var(--pf-vet-surface));
}
.usecase-detail__steps {
  padding-left: 1.35rem;
  font-size: 1rem;
  line-height: 1.55;
}
.usecase-detail__steps li + li {
  margin-top: 0.55rem;
}
.usecase-detail__actor {
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.usecase-detail__actor:last-of-type {
  border-bottom: none;
}
.usecase-detail__actor-role {
  margin: 0 0 0.35rem;
  font-weight: 650;
}
.usecase-detail__creds {
  margin: 0;
  display: grid;
  gap: 0.35rem;
}
.usecase-detail__creds div {
  display: grid;
  grid-template-columns: 7.5rem 1fr;
  gap: 0.5rem;
  align-items: center;
}
.usecase-detail__creds dt {
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
}
.usecase-detail__creds dd {
  margin: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.usecase-detail__creds code {
  font-size: 0.875rem;
  word-break: break-all;
}
.usecase-detail__copy {
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-bg, #fff);
  border-radius: 4px;
  padding: 0.15rem 0.45rem;
  font-size: 0.75rem;
  cursor: pointer;
}
.usecase-detail__copied {
  margin: 0.5rem 0 0;
  font-size: 0.8rem;
  color: var(--pf-vet-accent);
}
.usecase-detail__check {
  list-style: none;
  padding: 0;
}
.usecase-detail__check li {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding: 0.55rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.usecase-detail__check-label {
  font-weight: 550;
}
.usecase-detail__check-opts {
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
}
@media (max-width: 640px) {
  .usecase-detail__creds div {
    grid-template-columns: 1fr;
  }
}
</style>
