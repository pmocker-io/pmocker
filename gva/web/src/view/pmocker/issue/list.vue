<template>
  <div>
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
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable>
            <el-option label="待处理" value="open" />
            <el-option label="处理中" value="in_progress" />
            <el-option label="已解决" value="resolved" />
            <el-option label="已关闭" value="closed" />
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
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="ID" prop="ID" width="80" />
        <el-table-column label="问题标题" prop="title" min-width="200" />
        <el-table-column label="优先级" width="100">
          <template #default="{ row }">
            <el-tag :type="priorityType(getAttr(row, 'priority'))">{{ priorityLabel(getAttr(row, 'priority')) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="严重性" width="100">
          <template #default="{ row }">
            <el-tag :type="severityType(getAttr(row, 'severity'))">{{ getAttr(row, 'severity') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="120">
          <template #default="{ row }">
            {{ getAttr(row, 'category') }}
          </template>
        </el-table-column>
        <el-table-column label="解决类型" width="120">
          <template #default="{ row }">
            {{ getAttr(row, 'resolution_type') }}
          </template>
        </el-table-column>
        <el-table-column label="状态" prop="status" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负责人" prop="assigneeName" width="120" />
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">编辑</el-button>
            <el-button v-if="row.status === 'open'" type="warning" link @click="handleAssign(row)">分配</el-button>
            <el-button v-if="row.status === 'in_progress'" type="success" link @click="handleResolve(row)">解决</el-button>
            <el-button v-if="row.status === 'resolved'" type="info" link @click="handleClose(row)">关闭</el-button>
            <el-button v-if="row.status === 'closed'" link @click="handleReopen(row)">重开</el-button>
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
        <DynamicForm v-model="form" entity-type="issue" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getIssueList, createIssue, updateIssue, deleteIssue, assignIssue, resolveIssue, closeIssue, reopenIssue } from '@/api/pmocker/issue'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerIssueList' })

const searchInfo = ref({})
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, title: '', status: 'open', attrs: {} })

const getAttr = (row, key) => (row.attrs && row.attrs[key] !== undefined ? row.attrs[key] : (row[key] || ''))

const priorityType = (p) => ({ urgent: 'danger', high: 'warning', medium: 'info', low: 'info' }[p] || 'info')
const priorityLabel = (p) => ({ urgent: '紧急', high: '高', medium: '中', low: '低' }[p] || p)
const severityType = (s) => ({ fatal: 'danger', critical: 'danger', major: 'warning', minor: 'info', trivial: 'info' }[s] || 'info')
const statusType = (s) => ({ open: 'info', in_progress: 'warning', resolved: 'success', closed: '' }[s] || 'info')
const statusLabel = (s) => ({ open: '待处理', in_progress: '处理中', resolved: '已解决', closed: '已关闭' }[s] || s)

const getTableData = async () => {
  const params = { page: page.value, pageSize: pageSize.value, ...searchInfo.value }
  const res = await getIssueList(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const onSubmit = () => { page.value = 1; getTableData() }
const onReset = () => { searchInfo.value = {}; page.value = 1; getTableData() }

const resetForm = () => {
  Object.assign(form, { ID: null, title: '', status: 'open', attrs: {} })
}

const openDialog = (row) => {
  resetForm()
  if (row) {
    dialogType.value = 'edit'
    dialogTitle.value = '编辑问题'
    form.ID = row.ID
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
    const res = await api(form)
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
      const res = await deleteIssue({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

const handleAssign = async (row) => {
  ElMessageBox.prompt('请输入负责人ID', '分配', { inputPlaceholder: '用户ID' })
    .then(async ({ value }) => {
      const res = await assignIssue({ ID: row.ID, assigneeId: Number(value) })
      if (res.code === 0) {
        ElMessage.success('已分配')
        getTableData()
      }
    })
    .catch(() => {})
}

const handleResolve = async (row) => {
  const res = await resolveIssue({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('已解决')
    getTableData()
  }
}

const handleClose = async (row) => {
  const res = await closeIssue({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('已关闭')
    getTableData()
  }
}

const handleReopen = async (row) => {
  const res = await reopenIssue({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('已重开')
    getTableData()
  }
}

getTableData()
</script>
