import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import { pharmacyErrorMessage } from '~/utils/pharmacy-error'
import { pharmacyPriceIsPersisted, type PharmacyWasteReason } from '~/utils/pharmacy-stock'

export type PharmacySupplier = {
  id: string
  name: string
  email: string
}

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

export type PharmacyMovementRow = {
  id: string
  batchId: string
  delta: number
  reason: string
  reasonDetail?: string
  lotNumber?: string
  medicationCnk?: string
  medicationName?: string
  createdAt: string
}

export type PharmacySettingsForm = {
  autoQuarantineExpired: boolean
  expiryDigestEnabled: boolean
  notifyOnAutoQuarantine: boolean
}

export type PharmacyDeposit = {
  id: string
  name: string
  code: string
  isDefault: boolean
}

export type PharmacyDepositForm = {
  name: string
  code: string
  isDefault: boolean
}

export type PharmacyPricingForm = {
  purchasePriceCents: number
  sellPriceCents: number
  vatPercent: number
  minQty: number
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
  const busyReceipt = ref(false)
  const busyOrder = ref(false)
  const busyInventory = ref(false)
  const busyBatchAction = ref(false)
  const busySettings = ref(false)
  const busyPricing = ref(false)
  const error = ref('')
  const softWarn = ref(false)
  const bandFilter = ref('all')
  const depositFilter = ref('')
  const q = ref('')
  const selectedMed = ref<ProComboboxItem | null>(null)
  const receipt = reactive({ lotNumber: '', expiresOn: '', qty: 1, noteNumber: '', supplierName: '', depositId: '' })
  const batches = ref<PharmacyBatchRow[]>([])
  const movements = ref<PharmacyMovementRow[]>([])
  const deposits = ref<PharmacyDeposit[]>([])
  const depositForm = ref<PharmacyDepositForm>({ name: '', code: '', isDefault: false })
  const depositMsg = ref('')
  const busyDeposit = ref(false)
  const reorderAlerts = ref<PharmacyReorderAlert[]>([])
  const suppliers = ref<PharmacySupplier[]>([])
  const orderSupplierId = ref('')
  const orderEmail = ref('')
  const orderSupplierName = ref('')
  const orderMsg = ref('')
  const invSession = ref<PharmacyInvSession | null>(null)
  const invCounts = ref<Record<string, number>>({})
  const invMsg = ref('')
  const settingsMsg = ref('')
  const pricingMsg = ref('')
  const settings = reactive<PharmacySettingsForm>({
    autoQuarantineExpired: true,
    expiryDigestEnabled: true,
    notifyOnAutoQuarantine: true,
  })
  const pricingMed = ref<ProComboboxItem | null>(null)
  const pricingForm = reactive<PharmacyPricingForm>({
    purchasePriceCents: 0,
    sellPriceCents: 0,
    vatPercent: 21,
    minQty: 5,
  })
  const pricingHadExisting = ref(false)
  const pricingLoadToken = ref(0)
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

  function setDepositFilter(id: string) {
    depositFilter.value = id
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
        depositId: depositFilter.value || undefined,
      },
    })
    batches.value = unwrapData<{ items?: PharmacyBatchRow[] }>(res)?.items ?? []
  }

  async function loadDeposits() {
    try {
      const res = await $fetch<any>('/api/vet/pharmacy/deposits')
      deposits.value = unwrapData<{ items?: PharmacyDeposit[] }>(res)?.items ?? []
      const def = deposits.value.find(d => d.isDefault) || deposits.value[0]
      if (def && !receipt.depositId) {
        receipt.depositId = def.id
      }
      else if (receipt.depositId && !deposits.value.some(d => d.id === receipt.depositId)) {
        receipt.depositId = def?.id || ''
      }
    }
    catch (e: any) {
      deposits.value = []
      error.value = pharmacyErr(e, 'pharmacy.stock.errorLoad')
    }
  }

  async function createDeposit() {
    if (!depositForm.value.name.trim() || !depositForm.value.code.trim()) return
    busyDeposit.value = true
    error.value = ''
    depositMsg.value = ''
    try {
      await $fetch('/api/vet/pharmacy/deposits', {
        method: 'POST',
        body: {
          name: depositForm.value.name.trim(),
          code: depositForm.value.code.trim(),
          isDefault: depositForm.value.isDefault,
        },
      })
      depositForm.value = { name: '', code: '', isDefault: false }
      depositMsg.value = t('pharmacy.stock.depositCreated')
      await loadDeposits()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorDeposit')
    }
    finally {
      busyDeposit.value = false
    }
  }

  async function loadMovements() {
    try {
      const res = await $fetch<any>('/api/vet/pharmacy/movements', { query: { limit: '50' } })
      movements.value = unwrapData<{ items?: PharmacyMovementRow[] }>(res)?.items ?? []
    }
    catch (e: any) {
      movements.value = []
      error.value = pharmacyErr(e, 'pharmacy.stock.errorLoad')
    }
  }

  async function loadSettings() {
    try {
      const res = await $fetch<any>('/api/vet/pharmacy/settings')
      const data = unwrapData<Record<string, any>>(res)
      settings.autoQuarantineExpired = data.autoQuarantineExpired !== false
      settings.expiryDigestEnabled = data.expiryDigestEnabled !== false
      settings.notifyOnAutoQuarantine = data.notifyOnAutoQuarantine !== false
    }
    catch {
      /* keep defaults */
    }
  }

  async function saveSettings() {
    busySettings.value = true
    error.value = ''
    settingsMsg.value = ''
    try {
      await $fetch('/api/vet/pharmacy/settings', {
        method: 'PATCH',
        body: {
          autoQuarantineExpired: settings.autoQuarantineExpired,
          expiryDigestEnabled: settings.expiryDigestEnabled,
          notifyOnAutoQuarantine: settings.notifyOnAutoQuarantine,
        },
      })
      settingsMsg.value = t('pharmacy.stock.settingsSaved')
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorSettings')
    }
    finally {
      busySettings.value = false
    }
  }

  async function loadPricingForMed(medicationId: string) {
    const token = ++pricingLoadToken.value
    pricingHadExisting.value = false
    pricingForm.purchasePriceCents = 0
    pricingForm.sellPriceCents = 0
    pricingForm.vatPercent = 21
    pricingForm.minQty = 5
    try {
      const [priceRes, thrRes] = await Promise.all([
        $fetch<any>(`/api/vet/pharmacy/prices/${medicationId}`),
        $fetch<any>('/api/vet/pharmacy/reorder-thresholds', { query: { medicationId } }),
      ])
      if (token !== pricingLoadToken.value) return
      const price = unwrapData<{
        purchasePriceCents?: number
        sellPriceCents?: number
        vatPercent?: number
        updatedAt?: string
      }>(priceRes)
      pricingForm.purchasePriceCents = price.purchasePriceCents ?? 0
      pricingForm.sellPriceCents = price.sellPriceCents ?? 0
      pricingForm.vatPercent = price.vatPercent && price.vatPercent > 0 ? price.vatPercent : 21
      pricingHadExisting.value = pharmacyPriceIsPersisted(price)
      const thr = unwrapData<{ found?: boolean, minQty?: number }>(thrRes)
      if (thr?.found && typeof thr.minQty === 'number') {
        pricingForm.minQty = thr.minQty
      }
    }
    catch (e: any) {
      if (token !== pricingLoadToken.value) return
      error.value = pharmacyErr(e, 'pharmacy.stock.errorPricing')
    }
  }

  async function onPricingMedChange(med: ProComboboxItem | null) {
    pricingMed.value = med
    if (!med?.id) {
      pricingHadExisting.value = false
      return
    }
    await loadPricingForMed(med.id)
  }

  async function savePricing() {
    if (!pricingMed.value?.id) return
    if (pricingForm.purchasePriceCents === 0 && pricingForm.sellPriceCents === 0) {
      if (pricingHadExisting.value) {
        if (!confirm(t('pharmacy.stock.pricingZeroConfirm'))) return
      }
      else if (!confirm(t('pharmacy.stock.pricingZeroNewConfirm'))) {
        return
      }
    }
    busyPricing.value = true
    error.value = ''
    pricingMsg.value = ''
    let priceSaved = false
    try {
      await $fetch(`/api/vet/pharmacy/prices/${pricingMed.value.id}`, {
        method: 'PUT',
        body: {
          purchasePriceCents: pricingForm.purchasePriceCents,
          sellPriceCents: pricingForm.sellPriceCents,
          vatPercent: pricingForm.vatPercent,
        },
      })
      priceSaved = true
      pricingHadExisting.value = true
      await $fetch('/api/vet/pharmacy/reorder-thresholds', {
        method: 'PUT',
        body: {
          medicationId: pricingMed.value.id,
          minQty: pricingForm.minQty,
        },
      })
      pricingMsg.value = t('pharmacy.stock.pricingSaved')
      await loadReorderAlerts()
    }
    catch (e: any) {
      if (priceSaved) {
        error.value = pharmacyErr(e, 'pharmacy.stock.errorPricingThreshold')
        pricingMsg.value = t('pharmacy.stock.pricingPartial')
      }
      else {
        error.value = pharmacyErr(e, 'pharmacy.stock.errorPricing')
      }
    }
    finally {
      busyPricing.value = false
    }
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

  async function loadSuppliers() {
    try {
      const res = await $fetch<any>('/api/vet/pharmacy/suppliers')
      const data = unwrapData<PharmacySupplier[]>(res)
      suppliers.value = Array.isArray(data) ? data : []
      if (orderSupplierId.value && !suppliers.value.some(s => s.id === orderSupplierId.value)) {
        orderSupplierId.value = ''
      }
      if (!orderSupplierId.value && suppliers.value.length === 1) {
        orderSupplierId.value = suppliers.value[0].id
      }
    }
    catch {
      suppliers.value = []
    }
  }

  async function ensureOrderSupplier(): Promise<{ id: string, email: string }> {
    if (orderSupplierId.value) {
      const su = suppliers.value.find(s => s.id === orderSupplierId.value)
      if (!su?.email) throw new Error('supplier')
      return { id: su.id, email: su.email }
    }
    const email = orderEmail.value.trim()
    const name = orderSupplierName.value.trim()
    if (!email || !name) throw new Error('supplier')
    const res = await $fetch<any>('/api/vet/pharmacy/suppliers', {
      method: 'POST',
      body: { name, email },
    })
    const su = unwrapData<PharmacySupplier>(res)
    if (!su?.id || !su.email) throw new Error('supplier')
    await loadSuppliers()
    orderSupplierId.value = su.id
    return { id: su.id, email: su.email }
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
    busyInventory.value = true
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
      busyInventory.value = false
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
    busyInventory.value = true
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
      busyInventory.value = false
    }
  }

  async function cancelInventory() {
    if (!invSession.value?.id) return
    busyInventory.value = true
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
      busyInventory.value = false
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
      await Promise.all([
        loadSummary(),
        loadBatches(),
        loadMovements(),
        loadReorderAlerts(),
        loadSuppliers(),
        loadOpenInventory(),
        loadSettings(),
        loadDeposits(),
      ])
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorLoad')
    }
    finally {
      busy.value = false
    }
  }

  async function createAndSendOrder() {
    busyOrder.value = true
    error.value = ''
    orderMsg.value = ''
    try {
      const supplier = await ensureOrderSupplier()
      const created = await $fetch<any>('/api/vet/pharmacy/orders', {
        method: 'POST',
        body: { fromAlerts: true, supplierId: supplier.id },
      })
      const order = unwrapData<{ id?: string }>(created)
      if (!order?.id) throw new Error('no order')
      await $fetch(`/api/vet/pharmacy/orders/${order.id}/send`, {
        method: 'POST',
        body: { toEmail: supplier.email, supplierId: supplier.id },
      })
      orderMsg.value = t('pharmacy.stock.orderSent')
      orderEmail.value = ''
      orderSupplierName.value = ''
      await loadReorderAlerts()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorOrder')
    }
    finally {
      busyOrder.value = false
    }
  }

  async function receive() {
    if (!canReceive.value || !selectedMed.value) return
    busyReceipt.value = true
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
            depositId: receipt.depositId || undefined,
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
      busyReceipt.value = false
    }
  }

  async function quarantine(id: string) {
    if (!confirm(t('pharmacy.stock.quarantineConfirm'))) return
    busyBatchAction.value = true
    error.value = ''
    try {
      await $fetch(`/api/vet/pharmacy/batches/${id}/quarantine`, { method: 'POST', body: { reason: 'manual' } })
      await refresh()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorAction')
    }
    finally {
      busyBatchAction.value = false
    }
  }

  async function waste(id: string, reason: PharmacyWasteReason = 'expired') {
    busyBatchAction.value = true
    error.value = ''
    try {
      await $fetch(`/api/vet/pharmacy/batches/${id}/waste`, { method: 'POST', body: { reason } })
      await refresh()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorAction')
    }
    finally {
      busyBatchAction.value = false
    }
  }

  async function adjust(id: string, delta: number) {
    busyBatchAction.value = true
    error.value = ''
    try {
      await $fetch(`/api/vet/pharmacy/batches/${id}/adjust`, {
        method: 'POST',
        body: { delta, detail: 'manual' },
      })
      await refresh()
    }
    catch (e: any) {
      error.value = pharmacyErr(e, 'pharmacy.stock.errorAction')
    }
    finally {
      busyBatchAction.value = false
    }
  }

  function exportCsv() {
    window.open('/api/vet/pharmacy/batches/export.csv', '_blank')
  }

  return {
    canWritePharmacy,
    busy,
    busyReceipt,
    busyOrder,
    busyInventory,
    busyBatchAction,
    busySettings,
    busyPricing,
    error,
    softWarn,
    bandFilter,
    depositFilter,
    q,
    selectedMed,
    receipt,
    batches,
    movements,
    deposits,
    depositForm,
    depositMsg,
    busyDeposit,
    reorderAlerts,
    suppliers,
    orderSupplierId,
    orderEmail,
    orderSupplierName,
    orderMsg,
    invSession,
    invCounts,
    invMsg,
    settings,
    settingsMsg,
    pricingMed,
    pricingForm,
    pricingMsg,
    summary,
    bandCards,
    canReceive,
    bandVariant,
    bandLabel,
    setBand,
    setDepositFilter,
    searchMedications,
    loadBatches,
    loadMovements,
    loadDeposits,
    createDeposit,
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
    adjust,
    saveSettings,
    savePricing,
    onPricingMedChange,
    exportCsv,
  }
}
