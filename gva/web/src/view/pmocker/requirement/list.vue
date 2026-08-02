<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="需求名称">
          <el-input v-model="searchInfo.keyword" placeholder="请输入需求名称" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" placeholder="请选择状态" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="评审中" value="reviewing" />
            <el-option label="已批准" value="approved" />
            <el-option label="已驳回" value="rejected" />
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
        <el-button type="primary" @click="openDialog('add')">
          <svg-icon icon="lucide:plus" /> 新增需求
        </el-button>
      </div>
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="ID" prop="ID" width="80" />
        <el-table-column label="需求名称" prop="title" min-width="200" />
        <el-table-column label="优先级" width="100">
          <template #default="{ row }">
            <el-tag :type="priorityType(getAttr(row, 'priority'))">{{ getAttr(row, 'priority') }}</el-tag>
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
        <el-table-column label="状态" prop="status" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提出人" prop="owner_name" width="120" />
        <el-table-column label="创建时间" prop="CreatedAt" width="180">
          <template #default="{ row }">
            {{ formatDate(row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog('edit', row)">编辑</el-button>
            <el-button v-if="row.status === 'draft'" type="warning" link @click="handleSubmitReview(row)">提交评审</el-button>
            <el-button v-if="row.status === 'reviewing'" type="success" link @click="handleApprove(row)">批准</el-button>
            <el-button v-if="row.status === 'reviewing'" type="danger" link @click="handleReject(row)">驳回</el-button>
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
import { ref, reactive } from 'vue'
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

defineOptions({ name: 'PmockerRequirementList' })

const searchInfo = ref({})
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({
  ID: null,
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

const statusType = (status) => {
  const map = { draft: 'info', reviewing: 'warning', approved: 'success', rejected: 'danger' }
  return map[status] || 'info'
}

const statusLabel = (status) => {
  const map = { draft: '草稿', reviewing: '评审中', approved: '已批准', rejected: '已驳回' }
  return map[status] || status
}

const getTableData = async () => {
  const params = { page: page.value, pageSize: pageSize.value, ...searchInfo.value }
  const res = await getRequirementList(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const onSubmit = () => {
  page.value = 1
  getTableData()
}

const onReset = () => {
  searchInfo.value = {}
  page.value = 1
  getTableData()
}

const resetForm = () => {
  Object.assign(form, { ID: null, title: '', status: 'draft', attrs: {} })
}

const openDialog = (type, row) => {
  dialogType.value = type
  dialogTitle.value = type === 'add' ? '新增需求' : '编辑需求'
  resetForm()
  if (type === 'edit' && row) {
    form.ID = row.ID
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
    const res = await api(form)
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
      const res = await deleteRequirement({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

const handleSubmitReview = async (row) => {
  const res = await submitRequirementReview({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('已提交评审')
    getTableData()
  }
}

const handleApprove = async (row) => {
  const res = await approveRequirement({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('已批准')
    getTableData()
  }
}

const handleReject = (row) => {
  ElMessageBox.prompt('请输入驳回原因', '驳回', { type: 'warning' })
    .then(async ({ value }) => {
      const res = await rejectRequirement({ ID: row.ID, reason: value })
      if (res.code === 0) {
        ElMessage.success('已驳回')
        getTableData()
      }
    })
    .catch(() => {})
}

getTableData()
</script>
