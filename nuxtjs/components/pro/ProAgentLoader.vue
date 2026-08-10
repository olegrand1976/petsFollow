<template>
  <div
    class="pro-agent-loader"
    data-testid="visit-report-advanced-improving"
    role="status"
    aria-busy="true"
  >
    <ProIcon name="auto_awesome" :size="24" class="pro-agent-loader__icon" />
    <div class="pro-agent-loader__body">
      <strong>{{ title }}</strong>
      <ol v-if="steps.length" class="pro-agent-loader__steps" data-testid="visit-report-advanced-steps">
        <li
          v-for="(s, i) in steps"
          :key="`${s.agent}-${i}`"
          class="pro-agent-loader__step"
          :data-agent="s.agent"
        >
          <span class="pro-agent-loader__agent">{{ s.agent }}</span>
          — {{ s.label }}
        </li>
      </ol>
      <p v-else class="pro-hint">{{ waitingLabel }}</p>
    </div>
    <ProButton
      v-if="cancelLabel"
      variant="ghost"
      test-id="visit-report-advanced-cancel"
      @click="emit('cancel')"
    >
      {{ cancelLabel }}
    </ProButton>
  </div>
</template>

<script setup lang="ts">
export type AgentStep = { agent: string; label: string; at?: string }

defineProps<{
  title: string
  waitingLabel: string
  steps: AgentStep[]
  cancelLabel?: string
}>()

const emit = defineEmits<{ cancel: [] }>()
</script>

<style scoped>
.pro-agent-loader {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  padding: 0.75rem 1rem;
  border-radius: var(--pf-radius-md, 8px);
  background: var(--pf-vet-surface-2, #f4f7f6);
  border: 1px solid var(--pf-vet-border, #d7e0dd);
}
.pro-agent-loader__icon {
  flex-shrink: 0;
  color: var(--pf-vet-primary, #0d7377);
}
.pro-agent-loader__body {
  flex: 1;
  min-width: 0;
}
.pro-agent-loader__steps {
  margin: 0.35rem 0 0;
  padding-left: 1.1rem;
  font-size: 0.875rem;
}
.pro-agent-loader__agent {
  font-weight: 600;
  font-family: var(--pf-font-mono, monospace);
}
</style>
