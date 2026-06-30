<template>
  <div class="record-list">
    <!-- Toolbar -->
    <el-card shadow="never" class="toolbar-card">
      <el-form :inline="true" :model="store.filters" size="default">
        <el-form-item label="状态">
          <el-select v-model="store.filters.status" clearable placeholder="全部" style="width: 130px">
            <el-option label="狩猎中" value="hunting" />
            <el-option label="已获得" value="obtained" />
            <el-option label="已放弃" value="abandoned" />
          </el-select>
        </el-form-item>
        <el-form-item label="游戏">
          <el-select v-model="store.filters.gameId" clearable placeholder="全部" style="width: 160px">
            <el-option v-for="g in games" :key="g.id" :label="g.nameCN" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="方式">
          <el-select v-model="store.filters.methodId" clearable placeholder="全部" style="width: 150px">
            <el-option v-for="m in methods" :key="m.id" :label="m.nameCN" :value="m.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="store.filters.keyword"
            placeholder="搜索宝可梦或备注"
            clearable
            style="width: 200px"
            @keyup.enter="search"
          >
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Table -->
    <el-card shadow="never" class="table-card">
      <el-table
        :data="store.records"
        v-loading="store.loading"
        stripe
        style="width: 100%"
        @row-click="goDetail"
      >
        <el-table-column label="ID" prop="id" width="70" />
        <el-table-column label="宝可梦" min-width="140">
          <template #default="{ row }">
            <div class="pokemon-cell">
              <span class="national-no">#{{ row.pokemon?.nationalNo }}</span>
              <span :style="{ color: getTypeColor(row.pokemon?.type1) }" class="poke-name">
                {{ formatPokemonName(row.pokemon) }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="属性" min-width="100">
          <template #default="{ row }">
            <div class="types-cell">
              <span
                v-if="row.pokemon?.type1"
                class="type-badge"
                :style="{ background: getTypeColor(row.pokemon.type1) }"
              >
                {{ row.pokemon.type1 }}
              </span>
              <span
                v-if="row.pokemon?.type2"
                class="type-badge"
                :style="{ background: getTypeColor(row.pokemon.type2) }"
              >
                {{ row.pokemon.type2 }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="游戏" min-width="130">
          <template #default="{ row }">
            <span>{{ row.game?.nameCN || row.game?.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="方式" min-width="110">
          <template #default="{ row }">
            <span>{{ row.method?.nameCN || row.method?.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="85">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small" effect="plain">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="遭遇数" prop="totalEncounters" width="85" sortable />
        <el-table-column label="结果" width="55">
          <template #default="{ row }">
            <span v-if="row.shinyAppearance && row.status === 'obtained'" style="font-size: 18px">✨</span>
          </template>
        </el-table-column>
        <el-table-column label="开始日期" min-width="110">
          <template #default="{ row }">
            <span>{{ formatDate(row.startDate) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click.stop="editRecord(row)">编辑</el-button>
            <el-popconfirm title="确定删除?" @confirm.stop="handleDelete(row.id)">
              <template #reference>
                <el-button text type="danger" size="small" @click.stop>删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="store.currentPage"
          v-model:page-size="store.pageSize"
          :total="store.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @change="store.fetchRecords()"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useRecordStore } from '@/stores/record'
import { listGames } from '@/api/game'
import { listMethods } from '@/api/method'
import { getTypeColor } from '@/utils/typeColors'
import { formatPokemonName } from '@/utils/pokemonFormat'
import type { Game, Method } from '@/types'
import dayjs from 'dayjs'

const router = useRouter()
const store = useRecordStore()

const games = ref<Game[]>([])
const methods = ref<Method[]>([])

function statusType(s: string) {
  const map: Record<string, string> = { hunting: 'warning', obtained: 'success', abandoned: 'info' }
  return map[s] || 'info'
}
function statusLabel(s: string) {
  const map: Record<string, string> = { hunting: '狩猎中', obtained: '已获得', abandoned: '已放弃' }
  return map[s] || s
}
function formatDate(d: string) {
  return d ? dayjs(d).format('YYYY-MM-DD') : '-'
}

function search() {
  store.filters.page = 1
  store.fetchRecords()
}

function resetFilters() {
  store.resetFilters()
  store.fetchRecords()
}

function goDetail(row: any) {
  router.push(`/records/${row.id}`)
}

function editRecord(row: any) {
  router.push(`/records/${row.id}`)
}

async function handleDelete(id: number) {
  try {
    await store.removeRecord(id)
    ElMessage.success('已删除')
    store.fetchRecords()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(async () => {
  store.fetchRecords()
  try {
    const [gamesRes, methodsRes] = await Promise.all([listGames(), listMethods()])
    games.value = gamesRes.data.data || []
    methods.value = methodsRes.data.data || []
  } catch {
    // ignore
  }
})
</script>

<style scoped>
.toolbar-card { margin-bottom: 16px; }
.table-card { margin-bottom: 16px; }
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.pokemon-cell {
  display: flex;
  flex-direction: column;
}

.national-no {
  font-size: 11px;
  color: #909399;
}

.poke-name {
  font-weight: 600;
  font-size: 14px;
}

.types-cell {
  display: flex;
  gap: 4px;
}

.type-badge {
  padding: 1px 8px;
  border-radius: 4px;
  color: white;
  font-size: 11px;
  font-weight: 500;
}
</style>
