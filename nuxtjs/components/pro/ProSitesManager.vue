<template>
  <ProCard
    class="pro-settings-card"
    data-testid="settings-sites"
  >
    <p class="pro-settings-hint">{{ $t('sites.settingsHint') }}</p>
    <ul class="settings-sites-list" data-testid="settings-sites-list">
      <li
        v-for="site in managedSites"
        :key="site.id"
        class="settings-sites-row"
        :data-testid="`settings-site-row-${site.id}`"
      >
        <div class="settings-sites-row__meta">
          <template v-if="renamingSiteId === site.id">
            <input
              v-model="renamingSiteName"
              type="text"
              class="pro-input"
              :disabled="sitesBusy"
              :data-testid="`settings-site-rename-input-${site.id}`"
              @keydown.enter.prevent="saveSiteRename(site.id)"
            >
          </template>
          <template v-else>
            <strong>{{ site.name }}</strong>
            <span v-if="site.city" class="text-muted">{{ site.city }}</span>
            <ProBadge v-if="site.isPrimary" variant="success">{{ $t('sites.primary') }}</ProBadge>
            <ProBadge v-if="!site.active" variant="warning">{{ $t('sites.inactive') }}</ProBadge>
          </template>
        </div>
        <div class="settings-sites-row__actions">
          <template v-if="renamingSiteId === site.id">
            <ProButton
              variant="secondary"
              type="button"
              :disabled="sitesBusy || !renamingSiteName.trim()"
              :data-testid="`settings-site-rename-save-${site.id}`"
              @click="saveSiteRename(site.id)"
            >
              {{ $t('sites.renameSave') }}
            </ProButton>
            <ProButton
              variant="ghost"
              type="button"
              :disabled="sitesBusy"
              :data-testid="`settings-site-rename-cancel-${site.id}`"
              @click="cancelSiteRename"
            >
              {{ $t('common.cancel') }}
            </ProButton>
          </template>
          <template v-else>
            <ProButton
              v-if="site.active"
              variant="ghost"
              type="button"
              :disabled="sitesBusy"
              :data-testid="`settings-site-rename-${site.id}`"
              @click="startSiteRename(site)"
            >
              {{ $t('sites.rename') }}
            </ProButton>
            <ProButton
              v-if="!site.isPrimary && site.active"
              variant="secondary"
              type="button"
              :disabled="sitesBusy"
              :data-testid="`settings-site-primary-${site.id}`"
              @click="makeSitePrimary(site.id)"
            >
              {{ $t('sites.makePrimary') }}
            </ProButton>
            <ProButton
              v-if="!site.isPrimary && site.active"
              variant="ghost"
              type="button"
              :disabled="sitesBusy"
              :data-testid="`settings-site-deactivate-${site.id}`"
              @click="deactivateSite(site.id)"
            >
              {{ $t('sites.deactivate') }}
            </ProButton>
            <ProButton
              v-if="!site.active"
              variant="secondary"
              type="button"
              :disabled="sitesBusy"
              :data-testid="`settings-site-reactivate-${site.id}`"
              @click="reactivateSite(site.id)"
            >
              {{ $t('sites.reactivate') }}
            </ProButton>
          </template>
        </div>
        <div
          v-if="site.active"
          class="settings-rooms"
          :data-testid="`settings-rooms-${site.id}`"
        >
          <h4 class="pro-settings-subtitle">{{ $t('rooms.title') }}</h4>
          <ul class="settings-rooms-list" :data-testid="`settings-rooms-list-${site.id}`">
            <li
              v-for="room in roomsBySite[site.id] || []"
              :key="room.id"
              class="settings-rooms-row"
              :data-testid="`settings-room-row-${room.id}`"
            >
              <template v-if="renamingRoomId === room.id">
                <input
                  v-model="renamingRoomName"
                  type="text"
                  class="pro-input"
                  :disabled="roomsBusy"
                  :data-testid="`settings-room-rename-input-${room.id}`"
                  @keydown.enter.prevent="saveRoomRename(site.id, room.id)"
                >
                <ProButton
                  variant="secondary"
                  type="button"
                  :disabled="roomsBusy || !renamingRoomName.trim()"
                  :data-testid="`settings-room-rename-save-${room.id}`"
                  @click="saveRoomRename(site.id, room.id)"
                >
                  {{ $t('rooms.renameSave') }}
                </ProButton>
                <ProButton
                  variant="ghost"
                  type="button"
                  :disabled="roomsBusy"
                  :data-testid="`settings-room-rename-cancel-${room.id}`"
                  @click="cancelRoomRename"
                >
                  {{ $t('common.cancel') }}
                </ProButton>
              </template>
              <template v-else>
                <span>
                  {{ room.name }}
                  <ProBadge
                    v-if="!room.active"
                    variant="warning"
                    :data-testid="`settings-room-inactive-${room.id}`"
                  >
                    {{ $t('rooms.inactive') }}
                  </ProBadge>
                </span>
                <div class="settings-rooms-row__actions">
                  <ProButton
                    v-if="room.active"
                    variant="ghost"
                    type="button"
                    :disabled="roomsBusy"
                    :data-testid="`settings-room-rename-${room.id}`"
                    @click="startRoomRename(room)"
                  >
                    {{ $t('rooms.rename') }}
                  </ProButton>
                  <ProButton
                    v-if="room.active"
                    variant="ghost"
                    type="button"
                    :disabled="roomsBusy"
                    :data-testid="`settings-room-deactivate-${room.id}`"
                    @click="deactivateRoom(site.id, room.id)"
                  >
                    {{ $t('rooms.deactivate') }}
                  </ProButton>
                  <ProButton
                    v-else
                    variant="secondary"
                    type="button"
                    :disabled="roomsBusy"
                    :data-testid="`settings-room-reactivate-${room.id}`"
                    @click="reactivateRoom(site.id, room.id)"
                  >
                    {{ $t('rooms.reactivate') }}
                  </ProButton>
                </div>
              </template>
            </li>
          </ul>
          <p
            v-if="!(roomsBySite[site.id] || []).length"
            class="text-muted"
            :data-testid="`settings-rooms-empty-${site.id}`"
          >
            {{ $t('rooms.empty') }}
          </p>
          <div class="sites-form-row settings-rooms-create">
            <input
              v-model="newRoomName[site.id]"
              type="text"
              class="pro-input"
              :placeholder="$t('rooms.name')"
              :data-testid="`settings-room-name-${site.id}`"
            >
            <ProButton
              variant="secondary"
              type="button"
              :loading="roomsBusy"
              :disabled="!(newRoomName[site.id] || '').trim()"
              :data-testid="`settings-room-create-${site.id}`"
              @click="createRoom(site.id)"
            >
              {{ $t('rooms.create') }}
            </ProButton>
          </div>
        </div>
      </li>
    </ul>
    <p v-if="!managedSites.length && !sitesLoading" class="text-muted">{{ $t('sites.empty') }}</p>
    <div class="sites-form-row settings-sites-create">
      <input
        v-model="newSiteName"
        type="text"
        class="pro-input"
        :placeholder="$t('sites.name')"
        data-testid="settings-site-name"
      >
      <input
        v-model="newSiteCity"
        type="text"
        class="pro-input"
        :placeholder="$t('sites.city')"
        data-testid="settings-site-city"
      >
      <ProButton
        variant="secondary"
        type="button"
        :loading="sitesBusy"
        :disabled="!newSiteName.trim()"
        data-testid="settings-site-create"
        @click="createSite"
      >
        {{ $t('sites.create') }}
      </ProButton>
    </div>
    <p v-if="sitesError" class="pro-field-error" role="alert">{{ sitesError }}</p>
    <p v-if="sitesSaved" class="text-muted" role="status">{{ sitesSaved }}</p>
  </ProCard>
