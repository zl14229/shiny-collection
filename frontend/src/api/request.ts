import axios from 'axios'
import type { ApiResponse } from '@/types'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// response interceptor
request.interceptors.response.use(
  (response) => {
    const data = response.data as ApiResponse
    if (data.code !== 0) {
      return Promise.reject(new Error(data.message || 'request failed'))
    }
    return response
  },
  (error) => {
    const msg = error.response?.data?.message || error.message || 'network error'
    console.error('API Error:', msg)
    return Promise.reject(new Error(msg))
  },
)

export default request
