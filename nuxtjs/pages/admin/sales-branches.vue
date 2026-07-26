<template>
  <div data-testid="admin-sales-branches-page">
    <ProPageHeader
      :title="$t('admin.branches.title')"
      :subtitle="$t('admin.branches.subtitle')"
    />

    <ProCard class="pro-mb-lg" data-testid="admin-branch-create">
      <h3 class="pro-mb-md">{{ $t('admin.branches.createTitle') }}</h3>
      <form class="pro-form" @submit.prevent="createBranch">
        <ProInput v-model="form.name" test-id="admin-branch-name" :label="$t('admin.branches.name')" required />
        <ProInput v-model="form.code" test-id="admin-branch-code" :label="$t('admin.branches.code')" required />
        <p v-if="createMsg" class="pro-hint">{{ createMsg }}</p>
        <p v-if="createError" class="pro-error">{{ createError }}</p>
        <ProButton type="submit" test-id="admin-branch-submit" :disabled="creating">
          {{ $t('admin.branches.createSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard class="pro-mb-lg" data-testid="admin-branch-assign">
      <h3 class="pro-mb-md">{{ $t('admin.branches.assignTitle') }}</h3>
      <form class="pro-form" @submit.prevent="assignBranch">
        <div class="pro-field">
          <label class="pro-label" for="assign-commercial">{{ $t('admin.branches.assignCommercial') }}</label>
          <select
            id="assign-commercial"
            v-model="assign.commercialId"
            class="pro-select"
            required
            data-testid="admin-branch-commercial"
          >
            <option value="" disabled>{{ $t('admin.branches.assignCommercialPlaceholder') }}</option>
            <option v-for="c in commercials" :key="c.userId" :value="c.userId">
              {{ c.fullName }} ({{ c.email }})
            </option>
          </select>
        </div>
        <div class="pro-field">
          <label class="pro-label" for="assign-branch">{{ $t('admin.branches.assignBranch') }}</label>
          <select
            id="assign-branch"
            v-model="assign.branchId"
            class="pro-select"
            data-testid="admin-branch-select"
          >
            <option value="">{{ $t('admin.branches.assignBranchNone') }}</option>
            <option v-for="b in branches" :key="b.id" :value="b.id">
              {{ b.name }} ({{ b.code }})
            </option>
          </select>
        </div>
        <p v-if="assignMsg" class="pro-hint">{{ assignMsg }}</p>
        <p v-if="assignError" class="pro-error">{{ assignError }}</p>
        <ProButton type="submit" test-id="admin-branch-assign-submit" :disabled="assigning || !assign.commercialId">
          {{ $t('admin.branches.assignSubmit') }}
        </ProButton>
      </form>
    </ProCard>

    <ProCard>
      <ProEmptyState v-if="loadError" :title="$t('admin.branches.loadError')" />
      <ProTable
        v-else
        :empty="!branches.length"
        :empty-title="$t('admin.branches.empty')"
      >
        <thead>
          <tr>
            <th>{{ $t('admin.branches.colName') }}</th>
            <th>{{ $t('admin.branches.colCode') }}</th>
            <th>{{ $t('admin.branches.colMembers') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in branches" :key="b.id" :data-testid="`admin-branch-row-${b.id}`">
            <td>{{ b.name }}</td>
            <td>{{ b.code }}</td>
            <td>{{ b.memberCount ?? 0 }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()

const branches = ref<any[]>([])
const commercials = ref<any[]>([])
const loadError = ref(false)
const form = reactive({ name: '', code: '' })
const creating = ref(false)
const createMsg = ref('')
const createError = ref('')
const assign = reactive({ commercialId: '', branchId: '' })
const assigning = ref(false)
const assignMsg = ref('')
const assignError = ref('')

async function load() {
  loadError.value = false
  try {
    const [bRes, cRes]: any[] = await Promise.all([
      $fetch('/api/admin/sales-branches'),
      $fetch('/api/admin/commercials'),
    ])
    branches.value = bRes.data ?? bRes ?? []
    commercials.value = cRes.data ?? cRes ?? []
  } catch {
    loadError.value = true
  }
}

async function createBranch() {
  creating.value = true
  createMsg.value = ''
  createError.value = ''
  try {
    await $fetch('/api/admin/sales-branches', {
      method: 'POST',
      body: { name: form.name, code: form.code },
    })
    createMsg.value = t('admin.branches.createSuccess')
    form.name = ''
    form.code = ''
    await load()
  } catch {
    createError.value = t('admin.branches.createFailed')
  } finally {
    creating.value = false
  }
}

async function assignBranch() {
  assigning.value = true
  assignMsg.value = ''
  assignError.value = ''
  try {
    await $fetch(`/api/admin/commercials/${assign.commercialId}/branch`, {
      method: 'PATCH',
      body: { branchId: assign.branchId },
    })
    assignMsg.value = t('admin.branches.assignSuccess')
    await load()
  } catch {
    assignError.value = t('admin.branches.assignFailed')
  } finally {
    assigning.value = false
  }
}

onMounted(load)
</script>
