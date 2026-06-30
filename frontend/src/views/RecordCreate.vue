<template>
  <div class="record-create">
    <el-card shadow="never">
      <template #header>
        <span class="page-title">📝 新增狩猎记录</span>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        size="default"
      >
        <!-- Pokemon -->
        <el-form-item label="宝可梦" prop="pokemonId">
          <el-select
            v-model="form.pokemonId"
            filterable
            remote
            :remote-method="searchPokemon"
            :loading="searching"
            placeholder="搜索宝可梦名称"
            style="width: 300px"
          >
            <el-option
              v-for="p in pokemonOptions"
              :key="p.id"
              :label="`#${p.nationalNo} ${p.nameCN || p.name}`"
              :value="p.id"
            />
          </el-select>
        </el-form-item>

        <!-- Game -->
        <el-form-item label="游戏版本" prop="gameId">
          <el-select v-model="form.gameId" placeholder="选择游戏" style="width: 300px">
            <el-option
              v-for="g in games"
              :key="g.id"
              :label="g.nameCN"
              :value="g.id"
            />
          </el-select>
        </el-form-item>

        <!-- Method -->
        <el-form-item label="狩猎方式" prop="methodId">
          <el-select v-model="form.methodId" placeholder="选择方式" style="width: 300px">
            <el-option
              v-for="m in methods"
              :key="m.id"
              :label="m.nameCN"
              :value="m.id"
            />
          </el-select>
        </el-form-item>

        <!-- Status -->
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="hunting">狩猎中</el-radio>
            <el-radio value="obtained">已获得</el-radio>
            <el-radio value="abandoned">已放弃</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- Encounters -->
        <el-form-item label="遭遇次数">
          <el-input-number v-model="form.totalEncounters" :min="0" :step="1" style="width: 200px" />
        </el-form-item>

        <!-- Dates -->
        <el-form-item label="开始日期" prop="startDate">
          <el-date-picker v-model="form.startDate" type="date" placeholder="选择日期" style="width: 200px" />
        </el-form-item>

        <el-form-item label="结束日期" v-if="form.status !== 'hunting'">
          <el-date-picker v-model="form.endDate" type="date" placeholder="选择日期" style="width: 200px" />
        </el-form-item>

        <!-- Shiny -->
        <el-form-item label="是否出闪">
          <el-switch v-model="form.shinyAppearance" />
        </el-form-item>

        <!-- Details -->
        <el-divider content-position="left">详细信息</el-divider>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="性格">
              <el-input v-model="form.nature" placeholder="如：胆小" style="width: 160px" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="性别">
              <el-select v-model="form.gender" placeholder="性别" style="width: 120px">
                <el-option label="♂" value="♂" />
                <el-option label="♀" value="♀" />
                <el-option label="无性别" value="无" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="精灵球">
              <el-input v-model="form.ballUsed" placeholder="如：纪念球" style="width: 160px" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="等级">
              <el-input-number v-model="form.level" :min="1" :max="100" style="width: 160px" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="头目(LA)">
              <el-switch v-model="form.isAlpha" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="证章">
              <el-switch v-model="form.isMarked" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="证章名" v-if="form.isMarked">
          <el-input v-model="form.markName" placeholder="如：财运之证" style="width: 200px" />
        </el-form-item>

        <el-form-item label="备注">
          <el-input v-model="form.notes" type="textarea" :rows="3" placeholder="记录一些备注..." style="width: 500px" />
        </el-form-item>

        <!-- Submit -->
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">提交</el-button>
          <el-button @click="router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useRecordStore } from '@/stores/record'
import { listGames } from '@/api/game'
import { listMethods } from '@/api/method'
import { listPokemon } from '@/api/pokemon'
import type { Game, Method, Pokemon, RecordStatus } from '@/types'
import dayjs from 'dayjs'

const router = useRouter()
const store = useRecordStore()

const formRef = ref()
const submitting = ref(false)
const searching = ref(false)
const pokemonOptions = ref<Pokemon[]>([])
const games = ref<Game[]>([])
const methods = ref<Method[]>([])

const form = reactive({
  pokemonId: undefined as number | undefined,
  gameId: undefined as number | undefined,
  methodId: undefined as number | undefined,
  status: 'hunting',
  totalEncounters: 0,
  startDate: new Date(),
  endDate: undefined as Date | undefined,
  shinyAppearance: false,
  nature: '',
  gender: '',
  ballUsed: '',
  level: 1,
  isAlpha: false,
  isMarked: false,
  markName: '',
  notes: '',
})

const rules = {
  pokemonId: [{ required: true, message: '请选择宝可梦', trigger: 'change' }],
  gameId: [{ required: true, message: '请选择游戏版本', trigger: 'change' }],
  methodId: [{ required: true, message: '请选择狩猎方式', trigger: 'change' }],
  startDate: [{ required: true, message: '请选择开始日期', trigger: 'change' }],
}

async function searchPokemon(keyword: string) {
  if (!keyword) return
  searching.value = true
  try {
    const res = await listPokemon(keyword)
    pokemonOptions.value = res.data.data || []
  } catch {
    pokemonOptions.value = []
  } finally {
    searching.value = false
  }
}

async function handleSubmit() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await store.addRecord({
      pokemonId: form.pokemonId!,
      gameId: form.gameId!,
      methodId: form.methodId!,
      status: form.status as RecordStatus,
      totalEncounters: form.totalEncounters,
      startDate: dayjs(form.startDate).format('YYYY-MM-DD'),
      endDate: form.endDate ? dayjs(form.endDate).format('YYYY-MM-DD') : undefined,
      shinyAppearance: form.shinyAppearance,
      nature: form.nature,
      gender: form.gender,
      ballUsed: form.ballUsed,
      level: form.level,
      isAlpha: form.isAlpha,
      isMarked: form.isMarked,
      markName: form.markName,
      notes: form.notes,
      tagIds: [],
    })
    ElMessage.success('创建成功！')
    router.push('/records')
  } catch (e: any) {
    ElMessage.error(e.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  try {
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
  }
})
</script>

<style scoped>
.record-create {
  max-width: 800px;
  margin: 0 auto;
}

.page-title {
  font-weight: 600;
  font-size: 18px;
}
</style>
