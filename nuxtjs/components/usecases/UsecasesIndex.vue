<template>
  <div class="usecases" data-testid="usecases-index">
    <ProPageHeader :title="$t('usecases.title')" :subtitle="$t('usecases.subtitle')">
      <template #actions>
        <ProBadge variant="warning">{{ $t('usecases.stagingOnly') }}</ProBadge>
      </template>
    </ProPageHeader>

    <ProCard class="usecases__howto">
      <h2 class="usecases__h2">{{ $t('usecases.howtoTitle') }}</h2>
      <ol class="usecases__howto-list">
        <li>{{ $t('usecases.howto1') }}</li>
        <li>{{ $t('usecases.howto2') }}</li>
        <li>{{ $t('usecases.howto3') }}</li>
        <li>{{ $t('usecases.howto4') }}</li>
      </ol>
      <p class="usecases__hint">{{ $t('usecases.howtoHint') }}</p>
    </ProCard>

    <ProCard class="usecases__demo">
      <h2 class="usecases__h2">{{ $t('usecases.demoSession') }}</h2>
      <p class="usecases__blurb">{{ $t('usecases.demoSessionBlurb') }}</p>
      <ol class="usecases__demo-list">
        <li v-for="id in demoSession" :key="id">
          <NuxtLink :to="`${basePath}/${id}`" class="usecases__link">
            {{ getCase(id)?.title || id }}
            <span class="usecases__id-soft">{{ id }}</span>
          </NuxtLink>
        </li>
      </ol>
    </ProCard>

    <div v-for="folder in folders" :key="folder.id" class="usecases__folder">
      <h2 class="usecases__h2">{{ folder.label }}</h2>
      <p v-if="folder.blurb" class="usecases__blurb">{{ folder.blurb }}</p>
      <ul class="usecases__list">
        <li v-for="id in folder.items" :key="id">
          <NuxtLink :to="`${basePath}/${id}`" class="usecases__card-link" :data-testid="`usecase-link-${id}`">
            <div class="usecases__card-main">
              <span class="usecases__card-title">{{ getCase(id)?.title }}</span>
              <span class="usecases__card-obj">{{ shortObjective(id) }}</span>
            </div>
            <span class="usecases__meta">
              <ProBadge v-if="getCase(id)?.destructive" variant="danger">{{ $t('usecases.destructiveShort') }}</ProBadge>
              <span v-if="getCase(id)?.duration">{{ getCase(id)?.duration }}</span>
              <span class="usecases__id-soft">{{ id }}</span>
            </span>
          </NuxtLink>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ basePath: string }>()

const { folders, demoSession, getCase } = useUsecasesCatalog()
const basePath = computed(() => props.basePath.replace(/\/$/, ''))

function shortObjective(id: string) {
  const o = getCase(id)?.objective || ''
  if (o.length <= 110) return o
  return `${o.slice(0, 107).trim()}…`
}
</script>

<style scoped>
.usecases__howto,
.usecases__demo {
  margin-bottom: 1.25rem;
}
.usecases__h2 {
  margin: 0 0 0.5rem;
  font-size: 1.125rem;
  color: var(--pf-vet-primary);
}
.usecases__blurb {
  margin: 0 0 0.75rem;
  font-size: 0.9rem;
  color: var(--pf-vet-muted, #64748b);
  line-height: 1.45;
}
.usecases__howto-list,
.usecases__demo-list {
  margin: 0;
  padding-left: 1.25rem;
  line-height: 1.55;
}
.usecases__hint {
  margin: 0.75rem 0 0;
  font-size: 0.875rem;
  color: var(--pf-vet-muted, #64748b);
}
.usecases__link {
  color: var(--pf-vet-accent);
  text-decoration: none;
  display: inline-flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: baseline;
}
.usecases__link:hover {
  text-decoration: underline;
}
.usecases__folder {
  margin-bottom: 1.75rem;
}
.usecases__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.5rem;
}
.usecases__card-link {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
  padding: 0.85rem 1rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-md, 8px);
  background: var(--pf-vet-surface);
  text-decoration: none;
  color: inherit;
}
.usecases__card-link:hover {
  box-shadow: var(--pf-vet-shadow-sm);
  border-color: var(--pf-vet-accent);
}
.usecases__card-main {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}
.usecases__card-title {
  font-size: 1rem;
  font-weight: 650;
  color: var(--pf-vet-primary);
}
.usecases__card-obj {
  font-size: 0.875rem;
  color: var(--pf-vet-muted, #64748b);
  line-height: 1.4;
}
.usecases__meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.35rem;
  font-size: 0.75rem;
  color: var(--pf-vet-muted, #64748b);
  flex-shrink: 0;
}
.usecases__id-soft {
  font-size: 0.7rem;
  opacity: 0.7;
  font-variant-numeric: tabular-nums;
}
@media (max-width: 720px) {
  .usecases__card-link {
    flex-direction: column;
  }
  .usecases__meta {
    align-items: flex-start;
    flex-direction: row;
    flex-wrap: wrap;
  }
}
</style>
