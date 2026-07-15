export function useListView(storageKey: string, defaultView: 'table' | 'kanban' = 'table') {
  const viewMode = ref<'table' | 'kanban'>(defaultView)

  onMounted(() => {
    if (!import.meta.client) return
    const saved = localStorage.getItem(storageKey)
    if (saved === 'table' || saved === 'kanban') viewMode.value = saved
  })

  watch(viewMode, (val) => {
    if (import.meta.client) localStorage.setItem(storageKey, val)
  })

  return { viewMode }
}
