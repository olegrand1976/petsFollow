export type VetLinkRequest = {
  id: string
  clientName?: string
  clientEmail?: string
}

type Options = {
  /** Permission `shares.manage` — sans elle, aucune invitation n'est chargée. */
  canManage: () => boolean
  /** Appelé après une acceptation (recharger la liste clients). */
  onAccepted?: () => unknown
  /** Appelé après acceptation ou refus (badges de navigation). */
  onChanged?: () => unknown
}

/** État + actions des invitations client (link-requests) côté Pro. */
export function useVetLinkRequests(opts: Options) {
  const { mapError } = useApiError()
  const open = ref(false)
  const items = ref<VetLinkRequest[]>([])
  const busyId = ref('')
  const error = ref('')

  async function load() {
    if (!opts.canManage()) {
      items.value = []
      return
    }
    error.value = ''
    try {
      const res: any = await $fetch('/api/vet/link-requests')
      const raw = res?.data ?? res
      items.value = Array.isArray(raw) ? raw : []
    } catch (e: any) {
      items.value = []
      error.value = mapError(e)
    }
  }

  /** Plus aucune invitation à traiter → la modale n'a plus de raison de rester ouverte. */
  function closeIfEmpty() {
    if (!error.value && !items.value.length) open.value = false
  }

  async function act(id: string, action: 'accept' | 'reject') {
    busyId.value = id
    error.value = ''
    try {
      await $fetch(`/api/vet/link-requests/${id}/${action}`, { method: 'POST' })
      const tasks: unknown[] = [load(), opts.onChanged?.()]
      if (action === 'accept') tasks.push(opts.onAccepted?.())
      await Promise.all(tasks)
      closeIfEmpty()
    } catch (e: any) {
      error.value = mapError(e)
    } finally {
      busyId.value = ''
    }
  }

  return {
    open,
    items,
    busyId,
    error,
    load,
    accept: (id: string) => act(id, 'accept'),
    reject: (id: string) => act(id, 'reject'),
  }
}