</template>

<script setup lang="ts">
const { t } = useI18n()
const { mapError } = useApiError()
const { fetchUser } = useProUser()
const { concreteSiteId } = usePracticeSites()

type ManagedSite = {
  id: string
  name: string
  city: string
  isPrimary: boolean
  active: boolean
}
type ManagedRoom = { id: string; name: string; active: boolean }

const managedSites = ref<ManagedSite[]>([])
const sitesLoading = ref(false)
const sitesBusy = ref(false)
const sitesError = ref('')
const sitesSaved = ref('')
const newSiteName = ref('')
const newSiteCity = ref('')
const renamingSiteId = ref('')
const renamingSiteName = ref('')

const roomsBySite = ref<Record<string, ManagedRoom[]>>({})
const roomsBusy = ref(false)
const newRoomName = reactive<Record<string, string>>({})
const renamingRoomId = ref('')
const renamingRoomName = ref('')

function startSiteRename(site: ManagedSite) {
  renamingSiteId.value = site.id
  renamingSiteName.value = site.name
}

function cancelSiteRename() {
  renamingSiteId.value = ''
  renamingSiteName.value = ''
}

async function saveSiteRename(siteId: string) {
  const name = renamingSiteName.value.trim()
  if (!name) return
  sitesBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}`, {
      method: 'PATCH',
      body: { name },
    })
    sitesSaved.value = t('sites.renamed')
    cancelSiteRename()
    await refreshMeAndSites()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('sites.actionFailed')
  } finally {
    sitesBusy.value = false
  }
}

async function loadManagedSites() {
  sitesLoading.value = true
  sitesError.value = ''
  try {
    const res: any = await $fetch('/api/vet/sites', { query: { includeInactive: '1' } })
    const list = res?.data ?? res
    managedSites.value = (Array.isArray(list) ? list : []).map((s: any) => ({
      id: String(s.id || ''),
      name: String(s.name || ''),
      city: String(s.city || ''),
      isPrimary: !!s.isPrimary,
      active: s.active !== false,
    })).filter((s: ManagedSite) => !!s.id)
    await loadAllRooms()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('sites.loadFailed')
    managedSites.value = []
  } finally {
    sitesLoading.value = false
  }
}

async function loadAllRooms() {
  const active = managedSites.value.filter((s) => s.active)
  const next: Record<string, ManagedRoom[]> = { ...roomsBySite.value }
  await Promise.all(active.map(async (site) => {
    try {
      const res: any = await $fetch(`/api/vet/sites/${encodeURIComponent(site.id)}/rooms`, {
        query: { includeInactive: '1' },
      })
      const list = res?.data ?? res
      next[site.id] = (Array.isArray(list) ? list : [])
        .filter((r: any) => r?.id)
        .map((r: any) => ({
          id: String(r.id),
          name: String(r.name || ''),
          active: r.active !== false,
        }))
    } catch {
      next[site.id] = next[site.id] || []
    }
  }))
  roomsBySite.value = next
}

function startRoomRename(room: ManagedRoom) {
  renamingRoomId.value = room.id
  renamingRoomName.value = room.name
}

function cancelRoomRename() {
  renamingRoomId.value = ''
  renamingRoomName.value = ''
}

async function createRoom(siteId: string) {
  const name = (newRoomName[siteId] || '').trim()
  if (!name) return
  roomsBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}/rooms`, {
      method: 'POST',
      body: { name },
    })
    newRoomName[siteId] = ''
    sitesSaved.value = t('rooms.created')
    await loadAllRooms()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('rooms.actionFailed')
  } finally {
    roomsBusy.value = false
  }
}

