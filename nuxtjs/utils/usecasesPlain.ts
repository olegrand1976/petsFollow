/** Retire le markdown / jargon d’affichage pour une lecture néophyte. */
export function usecasePlain(text: string | null | undefined): string {
  if (!text) return ''
  return text
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '$1')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    .replace(/\s+/g, ' ')
    .trim()
}

const SURFACE_FRIENDLY: Record<string, string> = {
  'Web VetPro': 'Site web du cabinet',
  'Web Commercial': 'Site web commercial',
  'Web Commercial manager': 'Site web responsable commercial',
  'Web Admin': 'Site web administration',
  'Flutter Client': 'Application mobile du propriétaire',
  'Flutter Pro Light': 'Application mobile terrain',
  'Flutter Pro Light / Client': 'Application mobile (terrain ou propriétaire)',
  'Web VetPro + Flutter Client': 'Site web cabinet + app propriétaire',
  'Flutter Client → Web VetPro': 'App propriétaire puis site web cabinet',
  'Flutter Client + Web VetPro': 'App propriétaire + site web cabinet',
  'Web VetPro → Flutter Pro Light': 'Site web cabinet puis app terrain',
  'Web Commercial (+ Client / VetPro si activation réelle)': 'Site web commercial (parfois aussi app / cabinet)',
}

export function usecaseFriendlySurface(surface: string | null | undefined): string {
  if (!surface) return ''
  const key = surface.trim()
  return SURFACE_FRIENDLY[key] || usecasePlain(key)
}

export function usecaseFriendlyPriority(priority: string | null | undefined): string {
  const p = (priority || '').toLowerCase()
  if (p.includes('démo') || p.includes('demo')) return 'demo'
  if (p.includes('important')) return 'important'
  if (p.includes('secondaire')) return 'secondary'
  return 'other'
}
