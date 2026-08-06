<template>
  <div data-testid="commercial-prospect-detail-page">
    <ProPageHeader
      :title="prospect?.practiceName || $t('commercial.crm.ficheTitle')"
      :subtitle="$t('commercial.crm.ficheSubtitle')"
    >
      <template #actions>
        <NuxtLink to="/commercial/prospects" class="pro-link">{{ $t('commercial.crm.backList') }}</NuxtLink>
      </template>
    </ProPageHeader>

    <p v-if="loadError" class="pro-field-error" role="alert">{{ loadError }}</p>
    <p v-else-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>

    <template v-else-if="prospect">
      <ProCard class="pro-mb-lg" data-testid="prospect-fiche-header">
        <div class="pf-fiche-grid">
          <div>
            <p class="pro-hint">{{ $t('commercial.prospects.contactName') }}</p>
            <strong>{{ prospect.contactName || '—' }}</strong>
            <p>{{ prospect.contactEmail || '—' }} · {{ prospect.contactPhone || '—' }}</p>
            <p>{{ prospect.city || '—' }}</p>
          </div>
          <div>
            <label class="pro-label">{{ $t('commercial.prospects.statusLabel') }}</label>
            <select
              class="pro-select"
              data-testid="prospect-fiche-status"
              :value="prospect.status"
              @change="(e) => updateStatus((e.target as HTMLSelectElement).value)"
            >
              <option v-for="s in statuses" :key="s" :value="s">{{ $t(`commercial.prospects.status.${s}`) }}</option>
            </select>
            <label class="pro-label pro-mt-md">{{ $t('commercial.prospects.appointmentAt') }}</label>
            <input
              class="pro-input"
              type="datetime-local"
              data-testid="prospect-fiche-appt"
              :value="toLocalInput(prospect.appointmentAt)"
              @change="(e) => onAppt((e.target as HTMLInputElement).value)"
            >
          </div>
          <div class="pf-fiche-actions">
            <ProButton
              v-if="prospect.contactEmail && !prospect.emailOptOut"
              test-id="prospect-fiche-email"
              @click="openMailModal"
            >
              {{ $t('commercial.prospects.sendEmail') }}
            </ProButton>
            <NuxtLink
              v-if="prospect.status !== 'converted'"
              :to="`/commercial/vets?prospectId=${prospect.id}`"
              class="pro-link"
              data-testid="prospect-fiche-encode"
            >
              {{ $t('commercial.prospects.encode') }}
            </NuxtLink>
          </div>
        </div>
        <p v-if="actionError" class="pro-field-error pro-mt-md" role="alert">{{ actionError }}</p>
      </ProCard>

      <div class="pf-fiche-cols">
        <ProCard class="pro-mb-lg" data-testid="prospect-fiche-activities">
          <h3 class="pro-mb-md">{{ $t('commercial.crm.activitiesTitle') }}</h3>
          <form class="pro-form pf-inline-form pro-mb-md" @submit.prevent="createActivity">
            <ProInput v-model="actForm.title" test-id="prospect-act-title" :label="$t('commercial.crm.activityTitle')" required />
            <label class="pro-label">{{ $t('commercial.crm.activityKind') }}</label>
            <select v-model="actForm.kind" class="pro-select" data-testid="prospect-act-kind">
              <option v-for="k in activityKinds" :key="k" :value="k">{{ $t(`commercial.crm.kind.${k}`) }}</option>
            </select>
            <label class="pro-label">{{ $t('commercial.crm.dueAt') }}</label>
            <input v-model="actForm.dueAt" class="pro-input" type="datetime-local" data-testid="prospect-act-due">
            <ProButton type="submit" test-id="prospect-act-submit" :loading="actSaving">
              {{ $t('commercial.crm.addActivity') }}
            </ProButton>
          </form>
          <ProTable :empty="!openActivities.length" :empty-title="$t('commercial.crm.activitiesEmpty')">
            <thead>
              <tr>
                <th>{{ $t('commercial.crm.activityTitle') }}</th>
                <th>{{ $t('commercial.crm.dueAt') }}</th>
                <th />
              </tr>
            </thead>
            <tbody>
              <tr v-for="a in openActivities" :key="a.id" :data-testid="`prospect-act-${a.id}`">
                <td>{{ a.title }} <span class="pro-hint">({{ $t(`commercial.crm.kind.${a.kind}`) }})</span></td>
                <td>{{ formatDt(a.dueAt) }}</td>
                <td>
                  <ProButton variant="ghost" :test-id="`prospect-act-done-${a.id}`" @click="markDone(a.id)">
                    {{ $t('commercial.crm.markDone') }}
                  </ProButton>
                </td>
              </tr>
            </tbody>
          </ProTable>
        </ProCard>

        <ProCard class="pro-mb-lg" data-testid="prospect-fiche-timeline">
          <h3 class="pro-mb-md">{{ $t('commercial.crm.timelineTitle') }}</h3>
          <form class="pro-form pf-inline-form pro-mb-md" @submit.prevent="addNote">
            <ProInput v-model="noteBody" test-id="prospect-note-body" :label="$t('commercial.crm.noteBody')" required />
            <label class="pro-label">{{ $t('commercial.crm.eventKind') }}</label>
            <select v-model="noteKind" class="pro-select" data-testid="prospect-note-kind">
              <option value="note">{{ $t('commercial.crm.event.note') }}</option>
              <option value="call">{{ $t('commercial.crm.event.call') }}</option>
              <option value="meeting">{{ $t('commercial.crm.event.meeting') }}</option>
            </select>
            <ProButton type="submit" test-id="prospect-note-submit" :loading="noteSaving">
              {{ $t('commercial.crm.addNote') }}
            </ProButton>
          </form>
          <ul class="pf-timeline" data-testid="prospect-timeline-list">
            <li v-for="ev in events" :key="ev.id" class="pf-timeline__item" :data-testid="`prospect-event-${ev.id}`">
              <span class="pf-timeline__kind">{{ $t(`commercial.crm.event.${ev.kind}`, ev.kind) }}</span>
              <span class="pf-timeline__date">{{ formatDt(ev.createdAt) }}</span>
              <p>{{ ev.body }}</p>
              <p v-if="ev.actorName" class="pro-hint">{{ ev.actorName }}</p>
            </li>
          </ul>
          <p v-if="!events.length" class="pro-hint">{{ $t('commercial.crm.timelineEmpty') }}</p>
        </ProCard>
      </div>
    </template>

    <ProModal
      :open="mailOpen"
      :title="prospect ? $t('commercial.mail.sendTitle', { practice: prospect.practiceName }) : ''"
      :prevent-close="mailSending"
      size="lg"
      test-id="prospect-fiche-mail-modal"
      @update:open="(v) => { mailOpen = v }"
    >
      <p v-if="mailOk" class="pro-hint" role="status">{{ $t('commercial.mail.sent') }}</p>
      <p v-if="mailError" class="pro-field-error" role="alert">{{ mailError }}</p>
      <form class="pro-form" @submit.prevent="sendMail">
        <label class="pro-label">{{ $t('commercial.mail.template') }}</label>
        <select v-model="mailForm.templateId" class="pro-select" data-testid="prospect-fiche-mail-template" required @change="onMailTemplateChange">
          <option value="" disabled>—</option>
          <option v-for="tpl in mailTemplates" :key="tpl.id" :value="tpl.id">{{ tpl.name }}</option>
        </select>
        <ProInput v-model="mailForm.subject" test-id="prospect-fiche-mail-subject" :label="$t('commercial.mail.subject')" />
        <label class="pro-label">{{ $t('commercial.mail.body') }}</label>
        <textarea v-model="mailForm.bodyHtml" class="pro-input" rows="8" data-testid="prospect-fiche-mail-body" />
        <ProButton type="submit" test-id="prospect-fiche-mail-send" :loading="mailSending">{{ $t('commercial.mail.sendAction') }}</ProButton>
      </form>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { t } = useI18n()
