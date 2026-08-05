<template>
  <ProConsultationModal
    v-if="clientId"
    :open="open"
    :client-id="clientId"
    :resume-visit-id="resumeVisitId || undefined"
    :resume-pet-id="resumePetId || undefined"
    @update:open="onOpen"
    @closed="onClosed"
  />
</template>

<script setup lang="ts">
import { useActiveConsultation } from '~/composables/useActiveConsultation'

const active = useActiveConsultation()
const { user } = useProUser()
const open = active.open
const clientId = active.clientId
const resumeVisitId = active.resumeVisitId
const resumePetId = active.resumePetId

function onOpen(v: boolean) {
  if (!v) active.close()
  else open.value = true
}

function onClosed() {
  active.close()
}

function onVisibility() {
  if (document.visibilityState !== 'hidden') return
  void active.autosaveOnly()
}

watch(
  () => user.value?.email,
  (email) => {
    if (!email || active.open.value) return
    active.tryResumeForCurrentUser(email)
  },
  { immediate: true },
)

onMounted(() => {
  if (!import.meta.client) return
  document.addEventListener('visibilitychange', onVisibility)
  window.addEventListener('pagehide', onVisibility)
  active.tryResumeForCurrentUser(user.value?.email)
})

onBeforeUnmount(() => {
  if (!import.meta.client) return
  document.removeEventListener('visibilitychange', onVisibility)
  window.removeEventListener('pagehide', onVisibility)
})
</script>
