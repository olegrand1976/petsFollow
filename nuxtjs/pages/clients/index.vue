<template>
  <div data-testid="clients-page">
    <ProPageHeader :title="$t('clients.title')" :subtitle="$t('clients.subtitle')" />
    <p v-if="appLinkFeedback" class="pro-inline-feedback" role="status">{{ appLinkFeedback }}</p>
    <ProCard class="pro-mb-lg" data-testid="vet-create-client-form">
      <h3 class="pro-mb-md">{{ $t('clients.create.title') }}</h3>
      <p class="pro-hint pro-mb-md">{{ $t('clients.create.hint') }}</p>
      <form class="pro-form" @submit.prevent="createClient">
        <ProInput v-model="clientForm.fullName" test-id="create-client-name" :label="$t('clients.create.fullName')" required />
        <ProInput v-model="clientForm.email" test-id="create-client-email" type="email" :label="$t('clients.create.email')" required />
        <ProInput v-model="clientForm.password" test-id="create-client-password" type="password" :label="$t('clients.create.password')" required />
        <p v-if="clientMsg" class="pro-hint" data-testid="create-client-msg">{{ clientMsg }}</p>
        <p v-if="clientError" class="pro-error">{{ clientError }}</p>
        <ProButton type="submit" test-id="create-client-submit" :disabled="clientSaving">{{ $t('clients.create.submit') }}</ProButton>
      </form>
    </ProCard>
    <ProCard class="pro-mb-lg" data-testid="vet-referral-form">
      <h3 class="pro-mb-md">{{ $t('clients.referral.title') }}</h3>
      <p class="pro-hint pro-mb-md">{{ $t('clients.referral.hint') }}</p>
      <form class="pro-form" @submit.prevent="submitReferral">
        <ProInput v-model="referral.practiceName" test-id="referral-practice" :label="$t('clients.referral.practiceName')" required />
        <ProInput v-model="referral.contactName" test-id="referral-contact" :label="$t('clients.referral.contactName')" />
        <ProInput v-model="referral.contactEmail" test-id="referral-email" type="email" :label="$t('clients.referral.contactEmail')" />
        <ProInput v-model="referral.city" test-id="referral-city" :label="$t('clients.referral.city')" />
        <ProInput v-model="referral.notes" test-id="referral-notes" :label="$t('clients.referral.notes')" />
        <p v-if="referralMsg" class="pro-hint" data-testid="referral-msg">{{ referralMsg }}</p>
        <p v-if="referralError" class="pro-error">{{ referralError }}</p>
        <ProButton type="submit" test-id="referral-submit" :disabled="referralSaving">{{ $t('clients.referral.submit') }}</ProButton>
      </form>
    </ProCard>

    <ProCard>
      <ProListToolbar v-model:view-mode="viewMode">
        <template #filters>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="client-search">{{ $t('clients.search') }}</label>
            <input
              id="client-search"
              v-model="query"
              type="search"
              class="pro-input"
              :placeholder="$t('clients.searchPlaceholder')"
            />
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="pet-filter">{{ $t('clients.petsFilter') }}</label>
            <select id="pet-filter" v-model="petFilter" class="pro-select">
              <option value="all">{{ $t('clients.petsAll') }}</option>
              <option value="none">{{ $t('clients.petsNone') }}</option>
              <option value="with">{{ $t('clients.petsWith') }}</option>
            </select>
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="sort-by">{{ $t('clients.sort') }}</label>
            <select id="sort-by" v-model="sortBy" class="pro-select">
              <option value="name">{{ $t('clients.sortName') }}</option>
              <option value="pets">{{ $t('clients.sortPets') }}</option>
            </select>
          </div>
        </template>
      </ProListToolbar>

      <ProTable
        v-if="viewMode === 'table'"
        :empty="!filtered.length"
        :empty-title="$t('clients.emptyTitle')"
        :empty-description="$t('clients.emptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('clients.columnClient') }}</th>
            <th>{{ $t('clients.columnEmail') }}</th>
            <th>{{ $t('clients.columnPets') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in filtered" :key="c.userId">
            <td>
              <ProAvatar :src="c.avatarUrl" :name="c.fullName" class="client-avatar" />
              {{ c.fullName }}
            </td>
            <td>{{ c.email }}</td>
            <td>{{ c.petCount }}</td>
            <td>
              <div class="pro-client-actions">
                <NuxtLink :to="`/clients/${c.userId}`">{{ $t('common.profile') }}</NuxtLink>
                <button
                  type="button"
                  class="pro-link-btn"
                  :disabled="sendingId === c.userId"
                  :data-testid="`send-app-link-${c.userId}`"
                  @click="sendAppLink(c)"
                >
                  {{ sendingId === c.userId ? $t('clients.sendingAppLink') : $t('clients.sendAppLink') }}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>

      <ProKanban v-else>
        <ProKanbanColumn
          v-for="col in kanbanColumns"
          :key="col.key"
          :title="col.title"
          :count="col.items.length"
          :empty="!col.items.length"
          :empty-title="$t('common.empty')"
        >
          <NuxtLink
            v-for="c in col.items"
            :key="c.userId"
            :to="`/clients/${c.userId}`"
            class="pro-kanban-card"
          >
            <ProAvatar :src="c.avatarUrl" :name="c.fullName" class="client-avatar" />
            <strong>{{ c.fullName }}</strong>
            <p class="pro-kanban-card__meta">{{ c.email }}</p>
            <ProBadge variant="neutral">{{ c.petCount }} {{ petLabel(c.petCount) }}</ProBadge>
            <button
              type="button"
              class="pro-link-btn pro-kanban-card__action"
              :disabled="sendingId === c.userId"
              @click.prevent="sendAppLink(c)"
            >
              {{ sendingId === c.userId ? $t('clients.sendingAppLink') : $t('clients.sendAppLink') }}
            </button>
          </NuxtLink>
        </ProKanbanColumn>
      </ProKanban>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

type ClientRow = {
  userId: string
  email: string
  fullName: string
  avatarUrl?: string
  petCount: number
}

const { t } = useI18n()
const { compareStrings } = useFormatters()

const clients = ref<ClientRow[]>([])
const query = ref('')
const petFilter = ref<'all' | 'none' | 'with'>('all')
const sortBy = ref<'name' | 'pets'>('name')
const sendingId = ref('')
const appLinkFeedback = ref('')
const { viewMode } = useListView('pf-clients-view', 'table')

const referral = reactive({
  practiceName: '',
  contactName: '',
  contactEmail: '',
  city: '',
  notes: '',
})
const referralSaving = ref(false)
const referralMsg = ref('')
const referralError = ref('')

const clientForm = reactive({ fullName: '', email: '', password: '' })
const clientSaving = ref(false)
const clientMsg = ref('')
const clientError = ref('')

async function createClient() {
  clientSaving.value = true
  clientMsg.value = ''
  clientError.value = ''
  try {
    await $fetch('/api/vet/clients', { method: 'POST', body: { ...clientForm } })
    clientMsg.value = t('clients.create.success')
    Object.assign(clientForm, { fullName: '', email: '', password: '' })
    const res: any = await $fetch('/api/clients')
    clients.value = res.data ?? res ?? []
  } catch {
    clientError.value = t('clients.create.error')
  } finally {
    clientSaving.value = false
  }
}

async function submitReferral() {
  referralSaving.value = true
  referralMsg.value = ''
  referralError.value = ''
  try {
    await $fetch('/api/vet/prospects', { method: 'POST', body: { ...referral } })
    referralMsg.value = t('clients.referral.success')
    Object.assign(referral, { practiceName: '', contactName: '', contactEmail: '', city: '', notes: '' })
  } catch {
    referralError.value = t('clients.referral.error')
  } finally {
    referralSaving.value = false
  }
}

function petLabel(count: number) {
  return count > 1 ? t('common.pets') : t('common.pet')
}

async function sendAppLink(c: ClientRow) {
  if (sendingId.value) return
  sendingId.value = c.userId
  appLinkFeedback.value = ''
  try {
    const res: any = await $fetch(`/api/clients/${c.userId}/send-app-link`, { method: 'POST' })
    const data = res.data ?? res
    appLinkFeedback.value = data.message || t('clients.detail.sendAppLinkSuccess', { email: c.email })
  } catch {
    appLinkFeedback.value = t('clients.detail.sendAppLinkError')
  } finally {
    sendingId.value = ''
  }
}

const filtered = computed(() => {
  let list = [...clients.value]
  const q = query.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (c) =>
        c.fullName?.toLowerCase().includes(q) ||
        c.email?.toLowerCase().includes(q),
    )
  }
  if (petFilter.value === 'none') list = list.filter((c) => c.petCount === 0)
  if (petFilter.value === 'with') list = list.filter((c) => c.petCount > 0)
  if (sortBy.value === 'name') {
    list.sort((a, b) => compareStrings(a.fullName, b.fullName))
  } else {
    list.sort((a, b) => b.petCount - a.petCount)
  }
  return list
})

