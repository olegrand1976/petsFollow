import { absorbAuthTokens, clearAuthCookies, proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const res = await proxyApi<{ data?: { reauthRequired?: boolean } }>(
    event,
    '/api/v1/me/password',
    { method: 'PATCH', body },
  )

  // L'API n'a pas pu réémettre : le changement a bien eu lieu mais la session
  // est révoquée. Purger tout de suite vaut mieux que laisser l'utilisateur
  // tomber au prochain refresh, sans rapport apparent avec ce qu'il a fait.
  const payload = res?.data ?? (res as { reauthRequired?: boolean })
  if (payload?.reauthRequired) {
    clearAuthCookies(event)
    return res
  }

  // Sinon on adopte la paire réémise (cookies httpOnly) pour ne pas
  // déconnecter l'onglet qui vient de la demander.
  return absorbAuthTokens(event, res)
})
