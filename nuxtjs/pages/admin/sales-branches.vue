<template>
  <div data-testid="admin-sales-branches-page">
    <ProPageHeader
      :title="$t('admin.branches.title')"
      :subtitle="$t('admin.branches.subtitle')"
    >
      <template #actions>
        <ProButton
          variant="secondary"
          test-id="admin-branch-auto-run"
          :disabled="autoRunning || !pendingAuto.length"
          @click="runAuto"
        >
          <ProIcon name="auto_awesome" />
          {{ $t('admin.branches.autoRun') }}
        </ProButton>
        <ProButton
          variant="primary"
          test-id="admin-branch-create-toggle"
          @click="showCreate = !showCreate"
        >
          <ProIcon name="add" />
          {{ $t('admin.branches.createFallback') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p class="pro-hint pro-mb-lg" data-testid="admin-branch-schedule-hint">
      {{ $t('admin.branches.autoHint') }}
    </p>

    <p v-if="autoMsg" class="pro-hint pro-mb-md" data-testid="admin-branch-auto-msg">{{ autoMsg }}</p>
    <p v-if="autoError" class="pro-error pro-mb-md">{{ autoError }}</p>

    <ProCard v-if="showCreate" class="pro-mb-lg" data-testid="admin-branch-create">
      <h3 class="pro-mb-md">{{ $t('admin.branches.createTitle') }}</h3>
      <form class="pro-form pf-branch-create" @submit.prevent="createBranch">
        <ProInput v-model="form.name" test-id="admin-branch-name" :label="$t('admin.branches.name')" required />
        <ProInput v-model="form.code" test-id="admin-branch-code" :label="$t('admin.branches.code')" required />
        <p v-if="createMsg" class="pro-hint">{{ createMsg }}</p>
        <p v-if="createError" class="pro-error">{{ createError }}</p>
        <div class="pf-branch-create__actions">
          <ProButton type="submit" test-id="admin-branch-submit" :disabled="creating">
            {{ $t('admin.branches.createSubmit') }}
          </ProButton>
          <ProButton type="button" variant="ghost" @click="showCreate = false">
            {{ $t('admin.branches.createCancel') }}
          </ProButton>
        </div>
      </form>
    </ProCard>

    <ProCard v-if="pendingAuto.length" class="pro-mb-lg" data-testid="admin-branch-pending">
      <div class="pf-branch-section-head">
        <h3>{{ $t('admin.branches.pendingTitle') }}</h3>
        <ProBadge>{{ pendingAuto.length }}</ProBadge>
      </div>
      <p class="pro-hint pro-mb-md">{{ $t('admin.branches.pendingHint') }}</p>
      <ul class="pf-branch-pending-list">
        <li v-for="c in pendingAuto" :key="c.userId" class="pf-branch-pending-item">
          <div>
            <strong>{{ c.fullName }}</strong>
            <span class="pro-hint">{{ c.email }}</span>
          </div>
          <span class="pf-branch-pending-preview">{{ c.suggestedCode || '—' }}</span>
        </li>
      </ul>
    </ProCard>

    <ProCard data-testid="admin-branch-list">
      <div class="pf-branch-section-head pro-mb-md">
        <h3>{{ $t('admin.branches.listTitle') }}</h3>
        <ProBadge v-if="branches.length">{{ branches.length }}</ProBadge>
      </div>
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
            <th>{{ $t('admin.branches.colAssign') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in branches" :key="b.id" :data-testid="`admin-branch-row-${b.id}`">
            <td>
              <strong>{{ b.name }}</strong>
            </td>
            <td><code class="pf-branch-code">{{ b.code }}</code></td>
            <td>{{ b.memberCount ?? 0 }}</td>
            <td>
              <div class="pf-branch-actions">
                <div class="pf-branch-assign-inline">
                  <select
                    v-model="assignByBranch[b.id]"
                    class="pro-select"
                    :data-testid="`admin-branch-assign-select-${b.id}`"
                  >
                    <option value="">{{ $t('admin.branches.assignCommercialPlaceholder') }}</option>
                    <option v-for="c in commercials" :key="c.userId" :value="c.userId">
                      {{ c.fullName }}{{ c.branchId === b.id ? ' ✓' : '' }}
                    </option>
                  </select>
                  <ProButton
                    variant="secondary"
                    :disabled="!assignByBranch[b.id] || busyBranchId === b.id"
                    :test-id="`admin-branch-assign-btn-${b.id}`"
                    @click="assignToBranch(b.id)"
                  >
                    {{ $t('admin.branches.assignSubmit') }}
                  </ProButton>
                </div>
                <div v-if="membersOf(b.id).length" class="pf-branch-assign-inline">
                  <select
                    v-model="detachByBranch[b.id]"
                    class="pro-select"
                    :data-testid="`admin-branch-detach-select-${b.id}`"
                  >
                    <option value="">{{ $t('admin.branches.detachPlaceholder') }}</option>
                    <option v-for="c in membersOf(b.id)" :key="c.userId" :value="c.userId">
                      {{ c.fullName }}
                    </option>
                  </select>
                  <ProButton
                    variant="ghost"
                    :disabled="!detachByBranch[b.id] || busyBranchId === b.id"
                    :test-id="`admin-branch-detach-btn-${b.id}`"
                    @click="detachFromBranch(b.id)"
                  >
                    {{ $t('admin.branches.detachSubmit') }}
                  </ProButton>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
      <p v-if="assignMsg" class="pro-hint pro-mt-md">{{ assignMsg }}</p>
      <p v-if="assignError" class="pro-error pro-mt-md">{{ assignError }}</p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()

type Branch = { id: string; name: string; code: string; memberCount?: number }
type Commercial = { userId: string; fullName: string; email: string; branchId?: string; sponsorRole?: string }
type Pending = { userId: string; fullName: string; email: string; suggestedCode?: string; suggestedName?: string }

const branches = ref<Branch[]>([])
const commercials = ref<Commercial[]>([])
const pendingAuto = ref<Pending[]>([])
const loadError = ref(false)
const showCreate = ref(false)
const form = reactive({ name: '', code: '' })
const creating = ref(false)
const createMsg = ref('')
const createError = ref('')
const assignByBranch = reactive<Record<string, string>>({})
const detachByBranch = reactive<Record<string, string>>({})
const busyBranchId = ref('')
const assignMsg = ref('')
const assignError = ref('')
const autoRunning = ref(false)
const autoMsg = ref('')
const autoError = ref('')

function membersOf(branchId: string): Commercial[] {
  return commercials.value.filter((c) => c.branchId === branchId)
}

async function load() {
  loadError.value = false
  try {
    const [bRes, cRes]: any[] = await Promise.all([
      $fetch('/api/admin/sales-branches'),
      $fetch('/api/admin/commercials'),
    ])
    const payload = bRes.data ?? bRes ?? {}
    if (Array.isArray(payload)) {
      branches.value = payload
      pendingAuto.value = []
    } else {
      branches.value = payload.branches ?? []
      pendingAuto.value = payload.pendingAuto ?? []
    }
    commercials.value = cRes.data ?? cRes ?? []
    for (const b of branches.value) {
      if (!(b.id in assignByBranch)) assignByBranch[b.id] = ''
      if (!(b.id in detachByBranch)) detachByBranch[b.id] = ''
    }
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
    showCreate.value = false
    await load()
  } catch {
    createError.value = t('admin.branches.createFailed')
  } finally {
    creating.value = false
  }
}

async function assignToBranch(branchId: string) {
  const commercialId = assignByBranch[branchId]
  if (!commercialId) return
  busyBranchId.value = branchId
  assignMsg.value = ''
  assignError.value = ''
  try {
    await $fetch(`/api/admin/commercials/${commercialId}/branch`, {
      method: 'PATCH',
      body: { branchId },
    })
    assignMsg.value = t('admin.branches.assignSuccess')
    assignByBranch[branchId] = ''
    await load()
  } catch {
    assignError.value = t('admin.branches.assignFailed')
  } finally {
    busyBranchId.value = ''
  }
}

async function detachFromBranch(branchId: string) {
  const commercialId = detachByBranch[branchId]
  if (!commercialId) return
  busyBranchId.value = branchId
  assignMsg.value = ''
  assignError.value = ''
  try {
    await $fetch(`/api/admin/commercials/${commercialId}/branch`, {
      method: 'PATCH',
      body: { branchId: '' },
    })
    assignMsg.value = t('admin.branches.detachSuccess')
    detachByBranch[branchId] = ''
    await load()
  } catch {
    assignError.value = t('admin.branches.detachFailed')
  } finally {
    busyBranchId.value = ''
  }
}

async function runAuto() {
  const nPending = pendingAuto.value.length
  if (!nPending) return
  if (!window.confirm(t('admin.branches.autoRunConfirm', { count: nPending }))) return
  autoRunning.value = true
  autoMsg.value = ''
  autoError.value = ''
  try {
    const res: any = await $fetch('/api/admin/sales-branches/auto-run', { method: 'POST' })
    const data = res.data ?? res ?? {}
    const created = Number(data.created ?? 0)
    const skipped = Number(data.skipped ?? 0)
    autoMsg.value = t('admin.branches.autoRunSuccess', { count: created, skipped })
    await load()
  } catch {
    autoError.value = t('admin.branches.autoRunFailed')
  } finally {
    autoRunning.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.pf-branch-section-head {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}
.pf-branch-section-head h3 {
  margin: 0;
  font-size: 1.05rem;
}
.pf-branch-create {
  max-width: 28rem;
}
.pf-branch-create__actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}
.pf-branch-pending-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.pf-branch-pending-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.65rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius-md, 8px);
  background: var(--pf-vet-surface);
}
.pf-branch-pending-item strong {
  display: block;
}
.pf-branch-pending-preview {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.85rem;
  color: var(--pf-vet-accent);
  font-weight: 600;
  letter-spacing: 0.02em;
}
.pf-branch-code {
  font-size: 0.85rem;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
  background: color-mix(in srgb, var(--pf-vet-primary) 8%, transparent);
}
.pf-branch-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 16rem;
}
.pf-branch-assign-inline {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.pf-branch-assign-inline .pro-select {
  flex: 1;
  min-width: 0;
}
@media (max-width: 720px) {
  .pf-branch-pending-item,
  .pf-branch-assign-inline {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
