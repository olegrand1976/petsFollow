<template>
  <div data-testid="admin-commercials-page">
    <ProPageHeader :title="$t('admin.commercials.title')" :subtitle="$t('admin.commercials.subtitle')" />

    <ProCard class="pro-mb-lg" data-testid="admin-assign-vet">
      <h3 class="pro-mb-md">{{ $t('admin.commercials.assignTitle') }}</h3>
      <form class="pro-form" @submit.prevent="assignVet">
        <div class="pro-field">
          <label class="pro-label" for="assign-commercial">{{ $t('admin.commercials.assignCommercial') }}</label>
          <select
            id="assign-commercial"
            v-model="assignForm.commercialId"
            class="pro-select"
            data-testid="admin-assign-commercial"
            required
          >
            <option value="" disabled>{{ $t('admin.commercials.assignCommercialPlaceholder') }}</option>
            <option v-for="c in rows" :key="c.userId" :value="c.userId">
              {{ c.fullName }} ({{ c.email }})
            </option>
          </select>
        </div>
        <div class="pro-field">
          <label class="pro-label" for="assign-vet">{{ $t('admin.commercials.assignVet') }}</label>
          <select
            id="assign-vet"
            v-model="assignForm.vetUserId"
            class="pro-select"
            data-testid="admin-assign-vet"
            required
          >
            <option value="" disabled>{{ $t('admin.commercials.assignVetPlaceholder') }}</option>
            <option v-for="v in vets" :key="v.id" :value="v.id">
              {{ v.fullName }} ({{ v.email }})
            </option>
          </select>
        </div>
        <p v-if="assignMsg" class="pro-hint" data-testid="admin-assign-msg">{{ assignMsg }}</p>
        <p v-if="assignError" class="pro-error">{{ assignError }}</p>
        <ProButton type="submit" test-id="admin-assign-submit" :disabled="assignSaving || !assignForm.commercialId || !assignForm.vetUserId">
          {{ $t('admin.commercials.assignSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard>
      <ProTable :empty="!rows.length" :empty-title="$t('admin.commercials.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commercials.columnName') }}</th>
            <th>{{ $t('admin.commercials.columnEmail') }}</th>
            <th>{{ $t('admin.commercials.columnVets') }}</th>
            <th>{{ $t('admin.commercials.columnDue') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in rows" :key="r.userId">
            <td>{{ r.fullName }}</td>
            <td>{{ r.email }}</td>
            <td>{{ r.clientCount }}</td>
            <td>{{ formatDue(r.userId) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t, locale } = useI18n()

const rows = ref<any[]>([])
const vets = ref<any[]>([])
const dueByCommercial = ref<Record<string, number>>({})
const assignForm = reactive({ commercialId: '', vetUserId: '' })
const assignSaving = ref(false)
const assignMsg = ref('')
const assignError = ref('')

function formatDue(commercialId: string) {
  const cents = dueByCommercial.value[commercialId]
  if (cents == null) return '—'
  return new Intl.NumberFormat(locale.value, { style: 'currency', currency: 'EUR' }).format(cents / 100)
}

async function loadCommissions(commercialId: string) {
  try {
    const res: any = await $fetch(`/api/admin/commercials/${commercialId}/commissions`)
    const data = res.data ?? res
    dueByCommercial.value[commercialId] = Number(data.lifetimeEarnedCents ?? 0)
  } catch {
    dueByCommercial.value[commercialId] = 0
  }
}

async function load() {
  const [commercialsRes, usersRes]: any[] = await Promise.all([
    $fetch('/api/admin/commercials'),
    $fetch('/api/admin/users', { query: { role: 'vet' } }),
  ])
  rows.value = commercialsRes.data ?? commercialsRes ?? []
  const users = usersRes.data ?? usersRes ?? []
  vets.value = Array.isArray(users) ? users : []
  await Promise.all(rows.value.map((r: any) => loadCommissions(r.userId)))
}

async function assignVet() {
  assignSaving.value = true
  assignMsg.value = ''
  assignError.value = ''
  try {
    await $fetch(`/api/admin/commercials/${assignForm.commercialId}/assign`, {
      method: 'PATCH',
      body: { vetUserId: assignForm.vetUserId },
    })
    assignMsg.value = t('admin.commercials.assignSuccess')
    assignForm.vetUserId = ''
    await load()
  } catch {
    assignError.value = t('admin.commercials.assignFailed')
  } finally {
    assignSaving.value = false
  }
}

onMounted(load)
</script>
