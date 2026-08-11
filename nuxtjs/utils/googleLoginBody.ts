/** Body POST /api/auth/google — consent only required for create-if-absent (API). */
export function googleLoginBody(idToken: string, consent: boolean): {
  idToken: string
  consent: boolean
} {
  return { idToken, consent: Boolean(consent) }
}