async function saveRoomRename(siteId: string, roomId: string) {
  const name = renamingRoomName.value.trim()
  if (!name) return
  roomsBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}/rooms/${encodeURIComponent(roomId)}`, {
      method: 'PATCH',
      body: { name },
    })
    cancelRoomRename()
    sitesSaved.value = t('rooms.renamed')
    await loadAllRooms()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('rooms.actionFailed')
  } finally {
    roomsBusy.value = false
  }
}

async function deactivateRoom(siteId: string, roomId: string) {
  roomsBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}/rooms/${encodeURIComponent(roomId)}/deactivate`, {
      method: 'POST',
    })
    sitesSaved.value = t('rooms.deactivated')
    await loadAllRooms()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('rooms.actionFailed')
  } finally {
    roomsBusy.value = false
  }
}

async function reactivateRoom(siteId: string, roomId: string) {
  roomsBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}/rooms/${encodeURIComponent(roomId)}`, {
      method: 'PATCH',
      body: { active: true },
    })
    sitesSaved.value = t('rooms.reactivated')
    await loadAllRooms()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('rooms.actionFailed')
  } finally {
    roomsBusy.value = false
  }
}

async function refreshMeAndSites() {
  await fetchUser(true)
  await loadManagedSites()
}

async function createSite() {
  const name = newSiteName.value.trim()
  if (!name) return
  sitesBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch('/api/vet/sites', {
      method: 'POST',
      body: {
        name,
        city: newSiteCity.value.trim() || undefined,
        copyScheduleFromSiteId: concreteSiteId.value || undefined,
      },
    })
    newSiteName.value = ''
    newSiteCity.value = ''
    sitesSaved.value = t('sites.created')
    await refreshMeAndSites()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('sites.actionFailed')
  } finally {
    sitesBusy.value = false
  }
}

async function deactivateSite(siteId: string) {
  sitesBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}/deactivate`, { method: 'POST' })
    sitesSaved.value = t('sites.deactivated')
    await refreshMeAndSites()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('sites.actionFailed')
  } finally {
    sitesBusy.value = false
  }
}

