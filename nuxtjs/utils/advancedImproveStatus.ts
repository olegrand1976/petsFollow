/** État du run côté orchestrateur (SSE `status`) — codes stables traduits par l'UI. */
export type AdvancedImproveState = 'crew_warming' | 'crew_ready' | 'running'

export type AdvancedImproveControlStep = {
  state?: AdvancedImproveState | string
}

/** Steps de contrôle warm-up / ready / running — affichés via la ligne d'état, pas la liste agents. */
export function isAdvancedImproveControlStep(step: AdvancedImproveControlStep): boolean {
  return step.state === 'crew_warming' || step.state === 'crew_ready' || step.state === 'running'
}

export function advancedImproveStateFromStep(
  step: AdvancedImproveControlStep,
): AdvancedImproveState | null {
  if (step.state === 'crew_warming' || step.state === 'crew_ready' || step.state === 'running') {
    return step.state
  }
  return null
}