const kanbanColumns = computed(() => [
  {
    key: 'none',
    title: t('clients.kanbanNone'),
    items: filtered.value.filter((c) => c.petCount === 0),
  },
  {
    key: 'one',
    title: t('clients.kanbanOne'),
    items: filtered.value.filter((c) => c.petCount === 1),
  },
  {
    key: 'multi',
    title: t('clients.kanbanMulti'),
    items: filtered.value.filter((c) => c.petCount > 1),
  },
])

onMounted(async () => {
  const res: any = await $fetch('/api/clients')
  clients.value = res.data ?? res ?? []
})
</script>

<style scoped>
.client-avatar {
  margin-right: 0.5rem;
  vertical-align: middle;
}
.pro-client-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}
.pro-link-btn {
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--pf-vet-accent);
  font: inherit;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
}
.pro-link-btn:disabled {
  opacity: 0.6;
  cursor: wait;
}
.pro-kanban-card__action {
  display: block;
  margin-top: 0.5rem;
  text-align: left;
}
.pro-inline-feedback {
  margin: 0 0 1rem;
  padding: 0.75rem 1rem;
  border-radius: var(--pf-vet-radius);
  background: color-mix(in srgb, var(--pf-vet-accent) 10%, var(--pf-vet-surface));
  border: 1px solid color-mix(in srgb, var(--pf-vet-accent) 30%, transparent);
}
</style>
