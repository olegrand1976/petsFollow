export type ContactPhoneProfile = {
  role?: string
  contactPhone?: string | null
  mustChangePassword?: boolean | null
}

/** Commercial / manager must set a phone (unless forced password change first). */
export function needsContactPhone(me: ContactPhoneProfile): boolean {
  if (me.mustChangePassword === true) return false
  if (me.role !== 'commercial' && me.role !== 'commercial_manager') return false
  return !String(me.contactPhone || '').trim()
}
