import type { Pokemon } from '@/types'

const FORM_LABELS: Record<string, string> = {
  Alola: '阿罗拉',
  Galar: '伽勒尔',
  Hisui: '洗翠',
  Paldea: '帕底亚',
}

export function formatPokemonName(p: Pokemon | null | undefined): string {
  if (!p) return ''
  const base = p.nameCN
  if (p.form && FORM_LABELS[p.form]) {
    return `${base}（${FORM_LABELS[p.form]}的样子）`
  }
  return base
}

export function pokemonSearchLabel(p: Pokemon): string {
  const base = `#${p.nationalNo} ${p.nameCN || p.name}`
  if (p.form && FORM_LABELS[p.form]) {
    return `${base}（${FORM_LABELS[p.form]}的样子）`
  }
  return base
}
