<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="需求名称">
          <el-input v-model="searchInfo.keyword" placeholder="请输入需求名称" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSubmit">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog('add')">
          <svg-icon icon="lucide:plus" /> 新增需求
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
          <el-table-column label="需求名称" prop="title" min-width="200" />
          <el-table-column label="优先级" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="priorityType(getAttr(row, 'priority'))">{{ getAttr(row, 'priority') }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="MoSCoW" width="120">
            <template #default="{ row }">
              {{ getAttr(row, 'moscow_priority') }}
            </template>
          </el-table-column>
          <el-table-column label="需求类型" width="120">
            <template #default="{ row }">
              {{ getAttr(row, 'requirement_type') }}
            </template>
          </el-table-column>
          <el-table-column label="故事点" width="100">
            <template #default="{ row }">
              {{ getAttr(row, 'story_points') }}
            </template>
          </el-table-column>
          <el-table-column label="提出人" width="120">
            <template #default="{ row }">
              {{ row.ownerName || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="创建时间" prop="CreatedAt" width="180">
            <template #default="{ row }">
              {{ formatDate(row.CreatedAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="openDialog('edit', row)">编辑</el-button>
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
        <DynamicForm v-model="form" entity-type="requirement" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getRequirementList,
  createRequirement,
  updateRequirement,
  deleteRequirement,
  submitRequirementReview,
  approveRequirement,
  rejectRequirement
} from '@/api/pmocker/requirement'
import DynamicForm from '../components/DynamicForm.vue'
import ProjectSelector from '../components/ProjectSelector.vue'
import VerticalTabLayout from '../components/VerticalTabLayout.vue'
import { useProjectStore } from '@/pinia'
import { getTransitions } from '../components/statusTransitions.js'

defineOptions({ name: 'PmockerRequirementList' })

const projectStore = useProjectStore()
const onProjectChange = () => { getTableData() }

const searchInfo = ref({})
const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const selectedRows = ref([])

const form = reactive({
  id: null,
  title: '',
  status: 'draft',
  attrs: {}
})

const getAttr = (row, key) => (row.attrs && row.attrs[key] !== undefined ? row.attrs[key] : (row[key] || ''))

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
}

const priorityType = (priority) => {
  const map = { high: 'danger', medium: 'warning', low: 'info' }
  return map[priority] || 'info'
}

const activeTab = ref('draft')
const transitionConfig = getTransitions('requirement')

const tabs = computed(() => transitionConfig.map(group => ({
  name: group.status,
  label: group.label,
  count: tableData.value.filter(item => item.status === group.status).length,
})))

const currentGroup = computed(() => transitionConfig.find(group => group.status === activeTab.value))

const currentItems = computed(() => tableData.value.filter(item => item.status === activeTab.value))

// API 函数映射
const apiMap = { submitRequirementReview, approveRequirement, rejectRequirement, updateRequirement }

const getTableData = async () => {
  const params = { page: 1, pageSize: 999, projectId: projectStore.projectId, ...searchInfo.value }
  const res = await getRequirementList(params)
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
      if (action.apiFn === 'updateRequirement') {
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
  Object.assign(form, { id: null, title: '', status: 'draft', attrs: {} })
}

const openDialog = (type, row) => {
  dialogType.value = type
  dialogTitle.value = type === 'add' ? '新增需求' : '编辑需求'
  resetForm()
  if (type === 'edit' && row) {
    form.id = row.id
    form.title = row.title
    form.status = row.status || 'draft'
    form.attrs = row.attrs ? { ...row.attrs } : {}
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const api = dialogType.value === 'add' ? createRequirement : updateRequirement
    const res = await api({ ...form, projectId: projectStore.projectId })
    if (res.code === 0) {
      ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
      dialogVisible.value = false
      getTableData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该需求吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteRequirement({ id: row.id })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

getTableData()
</script>

<style scoped>
.batch-actions {
  display: flex;
  gap: 8px;
}
</style>
