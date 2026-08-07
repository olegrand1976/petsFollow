import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return proxyBinary(event, '/api/v1/vet/pharmacy/batches/export.csv', {
    contentDispositionFallback: 'attachment; filename="pharmacy-stock.csv"',
  })
})
