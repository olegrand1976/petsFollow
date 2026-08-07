/**
 * Les documents du dossier animal sont du PHI : ils vivent dans un namespace de
 * stockage privé et n'ont pas d'URL publique. Le seul chemin de lecture est la
 * route BFF authentifiée, qui relaie le flux de l'API Go.
 */
export function petDocumentHref(petId: string, documentId: string) {
  return `/api/pets/${encodeURIComponent(petId)}/documents/${encodeURIComponent(documentId)}/download`
}
