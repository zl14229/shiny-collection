import request from './request'
import type { ApiResponse, Pokemon } from '@/types'

export function listPokemon(keyword?: string) {
  return request.get<ApiResponse<Pokemon[]>>('/pokemon', {
    params: keyword ? { keyword } : {},
  })
}

export function getPokemon(id: number) {
  return request.get<ApiResponse<Pokemon>>(`/pokemon/${id}`)
}

export function createPokemon(data: { name: string; nationalNo: number; type1: string; type2?: string }) {
  return request.post<ApiResponse<Pokemon>>('/pokemon', data)
}
