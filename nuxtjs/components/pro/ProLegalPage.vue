<template>
  <div class="pro-legal">
    <ProPageHeader :title="title" />
    <ProCard>
      <p class="pro-legal__body">{{ body }}</p>
      <p class="pro-legal__updated text-muted">{{ $t('legal.updated') }}</p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  doc: 'terms' | 'privacy' | 'mentions'
}>()

const { t } = useI18n()

const title = computed(() => t(`legal.${props.doc}.title`))
const body = computed(() => t(`legal.${props.doc}.body`))
</script>

<style scoped>
.pro-legal {
  max-width: 48rem;
  margin: 0 auto;
  padding: 2rem 1.5rem 3rem;
}

.pro-legal__body {
  white-space: pre-line;
  line-height: 1.6;
}

.pro-legal__updated {
  margin-top: 1.5rem;
  font-size: 0.9rem;
}
</style>
