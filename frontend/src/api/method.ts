import request from './request'
import type { ApiResponse, Method } from '@/types'

export function listMethods() {
  return request.get<ApiResponse<Method[]>>('/methods')
}
