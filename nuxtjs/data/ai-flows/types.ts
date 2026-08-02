export type AiFlowProfileId = 'vet' | 'commercial' | 'care_pro'

export type AiFlowDiagram = {
  id: string
  profileId: AiFlowProfileId
  /** Texte Mermaid (nœuds en anglais court / sans accents pour le parser). */
  diagram: string
}

export type AiFlowsCatalog = {
  profiles: Array<{
    id: AiFlowProfileId
    icon: string
  }>
  diagrams: AiFlowDiagram[]
}
