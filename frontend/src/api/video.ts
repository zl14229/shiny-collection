import request from './request'
import type { ApiResponse } from '@/types'

export function uploadVideo(recordId: number, file: File) {
  const formData = new FormData()
  formData.append('video', file)
  return request.post<ApiResponse<{ filename: string; url: string }>>(
    `/records/${recordId}/video`,
    formData,
    {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 300000, // 5min for large uploads
    },
  )
}

export function deleteVideo(recordId: number) {
  return request.delete<ApiResponse<null>>(`/records/${recordId}/video`)
}
