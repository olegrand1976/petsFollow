import type { AiFlowProfileId, AiFlowsCatalog } from './types'

/**
 * Flux IA / automatisations — diagrammes Mermaid.
 * Libellés des nœuds via i18n `aiFlows.diagrams.<id>.nodes.*` (placeholders `{{key}}`).
 * Profil véto en tête (défaut `?profile=vet`).
 */
export const aiFlowsCatalog: AiFlowsCatalog = {
  profiles: [
    { id: 'vet', icon: 'medical_services' },
    { id: 'commercial', icon: 'campaign' },
    { id: 'care_pro', icon: 'handshake' },
  ],
  diagrams: [
    {
      id: 'vet_cr_ai',
      profileId: 'vet',
      diagram: `flowchart TD
  Start["{{start}}"] --> Dictate["{{dictate}}"]
  Dictate --> Transcribe["{{transcribe}}"]
  Transcribe --> Improve["{{improve}}"]
  Improve --> Review["{{review}}"]
  Review --> Finalize["{{finalize}}"]
  Finalize --> Record["{{record}}"]
  Record --> Client["{{client}}"]`,
    },
    {
      id: 'vet_ai_adoption',
      profileId: 'vet',
      diagram: `flowchart TD
  Track["{{track}}"] --> J0["{{j0}}"]
  J0 --> Usage{"{{usage}}"}
  Usage -->|"{{no}}"| Nudge["{{nudge}}"]
  Nudge --> Usage
  Usage -->|"{{yes}}"| Digest["{{digest}}"]
  Digest --> NPS["{{nps}}"]
  NPS --> ROI["{{roi}}"]
  NPS --> Friction["{{friction}}"]`,
    },
    {
      id: 'vet_care_continuity',
      profileId: 'vet',
      diagram: `flowchart TD
  CR["{{cr}}"] --> Msg["{{msg}}"]
  Msg --> Push["{{push}}"]
  Push --> Reply["{{reply}}"]
  Reply --> Thread["{{thread}}"]
  Thread --> Dossier["{{dossier}}"]
  Dossier --> RDV["{{rdv}}"]`,
    },
    {
      id: 'co_encode_activate',
      profileId: 'commercial',
      diagram: `flowchart TD
  Prospect["{{prospect}}"] --> Encode["{{encode}}"]
  Encode --> Onboard["{{onboard}}"]
  Onboard --> TrackAI["{{trackAi}}"]
  TrackAI --> Pets["{{pets}}"]
  Pets --> Ledger["{{ledger}}"]`,
    },
    {
      id: 'care_share_notes',
      profileId: 'care_pro',
      diagram: `flowchart TD
  Login["{{login}}"] --> Agenda["{{agenda}}"]
  Agenda --> Visit["{{visit}}"]
  Visit --> Share["{{share}}"]
  Share --> Vet["{{vet}}"]
  Visit --> Dictate["{{dictate}}"]
  Dictate --> WebAI["{{webAi}}"]`,
    },
  ],
}

const profileIdSet = new Set<AiFlowProfileId>(
  aiFlowsCatalog.profiles.map((p) => p.id),
)

export function parseAiFlowProfileId(raw: unknown): AiFlowProfileId | null {
  const value = Array.isArray(raw) ? raw[0] : raw
  if (typeof value !== 'string' || !value) return null
  return profileIdSet.has(value as AiFlowProfileId) ? (value as AiFlowProfileId) : null
}

export function diagramsForProfile(id: AiFlowProfileId) {
  return aiFlowsCatalog.diagrams.filter((d) => d.profileId === id)
}

/** Remplace `{{key}}` par les libellés i18n (échappement quotes Mermaid). */
export function fillMermaidTemplate(
  template: string,
  labels: Record<string, string>,
): string {
  return template.replace(/\{\{(\w+)\}\}/g, (_m, key: string) => {
    const raw = labels[key] ?? key
    return raw.replace(/"/g, "'")
  })
}
