export type UseCaseActor = {
  role: string
  account: string
  password: string
}

export type UseCaseItem = {
  id: string
  folderId: string
  slug: string
  title: string
  duration: string
  priority: string
  surface: string
  destructive: boolean
  objective: string
  actors: UseCaseActor[]
  prerequisites: string[]
  steps: string[]
  expected: string[]
  checklist: string[]
}

export type UseCaseFolder = {
  id: string
  dir: string
  label: string
  blurb?: string
  items: string[]
}

export type UseCaseCatalog = {
  generatedAt: string
  source: string
  demoSession: string[]
  folders: UseCaseFolder[]
  cases: Record<string, UseCaseItem>
}
