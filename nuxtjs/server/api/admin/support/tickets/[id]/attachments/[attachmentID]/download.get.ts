import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const attachmentID = getRouterParam(event, 'attachmentID')
  return proxyBinary(
    event,
    `/api/v1/admin/support/tickets/${id}/attachments/${attachmentID}/download`,
  )
})
