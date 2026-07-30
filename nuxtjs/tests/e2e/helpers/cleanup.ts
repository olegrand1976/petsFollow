import type { APIRequestContext, Page } from '@playwright/test'

type Deleter = Pick<APIRequestContext, 'delete'>

function asDeleter(pageOrReq: Page | APIRequestContext): Deleter {
  if ('request' in pageOrReq && pageOrReq.request) {
    return pageOrReq.request
  }
  return pageOrReq as APIRequestContext
}

/** Soft-delete walk-in consultation (hidden from /consultations). Best-effort. */
export async function softDeleteVisit(pageOrReq: Page | APIRequestContext, visitId: string) {
  const id = String(visitId || '').trim()
  if (!id) return
  await asDeleter(pageOrReq).delete(`/api/visits/${id}`).catch(() => undefined)
}

export async function softDeleteVisits(
  pageOrReq: Page | APIRequestContext,
  visitIds: Iterable<string>,
) {
  for (const id of visitIds) {
    await softDeleteVisit(pageOrReq, id)
  }
}

/** Hard-delete support ticket (admin/dev session). Best-effort. */
export async function deleteSupportTicket(pageOrReq: Page | APIRequestContext, ticketId: string) {
  const id = String(ticketId || '').trim()
  if (!id) return
  await asDeleter(pageOrReq)
    .delete(`/api/admin/support/tickets/${id}`)
    .catch(() => undefined)
}
