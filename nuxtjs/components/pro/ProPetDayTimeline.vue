<template>
  <ProCard :title="title" data-testid="pet-timeline-card">
    <div v-if="dayGroups.length" class="pro-pet-history">
      <section
        v-for="group in dayGroups"
        :key="group.dayKey"
        class="pro-pet-history__day"
        data-testid="pet-history-day"
      >
        <header class="pro-pet-history__day-head">
          <h3 class="pro-pet-history__day-title">{{ formatDay(group.items[0].createdAt) }}</h3>
          <ProBadge variant="neutral">
            {{ $t('clients.pet.timelineDayCount', { n: group.items.length }) }}
          </ProBadge>
        </header>
        <ul class="pro-pet-history__grid">
          <li
            v-for="item in group.items"
            :key="item.id"
            class="pro-pet-history__tile"
            :data-type="item.type || 'event'"
          >
            <div class="pro-pet-history__tile-top">
              <span class="pro-pet-history__icon" aria-hidden="true">
                <ProIcon :name="iconForType(item.type)" :size="18" />
              </span>
              <time class="pro-pet-history__time" :datetime="iso(item.createdAt)">
                {{ formatTime(item.createdAt) }}
              </time>
            </div>
            <strong class="pro-pet-history__title">{{ itemTitle(item) }}</strong>
            <p v-if="item.body" class="pro-pet-history__body">{{ item.body }}</p>
          </li>
        </ul>
      </section>
    </div>
    <ProEmptyState
      v-else
      :title="$t('clients.pet.timelineEmptyTitle')"
      :description="$t('clients.pet.timelineEmptyDescription')"
    />
  </ProCard>
</template>

<script setup lang="ts">
import { groupTimelineByDay, type TimelineDayItem } from '~/utils/groupTimelineByDay'

const props = withDefaults(
  defineProps<{
    items: TimelineDayItem[]
    title?: string
  }>(),
  {
    title: undefined,
  },
)

const { t } = useI18n()
const { formatDay, formatTime } = useFormatters()

const title = computed(() => props.title ?? t('clients.pet.timelineTitle'))

const dayGroups = computed(() => groupTimelineByDay(props.items ?? []))

function iso(value: string | Date) {
  const d = value instanceof Date ? value : new Date(value)
  return Number.isNaN(d.getTime()) ? '' : d.toISOString()
}

function iconForType(type?: string) {
  switch (type) {
    case 'heartrate':
      return 'favorite'
    case 'weight':
      return 'monitor_weight'
    case 'message':
      return 'chat'
    case 'care':
      return 'medical_services'
    case 'visit':
      return 'event'
    case 'event':
    case undefined:
    case '':
      return 'timeline'
    default:
      return 'timeline'
  }
}

function itemTitle(item: TimelineDayItem) {
  if (item.title?.trim()) return item.title
  const type = item.type ?? ''
  const keyByType: Record<string, string> = {
    heartrate: 'clients.pet.timelineTypeHeartrate',
    weight: 'clients.pet.timelineTypeWeight',
    message: 'clients.pet.timelineTypeMessage',
    care: 'clients.pet.timelineTypeCare',
    visit: 'clients.pet.timelineTypeVisit',
    event: 'clients.pet.timelineTypeEvent',
  }
  const key = keyByType[type]
  if (key) return t(key)
  return type || t('clients.pet.timelineTypeEvent')
}
</script>

<style scoped>
.pro-pet-history {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.pro-pet-history__day-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--pf-vet-border);
}

.pro-pet-history__day-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 650;
  color: var(--pf-vet-primary);
  text-transform: capitalize;
}

.pro-pet-history__grid {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr));
  gap: 0.75rem;
}

.pro-pet-history__tile {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 0.85rem 0.9rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-bg, #f8fafc);
  border-left: 3px solid var(--pf-vet-accent);
}

.pro-pet-history__tile[data-type='heartrate'] {
  border-left-color: var(--pf-vet-alert);
}

.pro-pet-history__tile[data-type='weight'] {
  border-left-color: var(--pf-vet-primary);
}

.pro-pet-history__tile[data-type='visit'],
.pro-pet-history__tile[data-type='care'] {
  border-left-color: var(--pf-vet-accent);
}

.pro-pet-history__tile[data-type='message'] {
  border-left-color: color-mix(in srgb, var(--pf-vet-primary) 65%, var(--pf-vet-accent));
}

.pro-pet-history__tile-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.pro-pet-history__icon {
  display: inline-flex;
  color: var(--pf-vet-primary);
}

.pro-pet-history__time {
  font-size: 0.85rem;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
  color: var(--pf-vet-primary);
}

.pro-pet-history__title {
  font-size: 0.9rem;
  line-height: 1.3;
}

.pro-pet-history__body {
  margin: 0;
  font-size: 0.82rem;
  color: var(--pf-vet-muted, #64748b);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
