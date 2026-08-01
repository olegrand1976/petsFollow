export default defineEventHandler(async (event) => {
  return proxyApi(event, '/api/v1/admin/pacs/prune-orphans', {
    method: 'POST',
    body: {},
    query: getQuery(event),
  })
})
