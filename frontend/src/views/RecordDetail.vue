<template>
  <div class="record-detail" v-loading="loading">
    <!-- Not Found -->
    <el-empty v-if="!record && !loading" description="记录不存在" />

    <template v-if="record">
      <el-card shadow="never">
        <template #header>
          <div class="detail-header">
            <span class="page-title">
              {{ formatPokemonName(record.pokemon) }}
              <span v-if="record.shinyAppearance" style="margin-left: 8px">✨</span>
            </span>
            <div class="header-actions">
              <el-button v-if="!editing" type="primary" @click="editing = true">编辑</el-button>
              <el-button v-else @click="cancelEdit">取消</el-button>
              <el-button @click="router.back()">返回</el-button>
            </div>
          </div>
        </template>

        <!-- Read Mode -->
        <div v-if="!editing" class="detail-body">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="宝可梦" :span="1">
              #{{ record.pokemon?.nationalNo }} {{ formatPokemonName(record.pokemon) }}
            </el-descriptions-item>
            <el-descriptions-item label="属性">
              <span
                v-if="record.pokemon?.type1"
                class="type-badge"
                :style="{ background: getTypeColor(record.pokemon.type1) }"
              >{{ record.pokemon.type1 }}</span>
              <span
                v-if="record.pokemon?.type2"
                class="type-badge"
                :style="{ background: getTypeColor(record.pokemon.type2) }"
              >{{ record.pokemon.type2 }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="游戏版本">{{ record.game?.nameCN || record.game?.name }}</el-descriptions-item>
            <el-descriptions-item label="狩猎方式">{{ record.method?.nameCN || record.method?.name }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="statusType(record.status)" size="small">{{ statusLabel(record.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="遭遇次数">{{ record.totalEncounters.toLocaleString() }}</el-descriptions-item>
            <el-descriptions-item label="开始日期">{{ formatDate(record.startDate) }}</el-descriptions-item>
            <el-descriptions-item label="结束日期">{{ record.endDate ? formatDate(record.endDate) : '-' }}</el-descriptions-item>
            <el-descriptions-item label="性格">{{ record.nature || '-' }}</el-descriptions-item>
            <el-descriptions-item label="性别">{{ record.gender || '-' }}</el-descriptions-item>
            <el-descriptions-item label="精灵球">{{ record.ballUsed || '-' }}</el-descriptions-item>
            <el-descriptions-item label="等级">{{ record.level }}</el-descriptions-item>
            <el-descriptions-item label="头目(LA)">{{ record.isAlpha ? '是' : '否' }}</el-descriptions-item>
            <el-descriptions-item label="证章">{{ record.isMarked ? `是 (${record.markName})` : '否' }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="record.notes" class="notes-section">
            <h4>备注</h4>
            <p>{{ record.notes }}</p>
          </div>

          <!-- 出闪时刻视频 -->
          <div v-if="record.shinyVideo || !editing" class="video-section">
            <h4>📹 出闪时刻</h4>
            <div v-if="videoUrl" class="video-wrapper">
              <video :src="videoUrl" controls preload="metadata" class="shiny-video">
                您的浏览器不支持视频播放
              </video>
              <div class="video-actions">
                <el-button type="danger" size="small" @click="handleVideoDelete">删除视频</el-button>
              </div>
            </div>
            <div v-else class="upload-area">
              <el-upload
                :show-file-list="false"
                :before-upload="onVideoSelect"
                accept=".mp4,.webm,.mov,.avi,.mkv"
              >
                <el-button type="primary" :loading="uploading">
                  {{ uploading ? '上传中...' : '上传出闪视频' }}
                </el-button>
              </el-upload>
              <p class="upload-hint">支持 mp4 / webm / mov / avi / mkv，最大 500MB</p>
            </div>
          </div>

          <div class="time-info">
            <span>创建于: {{ formatDateTime(record.createdAt) }}</span>
            <span>更新于: {{ formatDateTime(record.updatedAt) }}</span>
          </div>
        </div>

        <!-- Edit Mode -->
        <div v-else class="edit-body">
          <el-form
            ref="formRef"
            :model="editForm"
            label-width="120px"
            size="default"
          >
            <el-form-item label="宝可梦" prop="pokemonId">
              <el-select
                v-model="editForm.pokemonId"
                filterable
                remote
                :remote-method="searchPokemon"
                :loading="searching"
                placeholder="搜索宝可梦"
                style="width: 300px"
              >
                <el-option
                  v-for="p in pokemonOptions"
                  :key="p.id"
                  :label="pokemonSearchLabel(p)"
                  :value="p.id"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="游戏版本">
              <el-select v-model="editForm.gameId" style="width: 300px">
                <el-option v-for="g in games" :key="g.id" :label="g.nameCN" :value="g.id" />
              </el-select>
            </el-form-item>

            <el-form-item label="狩猎方式">
              <el-select v-model="editForm.methodId" style="width: 300px">
                <el-option v-for="m in methods" :key="m.id" :label="m.nameCN" :value="m.id" />
              </el-select>
            </el-form-item>

            <el-form-item label="状态">
              <el-radio-group v-model="editForm.status">
                <el-radio value="hunting">狩猎中</el-radio>
                <el-radio value="obtained">已获得</el-radio>
                <el-radio value="abandoned">已放弃</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="遭遇次数">
              <el-input-number v-model="editForm.totalEncounters" :min="0" />
            </el-form-item>

            <el-form-item label="是否出闪">
              <el-switch v-model="editForm.shinyAppearance" />
            </el-form-item>

            <el-form-item label="性格">
              <el-input v-model="editForm.nature" style="width: 200px" />
            </el-form-item>

            <el-form-item label="性别">
              <el-select v-model="editForm.gender" style="width: 120px">
                <el-option label="♂" value="♂" />
                <el-option label="♀" value="♀" />
                <el-option label="无性别" value="无" />
              </el-select>
            </el-form-item>

            <el-form-item label="精灵球">
              <el-input v-model="editForm.ballUsed" style="width: 200px" />
            </el-form-item>

            <el-form-item label="备注">
              <el-input v-model="editForm.notes" type="textarea" :rows="3" style="width: 500px" />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="saveEdit" :loading="saving">保存</el-button>
              <el-button @click="cancelEdit">取消</el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useRecordStore } from '@/stores/record'
import { listGames } from '@/api/game'
import { listMethods } from '@/api/method'
import { listPokemon } from '@/api/pokemon'
import { uploadVideo, deleteVideo } from '@/api/video'
import { getTypeColor } from '@/utils/typeColors'
import { formatPokemonName, pokemonSearchLabel } from '@/utils/pokemonFormat'
import type { Game, Method, Pokemon, RecordStatus } from '@/types'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const store = useRecordStore()

const loading = ref(true)
const editing = ref(false)
const saving = ref(false)
const searching = ref(false)
const uploading = ref(false)
const videoUrl = computed(() => {
  if (record.value?.shinyVideo) {
    return `/uploads/videos/${record.value.shinyVideo}`
  }
  return null
})

async function handleVideoUpload(file: File) {
  if (!record.value) return
  uploading.value = true
  try {
    await uploadVideo(record.value.id, file)
    ElMessage.success('视频上传成功')
    await store.fetchRecord(record.value.id)
  } catch (e: any) {
    ElMessage.error(e.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

async function handleVideoDelete() {
  if (!record.value) return
  try {
    await deleteVideo(record.value.id)
    ElMessage.success('视频已删除')
    await store.fetchRecord(record.value.id)
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

function onVideoSelect(file: File) {
  const ext = file.name.split('.').pop()?.toLowerCase()
  if (!['mp4', 'webm', 'mov', 'avi', 'mkv'].includes(ext || '')) {
    ElMessage.error('不支持的文件格式，支持: mp4/webm/mov/avi/mkv')
    return false
  }
  if (file.size > 500 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过 500MB')
    return false
  }
  handleVideoUpload(file)
  return false
}

const formRef = ref()

const games = ref<Game[]>([])
const methods = ref<Method[]>([])
const pokemonOptions = ref<Pokemon[]>([])

const record = ref<typeof store.currentRecord>(null)

const editForm = reactive({
  pokemonId: 0,
  gameId: 0,
  methodId: 0,
  status: 'hunting' as string,
  totalEncounters: 0,
  shinyAppearance: false,
  nature: '',
  gender: '',
  ballUsed: '',
  notes: '',
})

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

function formatDateTime(d: string) {
  return d ? dayjs(d).format('YYYY-MM-DD HH:mm') : '-'
}

function initEditForm() {
  if (!record.value) return
  editForm.pokemonId = record.value.pokemonId
  editForm.gameId = record.value.gameId
  editForm.methodId = record.value.methodId
  editForm.status = record.value.status
  editForm.totalEncounters = record.value.totalEncounters
  editForm.shinyAppearance = record.value.shinyAppearance
  editForm.nature = record.value.nature
  editForm.gender = record.value.gender
  editForm.ballUsed = record.value.ballUsed
  editForm.notes = record.value.notes
}

function cancelEdit() {
  editing.value = false
  initEditForm() // reset form
}

async function searchPokemon(keyword: string) {
  if (!keyword) return
  searching.value = true
  try {
    const res = await listPokemon(keyword)
    pokemonOptions.value = res.data.data || []
  } finally {
    searching.value = false
  }
}

async function saveEdit() {
  if (!record.value) return
  saving.value = true
  try {
    await store.editRecord(record.value.id, { ...editForm, status: editForm.status as RecordStatus })
    ElMessage.success('保存成功')
    editing.value = false
    // reload
    await store.fetchRecord(record.value.id)
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    if (!id) {
      loading.value = false
      return
    }
    await store.fetchRecord(id)
    record.value = store.currentRecord

    if (record.value) {
      initEditForm()
    }

    // load options for edit mode
    const [gamesRes, methodsRes, pokemonRes] = await Promise.all([
      listGames(),
      listMethods(),
      listPokemon(),
    ])
    games.value = gamesRes.data.data || []
    methods.value = methodsRes.data.data || []
    pokemonOptions.value = pokemonRes.data.data || []
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.record-detail {
  max-width: 900px;
  margin: 0 auto;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  font-weight: 600;
  font-size: 18px;
}

.detail-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.type-badge {
  padding: 2px 10px;
  border-radius: 4px;
  color: white;
  font-size: 12px;
  font-weight: 500;
  margin-right: 4px;
}

.notes-section {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 8px;
}

.notes-section h4 {
  margin-bottom: 8px;
  font-size: 14px;
  color: #606266;
}

.notes-section p {
  white-space: pre-wrap;
  color: #303133;
  font-size: 14px;
}

.time-info {
  display: flex;
  gap: 24px;
  font-size: 12px;
  color: #909399;
}

.edit-body {
  padding: 20px 0;
}

.video-section {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 8px;
}

.video-section h4 {
  margin-bottom: 12px;
  font-size: 14px;
  color: #606266;
}

.video-wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}

.shiny-video {
  max-width: 100%;
  max-height: 400px;
  border-radius: 8px;
  background: #000;
}

.video-actions {
  display: flex;
  gap: 8px;
}

.upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 24px;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  background: #fff;
}

.upload-hint {
  font-size: 12px;
  color: #909399;
  margin: 0;
}
</style>
