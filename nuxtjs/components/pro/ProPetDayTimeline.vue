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
          <h3 class="pro-pet-history__day-title">{{ formatDay(group.items[0].item.createdAt) }}</h3>
          <ProBadge variant="neutral">
            {{ $t('clients.pet.timelineDayCount', { n: group.items.length }) }}
          </ProBadge>
        </header>
        <ul class="pro-pet-history__grid">
          <li
            v-for="tile in group.items"
            :key="tile.item.id"
            class="pro-pet-history__tile"
            :class="{ 'pro-pet-history__tile--clickable': tile.clickable }"
            :data-type="tile.item.type || 'event'"
            :data-testid="tile.clickable ? 'pet-history-visit-report' : undefined"
            :role="tile.clickable ? 'button' : undefined"
            :tabindex="tile.clickable ? 0 : undefined"
            :aria-label="tile.clickable ? tile.ariaLabel : undefined"
            v-on="tile.clickable ? tile.listeners : {}"
          >
            <div class="pro-pet-history__tile-top">
              <span class="pro-pet-history__icon" aria-hidden="true">
                <ProIcon :name="tile.icon" :size="18" />
              </span>
              <time class="pro-pet-history__time" :datetime="tile.iso">
                {{ formatTime(tile.item.createdAt) }}
              </time>
            </div>
            <div class="pro-pet-history__title-row">
              <strong class="pro-pet-history__title">{{ tile.title }}</strong>
              <ProBadge
                v-if="tile.reportStatus === 'final' || tile.reportStatus === 'draft'"
                :variant="tile.reportStatus === 'final' ? 'success' : 'warning'"
                data-testid="pet-history-report-badge"
              >
                {{ tile.reportBadge }}
              </ProBadge>
            </div>
            <p v-if="tile.item.body" class="pro-pet-history__body">{{ tile.item.body }}</p>
            <span v-if="tile.clickable" class="pro-pet-history__open-hint">
              <ProIcon name="description" :size="14" />
              {{ $t('clients.pet.timelineOpenReport') }}
            </span>
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

export type TimelineHistoryItem = TimelineDayItem & {
  meta?: {
    hasReport?: boolean
    reportStatus?: string
    visitId?: string
    status?: string
    source?: string
    [key: string]: unknown
  }
}

type HistoryTile = {
  item: TimelineHistoryItem
  clickable: boolean
  icon: string
  iso: string
  title: string
  reportStatus: string
  reportBadge: string
  ariaLabel: string
  listeners: Record<string, (e: Event) => void>
}

const props = withDefaults(
  defineProps<{
    items: TimelineHistoryItem[]
    title?: string
  }>(),
  {
    title: undefined,
  },
)

const emit = defineEmits<{
  'open-visit': [payload: { id: string, scheduledAt?: string }]
}>()

const { t } = useI18n()
const { formatDay, formatTime } = useFormatters()

const title = computed(() => props.title ?? t('clients.pet.timelineTitle'))

const dayGroups = computed(() =>
  groupTimelineByDay(props.items ?? []).map((group) => ({
    dayKey: group.dayKey,
    items: group.items.map(buildTile),
  })),
)

function iso(value: string | Date) {
  const d = value instanceof Date ? value : new Date(value)
  return Number.isNaN(d.getTime()) ? '' : d.toISOString()
}

function hasVisitReport(item: TimelineHistoryItem) {
  return item.type === 'visit' && Boolean(item.meta?.hasReport)
}

function visitIdFor(item: TimelineHistoryItem) {
  const fromMeta = item.meta?.visitId
  if (typeof fromMeta === 'string' && fromMeta.trim()) return fromMeta.trim()
  return item.id
}

function reportStatusOf(item: TimelineHistoryItem) {
  const raw = item.meta?.reportStatus
  return typeof raw === 'string' ? raw : ''
}

function itemTitle(item: TimelineHistoryItem) {
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

function buildTile(item: TimelineHistoryItem): HistoryTile {
  const clickable = hasVisitReport(item)
  const titleText = itemTitle(item)
  const reportStatus = reportStatusOf(item)
  const reportBadge = reportStatus === 'final'
    ? t('clients.pet.timelineReportFinal')
    : t('clients.pet.timelineReportDraft')
  const open = () => emit('open-visit', {
    id: visitIdFor(item),
    scheduledAt: typeof item.createdAt === 'string' ? item.createdAt : iso(item.createdAt),
  })
  return {
    item,
    clickable,
    icon: iconForType(item.type),
    iso: iso(item.createdAt),
    title: titleText,
    reportStatus,
    reportBadge,
    ariaLabel: `${titleText} — ${reportBadge}`,
    listeners: clickable
      ? {
          click: () => open(),
          keydown: (e: Event) => {
            const ke = e as KeyboardEvent
            if (ke.key === 'Enter' || ke.key === ' ') {
              ke.preventDefault()
              open()
            }
          },
        }
      : {},
  }
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

.pro-pet-history__tile--clickable {
  cursor: pointer;
  transition: box-shadow 0.15s ease, border-color 0.15s ease, background 0.15s ease;
}

.pro-pet-history__tile--clickable:hover {
  background: var(--pf-vet-surface, #fff);
  box-shadow: var(--pf-vet-shadow-sm);
  border-color: color-mix(in srgb, var(--pf-vet-accent) 45%, var(--pf-vet-border));
}

.pro-pet-history__tile--clickable:focus-visible {
  outline: none;
  box-shadow: var(--pf-vet-focus-ring);
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

.pro-pet-history__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
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

.pro-pet-history__open-hint {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.15rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--pf-vet-accent);
}
</style>
