import type { UseCaseCatalog, UseCaseItem } from '~/data/usecases/types'
import catalogJson from '~/data/usecases/catalog.json'

const catalog = catalogJson as UseCaseCatalog

export function useUsecasesCatalog() {
  return {
    catalog,
    folders: catalog.folders,
    demoSession: catalog.demoSession,
    getCase(id: string): UseCaseItem | undefined {
      return catalog.cases[id]
    },
  }
}
