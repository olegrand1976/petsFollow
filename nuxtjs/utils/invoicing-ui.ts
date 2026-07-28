/**
 * Kill-switch unique pour l’UI facturation Pro (page `/invoicing` + CTA Facturer).
 * `false` = écran « En cours de développement » ; API Billit inchangée (`BILLIT_ENABLED`).
 * Remettre à `false` pour geler l’UI métier sans couper l’API mock/CI.
 */
export const INVOICING_UI_ENABLED = true
