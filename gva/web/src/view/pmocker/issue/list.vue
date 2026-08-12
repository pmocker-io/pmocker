<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="问题名称">
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
          <svg-icon icon="lucide:plus" /> 新增问题
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
          <el-table-column label="问题标题" prop="title" min-width="200" />
          <el-table-column label="优先级" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="priorityType(getAttr(row, 'priority'))">{{ priorityLabel(getAttr(row, 'priority')) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="严重性" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="severityType(getAttr(row, 'severity'))">{{ getAttr(row, 'severity') }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="负责人" width="100">
            <template #default="{ row }">
              {{ getAttr(row, 'assignee_name') || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="openDialog(row)">编辑</el-button>
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
        <DynamicForm v-model="form" entity-type="issue" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 分配对话框（批量分配） -->
    <el-dialog v-model="assignDialogVisible" title="分配负责人" width="400px">
      <p class="mb-3">将为选中的 {{ assignTargetIds.length }} 条问题分配负责人</p>
      <el-form label-width="80px">
        <el-form-item label="负责人">
          <el-select v-model="assignUserId" filterable clearable placeholder="请选择用户" class="w-full">
            <el-option
              v-for="u in userList"
              :key="u.ID"
              :label="u.nickName + ' (' + u.userName + ')'"
              :value="u.ID"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmAssign">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getIssueList, createIssue, updateIssue, deleteIssue, assignIssue, resolveIssue, closeIssue, reopenIssue } from '@/api/pmocker/issue'
import { getUserList } from '@/api/user'
import DynamicForm from '../components/DynamicForm.vue'
import ProjectSelector from '../components/ProjectSelector.vue'
import VerticalTabLayout from '../components/VerticalTabLayout.vue'
import { useProjectStore } from '@/pinia'
import { getTransitions } from '../components/statusTransitions.js'

defineOptions({ name: 'PmockerIssueList' })

const projectStore = useProjectStore()
const onProjectChange = () => { getTableData() }

const searchInfo = ref({})
const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')
const userList = ref([])
const assignDialogVisible = ref(false)
const assignUserId = ref(null)
const assignTargetIds = ref([])

const selectedRows = ref([])

const form = reactive({ id: null, title: '', status: 'open', attrs: {} })

const getAttr = (row, key) => (row.attrs && row.attrs[key] !== undefined ? row.attrs[key] : (row[key] || ''))

const priorityType = (p) => ({ urgent: 'danger', high: 'warning', medium: 'info', low: 'info' }[p] || 'info')
const priorityLabel = (p) => ({ urgent: '紧急', high: '高', medium: '中', low: '低' }[p] || p)
const severityType = (s) => ({ fatal: 'danger', critical: 'danger', major: 'warning', minor: 'info', trivial: 'info' }[s] || 'info')

const activeTab = ref('open')
const transitionConfig = getTransitions('issue')

const tabs = computed(() => transitionConfig.map(g => ({
  name: g.status,
  label: g.label,
  count: tableData.value.filter(item => item.status === g.status).length
})))

const currentGroup = computed(() => transitionConfig.find(g => g.status === activeTab.value) || { actions: [] })

const currentItems = computed(() => tableData.value.filter(item => item.status === activeTab.value))

const getTableData = async () => {
  const params = { page: 1, pageSize: 999, projectId: projectStore.projectId, ...searchInfo.value }
  const res = await getIssueList(params)
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

  // 分配操作需要先选择用户
  if (action.apiFn === 'assignIssue') {
    assignTargetIds.value = selected.map(r => r.id)
    assignUserId.value = null
    assignDialogVisible.value = true
    return
  }

  try {
    await ElMessageBox.confirm(`确认将选中的 ${selected.length} 条记录执行「${action.label}」操作？`, '提示', { type: 'warning' })

    const apiFn = apiMap[action.apiFn]
    const promises = selected.map(row => {
      if (action.apiFn === 'updateIssue') {
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

// API 函数映射
const apiMap = { assignIssue, updateIssue, resolveIssue, closeIssue, reopenIssue }

const confirmAssign = async () => {
  if (!assignUserId.value) {
    ElMessage.warning('请选择负责人')
    return
  }
  const promises = assignTargetIds.value.map(id => assignIssue({ id, assigneeId: assignUserId.value }))
  await Promise.all(promises)
  ElMessage.success('已分配')
  assignDialogVisible.value = false
  selectedRows.value = []
  getTableData()
}

const resetForm = () => {
  Object.assign(form, { id: null, title: '', status: 'open', attrs: {} })
}

const openDialog = (row) => {
  resetForm()
  if (row) {
    dialogType.value = 'edit'
    dialogTitle.value = '编辑问题'
    form.id = row.id
    form.title = row.title
    form.status = row.status || 'open'
    form.attrs = row.attrs ? { ...row.attrs } : {}
  } else {
    dialogType.value = 'add'
    dialogTitle.value = '新增问题'
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const api = dialogType.value === 'add' ? createIssue : updateIssue
    const res = await api({ ...form, projectId: projectStore.projectId })
    if (res.code === 0) {
      ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
      dialogVisible.value = false
      getTableData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该问题吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteIssue({ id: row.id })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

const loadUsers = async () => {
  try {
    const res = await getUserList({ page: 1, pageSize: 999 })
    if (res.code === 0) {
      userList.value = res.data.list || []
    }
  } catch (e) {
    console.error('loadUsers error:', e)
  }
}

onMounted(() => {
  loadUsers()
})

getTableData()
</script>

<style scoped>
.batch-actions {
  display: flex;
  gap: 8px;
}
</style>