const route = useRoute()
const { formatDate } = useFormatters()
const { user } = useProUser()
const prospectId = computed(() => String(route.params.id || ''))
const isManager = computed(() => user.value?.role === 'commercial_manager')
const prospectPatchPath = computed(() =>
  isManager.value
    ? `/api/commercial-manager/prospects/${prospectId.value}`
    : `/api/commercial/prospects/${prospectId.value}`,
)

const statuses = ['new', 'contacted', 'qualified', 'converted', 'lost']
const activityKinds = ['call', 'email', 'follow_up', 'meeting', 'other']

const loading = ref(true)
const loadError = ref('')
const actionError = ref('')
const prospect = ref<any>(null)
const events = ref<any[]>([])
const openActivities = ref<any[]>([])

const noteBody = ref('')
const noteKind = ref('note')
const noteSaving = ref(false)

const actForm = reactive({ title: '', kind: 'follow_up', dueAt: '' })
const actSaving = ref(false)

const mailOpen = ref(false)
const mailSending = ref(false)
const mailOk = ref(false)
const mailError = ref('')
const mailTemplates = ref<any[]>([])
const mailForm = reactive({ templateId: '', subject: '', bodyHtml: '' })

function mapError(e: any) {
  return e?.data?.error?.message || e?.data?.message || e?.message || t('commercial.crm.loadError')
}

