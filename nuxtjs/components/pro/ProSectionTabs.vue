<template>
  <div
    class="pro-section-tabs"
    role="tablist"
    :aria-label="ariaLabel"
  >
    <ProButton
      v-for="tab in tabs"
      :id="`tab-${tab.id}`"
      :key="tab.id"
      role="tab"
      :variant="modelValue === tab.id ? 'primary' : 'secondary'"
      :aria-selected="modelValue === tab.id"
      :tabindex="modelValue === tab.id ? 0 : -1"
      :test-id="`section-tab-${tab.id}`"
      @click="select(tab.id)"
    >
      {{ tab.label }}
      <span v-if="tab.count != null && tab.count !== ''" class="pro-tab-count">
        ({{ tab.count }})
      </span>
    </ProButton>
  </div>
</template>

<script setup lang="ts">
export type ProSectionTab = {
  id: string
  label: string
  count?: number | string
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    tabs: ProSectionTab[]
    /** When set, sync active tab to `?{queryKey}=` for deep-link. */
    queryKey?: string
    ariaLabel?: string
  }>(),
  {
    queryKey: 'tab',
    ariaLabel: undefined,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const route = useRoute()
const router = useRouter()

const validIds = computed(() => new Set(props.tabs.map(t => t.id)))

function select(id: string) {
  if (!validIds.value.has(id) || id === props.modelValue) return
  emit('update:modelValue', id)
  if (!props.queryKey) return
  const nextQuery = { ...route.query, [props.queryKey]: id }
  router.replace({ query: nextQuery })
}

onMounted(() => {
  if (!props.queryKey) return
  const fromQuery = route.query[props.queryKey]
  const id = typeof fromQuery === 'string' ? fromQuery : ''
  if (id && validIds.value.has(id) && id !== props.modelValue) {
    emit('update:modelValue', id)
  }
})

watch(
  () => route.query[props.queryKey || ''],
  (q) => {
    if (!props.queryKey) return
    const id = typeof q === 'string' ? q : ''
    if (id && validIds.value.has(id) && id !== props.modelValue) {
      emit('update:modelValue', id)
    }
  },
)
</script>

<style scoped>
.pro-section-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0 0 1.25rem;
}
.pro-tab-count {
  margin-left: 0.25rem;
  font-weight: 500;
  opacity: 0.85;
}
</style>
