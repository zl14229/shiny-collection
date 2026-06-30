import request from './request'
import type { ApiResponse, Game, StatsOverview, GameStat } from '@/types'

export function listGames() {
  return request.get<ApiResponse<Game[]>>('/games')
}

export function getGame(id: number) {
  return request.get<ApiResponse<Game>>(`/games/${id}`)
}
