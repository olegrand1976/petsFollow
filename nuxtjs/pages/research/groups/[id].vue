<template>
  <div data-testid="research-group-detail-page">
    <ProPageHeader
      :title="group?.name || $t('research.groups.title')"
      :subtitle="group?.description || $t('research.groups.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
        <NuxtLink to="/research/groups" class="pro-link">{{ $t('research.groups.back') }}</NuxtLink>
      </template>
    </ProPageHeader>

    <p v-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>
    <ProEmptyState v-else-if="loadError" :title="$t('research.loadError')" />
    <template v-else-if="group">
      <ProCard v-if="group.memberRole === 'owner'" :title="$t('research.groups.addMember')" class="pro-mb-md">
        <div class="pro-filters">
          <ProInput v-model="memberEmail" :label="$t('research.groups.memberEmail')" test-id="research-member-email" />
          <ProButton :loading="adding" test-id="research-member-add" @click="addMember">
            {{ $t('research.groups.add') }}
          </ProButton>
        </div>
        <p v-if="memberError" class="pro-error" role="alert">{{ memberError }}</p>
      </ProCard>

      <ProTable :empty="!members.length" :empty-title="$t('research.groups.noMembers')" data-testid="research-members-table">
        <thead>
          <tr>
            <th>{{ $t('research.groups.colName') }}</th>
            <th>{{ $t('research.groups.colEmail') }}</th>
            <th>{{ $t('research.groups.role') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in members" :key="m.userId">
            <td>{{ m.fullName || '—' }}</td>
            <td>{{ m.email }}</td>
            <td>{{ m.memberRole }}</td>
            <td>
              <ProButton
                v-if="group.memberRole === 'owner' && m.memberRole !== 'owner'"
                variant="ghost"
                @click="removeMember(m.userId)"
              >
                {{ $t('common.delete') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'research',
  middleware: ['research-only'],
})

const { t } = useI18n()
const route = useRoute()
const groupID = computed(() => String(route.params.id || ''))

type Group = {
  id: string
  name: string
  description: string
  memberRole: string
}
type Member = {
  userId: string
  email: string
  fullName: string
  memberRole: string
}

const group = ref<Group | null>(null)
const members = ref<Member[]>([])
const loading = ref(false)
const loadError = ref(false)
const memberEmail = ref('')
const adding = ref(false)
const memberError = ref('')

async function load() {
  loading.value = true
  loadError.value = false
  try {
    const [gRes, mRes]: any[] = await Promise.all([
      $fetch(`/api/research/groups/${groupID.value}`),
      $fetch(`/api/research/groups/${groupID.value}/members`),
    ])
    group.value = (gRes?.data ?? gRes) as Group
    const md = mRes?.data ?? mRes
    members.value = Array.isArray(md?.items) ? md.items : []
  } catch {
    loadError.value = true
    group.value = null
    members.value = []
  } finally {
    loading.value = false
  }
}

async function addMember() {
  adding.value = true
  memberError.value = ''
  try {
    await $fetch(`/api/research/groups/${groupID.value}/members`, {
      method: 'POST',
      body: { email: memberEmail.value },
    })
    // Uniform API success (anti-enumeration) — refresh shows whether the member was added.
    memberEmail.value = ''
    await load()
  } catch {
    memberError.value = t('research.groups.memberError')
  } finally {
    adding.value = false
  }
}

async function removeMember(userId: string) {
  try {
    await $fetch(`/api/research/groups/${groupID.value}/members/${userId}`, { method: 'DELETE' })
    await load()
  } catch {
    memberError.value = t('research.groups.memberError')
  }
}

onMounted(() => {
  void load()
})
</script>