async function reactivateSite(siteId: string) {
  sitesBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}`, {
      method: 'PATCH',
      body: { active: true },
    })
    sitesSaved.value = t('sites.reactivated')
    await refreshMeAndSites()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('sites.actionFailed')
  } finally {
    sitesBusy.value = false
  }
}

async function makeSitePrimary(siteId: string) {
  sitesBusy.value = true
  sitesError.value = ''
  sitesSaved.value = ''
  try {
    await $fetch(`/api/vet/sites/${encodeURIComponent(siteId)}`, {
      method: 'PATCH',
      body: { isPrimary: true },
    })
    sitesSaved.value = t('sites.primaryUpdated')
    await refreshMeAndSites()
  } catch (e: any) {
    sitesError.value = mapError(e) || t('sites.actionFailed')
  } finally {
    sitesBusy.value = false
  }
}

onMounted(() => {
  loadManagedSites()
})
</script>

<style scoped>
.pro-settings-card {
  margin-bottom: 1.5rem;
}
.pro-settings-hint {
  color: var(--pf-vet-text-muted);
  font-size: 0.9rem;
  margin-bottom: 1rem;
}
.pro-settings-subtitle {
  margin: 0 0 0.75rem;
  font-size: 1rem;
}
.sites-form-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;
}
.settings-sites-list {
  list-style: none;
  margin: 0 0 1rem;
  padding: 0;
}
.settings-sites-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  justify-content: space-between;
  align-items: flex-start;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.settings-sites-row__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.settings-sites-row__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
.settings-rooms {
  flex: 1 1 100%;
  margin-top: 0.35rem;
  padding: 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius);
  background: var(--pf-vet-bg);
}
.settings-rooms-list {
  list-style: none;
  margin: 0 0 0.5rem;
  padding: 0;
}
.settings-rooms-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  justify-content: space-between;
  padding: 0.35rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.settings-rooms-row:last-child {
  border-bottom: 0;
}
.settings-rooms-row__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
.settings-rooms-create {
  margin-top: 0.5rem;
}
.settings-sites-create {
  margin-top: 0.75rem;
}
</style>
