<template>
  <div data-testid="vet-team-page">
    <ProPageHeader :title="$t('team.title')" :subtitle="$t('team.subtitle')" />

    <ProCard v-if="canManageTeam" class="pro-mb-lg">
      <h3 class="pro-mb-md">{{ $t('team.inviteTitle') }}</h3>
      <form class="pro-form" @submit.prevent="invite">
        <ProInput v-model="form.email" type="email" :label="$t('team.email')" required />
        <ProInput v-model="form.fullName" :label="$t('team.fullName')" required />
        <ProInput v-model="form.password" type="password" :label="$t('team.tempPassword')" required />
        <div class="pro-field">
          <label class="pro-label">{{ $t('team.role') }}</label>
          <select v-model="form.teamRole" class="pro-select">
            <option value="vet">{{ $t('team.roleVet') }}</option>
            <option value="vet_assistant">{{ $t('team.roleAssistant') }}</option>
            <option value="secretary">{{ $t('team.roleSecretary') }}</option>
          </select>
        </div>
        <p v-if="inviteMsg" class="pro-hint">{{ inviteMsg }}</p>
        <ProButton type="submit" :disabled="saving">{{ $t('team.invite') }}</ProButton>
      </form>
    </ProCard>

    <ProCard>
      <ProTable :empty="!members.length" :empty-title="$t('team.empty')">
        <thead>
          <tr>
            <th>{{ $t('team.columnName') }}</th>
            <th>{{ $t('team.columnEmail') }}</th>
            <th>{{ $t('team.columnRole') }}</th>
            <th>{{ $t('team.columnRights') }}</th>
            <th v-if="canManageTeam" />
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in members" :key="m.id">
            <td>{{ m.fullName }}</td>
            <td>{{ m.email }}</td>
            <td>{{ roleLabel(m.teamRole) }}</td>
            <td>
              <div v-if="canManageTeam && m.teamRole !== 'reference_vet'" class="team-perms" data-testid="team-perms">
                <label
                  v-for="key in TEAM_PERM_KEYS"
                  :key="key"
                  class="team-perm"
                  :class="{ 'team-perm--denied': isHardDenied(m.teamRole, key) }"
                  :title="permTip(key)"
                >
                  <input
                    type="checkbox"
                    :checked="!!m.permissions?.[key]"
                    :disabled="isHardDenied(m.teamRole, key)"
                    :aria-describedby="`team-perm-tip-${m.id}-${key}`"
                    @change="togglePerm(m, key, ($event.target as HTMLInputElement).checked)"
                  >
                  <span>{{ permLabel(key) }}</span>
                  <span :id="`team-perm-tip-${m.id}-${key}`" class="visually-hidden">{{ permTip(key) }}</span>
                </label>
              </div>
              <span v-else class="pro-hint">{{ $t('team.defaults') }}</span>
            </td>
            <td v-if="canManageTeam">
              <ProIconAction
                v-if="m.teamRole !== 'reference_vet'"
                icon="person_off"
                variant="danger"
                :label="$t('team.revoke')"
                @click="revoke(m.id)"
              />
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

/** Stable order aligned with go/internal/store/team.go DefaultTeamPermissions. */
const TEAM_PERM_KEYS = [
  'clients.read',
  'clients.write',
  'pets.read',
  'pets.write_clinical',
  'heartrate.validate',
  'messaging',
  'calendar.manage',
  'consultations.history.read',
  'care.manage',
  'shares.read',
  'shares.manage',
  'pharmacy.read',
  'pharmacy.write',
  'practice.settings',
  'team.manage',
  'commissions.view',
] as const

type TeamPermKey = (typeof TEAM_PERM_KEYS)[number]

const HARD_DENIED: Record<string, ReadonlySet<string>> = {
  secretary: new Set([
    'pets.write_clinical',
    'pharmacy.write',
    'heartrate.validate',
    'shares.manage',
    'practice.settings',
    'team.manage',
    'commissions.view',
  ]),
  vet_assistant: new Set([
    'heartrate.validate',
    'shares.manage',
    'practice.settings',
    'team.manage',
    'commissions.view',
  ]),
  vet: new Set(['practice.settings', 'team.manage', 'commissions.view']),
}

type TeamMember = {
  id: string
  fullName: string
  email: string
  teamRole: string
  permissions: Record<string, boolean>
}

const { t, te } = useI18n()
const { fetchUser } = useProUser()
const { canPractice } = usePracticePerms()
const canManageTeam = computed(() => canPractice('team.manage'))
const members = ref<TeamMember[]>([])
const saving = ref(false)
const inviteMsg = ref('')
const form = reactive({
  email: '',
  fullName: '',
  password: '',
  teamRole: 'vet_assistant',
})

function roleLabel(role: string) {
  const map: Record<string, string> = {
    reference_vet: t('team.roleReference'),
    vet: t('team.roleVet'),
    vet_assistant: t('team.roleAssistant'),
    secretary: t('team.roleSecretary'),
  }
  return map[role] ?? role
}

function permI18nPath(key: TeamPermKey, field: 'label' | 'tip') {
  // clients.read → team.perms.clients.read.label
  return `team.perms.${key}.${field}`
}

function permLabel(key: TeamPermKey) {
  const path = permI18nPath(key, 'label')
  return te(path) ? t(path) : key
}

function permTip(key: TeamPermKey) {
  const path = permI18nPath(key, 'tip')
  return te(path) ? t(path) : ''
}

function isHardDenied(role: string, key: string) {
  return HARD_DENIED[role]?.has(key) ?? false
}

async function load() {
  await fetchUser(true)
  const res = await $fetch<{ data?: { members?: TeamMember[]; deskIdleMinutes?: number } | TeamMember[] } | TeamMember[]>('/api/vet/team')
  const payload = Array.isArray(res) ? res : (res.data ?? [])
  members.value = Array.isArray(payload)
    ? payload
    : (Array.isArray(payload.members) ? payload.members : [])
}

async function invite() {
  saving.value = true
  inviteMsg.value = ''
  try {
    await $fetch('/api/vet/team', { method: 'POST', body: { ...form } })
    inviteMsg.value = t('team.inviteOk')
    form.email = ''
    form.fullName = ''
    form.password = ''
    await load()
  } catch {
    inviteMsg.value = t('team.inviteErr')
  } finally {
    saving.value = false
  }
}

async function togglePerm(m: TeamMember, key: string, checked: boolean) {
  if (isHardDenied(m.teamRole, key)) return
  const permissions = { ...m.permissions, [key]: checked }
  await $fetch(`/api/vet/team/${m.id}`, { method: 'PATCH', body: { permissions } })
  await load()
}

async function revoke(id: string) {
  if (!window.confirm(t('team.revokeConfirm'))) return
  await $fetch(`/api/vet/team/${id}`, { method: 'DELETE' })
  await load()
}

onMounted(load)
</script>

<style scoped>
.team-perms {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem 0.75rem;
  max-width: 36rem;
}
.team-perm {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.75rem;
  cursor: help;
}
.team-perm--denied {
  opacity: 0.45;
  cursor: not-allowed;
}
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
