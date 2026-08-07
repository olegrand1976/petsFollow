import { beforeEach, describe, expect, it, vi } from 'vitest'

/**
 * Changer son mot de passe incrémente token_version côté API : toutes les
 * sessions du compte tombent au prochain refresh. L'API réémet donc une paire
 * pour l'appelant, que la BFF doit reposer en cookies httpOnly — sinon l'onglet
 * qui vient de changer le mot de passe se déconnecte tout seul.
 */

const setCookieCalls: Array<{ name: string, value: string, options: any }> = []
const deletedCookies: string[] = []

vi.stubGlobal('setCookie', (_event: unknown, name: string, value: string, options: any) => {
  setCookieCalls.push({ name, value, options })
})
vi.stubGlobal('deleteCookie', (_event: unknown, name: string) => {
  deletedCookies.push(name)
})

const { absorbAuthTokens } = await import('../../server/utils/api')

describe('absorbAuthTokens sur le changement de mot de passe', () => {
  beforeEach(() => {
    setCookieCalls.length = 0
  })

  it('repose la paire réémise en cookies httpOnly', () => {
    absorbAuthTokens({} as any, {
      data: { ok: true, accessToken: 'new-access', refreshToken: 'new-refresh' },
    })

    const token = setCookieCalls.find(c => c.name === 'pf_token')
    const refresh = setCookieCalls.find(c => c.name === 'pf_refresh')
    expect(token?.value).toBe('new-access')
    expect(refresh?.value).toBe('new-refresh')
    expect(token?.options.httpOnly).toBe(true)
    expect(refresh?.options.httpOnly).toBe(true)
  })

  it('ne renvoie jamais les tokens au JS client', () => {
    const out = absorbAuthTokens({} as any, {
      data: { ok: true, accessToken: 'new-access', refreshToken: 'new-refresh' },
    }) as { data: Record<string, unknown> }

    expect(out.data.accessToken).toBeUndefined()
    expect(out.data.refreshToken).toBeUndefined()
    expect(out.data.ok).toBe(true)
  })

  it('laisse passer une réponse sans token sans poser de cookie', () => {
    const out = absorbAuthTokens({} as any, { data: { ok: true } }) as { data: Record<string, unknown> }

    expect(setCookieCalls).toHaveLength(0)
    expect(out.data.ok).toBe(true)
  })
})

/**
 * Si l'API n'a pas pu réémettre, le mot de passe a quand même changé et la
 * session est révoquée : la BFF purge les cookies pour renvoyer au login tout
 * de suite, au lieu d'une déconnexion différée sans rapport apparent.
 */
describe('clearAuthCookies sur reauthRequired', () => {
  it('purge le token, le refresh et le marqueur de session', async () => {
    const { clearAuthCookies } = await import('../../server/utils/api')
    deletedCookies.length = 0

    clearAuthCookies({ node: { req: {} } } as any)

    expect(deletedCookies).toEqual(expect.arrayContaining(['pf_token', 'pf_refresh', 'pf_session']))
  })
})
