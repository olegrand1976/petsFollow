import { apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const body = await readBody(event)
  return $fetch(`${config.apiBase}/api/v1/commercial/vets`, {
    method: 'POST',
    headers: apiHeaders(event),
    body,
  })
})
