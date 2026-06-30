import request from './request'
import type { ApiResponse, StatsOverview, GameStat } from '@/types'

export function getStatsOverview() {
  return request.get<ApiResponse<StatsOverview>>('/stats/overview')
}

export function getStatsByGame() {
  return request.get<ApiResponse<GameStat[]>>('/stats/by-game')
}
