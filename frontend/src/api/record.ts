import request from './request'
import type { ApiResponse, PageInfo, HuntRecord, CreateRecordForm, RecordQuery } from '@/types'

export function listRecords(params: RecordQuery) {
  return request.get<ApiResponse<PageInfo<HuntRecord>>>('/records', { params })
}

export function getRecord(id: number) {
  return request.get<ApiResponse<HuntRecord>>(`/records/${id}`)
}

export function createRecord(data: CreateRecordForm) {
  return request.post<ApiResponse<HuntRecord>>('/records', data)
}

export function updateRecord(id: number, data: Partial<CreateRecordForm>) {
  return request.put<ApiResponse<HuntRecord>>(`/records/${id}`, data)
}

export function deleteRecord(id: number) {
  return request.delete<ApiResponse<null>>(`/records/${id}`)
}
