<template>
  <div class="config-editor">
    <el-page-header :content="`配置包编辑：${seed.name || ''}`" @back="goBack">
      <template #extra>
        <el-button type="primary" @click="saveAll">保存</el-button>
        <el-button type="success" @click="handlePublish">发布</el-button>
      </template>
    </el-page-header>

    <el-alert title="层级配置：配置包 → 模块 → 字段/状态/流转/项目种子，像代码生成器一样逐级定义。" type="info" :closable="false" class="mt-4" />

    <el-card shadow="never" class="mt-4">
      <template #header>配置包基本信息</template>
      <el-form :model="seed" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="seed.name" placeholder="配置包名称，如 PMBOK 第六版混合型配置" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="seed.description" type="textarea" :rows="2" placeholder="配置包描述" />
        </el-form-item>
        <el-form-item label="状态">
          <el-tag :type="statusTag(pkg.status)" size="small">{{ statusLabel(pkg.status) }}</el-tag>
          <el-tag size="small" type="info" style="margin-left: 8px">v{{ pkg.version }}</el-tag>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 模块列表 -->
    <el-card shadow="never" class="mt-4">
      <template #header>
        <div class="flex items-center justify-between">
          <span>模块列表（{{ moduleKeys.length }} 个模块）</span>
          <el-button type="primary" size="small" @click="openAddModule">
            <svg-icon icon="lucide:plus" /> 添加模块
          </el-button>
        </div>
      </template>
      <el-collapse v-model="activeModules">
        <el-collapse-item v-for="key in moduleKeys" :key="key" :name="key">
          <template #title>
            <div class="flex items-center gap-3 w-full pr-4">
              <svg-icon :icon="moduleIcon(key)" class="text-base" />
              <span class="font-medium">{{ moduleName(key) }}</span>
              <el-tag size="small" type="info">{{ seed.modules[key].entityType }}</el-tag>
              <el-tag size="small">{{ seed.modules[key].fields?.length || 0 }} 字段</el-tag>
              <el-tag size="small" type="success">{{ (seed.modules[key].states?.length || 0) }} 状态</el-tag>
              <span class="ml-auto flex gap-2">
                <el-button link type="danger" size="small" @click.stop="removeModule(key)">删除模块</el-button>
              </span>
            </div>
          </template>

          <!-- 模块详情 -->
          <div class="module-detail">
            <el-form label-width="100px" class="mb-4">
              <el-row :gutter="16">
                <el-col :span="8">
                  <el-form-item label="模块编码">
                    <el-input v-model="seed.modules[key].module" :disabled="true" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="实体类型">
                    <el-input v-model="seed.modules[key].entityType" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="显示名">
                    <el-input v-model="seed.modules[key].name" />
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form>

            <el-tabs>
              <!-- 字段 -->
              <el-tab-pane label="字段定义">
                <div class="flex justify-between mb-2">
                  <span class="text-sm text-gray-500">模块字段（{{ seed.modules[key].fields?.length || 0 }}）</span>
                  <el-button type="primary" size="small" @click="openAddField(key)">添加字段</el-button>
                </div>
                <el-table :data="seed.modules[key].fields || []" size="small" border>
                  <el-table-column prop="key" label="字段键" width="140" />
                  <el-table-column prop="label" label="显示名" width="140" />
                  <el-table-column prop="dataType" label="类型" width="100">
                    <template #default="{ row }">
                      <el-tag size="small">{{ row.dataType }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="选项" min-width="140">
                    <template #default="{ row }">
                      <span class="text-xs text-gray-500">{{ (row.options || []).join(', ') }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="default" label="默认值" width="90" />
                  <el-table-column label="操作" width="120">
                    <template #default="{ row, $index }">
                      <el-button link type="primary" size="small" @click="openEditField(key, $index)">编辑</el-button>
                      <el-button link type="danger" size="small" @click="removeField(key, $index)">删除</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-tab-pane>

              <!-- 状态 -->
              <el-tab-pane label="状态定义">
                <div class="flex justify-between mb-2">
                  <span class="text-sm text-gray-500">模块状态（{{ seed.modules[key].states?.length || 0 }}）</span>
                  <el-button type="primary" size="small" @click="addState(key)">添加状态</el-button>
                </div>
                <el-table :data="seed.modules[key].states || []" size="small" border>
                  <el-table-column prop="status" label="状态值" width="140" />
                  <el-table-column prop="label" label="显示名" width="140" />
                  <el-table-column prop="tagType" label="标签类型" width="120" />
                  <el-table-column label="操作" width="120">
                    <template #default="{ row, $index }">
                      <el-button link type="danger" size="small" @click="removeState(key, $index)">删除</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-tab-pane>

              <!-- 流转 -->
              <el-tab-pane label="流转规则">
                <div class="flex justify-between mb-2">
                  <span class="text-sm text-gray-500">流转规则（{{ seed.modules[key].transitions?.length || 0 }}）</span>
                  <el-button type="primary" size="small" @click="addTransition(key)">添加流转</el-button>
                </div>
                <el-table :data="seed.modules[key].transitions || []" size="small" border>
                  <el-table-column prop="from" label="源状态" width="120" />
                  <el-table-column prop="to" label="目标状态" width="120" />
                  <el-table-column prop="action" label="动作" width="120" />
                  <el-table-column label="退回" width="90">
                    <template #default="{ row }">
                      <el-tag v-if="row.rollback" type="warning" size="small">退回</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="120">
                    <template #default="{ row, $index }">
                      <el-button link type="danger" size="small" @click="removeTransition(key, $index)">删除</el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </el-tab-pane>

              <!-- 项目种子 -->
              <el-tab-pane label="项目种子">
                <div class="text-sm text-gray-500 mb-2">项目引用（{{ seed.modules[key].projects?.length || 0 }}）</div>
                <el-table :data="seed.modules[key].projects || []" size="small" border>
                  <el-table-column prop="projectCode" label="项目编码" width="120" />
                  <el-table-column prop="name" label="项目名" width="160" />
                  <el-table-column label="实体数" width="90">
                    <template #default="{ row }">
                      <el-tag size="small">{{ entityCount(row.entities) }}</el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="实体预览" min-width="200">
                    <template #default="{ row }">
                      <span class="text-xs text-gray-500">{{ entityPreview(row.entities) }}</span>
                    </template>
                  </el-table-column>
                </el-table>
              </el-tab-pane>
            </el-tabs>
          </div>
        </el-collapse-item>
      </el-collapse>
    </el-card>

    <!-- 添加模块弹窗 -->
    <el-dialog v-model="addModuleVisible" title="添加模块" width="480px">
      <el-form :model="newModule" label-width="90px">
        <el-form-item label="模块编码">
          <el-select v-model="newModule.key" filterable allow-create placeholder="选择或输入模块编码" style="width: 100%">
            <el-option v-for="m in availableModules" :key="m.key" :label="m.label" :value="m.key" />
          </el-select>
        </el-form-item>
        <el-form-item label="实体类型">
          <el-input v-model="newModule.entityType" placeholder="如 task / requirement / risk" />
        </el-form-item>
        <el-form-item label="显示名">
          <el-input v-model="newModule.name" placeholder="如 进度管理" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addModuleVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmAddModule">添加</el-button>
      </template>
    </el-dialog>

    <!-- 字段编辑弹窗 -->
    <el-dialog v-model="fieldDialogVisible" :title="fieldDialogTitle" width="560px">
      <el-form :model="fieldForm" label-width="100px">
        <el-form-item label="字段键">
          <el-input v-model="fieldForm.key" placeholder="如 priority" />
        </el-form-item>
        <el-form-item label="显示名">
          <el-input v-model="fieldForm.label" placeholder="如 优先级" />
        </el-form-item>
        <el-form-item label="数据类型">
          <el-select v-model="fieldForm.dataType" style="width: 100%">
            <el-option v-for="t in dataTypes" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="选项" v-if="fieldForm.dataType === 'enum'">
          <el-select v-model="fieldForm.options" multiple filterable allow-create default-first-option placeholder="输入选项后回车" style="width: 100%">
            <el-option v-for="o in fieldForm.options" :key="o" :label="o" :value="o" />
          </el-select>
        </el-form-item>
        <el-form-item label="默认值">
          <el-input v-model="fieldForm.default" placeholder="默认值" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="fieldDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmField">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPackage, getPackageSeedStruct, updatePackageSeedStruct, publishPackage } from '@/api/pmocker/config'

defineOptions({ name: 'PmockerConfigPackageEditor' })

const router = useRouter()
const route = useRoute()
const pkgId = Number(route.query.id)

const pkg = ref({ name: '', status: '', version: 0 })
const seed = ref({ name: '', description: '', modules: {} })
const loading = ref(false)
const activeModules = ref([])

const moduleKeys = computed(() => Object.keys(seed.value.modules || {}))
const moduleNames = {
  eps: '组织EPS', requirement: '需求管理', schedule: '进度管理', risk: '风险管理',
  issue: '问题管理', change: '变更管理', deliverable: '交付物管理', scope: '范围管理',
  cost: '成本管理', team: '团队管理'
}
const moduleIcons = {
  eps: 'lucide:git-branch', requirement: 'lucide:file-text', schedule: 'lucide:calendar',
  risk: 'lucide:alert-triangle', issue: 'lucide:message-circle', change: 'lucide:refresh-cw',
  deliverable: 'lucide:package', scope: 'lucide:list-tree', cost: 'lucide:wallet', team: 'lucide:users'
}
const dataTypes = ['string', 'text', 'int', 'decimal', 'date', 'datetime', 'bool', 'enum', 'ref', 'json', 'user']

const moduleName = (key) => seed.value.modules[key]?.name || moduleNames[key] || key
const moduleIcon = (key) => moduleIcons[key] || 'lucide:box'

const entityCount = (entities) => {
  if (!entities) return 0
  return Object.values(entities).reduce((sum, arr) => sum + (arr?.length || 0), 0)
}
const entityPreview = (entities) => {
  if (!entities) return ''
  return Object.entries(entities).map(([et, arr]) => `${et}:${arr?.length || 0}`).join(', ')
}

// 可用模块（未添加的）
const availableModules = computed(() => {
  const all = Object.keys(moduleNames).map(k => ({ key: k, label: moduleNames[k] }))
  return all.filter(m => !(m.key in (seed.value.modules || {})))
})

const loadDetail = async () => {
  if (!pkgId) { ElMessage.error('缺少配置包ID'); return }
  loading.value = true
  const pkgRes = await getPackage(pkgId)
  if (pkgRes.code === 0) pkg.value = pkgRes.data
  const seedRes = await getPackageSeedStruct(pkgId)
  if (seedRes.code === 0) {
    seed.value = seedRes.data || { name: '', description: '', modules: {} }
    // 默认展开第一个模块
    const keys = Object.keys(seed.value.modules || {})
    if (keys.length) activeModules.value = [keys[0]]
  }
  loading.value = false
}

const saveAll = async () => {
  const res = await updatePackageSeedStruct(pkgId, seed.value)
  if (res.code === 0) { ElMessage.success('已保存'); loadDetail() }
}

const handlePublish = async () => {
  await ElMessageBox.confirm('确认发布？发布将自动同步字段/状态/种子到数据库并生成版本快照。', '发布确认', { type: 'warning' })
  const res = await publishPackage(pkgId)
  if (res.code === 0) { ElMessage.success('发布成功'); loadDetail() }
}

// 添加模块
const addModuleVisible = ref(false)
const newModule = ref({ key: '', entityType: '', name: '' })
const openAddModule = () => {
  newModule.value = { key: '', entityType: '', name: '' }
  addModuleVisible.value = true
}
const confirmAddModule = () => {
  if (!newModule.value.key) { ElMessage.warning('请选择模块编码'); return }
  seed.value.modules[newModule.value.key] = {
    entityType: newModule.value.entityType || newModule.value.key,
    name: newModule.value.name || moduleNames[newModule.value.key] || newModule.value.key,
    fields: [], states: [], transitions: [], projects: []
  }
  activeModules.value.push(newModule.value.key)
  addModuleVisible.value = false
  ElMessage.success('模块已添加')
}
const removeModule = async (key) => {
  await ElMessageBox.confirm(`确认删除模块「${key}」？`, '提示', { type: 'warning' })
  delete seed.value.modules[key]
  activeModules.value = activeModules.value.filter(k => k !== key)
  ElMessage.success('模块已删除')
}

// 字段
const fieldDialogVisible = ref(false)
const fieldDialogTitle = ref('')
const fieldForm = ref({})
let fieldModuleKey = ''
let fieldIndex = -1

const openAddField = (key) => {
  fieldModuleKey = key
  fieldIndex = -1
  fieldForm.value = { key: '', label: '', dataType: 'string', options: [], default: '' }
  fieldDialogTitle.value = '添加字段'
  fieldDialogVisible.value = true
}
const openEditField = (key, idx) => {
  fieldModuleKey = key
  fieldIndex = idx
  const f = seed.value.modules[key].fields[idx]
  fieldForm.value = { ...f, options: f.options || [] }
  fieldDialogTitle.value = '编辑字段'
  fieldDialogVisible.value = true
}
const confirmField = () => {
  if (!fieldForm.value.key) { ElMessage.warning('字段键不能为空'); return }
  const field = { ...fieldForm.value }
  if (field.dataType !== 'enum') delete field.options
  const fields = seed.value.modules[fieldModuleKey].fields
  if (fieldIndex >= 0) fields[fieldIndex] = field
  else fields.push(field)
  fieldDialogVisible.value = false
  ElMessage.success(fieldIndex >= 0 ? '字段已更新' : '字段已添加')
}
const removeField = (key, idx) => {
  seed.value.modules[key].fields.splice(idx, 1)
  ElMessage.success('字段已删除')
}

// 状态
const addState = (key) => {
  seed.value.modules[key].states.push({ status: '', label: '', tagType: 'info' })
}
const removeState = (key, idx) => {
  seed.value.modules[key].states.splice(idx, 1)
}

// 流转
const addTransition = (key) => {
  seed.value.modules[key].transitions.push({ from: '', to: '', action: '', rollback: false })
}
const removeTransition = (key, idx) => {
  seed.value.modules[key].transitions.splice(idx, 1)
}

const goBack = () => router.push({ name: 'pmockerConfigPackageList' })
const statusLabel = (s) => ({ draft: '草稿', reviewing: '评审中', published: '已发布', archived: '已归档' }[s] || s)
const statusTag = (s) => ({ draft: 'info', reviewing: 'warning', published: 'success', archived: '' }[s] || 'info')

onMounted(loadDetail)
</script>
