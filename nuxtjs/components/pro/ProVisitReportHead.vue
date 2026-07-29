<template>
  <div class="pro-visit-report__head">
    <div class="pro-visit-report__meta">
      <p v-if="visitDateLabel" class="pro-hint" data-testid="visit-report-date">
        {{ $t('calendar.reportVisitDate') }} : {{ visitDateLabel }}
      </p>
      <p class="pro-hint" data-testid="visit-report-flow-hint">
        {{ $t('calendar.reportFlowHint') }}
      </p>
    </div>
    <div class="pro-visit-report__head-actions">
      <ProBadge
        v-if="reportStatus === 'final'"
        variant="success"
        data-testid="visit-report-status-final"
      >
        {{ $t('calendar.reportStatusFinal') }}
      </ProBadge>
      <details class="visit-report-howto" data-testid="visit-report-howto">
        <summary>{{ $t('calendar.reportHowItWorksTitle') }}</summary>
        <ol>
          <li>{{ $t('calendar.reportHowItWorksStep1') }}</li>
          <li>{{ $t('calendar.reportHowItWorksStep2') }}</li>
          <li>{{ $t('calendar.reportHowItWorksStep3') }}</li>
        </ol>
      </details>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  visitDateLabel?: string
  reportStatus?: string
}>()
</script>

<style scoped>
.pro-visit-report__head {
  flex-shrink: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.pro-visit-report__meta {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
}

.pro-visit-report__meta .pro-hint {
  margin: 0;
}

.pro-visit-report__head-actions {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  flex-shrink: 0;
}

.visit-report-howto {
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  padding: 0.35rem 0.65rem;
  background: var(--pf-vet-bg, #f8fafc);
  max-width: min(22rem, 100%);
}

.visit-report-howto > summary {
  cursor: pointer;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--pf-vet-primary);
}

.visit-report-howto ol {
  margin: 0.5rem 0 0.25rem;
  padding-left: 1.25rem;
  font-size: 0.85rem;
  line-height: 1.5;
}
</style>
