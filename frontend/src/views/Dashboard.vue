<template>
  <div class="dashboard">
    <!-- Stats Cards Row -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card shiny-card">
          <div class="stat-value">{{ stats?.totalShiny ?? '-' }}</div>
          <div class="stat-label">✨ 共获得异色</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card hunting-card">
          <div class="stat-value">{{ stats?.huntingRecords ?? '-' }}</div>
          <div class="stat-label">🎯 正在狩猎</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card encounter-card">
          <div class="stat-value">{{ formatEncounters(stats?.totalEncounters) }}</div>
          <div class="stat-label">⚡ 总遭遇次数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card total-card">
          <div class="stat-value">{{ stats?.totalRecords ?? '-' }}</div>
          <div class="stat-label">📋 总记录数</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Charts Row -->
    <el-row :gutter="20" class="charts-row">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span class="card-title">狩猎方式分布</span>
          </template>
          <div class="chart-placeholder">
            <div v-if="stats?.methodBreakdown?.length" class="method-chart">
              <div
                v-for="item in stats.methodBreakdown"
                :key="item.methodId"
                class="method-bar-row"
              >
                <span class="method-label">{{ item.methodName }}</span>
                <el-progress
                  :percentage="getMethodPercent(item.count)"
                  :stroke-width="20"
                  :text-inside="true"
                  :color="getMethodColor(item.methodId)"
                >
                  <span>{{ item.count }} 条</span>
                </el-progress>
              </div>
            </div>
            <el-empty v-else description="还没有狩猎记录" :image-size="80" />
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <span class="card-title">统计概览</span>
          </template>
          <div class="chart-placeholder">
            <div v-if="stats" class="quick-stats">
              <div class="quick-stat-item">
                <span class="qs-label">各游戏统计</span>
              </div>
              <div v-if="gameStats.length" class="game-stats-list">
                <div v-for="gs in gameStats" :key="gs.gameId" class="game-stat-row">
                  <span class="gs-game">{{ gs.gameName }}</span>
                  <span class="gs-counts">
                    <span class="gs-shiny">✨ {{ gs.shiny }}</span>
                    <span class="gs-divider">/</span>
                    <span class="gs-total">{{ gs.total }}</span>
                  </span>
                </div>
              </div>
              <el-empty v-else description="暂无游戏统计" :image-size="60" />
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Monthly Trend -->
    <el-card shadow="hover" class="trend-card">
      <template #header>
        <span class="card-title">📈 近月出闪趋势</span>
      </template>
      <div class="trend-chart">
        <div v-if="stats?.monthlyTrend?.length" class="monthly-bars">
          <div
            v-for="m in stats.monthlyTrend"
            :key="`${m.year}-${m.month}`"
            class="month-bar-wrapper"
          >
            <div class="month-bar" :style="{ height: getBarHeight(m.count) + 'px' }">
              <span class="bar-count">{{ m.count }}</span>
            </div>
            <span class="month-label">{{ m.year }}/{{ String(m.month).padStart(2, '0') }}</span>
          </div>
        </div>
        <el-empty v-else description="还没有出闪记录" :image-size="80" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRecordStore } from '@/stores/record'

const store = useRecordStore()
const stats = computed(() => store.stats)
const gameStats = computed(() => store.gameStats)

const METHOD_COLORS = [
  '#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399',
  '#B37FEB', '#36CFC9', '#F2A6B1', '#79BBFF', '#B3E19B',
]

function getMethodColor(index: number): string {
  return METHOD_COLORS[index % METHOD_COLORS.length]
}

function getMethodPercent(count: number): number {
  if (!stats.value?.totalRecords) return 0
  return Math.round((count / stats.value.totalRecords) * 100)
}

function getBarHeight(count: number): number {
  const max = Math.max(...(stats.value?.monthlyTrend?.map(m => m.count) || [1]), 1)
  return Math.max((count / max) * 120, 4)
}

function formatEncounters(n: number | undefined): string {
  if (n === undefined) return '-'
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
  return n.toLocaleString()
}

onMounted(() => {
  store.fetchStats()
})
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

.stats-row { margin-bottom: 20px; }

.stat-card {
  text-align: center;
  border-radius: 12px;
  border: none;
}

.stat-value {
  font-size: 36px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}

.charts-row { margin-bottom: 20px; }

.card-title {
  font-weight: 600;
  font-size: 16px;
}

.chart-placeholder {
  min-height: 200px;
}

/* Method bars */
.method-chart { display: flex; flex-direction: column; gap: 12px; }
.method-bar-row { display: flex; align-items: center; gap: 12px; }
.method-label {
  width: 100px;
  font-size: 13px;
  color: #606266;
  text-align: right;
  flex-shrink: 0;
}
.method-bar-row .el-progress { flex: 1; margin: 0; }

/* Quick stats */
.quick-stats { display: flex; flex-direction: column; }
.quick-stat-item { margin-bottom: 12px; }
.qs-label { font-weight: 600; font-size: 14px; color: #303133; }
.game-stats-list { display: flex; flex-direction: column; gap: 8px; }
.game-stat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  background: #f5f7fa;
  border-radius: 8px;
}
.gs-game { font-size: 13px; color: #606266; }
.gs-counts { display: flex; align-items: center; gap: 4px; }
.gs-shiny { color: #e6a23c; font-weight: 600; }
.gs-divider { color: #dcdfe6; }
.gs-total { color: #909399; }

/* Monthly trend */
.trend-card { margin-bottom: 20px; }

.monthly-bars {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  padding: 20px 0;
  min-height: 180px;
  justify-content: center;
}

.month-bar-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.month-bar {
  width: 40px;
  background: linear-gradient(180deg, #f6e05e, #e6a23c);
  border-radius: 8px 8px 0 0;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 4px;
  min-height: 4px;
  transition: height 0.3s ease;
}

.bar-count {
  font-size: 11px;
  color: #7c4a00;
  font-weight: 700;
}

.month-label {
  font-size: 11px;
  color: #909399;
  white-space: nowrap;
}
</style>
