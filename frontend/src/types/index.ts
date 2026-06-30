// ====== Data Models ======

export interface Pokemon {
  id: number
  name: string
  nameCN: string
  nationalNo: number
  type1: string
  type2: string
  imageUrl: string
  createdAt: string
  updatedAt: string
}

export interface Game {
  id: number
  name: string
  nameCN: string
  generation: number
  platform: string
  shortName: string
  releaseYear: number
}

export interface Method {
  id: number
  name: string
  nameCN: string
}

export interface Tag {
  id: number
  name: string
  color: string
}

export type RecordStatus = 'hunting' | 'obtained' | 'abandoned'

export interface HuntRecord {
  id: number
  pokemonId: number
  gameId: number
  methodId: number
  status: RecordStatus
  totalEncounters: number
  startDate: string
  endDate?: string
  shinyAppearance: boolean
  nature: string
  gender: string
  ballUsed: string
  level: number
  isAlpha: boolean
  isMarked: boolean
  markName: string
  notes: string
  shinyVideo: string
  tags: Tag[]
  pokemon?: Pokemon
  game?: Game
  method?: Method
  createdAt: string
  updatedAt: string
}

// ====== API Response ======

export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
}

export interface PageInfo<T = any> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// ====== Stats ======

export interface StatsOverview {
  totalRecords: number
  totalShiny: number
  huntingRecords: number
  totalEncounters: number
  methodBreakdown: MethodStat[]
  monthlyTrend: MonthlyStat[]
}

export interface MethodStat {
  methodId: number
  methodName: string
  count: number
}

export interface MonthlyStat {
  year: number
  month: number
  count: number
}

export interface GameStat {
  gameId: number
  gameName: string
  total: number
  shiny: number
}

// ====== Form ======

export interface CreateRecordForm {
  pokemonId: number
  gameId: number
  methodId: number
  status: RecordStatus
  totalEncounters: number
  startDate: string
  endDate?: string
  shinyAppearance: boolean
  nature: string
  gender: string
  ballUsed: string
  level: number
  isAlpha: boolean
  isMarked: boolean
  markName: string
  notes: string
  tagIds: number[]
}

// ====== Query ======

export interface RecordQuery {
  page?: number
  pageSize?: number
  status?: string
  gameId?: number
  methodId?: number
  pokemonId?: number
  tagId?: number
  keyword?: string
}
