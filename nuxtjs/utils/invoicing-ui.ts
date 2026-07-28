/**
 * Kill-switch unique pour l’UI facturation Pro (page `/invoicing` + CTA Facturer).
 * `false` = écran « En cours de développement » ; API Billit inchangée (`BILLIT_ENABLED`).
 * Remettre à `true` pour réactiver l’UI métier (déjà présente derrière le flag dans la page).
 */
export const INVOICING_UI_ENABLED = false
