#!/usr/bin/env node
/**
 * Sync useCase UC markdown files into nuxtjs/data/usecases/catalog.json
 * Usage: node scripts/sync-usecases.mjs
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const useCaseRoot = path.join(root, 'useCase')
const outPath = path.join(root, 'nuxtjs/data/usecases/catalog.json')

const FOLDERS = [
  {
    dir: '01-vetpro',
    id: 'vetpro',
    label: 'Cabinet (site web)',
    blurb: 'Ce que voit le vétérinaire sur l’ordinateur du cabinet.',
  },
  {
    dir: '02-client',
    id: 'client',
    label: 'Propriétaire (app mobile)',
    blurb: 'Ce que voit le client avec son téléphone.',
  },
  {
    dir: '03-pro-light',
    id: 'pro-light',
    label: 'Terrain (app mobile)',
    blurb: 'Maréchal / véto en visite — application légère.',
  },
  {
    dir: '04-commercial',
    id: 'commercial',
    label: 'Commercial',
    blurb: 'Prospects, inscriptions et commissions.',
  },
  {
    dir: '05-commercial-manager',
    id: 'commercial-manager',
    label: 'Responsable commercial',
    blurb: 'Pilotage de l’équipe commerciale.',
  },
  {
    dir: '06-admin',
    id: 'admin',
    label: 'Administration',
    blurb: 'Outils internes (équipe petsFollow).',
  },
  {
    dir: '07-equipe-cabinet',
    id: 'equipe-cabinet',
    label: 'Équipe du cabinet',
    blurb: 'Collègue, assistante, secrétaire.',
  },
  {
    dir: '10-interactions',
    id: 'interactions',
    label: 'À deux (cabinet + app)',
    blurb: 'Scénarios où deux personnes se parlent via l’outil.',
  },
]

function section(md, heading) {
  const re = new RegExp(`## ${heading}\\s*\\n([\\s\\S]*?)(?=\\n## |\\n---|$)`, 'i')
  const m = md.match(re)
  return m ? m[1].trim() : ''
}

function meta(md, key) {
  const re = new RegExp(`\\| \\*\\*${key}\\*\\* \\| \`?([^|\`]+)\`? \\|`, 'i')
  const m = md.match(re)
  return m ? m[1].trim() : ''
}

function parseActors(block) {
  const actors = []
  for (const line of block.split('\n')) {
    if (!line.startsWith('|') || line.includes('---') || /Rôle|Compte|MDP/i.test(line)) continue
    const cols = line.split('|').map((c) => c.trim()).filter(Boolean)
    if (cols.length >= 3) {
      actors.push({ role: cols[0], account: cols[1].replace(/`/g, ''), password: cols[2].replace(/`/g, '') })
    }
  }
  return actors
}

function parseSteps(block) {
  const steps = []
  for (const line of block.split('\n')) {
    const m = line.match(/^\d+\.\s+(.+)/)
    if (m) steps.push(m[1].trim())
  }
  return steps
}

function parseChecklist(block) {
  const items = []
  for (const line of block.split('\n')) {
    if (!line.startsWith('|') || line.includes('---') || /Résultat/i.test(line)) continue
    const cols = line.split('|').map((c) => c.trim()).filter(Boolean)
    if (cols.length >= 1 && cols[0] !== '') items.push(cols[0])
  }
  return items
}

function parseBullets(block) {
  return block
    .split('\n')
    .map((l) => l.replace(/^[-*]\s+/, '').trim())
    .filter((l) => l && !l.startsWith('→') && !l.startsWith('|'))
}

function stripMd(text) {
  return String(text || '')
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '$1')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    .replace(/\s+/g, ' ')
    .trim()
}

function parseUc(filePath, folderId) {
  const md = fs.readFileSync(filePath, 'utf8')
  const titleLine = md.match(/^#\s+(.+)$/m)?.[1]?.trim() || path.basename(filePath)
  const id = meta(md, 'ID').replace(/`/g, '') || path.basename(filePath, '.md').split('-').slice(0, 3).join('-').toUpperCase()
  const destructive = /Destructif/i.test(meta(md, 'Destructif')) || /\*\*Destructif\*\*/i.test(md.slice(0, 800))
  const objective = stripMd(section(md, 'Objectif'))
  return {
    id,
    folderId,
    slug: path.basename(filePath, '.md'),
    title: stripMd(titleLine.replace(/^UC-[A-Z]+-\d+\s+[—–-]\s*/, '')),
    duration: stripMd(meta(md, 'Durée')),
    priority: stripMd(meta(md, 'Priorité')),
    surface: stripMd(meta(md, 'Surface') || meta(md, 'Surfaces')),
    destructive,
    objective,
    actors: parseActors(section(md, 'Acteurs')).map((a) => ({
      role: stripMd(a.role),
      account: stripMd(a.account),
      password: stripMd(a.password),
    })),
    prerequisites: parseBullets(section(md, 'Prérequis')).map(stripMd),
    steps: parseSteps(section(md, 'Étapes')).map(stripMd),
    expected: parseBullets(section(md, 'Résultat attendu')).map(stripMd),
    checklist: parseChecklist(section(md, 'Checklist')).map(stripMd),
  }
}

const folders = []
const byId = {}

for (const f of FOLDERS) {
  const dirPath = path.join(useCaseRoot, f.dir)
  if (!fs.existsSync(dirPath)) continue
  const files = fs.readdirSync(dirPath).filter((n) => n.startsWith('UC-') && n.endsWith('.md')).sort()
  const items = files.map((name) => parseUc(path.join(dirPath, name), f.id))
  for (const item of items) byId[item.id] = item
  folders.push({ id: f.id, dir: f.dir, label: f.label, blurb: f.blurb, items: items.map((i) => i.id) })
}

const catalog = {
  source: 'useCase/',
  demoSession: ['UC-VP-01', 'UC-X-01', 'UC-X-02', 'UC-X-05', 'UC-PL-01', 'UC-CL-02', 'UC-CO-02'],
  folders,
  cases: byId,
}

fs.mkdirSync(path.dirname(outPath), { recursive: true })
fs.writeFileSync(outPath, `${JSON.stringify(catalog, null, 2)}\n`)
console.log(`Wrote ${Object.keys(byId).length} use cases → ${path.relative(root, outPath)}`)
