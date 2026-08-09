/**
 * Garde-fou sur le liage props → flow de `ProConsultationWorkspace`.
 *
 * Sur une reprise agenda, la modale résout la date du RDV et l'animal par des
 * requêtes qui atterrissent après la première peinture. Tant qu'un seul watcher
 * couvrait identité **et** métadonnées, cette réponse tardive rejouait la remise
 * à zéro d'écran et refermait la confirmation de fermeture pendant que le véto
 * la lisait (e2e `03b-consultation` « visite conservée au close », instable
 * uniquement sur staging où le réseau est réel).
 *
 * Le composant n'est pas montable ici (Vitest en environnement `node`, sans DOM
 * ni auto-imports Nuxt) : on verrouille donc la forme du source, comme
 * `proAccordionSection.spec.ts`.
 */
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const src = readFileSync(
  join(dirname(fileURLToPath(import.meta.url)), '../../components/pro/ProConsultationWorkspace.vue'),
  'utf8',
)

/** Corps d'une fonction de premier niveau du `<script setup>`. */
function functionBody(name: string): string {
  const start = src.indexOf(`function ${name}() {`)
  expect(start, `function ${name}() introuvable`).toBeGreaterThan(-1)
  const end = src.indexOf('\n}', start)
  expect(end, `fin de ${name}() introuvable`).toBeGreaterThan(start)
  return src.slice(start, end)
}

describe('ProConsultationWorkspace — liage props', () => {
  it('ne remet l\'écran à zéro que sur un changement d\'identité', () => {
    expect(src).toContain('() => [props.visitId, props.clientId] as const')
    // La liste surveillée pour le reset ne doit pas réabsorber les métadonnées.
    expect(src).not.toMatch(/\[props\.visitId,[^\]]*props\.scheduledAt/)
    expect(src).not.toMatch(/\[props\.visitId,[^\]]*props\.petId/)
  })

  it('reflète les métadonnées tardives par un watcher séparé', () => {
    expect(src).toContain('() => [props.petId, props.scheduledAt, props.preserveVisit] as const')
    expect(src).toContain('syncFlowProps()')
  })

  it('syncFlowProps recopie les props sans toucher aux dialogues', () => {
    const body = functionBody('syncFlowProps')
    expect(body).toContain('flowScheduledAt.value')
    expect(body).toContain('flowPetId.value')
    expect(body).toContain('active.syncActiveVisit')
    for (const forbidden of [
      'leavePromptOpen',
      'nextPromptOpen',
      'nextStepsOpen',
      'reportSaved.value = false',
      'hydrateResumeSaved',
    ]) {
      expect(body, `syncFlowProps ne doit pas toucher ${forbidden}`).not.toContain(forbidden)
    }
  })

  it('bindFlow réutilise le miroir puis ferme les dialogues', () => {
    const body = functionBody('bindFlow')
    expect(body).toContain('syncFlowProps()')
    expect(body).toContain('leavePromptOpen.value = false')
    expect(body).toContain('reportSaved.value = false')
    expect(body).toContain('hydrateResumeSaved(props.visitId)')
  })
})
