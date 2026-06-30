import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { listRecords, getRecord, createRecord, updateRecord, deleteRecord } from '@/api/record'
import { getStatsOverview, getStatsByGame } from '@/api/stats'
import type { HuntRecord, StatsOverview, GameStat, CreateRecordForm, RecordQuery } from '@/types'

export const useRecordStore = defineStore('record', () => {
  // state
  const records = ref<HuntRecord[]>([])
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)
  const currentRecord = ref<HuntRecord | null>(null)
  const stats = ref<StatsOverview | null>(null)
  const gameStats = ref<GameStat[]>([])

  // query filters (not reactive by default in pinia set up)
  const filters = reactive<RecordQuery>({
    page: 1,
    pageSize: 20,
  })

  // actions
  async function fetchRecords() {
    loading.value = true
    try {
      const params = { ...filters, page: currentPage.value, pageSize: pageSize.value }
      const res = await listRecords(params)
      const data = res.data.data!
      records.value = data.list
      total.value = data.total
      currentPage.value = data.page
    } catch (e) {
      console.error('Failed to fetch records', e)
      records.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchRecord(id: number) {
    try {
      const res = await getRecord(id)
      currentRecord.value = res.data.data!
      return currentRecord.value
    } catch {
      currentRecord.value = null
      return null
    }
  }

  async function addRecord(data: CreateRecordForm) {
    const res = await createRecord(data)
    return res.data.data
  }

  async function editRecord(id: number, data: Partial<CreateRecordForm>) {
    const res = await updateRecord(id, data)
    return res.data.data
  }

  async function removeRecord(id: number) {
    await deleteRecord(id)
  }

  async function fetchStats() {
    try {
      const [overviewRes, gameStatsRes] = await Promise.all([
        getStatsOverview(),
        getStatsByGame(),
      ])
      stats.value = overviewRes.data.data!
      gameStats.value = gameStatsRes.data.data!
    } catch (e) {
      console.error('Failed to fetch stats', e)
    }
  }

  function resetFilters() {
    filters.page = 1
    filters.pageSize = 20
    filters.status = undefined
    filters.gameId = undefined
    filters.methodId = undefined
    filters.keyword = undefined
  }

  return {
    records,
    total,
    currentPage,
    pageSize,
    loading,
    currentRecord,
    stats,
    gameStats,
    filters,
    fetchRecords,
    fetchRecord,
    addRecord,
    editRecord,
    removeRecord,
    fetchStats,
    resetFilters,
  }
})
