<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="变更标题">
          <el-input v-model="searchInfo.keyword" placeholder="请输入" />
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="searchInfo.priority" clearable>
            <el-option label="紧急" value="urgent" />
            <el-option label="高" value="high" />
            <el-option label="中" value="medium" />
            <el-option label="低" value="low" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog()">
          <svg-icon icon="lucide:plus" /> 新增变更请求
        </el-button>
      </div>

      <VerticalTabLayout
        :active-tab="activeTab"
        :tabs="tabs"
        @tab-change="switchTab"
      >
        <template #toolbar>
          <div class="batch-actions">
            <el-button v-for="action in (currentGroup?.actions || [])" :key="action.label"
              :type="action.type" size="small"
              :disabled="selectedRows.length === 0"
              @click="handleBatchAction(action)">
              {{ action.label }}
            </el-button>
          </div>
        </template>
        <el-table :data="currentItems" row-key="id" size="small"
          @selection-change="onSelectionChange">
          <el-table-column type="selection" width="40" />
          <el-table-column label="ID" prop="id" width="70" />
          <el-table-column label="变更标题" prop="title" min-width="200" />
          <el-table-column label="优先级" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="priorityType(getAttr(row, 'priority'))">{{ priorityLabel(getAttr(row, 'priority')) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="变更性质" width="110">
            <template #default="{ row }">
              {{ getAttr(row, 'change_nature') }}
            </template>
          </el-table-column>
          <el-table-column label="紧急变更" width="100">
            <template #default="{ row }">
              <el-tag v-if="getAttr(row, 'is_emergency')" type="danger">紧急</el-tag>
              <el-tag v-else type="info">否</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="受影响基线" width="160">
            <template #default="{ row }">
              {{ formatBaselines(getAttr(row, 'affected_baselines')) }}
            </template>
          </el-table-column>
          <el-table-column label="提出人" width="120">
            <template #default="{ row }">
              {{ getAttr(row, 'proposer_name') || row.createdByName || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="负责人" width="120">
            <template #default="{ row }">
              {{ getAttr(row, 'assignee_name') || row.ownerName || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="创建时间" prop="CreatedAt" width="180">
            <template #default="{ row }">{{ formatDate(row.CreatedAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="openDialog(row)">编辑</el-button>
              <el-button type="primary" link size="small" @click="showAuditLogs(row.id)">审计</el-button>
              <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="currentItems.length === 0" description="暂无数据" />
      </VerticalTabLayout>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="800px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="名称" prop="title" :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]">
          <el-input v-model="form.title" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="change_request" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="logDrawerVisible" title="变更审计追溯" size="50%">
      <el-table :data="changeLogs" border stripe size="small">
        <el-table-column prop="fieldKey" label="字段" width="140" />
        <el-table-column prop="oldValue" label="旧值" show-overflow-tooltip />
        <el-table-column prop="newValue" label="新值" show-overflow-tooltip />
        <el-table-column prop="changedBy" label="变更人" width="100" />
        <el-table-column prop="CreatedAt" label="时间" width="170" />
      </el-table>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getChangeList, createChange, updateChange, deleteChange, analyzeChange, ccbReviewChange, approveChange, rejectChange, implementChange, verifyChange, closeChange } from '@/api/pmocker/change'
import { listChangeLogs } from '@/api/pmocker/changeLog'
import DynamicForm from '../components/DynamicForm.vue'
import ProjectSelector from '../components/ProjectSelector.vue'
import VerticalTabLayout from '../components/VerticalTabLayout.vue'
import { useProjectStore } from '@/pinia'
import { getTransitions } from '../components/statusTransitions.js'

defineOptions({ name: 'PmockerChangeList' })

const projectStore = useProjectStore()
const onProjectChange = () => { getTableData() }

const searchInfo = ref({})
const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const selectedRows = ref([])

const form = reactive({ id: null, title: '', status: 'submitted', attrs: {} })

const getAttr = (row, key) => (row.attrs && row.attrs[key] !== undefined ? row.attrs[key] : (row[key] || ''))

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : ''
const priorityType = (p) => ({ urgent: 'danger', high: 'warning', medium: 'info', low: 'info' }[p] || 'info')
const priorityLabel = (p) => ({ urgent: '紧急', high: '高', medium: '中', low: '低' }[p] || p)

const formatBaselines = (val) => {
  if (!val) return ''
  if (Array.isArray(val)) return val.join(', ')
  if (typeof val === 'object') return JSON.stringify(val)
  return val
}

const activeTab = ref('draft')
const transitionConfig = getTransitions('change_request')

const tabs = computed(() => transitionConfig.map(group => ({
  name: group.status,
  label: group.label,
  count: tableData.value.filter(item => item.status === group.status).length,
})))

const currentGroup = computed(() => transitionConfig.find(group => group.status === activeTab.value))

const currentItems = computed(() => tableData.value.filter(item => item.status === activeTab.value))

// API 函数映射
const apiMap = { analyzeChange, ccbReviewChange, approveChange, rejectChange, implementChange, verifyChange, closeChange, updateChange }

const getTableData = async () => {
  const params = { page: 1, pageSize: 999, projectId: projectStore.projectId, ...searchInfo.value }
  const res = await getChangeList(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const onSubmit = () => { getTableData() }
const onReset = () => { searchInfo.value = {}; getTableData() }

const onSelectionChange = (selection) => {
  selectedRows.value = selection
}

const switchTab = (name) => {
  activeTab.value = name
  selectedRows.value = []
}

// 批量状态流转
const handleBatchAction = async (action) => {
  const selected = selectedRows.value
  if (selected.length === 0) return
  try {
    await ElMessageBox.confirm(`确认将选中的 ${selected.length} 条记录执行「${action.label}」操作？`, '提示', { type: 'warning' })
    const apiFn = apiMap[action.apiFn]
    const promises = selected.map(row => {
      if (action.apiFn === 'updateChange') {
        return apiFn({ id: row.id, title: row.title, status: action.target, attrs: row.attrs || {} })
      }
      return apiFn({ id: row.id })
    })
    await Promise.all(promises)
    ElMessage.success(`成功${action.label} ${selected.length} 条记录`)
    selectedRows.value = []
    getTableData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

const resetForm = () => {
  Object.assign(form, { id: null, title: '', status: 'submitted', attrs: {} })
}

const openDialog = (row) => {
  resetForm()
  if (row) {
    dialogType.value = 'edit'
    dialogTitle.value = '编辑变更请求'
    form.id = row.id
    form.title = row.title
    form.status = row.status || 'submitted'
    form.attrs = row.attrs ? { ...row.attrs } : {}
    // 兼容旧数据：把顶层字段合并到 attrs
    if (row.priority && form.attrs.priority === undefined) form.attrs.priority = row.priority
    if (row.content && form.attrs.content === undefined) form.attrs.content = row.content
    if (row.reason && form.attrs.reason === undefined) form.attrs.reason = row.reason
  } else {
    dialogType.value = 'add'
    dialogTitle.value = '新增变更请求'
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const api = dialogType.value === 'add' ? createChange : updateChange
    const res = await api({ ...form, projectId: projectStore.projectId })
    if (res.code === 0) {
      ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
      dialogVisible.value = false
      getTableData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该变更请求吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteChange({ id: row.id })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

const logDrawerVisible = ref(false)
const changeLogs = ref([])
const showAuditLogs = async (entityId) => {
  const res = await listChangeLogs({ entityId })
  if (res.code === 0) changeLogs.value = res.data || []
  logDrawerVisible.value = true
}

getTableData()
</script>

<style scoped>
.batch-actions {
  display: flex;
  gap: 8px;
}
</style>
