<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="变更标题">
          <el-input v-model="searchInfo.keyword" placeholder="请输入" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable>
            <el-option label="已提交" value="submitted" />
            <el-option label="分析中" value="analyzing" />
            <el-option label="CCB评审中" value="ccb_review" />
            <el-option label="已批准" value="approved" />
            <el-option label="已驳回" value="rejected" />
            <el-option label="实施中" value="implementing" />
            <el-option label="验证中" value="verifying" />
            <el-option label="已关闭" value="closed" />
          </el-select>
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
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="ID" prop="ID" width="80" />
        <el-table-column label="变更标题" prop="title" min-width="200" />
        <el-table-column label="优先级" width="80">
          <template #default="{ row }">
            <el-tag :type="priorityType(getAttr(row, 'priority'))">{{ priorityLabel(getAttr(row, 'priority')) }}</el-tag>
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
        <el-table-column label="状态" prop="status" width="120">
          <template #default="{ row }">
            <el-tag :type="changeStatusType(row.status)">{{ changeStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提出人" prop="submitterName" width="120" />
        <el-table-column label="创建时间" prop="CreatedAt" width="180">
          <template #default="{ row }">{{ formatDate(row.CreatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">编辑</el-button>
            <el-button v-if="row.status === 'submitted'" type="warning" link @click="handleAnalyze(row)">影响分析</el-button>
            <el-button type="primary" link size="small" @click="showAuditLogs(row.ID)">审计</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        class="gva-pagination"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="getTableData"
        @current-change="getTableData"
      />
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
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getChangeList, createChange, updateChange, deleteChange, analyzeChange } from '@/api/pmocker/change'
import { listChangeLogs } from '@/api/pmocker/changeLog'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerChangeList' })

const searchInfo = ref({})
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, title: '', status: 'submitted', attrs: {} })

const getAttr = (row, key) => (row.attrs && row.attrs[key] !== undefined ? row.attrs[key] : (row[key] || ''))

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : ''
const priorityType = (p) => ({ urgent: 'danger', high: 'warning', medium: 'info', low: 'info' }[p] || 'info')
const priorityLabel = (p) => ({ urgent: '紧急', high: '高', medium: '中', low: '低' }[p] || p)
const changeStatusType = (s) => ({
  submitted: 'info', analyzing: 'warning', ccb_review: 'warning',
  approved: 'success', rejected: 'danger', implementing: 'primary',
  verifying: 'primary', closed: ''
}[s] || 'info')
const changeStatusLabel = (s) => ({
  submitted: '已提交', analyzing: '分析中', ccb_review: 'CCB评审中',
  approved: '已批准', rejected: '已驳回', implementing: '实施中',
  verifying: '验证中', closed: '已关闭'
}[s] || s)

const formatBaselines = (val) => {
  if (!val) return ''
  if (Array.isArray(val)) return val.join(', ')
  if (typeof val === 'object') return JSON.stringify(val)
  return val
}

const getTableData = async () => {
  const params = { page: page.value, pageSize: pageSize.value, projectId: projectStore.projectId, ...searchInfo.value }
  const res = await getChangeList(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const onSubmit = () => { page.value = 1; getTableData() }
const onReset = () => { searchInfo.value = {}; page.value = 1; getTableData() }

const resetForm = () => {
  Object.assign(form, { ID: null, title: '', status: 'submitted', attrs: {} })
}

const openDialog = (row) => {
  resetForm()
  if (row) {
    dialogType.value = 'edit'
    dialogTitle.value = '编辑变更请求'
    form.ID = row.ID
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
      const res = await deleteChange({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

const handleAnalyze = async (row) => {
  const res = await analyzeChange({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('分析完成')
    getTableData()
  }
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