function formatDt(v?: string) {
  if (!v) return '—'
  try { return formatDate(v) } catch { return v }
}

function toLocalInput(iso?: string) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const res: any = await $fetch(`/api/commercial/prospects/${prospectId.value}`)
    const data = res.data ?? res
    openActivities.value = (data.activities ?? []).filter((a: any) => a.status === 'open')
    events.value = data.events ?? []
    prospect.value = data.prospect ?? data
  } catch (e: any) {
    loadError.value = mapError(e)
    prospect.value = null
  } finally {
    loading.value = false
  }
}

async function patch(body: Record<string, unknown>) {
  actionError.value = ''
  try {
    await $fetch(prospectPatchPath.value, { method: 'PATCH', body })
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  }
}

async function updateStatus(status: string) {
  await patch({ status })
}

async function onAppt(value: string) {
  if (!value) {
    await patch({ clearAppointment: true })
    return
  }
  await patch({ appointmentAt: new Date(value).toISOString(), appointmentOutcome: 'scheduled' })
}

async function addNote() {
  noteSaving.value = true
  actionError.value = ''
  try {
    await $fetch(`/api/commercial/prospects/${prospectId.value}/events`, {
      method: 'POST',
      body: { kind: noteKind.value, body: noteBody.value },
    })
    noteBody.value = ''
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    noteSaving.value = false
  }
}

async function createActivity() {
  actSaving.value = true
  actionError.value = ''
  try {
    const body: Record<string, unknown> = { title: actForm.title, kind: actForm.kind }
    if (actForm.dueAt) body.dueAt = new Date(actForm.dueAt).toISOString()
    await $fetch(`/api/commercial/prospects/${prospectId.value}/activities`, { method: 'POST', body })
    actForm.title = ''
    actForm.dueAt = ''
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    actSaving.value = false
  }
}

async function markDone(id: string) {
  actionError.value = ''
  try {
    await $fetch(`/api/commercial/activities/${id}`, { method: 'PATCH', body: { status: 'done' } })
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  }
}

async function openMailModal() {
  mailOpen.value = true
  mailOk.value = false
  mailError.value = ''
  mailForm.templateId = ''
  mailForm.subject = ''
  mailForm.bodyHtml = ''
  try {
    const res: any = await $fetch('/api/commercial/email-templates', { query: { active: '1' } })
    mailTemplates.value = res.data ?? res ?? []
  } catch (e: any) {
    mailError.value = mapError(e)
  }
}

function onMailTemplateChange() {
  const tpl = mailTemplates.value.find((t) => t.id === mailForm.templateId)
  if (!tpl) return
  mailForm.subject = tpl.subject
  mailForm.bodyHtml = tpl.bodyHtml
}

async function sendMail() {
  if (!mailForm.templateId) return
  mailSending.value = true
  mailError.value = ''
  mailOk.value = false
  try {
    await $fetch(`/api/commercial/prospects/${prospectId.value}/emails`, {
      method: 'POST',
      body: {
        templateId: mailForm.templateId,
        subject: mailForm.subject || undefined,
        bodyHtml: mailForm.bodyHtml || undefined,
      },
    })
    mailOk.value = true
    await load()
  } catch (e: any) {
    mailError.value = mapError(e)
  } finally {
    mailSending.value = false
  }
}

watch(prospectId, () => { load() }, { immediate: true })
</script>

<style scoped>
.pf-fiche-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}
.pf-fiche-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: flex-start;
}
.pf-fiche-cols {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
}
.pf-inline-form {
  display: grid;
  gap: 0.5rem;
}
.pf-timeline {
  list-style: none;
  margin: 0;
  padding: 0;
}
.pf-timeline__item {
  border-left: 2px solid var(--pf-vet-border, #ddd);
  padding: 0.5rem 0 0.75rem 0.75rem;
  margin-bottom: 0.25rem;
}
.pf-timeline__kind {
  font-weight: 600;
  margin-right: 0.5rem;
}
.pf-timeline__date {
  color: var(--pf-vet-muted, #666);
  font-size: 0.85rem;
}
</style>
