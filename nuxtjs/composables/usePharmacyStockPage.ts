import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import { pharmacyErrorMessage } from '~/utils/pharmacy-error'

export type PharmacyReorderAlert = {
  medicationId: string
  medicationCnk: string
  medicationName: string
  minQty: number
  onHandQty: number
}

export type PharmacyInvLine = {
  id: string
  medicationName: string
  medicationCnk: string
  lotNumber: string
  systemQty: number
  countedQty?: number
}

export type PharmacyInvSession = {
  id: string
  status: string
  lineCount?: number
  countedCount?: number
  lines?: PharmacyInvLine[]
}

export type PharmacyBatchRow = {
  id: string
  medicationName: string
  medicationCnk: string
  depositCode: string
  lotNumber: string
  expiresOn: string
  qtyOnHand: number
  unit: string
  status: string
  expiryBand: string
}

type MedRow = {
  id: string
  cnk: string
  name: string
  pharmaceuticalForm?: string
  isAntibiotic?: boolean
}

function unwrapData<T>(res: any): T {
  return (res?.data ?? res) as T
}

export function usePharmacyStockPage() {
  const { t } = useI18n()
  const { canPractice } = usePracticePerms()
  const canWritePharmacy = computed(() => canPractice('pharmacy.write'))

  function pharmacyErr(e: any, fallbackKey: string): string {
    return pharmacyErrorMessage(t, e, fallbackKey)
  }

  const busy = ref(false)
  const error = ref('')
  const softWarn = ref(false)
  const bandFilter = ref('all')
  const q = ref('')
  const selectedMed = ref<ProComboboxItem | null>(null)
  const receipt = reactive({ lotNumber: '', expiresOn: '', qty: 1, noteNumber: '', supplierName: '' })
  const batches = ref<PharmacyBatchRow[]>([])
  const reorderAlerts = ref<PharmacyReorderAlert[]>([])
  const orderEmail = ref('')
  const orderMsg = ref('')
  const invSession = ref<PharmacyInvSession | null>(null)
  const invCounts = ref<Record<string, number>>({})
  const invMsg = ref('')
  const summary = ref<Record<string, number>>({
    ok: 0, soon: 0, return: 0, critical: 0, expired: 0, quarantine: 0,
  })

  const bandCards = computed(() => [
    { key: 'all', field: '_all', label: t('pharmacy.stock.bandAll') },
    { key: 'critical', field: 'critical', label: t('pharmacy.stock.bandCritical') },
    { key: 'return', field: 'return', label: t('pharmacy.stock.bandReturn') },
    { key: 'soon', field: 'soon', label: t('pharmacy.stock.bandSoon') },
    { key: 'expired', field: 'expired', label: t('pharmacy.stock.bandExpired') },
    { key: 'quarantine', field: 'quarantine', label: t('pharmacy.stock.bandQuarantine') },
  ])

  const canReceive = computed(() =>
    Boolean(selectedMed.value?.id && receipt.lotNumber && receipt.expiresOn && receipt.qty > 0),
  )

  function bandVariant(band: string): 'success' | 'warning' | 'danger' | 'neutral' {
    if (band === 'ok') return 'success'
    if (band === 'soon' || band === 'return') return 'warning'
    if (band === 'critical' || band === 'expired' || band === 'quarantine') return 'danger'
    return 'neutral'
  }

  function bandLabel(band: string) {
    const map: Record<string, string> = {
      ok: t('pharmacy.stock.bandOk'),
      soon: t('pharmacy.stock.bandSoon'),
      return: t('pharmacy.stock.bandReturn'),
      critical: t('pharmacy.stock.bandCritical'),
      expired: t('pharmacy.stock.bandExpired'),
      quarantine: t('pharmacy.stock.bandQuarantine'),
    }
    return map[band] || band
  }

  function setBand(key: string) {
    bandFilter.value = key
    loadBatches()
  }

  async function searchMedications(query: string): Promise<ProComboboxItem[]> {
    const res = await $fetch<any>('/api/vet/pharmacy/medications/search', { query: { q: query, limit: '20' } })
    const items = unwrapData<{ items?: MedRow[] }>(res)?.items ?? []
    return items.map((m) => ({
      id: m.id,
      label: m.name,
      hint: m.cnk + (m.pharmaceuticalForm ? ` · ${m.pharmaceuticalForm}` : ''),
      badge: m.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
      raw: m,
    }))
  }

  async function loadSummary() {
    const res = await $fetch<any>('/api/vet/pharmacy/expiry/summary')
    const data = unwrapData<Record<string, number>>(res)
    summary.value = {
      ok: data.ok ?? 0,
      soon: data.soon ?? 0,
      return: data.return ?? 0,
      critical: data.critical ?? 0,
      expired: data.expired ?? 0,
      quarantine: data.quarantine ?? 0,
      _all: (data.ok ?? 0) + (data.soon ?? 0) + (data.return ?? 0) + (data.critical ?? 0) + (data.expired ?? 0) + (data.quarantine ?? 0),
    }
  }

  async function loadBatches() {
    const res = await $fetch<any>('/api/vet/pharmacy/batches', {
      query: {
        band: bandFilter.value === 'all' ? undefined : bandFilter.value,
        q: q.value || undefined,
      },
    })
    batches.value = unwrapData<{ items?: PharmacyBatchRow[] }>(res)?.items ?? []
  }

  async function loadReorderAlerts() {
    try {
      const res = await $fetch<any>('/api/vet/pharmacy/reorder-alerts')
      const data = unwrapData<PharmacyReorderAlert[] | { data?: PharmacyReorderAlert[] }>(res)
      reorderAlerts.value = Array.isArray(data) ? data : []
    }
    catch (e: any) {
      reorderAlerts.value = []
      error.value = pharmacyErr(e, 'pharmacy.stock.errorReorderLoad')
    }
  }

  async function loadOpenInventory() {
    try {
      const res = await $fetch<any>('/api/vet/pharmacy/inventory/sessions')
      const list = unwrapData<PharmacyInvSession[]>(res)
      const open = Array.isArray(list) ? list.find(s => s.status === 'open') : null
      if (!open?.id) {
        invSession.value = null
        return
      }
      const detail = await $fetch<any>(`/api/vet/pharmacy/inventory/sessions/${open.id}`)
      invSession.value = unwrapData<PharmacyInvSession>(detail)
      const counts: Record<string, number> = {}
      for (const ln of invSession.value.lines || []) {
        counts[ln.id] = ln.countedQty ?? ln.systemQty
      }
      invCounts.value = counts
    }
    catch (e: any) {
      invSession.value = null
      error.value = pharmacyErr(e, 'pharmacy.stock.errorInvLoad')
    }
  }

  async function startInventory() {
    busy.value = true
    error.value = ''
    invMsg.value = ''
    try {
      const res = await $fetch<any>('/api/vet/pharmacy/inventory/sessions', { method: 'POST', body: {} })
      invSession.value = unwrapData<PharmacyInvSession>(res)
      const counts: Record<string, number> = {}
      for (const ln of invSession.value.lines || []) {
        counts[ln.id] = ln.systemQty
      }
      invCounts.value = counts
      invMsg.value = t('pharmacy.stock.invStarted')
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorInv')
    }
    finally {
      busy.value = false
    }
  }

  async function onInvCountInput(lineId: string, e: Event) {
    const el = e.target as HTMLInputElement | null
    await onInvCount(lineId, el?.value ?? '')
  }

  async function onInvCount(lineId: string, raw: string) {
    const n = Number(raw)
    if (Number.isNaN(n) || n < 0 || !invSession.value?.id) return
    invCounts.value[lineId] = n
    try {
      await $fetch(`/api/vet/pharmacy/inventory/sessions/${invSession.value.id}/lines/${lineId}`, {
        method: 'PATCH',
        body: { countedQty: n },
      })
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorInv')
    }
  }

  async function closeInventory() {
    if (!invSession.value?.id) return
    busy.value = true
    error.value = ''
    try {
      const counts = (invSession.value.lines || []).map(ln => ({
        lineId: ln.id,
        countedQty: invCounts.value[ln.id] ?? ln.countedQty ?? ln.systemQty,
      }))
      const res = await $fetch<any>(`/api/vet/pharmacy/inventory/sessions/${invSession.value.id}/close`, {
        method: 'POST',
        body: { counts },
      })
      invSession.value = unwrapData<PharmacyInvSession>(res)
      invMsg.value = t('pharmacy.stock.invClosed')
      await refresh()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorInv')
    }
    finally {
      busy.value = false
    }
  }

  async function cancelInventory() {
    if (!invSession.value?.id) return
    busy.value = true
    try {
      await $fetch(`/api/vet/pharmacy/inventory/sessions/${invSession.value.id}/cancel`, { method: 'POST' })
      invSession.value = null
      invCounts.value = {}
      invMsg.value = t('pharmacy.stock.invCancelled')
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorInv')
    }
    finally {
      busy.value = false
    }
  }

  async function exportInventory() {
    if (!invSession.value?.id) return
    const csv = await $fetch<string>(`/api/vet/pharmacy/inventory/sessions/${invSession.value.id}/export.csv`, {
      responseType: 'text',
    })
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `inventory-${invSession.value.id.slice(0, 8)}.csv`
    a.click()
    URL.revokeObjectURL(a.href)
  }

  async function refresh() {
    busy.value = true
    error.value = ''
    try {
      await Promise.all([loadSummary(), loadBatches(), loadReorderAlerts(), loadOpenInventory()])
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorLoad')
    }
    finally {
      busy.value = false
    }
  }

  async function createAndSendOrder() {
    if (!orderEmail.value) return
    busy.value = true
    error.value = ''
    orderMsg.value = ''
    try {
      const created = await $fetch<any>('/api/vet/pharmacy/orders', {
        method: 'POST',
        body: { fromAlerts: true },
      })
      const order = unwrapData<{ id?: string }>(created)
      if (!order?.id) throw new Error('no order')
      await $fetch(`/api/vet/pharmacy/orders/${order.id}/send`, {
        method: 'POST',
        body: { toEmail: orderEmail.value },
      })
      orderMsg.value = t('pharmacy.stock.orderSent')
      await loadReorderAlerts()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorOrder')
    }
    finally {
      busy.value = false
    }
  }

  async function receive() {
    if (!canReceive.value || !selectedMed.value) return
    busy.value = true
    error.value = ''
    softWarn.value = false
    try {
      const userBL = receipt.noteNumber.trim()
      const res = await $fetch<any>('/api/vet/pharmacy/delivery-notes', {
        method: 'POST',
        body: {
          noteNumber: userBL,
          supplierName: receipt.supplierName.trim(),
          notify: Boolean(userBL),
          items: [{
            medicationId: selectedMed.value.id,
            lotNumber: receipt.lotNumber,
            expiresOn: receipt.expiresOn,
            qty: receipt.qty,
          }],
        },
      })
      const data = unwrapData<{ softWarnings?: number }>(res)
      softWarn.value = (data.softWarnings ?? 0) > 0
      receipt.lotNumber = ''
      receipt.noteNumber = ''
      receipt.supplierName = ''
      receipt.qty = 1
      await refresh()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorReceive')
    }
    finally {
      busy.value = false
    }
  }

  async function quarantine(id: string) {
    busy.value = true
    error.value = ''
    try {
      await $fetch(`/api/vet/pharmacy/batches/${id}/quarantine`, { method: 'POST', body: { reason: 'manual' } })
      await refresh()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorAction')
    }
    finally {
      busy.value = false
    }
  }

  async function waste(id: string) {
    if (!confirm(t('pharmacy.stock.wasteConfirm'))) return
    busy.value = true
    error.value = ''
    try {
      await $fetch(`/api/vet/pharmacy/batches/${id}/waste`, { method: 'POST', body: { reason: 'expired' } })
      await refresh()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorAction')
    }
    finally {
      busy.value = false
    }
  }

  function exportCsv() {
    window.open('/api/vet/pharmacy/batches/export.csv', '_blank')
  }

  return {
    canWritePharmacy,
    busy,
    error,
    softWarn,
    bandFilter,
    q,
    selectedMed,
    receipt,
    batches,
    reorderAlerts,
    orderEmail,
    orderMsg,
    invSession,
    invCounts,
    invMsg,
    summary,
    bandCards,
    canReceive,
    bandVariant,
    bandLabel,
    setBand,
    searchMedications,
    loadBatches,
    startInventory,
    onInvCountInput,
    closeInventory,
    cancelInventory,
    exportInventory,
    refresh,
    createAndSendOrder,
    receive,
    quarantine,
    waste,
    exportCsv,
  }
}
