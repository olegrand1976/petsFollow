<template>
  <div
    v-if="visible"
    class="cal-chip__actions"
    data-testid="calendar-chip-actions"
    @click.stop
  >
    <button
      v-if="!visit.waitingRoomAt"
      type="button"
      class="cal-chip__action"
      :disabled="disabled"
      :aria-label="t('calendar.waitingRoomOn')"
      :title="t('calendar.waitingRoomOn')"
      data-testid="calendar-chip-waiting-on"
      @click="emit('action', 'mark_waiting_room')"
    >
      <ProIcon name="hourglass_top" :size="14" />
    </button>
    <button
      v-else
      type="button"
      class="cal-chip__action cal-chip__action--active"
      :disabled="disabled"
      :aria-label="t('calendar.waitingRoomOff')"
      :title="t('calendar.waitingRoomOff')"
      data-testid="calendar-chip-waiting-off"
      @click="emit('action', 'clear_waiting_room')"
    >
      <ProIcon name="hourglass_bottom" :size="14" />
    </button>
    <button
      type="button"
      class="cal-chip__action cal-chip__action--danger"
      :disabled="disabled"
      :aria-label="cancelLabel"
      :title="cancelLabel"
      data-testid="calendar-chip-cancel"
      @click="emit('action', 'cancel')"
    >
      <ProIcon name="cancel" :size="14" />
    </button>
  </div>
</template>

<script setup lang="ts">
import {
  canShowCalendarChipDeskActions,
  type CalendarDeskAction,
  type CalendarVisit,
} from '~/composables/useCalendarGrid'

const props = withDefaults(
  defineProps<{
    visit: CalendarVisit
    disabled?: boolean
  }>(),
  { disabled: false },
)

const emit = defineEmits<{ action: [CalendarDeskAction] }>()

const { t } = useI18n()
const { canPractice } = usePracticePerms()

const visible = computed(() => canShowCalendarChipDeskActions(props.visit))
const cancelLabel = computed(() =>
  canPractice('pets.write_clinical') ? t('calendar.cancel') : t('calendar.deleteVisit'),
)
</script>
