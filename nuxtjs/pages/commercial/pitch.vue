<template>
  <div data-testid="commercial-pitch-page">
    <ProPageHeader
      :title="$t('commercial.pitch.title')"
      :subtitle="$t('commercial.pitch.subtitle')"
    />

    <div class="pro-grid-2">
      <ProCard :title="$t('commercial.pitch.webTitle')">
        <ul class="pro-feature-list">
          <li v-for="item in webFeatures" :key="item">{{ item }}</li>
        </ul>
      </ProCard>
      <ProCard :title="$t('commercial.pitch.mobileTitle')">
        <ul class="pro-feature-list">
          <li v-for="item in mobileFeatures" :key="item">{{ item }}</li>
        </ul>
      </ProCard>
    </div>

    <ProCard class="pro-mt-lg" :title="$t('commercial.pitch.commissionTitle')">
      <p class="pro-mb-md">{{ $t('commercial.pitch.vetGrid') }}</p>
      <p class="pro-mb-md">{{ $t('commercial.pitch.commercialFlat') }}</p>
      <p>{{ $t('commercial.pitch.addons') }}</p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { tm, rt } = useI18n()

const webFeatures = computed(() => {
  const raw = tm('commercial.pitch.webFeatures') as string[]
  return Array.isArray(raw) ? raw.map((x) => (typeof x === 'string' ? x : rt(x as any))) : []
})
const mobileFeatures = computed(() => {
  const raw = tm('commercial.pitch.mobileFeatures') as string[]
  return Array.isArray(raw) ? raw.map((x) => (typeof x === 'string' ? x : rt(x as any))) : []
})
</script>

<style scoped>
.pro-feature-list {
  margin: 0;
  padding-left: 1.25rem;
  display: grid;
  gap: 0.5rem;
}
</style>
