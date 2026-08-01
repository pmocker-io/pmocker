<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="名称">
          <el-input v-model="searchInfo.keyword" placeholder="请输入" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable>
            <el-option label="草稿" value="draft" />
            <el-option label="评审中" value="reviewing" />
            <el-option label="已接收" value="accepted" />
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
        <el-button type="primary" @click="openDialog">
          <svg-icon icon="lucide:plus" /> 新增交付物
        </el-button>
      </div>
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="ID" prop="ID" width="80" />
        <el-table-column label="交付物名称" prop="title" min-width="200" />
        <el-table-column label="类型" prop="type" width="120">
          <template #default="{ row }"><el-tag>{{ typeLabel(row.type) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="版本" prop="version" width="80" />
        <el-table-column label="状态" prop="status" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负责人" prop="ownerName" width="120" />
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">编辑</el-button>
            <el-button v-if="row.status === 'draft'" type="warning" link @click="handleSubmit(row)">提交评审</el-button>
            <el-button v-if="row.status === 'reviewing'" type="success" link @click="handleAccept(row)">接收</el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="title">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type">
            <el-option label="文档" value="document" />
            <el-option label="代码" value="code" />
            <el-option label="设计稿" value="design" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="4" />
        </el-form-item>
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
import { getDeliverableList, createDeliverable, deleteDeliverable, submitDeliverableReview, acceptDeliverable, rejectDeliverable } from '@/api/pmocker/deliverable'

defineOptions({ name: 'PmockerDeliverableList' })

const searchInfo = ref({})
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, title: '', type: 'document', description: '' })
const rules = {
  title: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const typeLabel = (t) => ({ document: '文档', code: '代码', design: '设计稿', other: '其他' }[t] || t)
const statusType = (s) => ({ draft: 'info', reviewing: 'warning', accepted: 'success', rejected: 'danger' }[s] || 'info')
const statusLabel = (s) => ({ draft: '草稿', reviewing: '评审中', accepted: '已接收', rejected: '已驳回' }[s] || s)

const getTableData = async () => {
  const params = { page: page.value, pageSize: pageSize.value, ...searchInfo.value }
  const res = await getDeliverableList(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const onSubmit = () => { page.value = 1; getTableData() }
const onReset = () => { searchInfo.value = {}; page.value = 1; getTableData() }

const openDialog = (row) => {
  dialogType.value = row ? 'edit' : 'add'
  dialogTitle.value = row ? '编辑交付物' : '新增交付物'
  if (row) {
    Object.assign(form, row)
  } else {
    Object.assign(form, { ID: null, title: '', type: 'document', description: '' })
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await createDeliverable(form)
    if (res.code === 0) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      getTableData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该交付物吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteDeliverable({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

const handleSubmit = async (row) => {
  const res = await submitDeliverableReview({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('已提交评审')
    getTableData()
  }
}

const handleAccept = async (row) => {
  const res = await acceptDeliverable({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('已接收')
    getTableData()
  }
}

const handleReject = async (row) => {
  ElMessageBox.prompt('请输入驳回原因', '驳回', { type: 'warning' })
    .then(async ({ value }) => {
      const res = await rejectDeliverable({ ID: row.ID, reason: value })
      if (res.code === 0) {
        ElMessage.success('已驳回')
        getTableData()
      }
    })
    .catch(() => {})
}

getTableData()
</script>
