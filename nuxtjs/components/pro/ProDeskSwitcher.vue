<template>
  <div
    v-if="members.length"
    class="pro-desk-switcher"
    role="group"
    :aria-label="$t('desk.ariaSwitcher')"
    data-testid="pro-desk-switcher"
  >
    <button
      v-for="m in members"
      :key="m.email"
      type="button"
      class="pro-desk-switcher__btn"
      :class="{ 'pro-desk-switcher__btn--active': isCurrent(m.email) }"
      :title="m.fullName"
      :aria-label="m.fullName"
      :aria-current="isCurrent(m.email) ? 'true' : undefined"
      :data-testid="`pro-desk-user-${m.email}`"
      @click="onPick(m.email)"
    >
      <ProAvatar :name="m.fullName" size="sm" />
      <span v-if="isCurrent(m.email)" class="pro-desk-switcher__dot" :title="$t('desk.current')" />
    </button>
  </div>
</template>

<script setup lang="ts">
const { user } = useProUser()
const desk = useDeskSession()

const members = computed(() => desk.roster.value)

function isCurrent(email: string) {
  return !!user.value?.email && user.value.email.toLowerCase() === email.toLowerCase()
}

function onPick(email: string) {
  if (isCurrent(email)) return
  void desk.openSwitch(email)
}
</script>

<style scoped>
.pro-desk-switcher {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-right: 0.35rem;
  max-width: min(40vw, 18rem);
  overflow-x: auto;
  padding: 0.15rem 0;
}

.pro-desk-switcher__btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.15rem;
  border: 2px solid transparent;
  border-radius: 999px;
  background: transparent;
  cursor: pointer;
  flex-shrink: 0;
}

.pro-desk-switcher__btn:hover {
  border-color: var(--pf-vet-border);
}

.pro-desk-switcher__btn--active {
  border-color: var(--pf-vet-accent);
}

.pro-desk-switcher__dot {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 50%;
  background: var(--pf-vet-accent);
  border: 2px solid var(--pf-vet-surface);
}
</style>
