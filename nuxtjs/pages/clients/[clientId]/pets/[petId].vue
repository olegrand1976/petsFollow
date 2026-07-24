<template>
  <div data-testid="pet-detail-page">
    <nav class="pro-breadcrumb" :aria-label="$t('common.breadcrumb')">
      <NuxtLink to="/pets">{{ $t('nav.pets') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <NuxtLink to="/clients">{{ $t('nav.clients') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <NuxtLink :to="`/clients/${clientId}`">{{ $t('common.profile') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <span>{{ pet?.name || $t('clients.pet.title') }}</span>
    </nav>
    <ProPageHeader
      :title="pet?.name || $t('clients.pet.title')"
      :subtitle="petSubtitle"
    >
      <template #actions>
        <div class="pro-pet-header-actions">
          <ProBadge v-if="isPrimaryPractice" variant="success">{{ $t('clients.pet.primaryBadge') }}</ProBadge>
          <ProButton
            variant="secondary"
            test-id="pet-open-messages"
            :disabled="messagingBusy"
            @click="openMessaging"
          >
            <ProIcon name="chat" :size="18" />
            {{ $t('clients.pet.openMessages') }}
          </ProButton>
          <ProAvatar :src="pet?.photoUrl" :name="pet?.name || ''" size="lg" />
        </div>
      </template>
    </ProPageHeader>
    <p v-if="pageError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ pageError }}</p>
    <p v-if="messagingError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ messagingError }}</p>

    <ProCard :title="$t('clients.pet.photoTitle')" class="pro-settings-card">
      <ProAvatarUpload
        v-model="petPhotoUrl"
        :name="pet?.name || ''"
        :upload-url="`/api/pets/${petId}/photo`"
        :label="$t('clients.pet.photoChange')"
        :hint="$t('clients.pet.photoHint')"
        @uploaded="onPetPhotoUploaded"
      />
    </ProCard>

    <div v-if="recentAlert" class="pro-alert-banner" role="status" data-testid="pet-alert-banner">
      <ProBadge variant="danger">{{ $t('clients.pet.alert') }}</ProBadge>
      <span>{{ $t('clients.pet.recentAlertBanner') }}</span>
    </div>

    <ProCard v-if="hasChartData" :title="$t('clients.pet.chartTitle')">
      <div class="pro-toggle pro-pet-filter" role="group" :aria-label="$t('clients.pet.chartRangeLabel')">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': chartRange === '3m' }"
          data-testid="pet-chart-range-3m"
          @click="chartRange = '3m'"
        >
          {{ $t('clients.pet.chartRange3m') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': chartRange === '6m' }"
          data-testid="pet-chart-range-6m"
          @click="chartRange = '6m'"
        >
          {{ $t('clients.pet.chartRange6m') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': chartRange === '1y' }"
          data-testid="pet-chart-range-1y"
          @click="chartRange = '1y'"
        >
          {{ $t('clients.pet.chartRange1y') }}
        </button>
      </div>
      <ProBpmChart
        v-if="chartValues.length"
        :values="chartValues"
        :alerts="chartAlerts"
        :dates="chartDates"
        :domain-start="chartDomain.start"
        :domain-end="chartDomain.end"
        :aria-label="$t('clients.pet.chartTitle')"
      />
      <p v-else class="text-muted" data-testid="pet-chart-empty-period">
        {{ $t('clients.pet.chartEmptyPeriod') }}
      </p>
    </ProCard>

    <ProCard :title="$t('clients.pet.heartrateTitle')">
      <div class="pro-toggle pro-pet-filter" role="group" :aria-label="$t('clients.pet.heartrateTitle')">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': sessionFilter === 'all' }"
          data-testid="pet-filter-all"
          @click="sessionFilter = 'all'"
        >
          {{ $t('clients.pet.filterAll') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': sessionFilter === 'alerts' }"
          data-testid="pet-filter-alerts"
          @click="sessionFilter = 'alerts'"
        >
          {{ $t('clients.pet.filterAlerts') }}
        </button>
      </div>
      <ProTable
        :empty="!filteredSessions.length"
        :empty-title="$t('clients.pet.heartrateEmptyTitle')"
        :empty-description="$t('clients.pet.heartrateEmptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.columnBpm') }}</th>
            <th>{{ $t('clients.pet.columnDuration') }}</th>
            <th>{{ $t('clients.pet.columnStatus') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="s in filteredSessions"
            :key="s.id"
            :class="{ 'pro-table-row--alert': s.isAlert }"
          >
            <td>
              <span class="pro-pet-reading-date">
                {{ formatDate(s.startedAt) }}
                <ProBadge
                  v-if="isReadingNew(s)"
                  variant="danger"
                  data-testid="reading-new-badge"
                >
                  {{ $t('clients.pet.newReading') }}
                </ProBadge>
              </span>
            </td>
            <td><code>{{ s.bpm }}</code></td>
            <td>{{ s.durationSec }}s</td>
            <td>
              <ProBadge :variant="s.isAlert ? 'danger' : 'success'">
                {{ s.isAlert ? $t('clients.pet.alert') : $t('clients.pet.ok') }}
              </ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('clients.pet.careTitle')" class="pro-mb-lg">
      <form class="pro-pet-inline-form" @submit.prevent="createCare">
        <input v-model="careDraft.title" class="pro-input" :placeholder="$t('clients.pet.careTitleField')" required />
        <select v-model="careDraft.type" class="pro-input" :aria-label="$t('clients.pet.careType')">
          <option value="vaccination">{{ $t('clients.pet.careTypeVaccination') }}</option>
          <option value="deworming">{{ $t('clients.pet.careTypeDeworming') }}</option>
          <option value="vet_check">{{ $t('clients.pet.careTypeVetCheck') }}</option>
          <option value="dental">{{ $t('clients.pet.careTypeDental') }}</option>
          <option value="farrier">{{ $t('clients.pet.careTypeFarrier') }}</option>
          <option value="fecal_egg">{{ $t('clients.pet.careTypeFecalEgg') }}</option>
          <option value="custom">{{ $t('clients.pet.careTypeCustom') }}</option>
        </select>
        <ProButton type="submit" :disabled="careBusy">{{ $t('clients.pet.careCreate') }}</ProButton>
      </form>
      <ProEmptyState
        v-if="!careReminders.length"
        :title="$t('clients.pet.careEmptyTitle')"
        :description="$t('clients.pet.careEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.pet.careType') }}</th>
            <th>{{ $t('clients.pet.careTitleField') }}</th>
            <th>{{ $t('clients.pet.careDue') }}</th>
            <th>{{ $t('clients.pet.columnStatus') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in careReminders" :key="c.id">
            <td>{{ careTypeLabel(c.type) }}</td>
            <td>{{ c.title }}</td>
            <td>
              {{ formatDate(c.dueAt) }}
              <ProBadge v-if="isCareOverdue(c)" variant="danger">{{ $t('clients.pet.careOverdue') }}</ProBadge>
            </td>
            <td>{{ c.status }}</td>
            <td>
              <div v-if="c.status === 'pending'" class="pro-flex-gap">
                <ProButton :disabled="careBusy" @click="markCareDone(c.id)">{{ $t('clients.pet.careDone') }}</ProButton>
                <ProButton variant="ghost" :disabled="careBusy" @click="postponeCare(c.id, 7)">{{ $t('clients.pet.carePostpone') }}</ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('clients.pet.visitsTitle')" class="pro-mb-lg">
      <form class="pro-pet-inline-form" @submit.prevent="proposeVisit(false)">
        <input v-model="visitDraft.scheduledAt" class="pro-input" type="datetime-local" :aria-label="$t('clients.pet.visitScheduledAt')" required />
        <input v-model="visitDraft.notes" class="pro-input" :placeholder="$t('clients.pet.visitNotes')" />
        <ProButton type="submit" :disabled="visitBusy" variant="secondary">
          {{ $t('clients.pet.visitPropose') }}
        </ProButton>
        <ProButton type="button" :disabled="visitBusy || !visitDraft.scheduledAt" @click="proposeVisit(true)">
          {{ $t('clients.pet.visitConfirmDirect') }}
        </ProButton>
      </form>
      <ProEmptyState
        v-if="!visits.length"
        :title="$t('clients.pet.visitsEmptyTitle')"
        :description="$t('clients.pet.visitsEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.visitNotes') }}</th>
            <th>{{ $t('clients.pet.visitStatus') }}</th>
            <th>{{ $t('clients.pet.visitSource') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="v in visits" :key="v.id">
            <td>{{ formatDate(v.scheduledAt || v.createdAt) }}</td>
            <td>{{ v.notes || '—' }}</td>
            <td>{{ v.status }}</td>
            <td>
              {{ v.source === 'vet' ? $t('clients.pet.visitSourceVet') : $t('clients.pet.visitSourceClient') }}
            </td>
            <td>
              <div class="pro-flex-gap">
                <ProButton
                  v-if="v.status === 'requested' && v.pendingActionBy === 'vet'"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'confirm')"
                >
                  {{ $t('clients.pet.visitConfirm') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'reschedule_pending' && v.pendingActionBy === 'vet'"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'accept_reschedule')"
                >
                  {{ $t('calendar.acceptReschedule') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'reschedule_pending' && v.pendingActionBy === 'vet'"
                  variant="ghost"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'reject_reschedule')"
                >
                  {{ $t('calendar.rejectReschedule') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'confirmed'"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'done')"
                >
                  {{ $t('clients.pet.visitDone') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'requested' || v.status === 'confirmed' || v.status === 'reschedule_pending'"
                  variant="ghost"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'cancel')"
                >
                  {{ $t('clients.pet.visitCancel') }}
                </ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('share.petTitle')" class="pro-mb-lg" data-testid="pet-shares-card">
      <p class="pro-hint pro-mb-md">{{ $t('share.petHint') }}</p>
      <form class="pro-pet-inline-form" @submit.prevent="addPetShare">
        <select v-model="shareColleagueId" class="pro-input" data-testid="pet-share-colleague">
          <option value="">{{ $t('share.colleaguePlaceholder') }}</option>
          <option v-for="c in colleagues" :key="c.userId" :value="c.userId">
            {{ c.fullName }} ({{ c.email }})
          </option>
        </select>
        <ProInput v-model="shareEmail" type="email" :label="$t('share.email')" test-id="pet-share-email" />
        <select v-model="sharePermission" class="pro-input" data-testid="pet-share-permission">
          <option value="read">{{ $t('share.permRead') }}</option>
          <option value="write_notes">{{ $t('share.permWriteNotes') }}</option>
          <option value="full">{{ $t('share.permFull') }}</option>
        </select>
        <select v-model="shareExpiresDays" class="pro-input" data-testid="pet-share-expires">
          <option value="">{{ $t('share.expiresNever') }}</option>
          <option value="7">{{ $t('share.expiresDays', { n: 7 }) }}</option>
          <option value="30">{{ $t('share.expiresDays', { n: 30 }) }}</option>
          <option value="90">{{ $t('share.expiresDays', { n: 90 }) }}</option>
        </select>
        <ProButton type="submit" :disabled="shareBusy" test-id="pet-share-submit">
          {{ $t('share.add') }}
        </ProButton>
      </form>
      <p v-if="shareError" class="pro-error">{{ shareError }}</p>
      <ProTable v-if="petShares.length">
        <thead>
          <tr>
            <th>{{ $t('share.columnName') }}</th>
            <th>{{ $t('share.columnEmail') }}</th>
            <th>{{ $t('share.columnPermission') }}</th>
            <th>{{ $t('share.columnExpires') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in petShares" :key="s.id">
            <td>{{ s.granteeName }}</td>
            <td>{{ s.granteeEmail }}</td>
            <td>{{ s.permission }}</td>
            <td>{{ s.expiresAt ? formatShareDate(s.expiresAt) : $t('share.expiresNever') }}</td>
            <td>
              <ProButton variant="ghost" :disabled="shareBusy" @click="revokePetShare(s.granteeUserId)">
                {{ $t('share.revoke') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('clients.pet.documentsTitle')" class="pro-mb-lg">
      <form class="pro-pet-inline-form" @submit.prevent="uploadDocument">
        <input
          ref="docInputEl"
          type="file"
          accept="application/pdf,image/jpeg,image/png,image/webp"
          class="pro-input"
          data-testid="pet-document-input"
          @change="onDocFile"
        />
        <input
          v-model="docTitle"
          class="pro-input"
          :placeholder="$t('clients.pet.documentTitlePlaceholder')"
          data-testid="pet-document-title"
        />
        <ProButton type="submit" :disabled="docBusy || !docFile" data-testid="pet-document-upload">
          {{ $t('clients.pet.documentUpload') }}
        </ProButton>
      </form>
      <p v-if="docError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ docError }}</p>
      <p class="pro-hint">{{ $t('clients.pet.documentsHint') }}</p>
      <ProEmptyState
        v-if="!documents.length"
        :title="$t('clients.pet.documentsEmptyTitle')"
        :description="$t('clients.pet.documentsEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.pet.documentName') }}</th>
            <th>{{ $t('clients.pet.documentType') }}</th>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.documentUploader') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="d in documents" :key="d.id">
            <td>
              <a :href="d.fileUrl" target="_blank" rel="noopener noreferrer">
                {{ d.title || d.fileName }}
              </a>
            </td>
            <td>{{ documentKindLabel(d.contentType) }}</td>
            <td>{{ formatDate(d.createdAt) }}</td>
            <td>{{ d.uploaderName || $t('common.dash') }}</td>
            <td>
              <div class="pro-flex-gap">
                <a :href="d.fileUrl" target="_blank" rel="noopener noreferrer" class="pro-link-btn">
                  {{ $t('clients.pet.documentOpen') }}
                </a>
                <ProButton variant="ghost" :disabled="docBusy" @click="deleteDocument(d.id)">
                  {{ $t('common.delete') }}
                </ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('clients.pet.timelineTitle')">
      <ul v-if="timeline.length" class="pro-timeline">
        <li v-for="item in timeline" :key="item.id" class="pro-timeline__item">
          <div class="pro-timeline__dot" aria-hidden="true" />
          <div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.body }}</p>
            <small class="text-muted">{{ formatDate(item.createdAt) }}</small>
          </div>
        </li>
      </ul>
      <ProEmptyState
        v-else
        :title="$t('clients.pet.timelineEmptyTitle')"
        :description="$t('clients.pet.timelineEmptyDescription')"
      />
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const route = useRoute()
const clientId = route.params.clientId as string
const petId = route.params.petId as string
const pet = ref<any>(null)
const petPhotoUrl = ref('')
const sessions = ref<any[]>([])
const timeline = ref<any[]>([])
const careReminders = ref<any[]>([])
const visits = ref<any[]>([])
const documents = ref<any[]>([])
const sessionFilter = ref<'all' | 'alerts'>('all')
const chartRange = ref<'3m' | '6m' | '1y'>('3m')
const highlightedNewIds = ref<Set<string>>(new Set())
const careBusy = ref(false)
const visitBusy = ref(false)
const messagingBusy = ref(false)
const messagingError = ref('')
const pageError = ref('')
const docBusy = ref(false)
const docError = ref('')
const docTitle = ref('')
const docFile = ref<File | null>(null)
const docInputEl = ref<HTMLInputElement | null>(null)
const petShares = ref<any[]>([])
const colleagues = ref<{ userId: string; fullName: string; email: string }[]>([])
const shareColleagueId = ref('')
const shareEmail = ref('')
const sharePermission = ref('write_notes')
const shareExpiresDays = ref('')
const shareBusy = ref(false)
const shareError = ref('')
const careDraft = reactive({ title: '', type: 'vaccination' })
const visitDraft = reactive({ scheduledAt: '', notes: '' })
let sessionsPollTimer: ReturnType<typeof setInterval> | null = null

const { formatDate } = useFormatters()
const { t } = useI18n()
const { mapError } = useApiError()
const { user, fetchUser } = useProUser()
const { refresh: refreshNavBadges } = useNavBadges()
const router = useRouter()

function isReadingNew(s: { id: string, isNew?: boolean }) {
  return highlightedNewIds.value.has(s.id) || !!s.isNew
}

async function loadSessions(markSeenAfter = false) {
  const sessionsRes: any = await $fetch(`/api/pets/${petId}/heartrate`)
  const list = sessionsRes.data ?? sessionsRes ?? []
  const next = new Set(highlightedNewIds.value)
  let hadNew = false
  for (const s of list) {
    if (s.isNew) {
      next.add(s.id)
      hadNew = true
    }
  }
  highlightedNewIds.value = next
  sessions.value = list
  if (markSeenAfter && hadNew) {
    try {
      await $fetch(`/api/pets/${petId}/heartrate/seen`, { method: 'POST' })
      await refreshNavBadges()
    } catch {
      // Non-blocking: unread badges may stay until next visit.
    }
  }
}
function careTypeLabel(type: string) {
  switch (type) {
    case 'vaccination':
      return t('clients.pet.careTypeVaccination')
    case 'deworming':
      return t('clients.pet.careTypeDeworming')
    case 'vet_check':
      return t('clients.pet.careTypeVetCheck')
    case 'dental':
      return t('clients.pet.careTypeDental')
    case 'farrier':
      return t('clients.pet.careTypeFarrier')
    case 'fecal_egg':
      return t('clients.pet.careTypeFecalEgg')
    case 'custom':
      return t('clients.pet.careTypeCustom')
    default:
      return type
  }
}

const petSubtitle = computed(() => {
  if (!pet.value) return ''
  return [pet.value.species, pet.value.breed].filter(Boolean).join(' · ')
})

const isPrimaryPractice = computed(() => {
  const practiceId = user.value?.practiceId
  return Boolean(practiceId && pet.value?.practiceId && practiceId === pet.value.practiceId)
})

const filteredSessions = computed(() => {
  if (sessionFilter.value === 'alerts') {
    return sessions.value.filter((s) => s.isAlert)
  }
  return sessions.value
})

const recentAlert = computed(() => sessions.value.length > 0 && !!sessions.value[0]?.isAlert)

const rangeMonths = computed(() =>
  chartRange.value === '3m' ? 3 : chartRange.value === '6m' ? 6 : 12,
)

const chartDomain = computed(() => {
  const end = new Date()
  const start = new Date(end)
  start.setMonth(start.getMonth() - rangeMonths.value)
  return { start: start.toISOString(), end: end.toISOString() }
})

const hasChartData = computed(() =>
  sessions.value.some(s => s.bpm != null && s.startedAt),
)

const chartSessions = computed(() => {
  const cutoff = new Date(chartDomain.value.start)
  return [...sessions.value]
    .filter(s => s.bpm != null && s.startedAt && new Date(s.startedAt) >= cutoff)
    .sort((a, b) => +new Date(a.startedAt) - +new Date(b.startedAt))
})
const chartValues = computed(() => chartSessions.value.map(s => s.bpm as number))
const chartAlerts = computed(() => chartSessions.value.map(s => !!s.isAlert))
const chartDates = computed(() => chartSessions.value.map(s => s.startedAt as string))

function onPetPhotoUploaded(data: any) {
  pet.value = { ...pet.value, ...data }
  petPhotoUrl.value = data?.photoUrl || petPhotoUrl.value
}

function isCareOverdue(c: any) {
  return c.status === 'pending' && c.dueAt && new Date(c.dueAt).getTime() < Date.now()
}

async function loadCareAndVisits() {
  const [careRes, visitsRes]: any[] = await Promise.all([
    $fetch(`/api/pets/${petId}/care-reminders`),
    $fetch(`/api/pets/${petId}/visits`),
  ])
  careReminders.value = careRes.data ?? careRes ?? []
  visits.value = visitsRes.data ?? visitsRes ?? []
}

async function loadDocuments() {
  const res: any = await $fetch(`/api/pets/${petId}/documents`)
  documents.value = res.data ?? res ?? []
}

async function loadPetShares() {
  const [sharesRes, colleaguesRes]: any[] = await Promise.all([
    $fetch(`/api/pets/${petId}/shares`),
    $fetch('/api/vet/colleagues').catch(() => ({ data: [] })),
  ])
  petShares.value = sharesRes.data ?? sharesRes ?? []
  colleagues.value = colleaguesRes.data ?? colleaguesRes ?? []
}

async function addPetShare() {
  shareBusy.value = true
  shareError.value = ''
  try {
    const body: Record<string, string> = { permission: sharePermission.value || 'write_notes' }
    if (shareColleagueId.value) body.granteeUserId = shareColleagueId.value
    else if (shareEmail.value.trim()) body.email = shareEmail.value.trim()
    else {
      shareError.value = t('share.fieldsRequired')
      return
    }
    if (shareExpiresDays.value) {
      const d = new Date()
      d.setDate(d.getDate() + Number(shareExpiresDays.value))
      body.expiresAt = d.toISOString()
    }
    await $fetch(`/api/pets/${petId}/shares`, { method: 'POST', body })
    shareColleagueId.value = ''
    shareEmail.value = ''
    sharePermission.value = 'write_notes'
    shareExpiresDays.value = ''
    await loadPetShares()
  } catch (e: any) {
    shareError.value = mapError(e)
  } finally {
    shareBusy.value = false
  }
}

function formatShareDate(iso: string) {
  try {
    return new Date(iso).toLocaleDateString()
  } catch {
    return iso
  }
}

async function revokePetShare(granteeUserId: string) {
  shareBusy.value = true
  try {
    await $fetch(`/api/pets/${petId}/shares/${granteeUserId}`, { method: 'DELETE' })
    await loadPetShares()
  } catch (e: any) {
    shareError.value = mapError(e)
  } finally {
    shareBusy.value = false
  }
}

function documentKindLabel(contentType: string) {
  if (contentType?.includes('pdf')) return t('clients.pet.documentTypePdf')
  if (contentType?.startsWith('image/')) return t('clients.pet.documentTypeImage')
  return contentType || t('common.dash')
}

function onDocFile(ev: Event) {
  const input = ev.target as HTMLInputElement
  docFile.value = input.files?.[0] ?? null
  docError.value = ''
}

async function uploadDocument() {
  if (!docFile.value) return
  docBusy.value = true
  docError.value = ''
  try {
    const fd = new FormData()
    fd.append('file', docFile.value)
    if (docTitle.value.trim()) fd.append('title', docTitle.value.trim())
    await $fetch(`/api/pets/${petId}/documents`, { method: 'POST', body: fd })
    docTitle.value = ''
    docFile.value = null
    if (docInputEl.value) docInputEl.value.value = ''
    await loadDocuments()
  } catch (e: any) {
    docError.value = mapError(e)
  } finally {
    docBusy.value = false
  }
}

async function deleteDocument(id: string) {
  docBusy.value = true
  docError.value = ''
  try {
    await $fetch(`/api/pets/documents/${id}`, { method: 'DELETE' })
    await loadDocuments()
  } catch (e: any) {
    docError.value = mapError(e)
  } finally {
    docBusy.value = false
  }
}

async function openMessaging() {
  const ownerId = pet.value?.ownerUserId || clientId
  if (!ownerId) return
  messagingBusy.value = true
  messagingError.value = ''
  try {
    const res: any = await $fetch('/api/messaging/threads', {
      method: 'POST',
      body: { clientUserId: ownerId },
    })
    const thread = res.data ?? res
    await router.push({ path: '/messages', query: { thread: thread.id } })
  } catch (e: any) {
    messagingError.value = mapError(e)
  } finally {
    messagingBusy.value = false
  }
}

async function createCare() {
  careBusy.value = true
  try {
    await $fetch(`/api/pets/${petId}/care-reminders`, {
      method: 'POST',
      body: { title: careDraft.title, type: careDraft.type, dueDays: 30 },
    })
    careDraft.title = ''
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function markCareDone(id: string) {
  careBusy.value = true
  try {
    await $fetch(`/api/care-reminders/${id}/done`, { method: 'POST' })
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function postponeCare(id: string, days: number) {
  careBusy.value = true
  try {
    await $fetch(`/api/care-reminders/${id}/postpone`, { method: 'POST', body: { days } })
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function proposeVisit(confirmDirect: boolean) {
  if (!visitDraft.scheduledAt) return
  visitBusy.value = true
  try {
    await $fetch(`/api/pets/${petId}/visits`, {
      method: 'POST',
      body: {
        notes: visitDraft.notes,
        confirmDirect,
        scheduledAt: new Date(visitDraft.scheduledAt).toISOString(),
      },
    })
    visitDraft.notes = ''
    visitDraft.scheduledAt = ''
    await loadCareAndVisits()
  } finally {
    visitBusy.value = false
  }
}

async function visitAction(id: string, action: string) {
  visitBusy.value = true
  pageError.value = ''
  try {
    await $fetch(`/api/visits/${id}`, { method: 'PATCH', body: { action } })
    await loadCareAndVisits()
  } catch (e: any) {
    pageError.value = mapError(e)
  } finally {
    visitBusy.value = false
  }
}

onMounted(async () => {
  pageError.value = ''
  try {
    await fetchUser()
    const petRes: any = await $fetch(`/api/pets/${petId}`)
    pet.value = petRes.data ?? petRes
    petPhotoUrl.value = pet.value?.photoUrl || ''

    const timelineRes: any = await $fetch(`/api/pets/${petId}/timeline`)
    timeline.value = timelineRes.data ?? timelineRes ?? []
    await loadSessions(true)
    sessionsPollTimer = setInterval(() => {
      loadSessions(true).catch(() => {})
    }, 8000)
  } catch (e: any) {
    pageError.value = mapError(e)
    return
  }

  try {
    await loadCareAndVisits()
  } catch (e: any) {
    pageError.value = mapError(e)
  }

  try {
    await loadDocuments()
  } catch (e: any) {
    docError.value = mapError(e)
  }

  try {
    await loadPetShares()
  } catch {
    petShares.value = []
  }
})

onBeforeUnmount(() => {
  if (sessionsPollTimer) clearInterval(sessionsPollTimer)
})
</script>

<style scoped>
.pro-alert-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  margin-bottom: 1rem;
  border-radius: var(--pf-vet-radius);
  border: 1px solid color-mix(in srgb, var(--pf-vet-alert) 35%, transparent);
  background: color-mix(in srgb, var(--pf-vet-alert) 8%, var(--pf-vet-surface));
}

.pro-pet-filter {
  margin-bottom: 1rem;
}

.pro-pet-reading-date {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pro-table-row--alert {
  background: color-mix(in srgb, var(--pf-vet-alert) 6%, transparent);
}

.pro-pet-header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.pro-pet-inline-form {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.pro-pet-inline-form .pro-input {
  flex: 1 1 10rem;
  min-width: 8rem;
}

.pro-timeline {
  list-style: none;
  margin: 0;
  padding: 0;
}

.pro-timeline__item {
  display: grid;
  grid-template-columns: 1rem 1fr;
  gap: 0.75rem 1rem;
  padding-bottom: 1.25rem;
  border-left: 2px solid var(--pf-vet-border);
  margin-left: 0.35rem;
  padding-left: 1.25rem;
  position: relative;
}

.pro-timeline__item:last-child {
  border-left-color: transparent;
  padding-bottom: 0;
}

.pro-timeline__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--pf-vet-accent);
  position: absolute;
  left: -6px;
  top: 0.35rem;
}

.pro-timeline__item p {
  margin: 0.25rem 0;
}
</style>
