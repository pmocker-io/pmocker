<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="名称">
          <el-input v-model="searchInfo.keyword" placeholder="请输入" />
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
          <svg-icon icon="lucide:plus" /> 新增交付物
        </el-button>
      </div>

      <!-- 按状态分组展示 -->
      <div v-for="group in groupedData" :key="group.status" class="status-group">
        <div class="group-header">
          <div class="group-title">
            <el-tag :type="group.tagType">{{ group.label }}</el-tag>
            <span class="count">({{ group.items.length }})</span>
          </div>
          <div class="group-actions" v-if="group.actions.length > 0">
            <el-button v-for="action in group.actions" :key="action.label"
              :type="action.type" size="small"
              :disabled="!selectedMap[group.status] || selectedMap[group.status].length === 0"
              @click="handleBatchAction(group, action)">
              {{ action.label }}
            </el-button>
          </div>
        </div>
        <el-table :data="group.items" row-key="id" size="small"
          @selection-change="(val) => onSelectionChange(group.status, val)">
          <el-table-column type="selection" width="40" />
          <el-table-column label="ID" prop="id" width="70" />
          <el-table-column label="交付物名称" prop="title" min-width="200" />
          <el-table-column label="类型" width="100">
            <template #default="{ row }">
              <el-tag>{{ typeLabel(getAttr(row, 'type') || row.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="版本" width="80">
            <template #default="{ row }">
              {{ getAttr(row, 'version') || row.version }}
            </template>
          </el-table-column>
          <el-table-column label="评审状态" width="110">
            <template #default="{ row }">
              <el-tag :type="reviewStatusType(getAttr(row, 'review_status'))">{{ getAttr(row, 'review_status') }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="保密级别" width="120">
            <template #default="{ row }">
              <el-tag :type="securityType(getAttr(row, 'security_classification'))">{{ getAttr(row, 'security_classification') }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="缺陷数" width="90">
            <template #default="{ row }">
              {{ getAttr(row, 'defect_count') }}
            </template>
          </el-table-column>
          <el-table-column label="负责人" width="120">
            <template #default="{ row }">
              {{ getAttr(row, 'reviewer_name') || row.ownerName || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="openDialog(row)">编辑</el-button>
              <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <el-empty v-if="groupedData.length === 0" description="暂无数据" />
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="800px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="名称" prop="title" :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]">
          <el-input v-model="form.title" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="deliverable" />
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
import { getDeliverableList, createDeliverable, updateDeliverable, deleteDeliverable, submitDeliverableReview, acceptDeliverable, rejectDeliverable } from '@/api/pmocker/deliverable'
import DynamicForm from '../components/DynamicForm.vue'
import ProjectSelector from '../components/ProjectSelector.vue'
import { useProjectStore } from '@/pinia'
import { groupByStatus } from '../components/statusTransitions.js'

defineOptions({ name: 'PmockerDeliverableList' })

const projectStore = useProjectStore()
const onProjectChange = () => { getTableData() }

const searchInfo = ref({})
const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const selectedMap = ref({})

const form = reactive({ id: null, title: '', status: 'draft', attrs: {} })

const getAttr = (row, key) => (row.attrs && row.attrs[key] !== undefined ? row.attrs[key] : (row[key] || ''))

const typeLabel = (t) => ({ document: '文档', code: '代码', design: '设计稿', other: '其他' }[t] || t || '')
const reviewStatusType = (s) => ({ pending: 'info', in_review: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info')
const securityType = (s) => ({ public: 'info', internal: '', confidential: 'warning', secret: 'danger' }[s] || 'info')

// 按状态分组
const groupedData = computed(() => groupByStatus(tableData.value, 'deliverable'))

const getTableData = async () => {
  const params = { page: 1, pageSize: 999, projectId: projectStore.projectId, ...searchInfo.value }
  const res = await getDeliverableList(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const onSubmit = () => { getTableData() }
const onReset = () => { searchInfo.value = {}; getTableData() }

const onSelectionChange = (status, selection) => {
  selectedMap.value = { ...selectedMap.value, [status]: selection }
}

// API 函数映射
const apiMap = { submitDeliverableReview, acceptDeliverable, rejectDeliverable, updateDeliverable }

// 批量状态流转
const handleBatchAction = async (group, action) => {
  const selected = selectedMap.value[group.status] || []
  if (selected.length === 0) return

  try {
    await ElMessageBox.confirm(`确认将选中的 ${selected.length} 条记录执行「${action.label}」操作？`, '提示', { type: 'warning' })

    const apiFn = apiMap[action.apiFn]
    const promises = selected.map(row => {
      if (action.apiFn === 'updateDeliverable') {
        return apiFn({ id: row.id, title: row.title, status: action.target, attrs: row.attrs || {} })
      }
      return apiFn({ id: row.id })
    })
    await Promise.all(promises)
    ElMessage.success(`成功${action.label} ${selected.length} 条记录`)
    selectedMap.value[group.status] = []
    getTableData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

const resetForm = () => {
  Object.assign(form, { id: null, title: '', status: 'draft', attrs: {} })
}

const openDialog = (row) => {
  resetForm()
  if (row) {
    dialogType.value = 'edit'
    dialogTitle.value = '编辑交付物'
    form.id = row.id
    form.title = row.title
    form.status = row.status || 'draft'
    form.attrs = row.attrs ? { ...row.attrs } : {}
    // 兼容旧数据：把顶层字段合并到 attrs
    if (row.type && form.attrs.type === undefined) form.attrs.type = row.type
    if (row.version && form.attrs.version === undefined) form.attrs.version = row.version
    if (row.description && form.attrs.description === undefined) form.attrs.description = row.description
  } else {
    dialogType.value = 'add'
    dialogTitle.value = '新增交付物'
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const api = dialogType.value === 'add' ? createDeliverable : updateDeliverable
    const res = await api({ ...form, projectId: projectStore.projectId })
    if (res.code === 0) {
      ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
      dialogVisible.value = false
      getTableData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该交付物吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteDeliverable({ id: row.id })
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
.status-group {
  margin-bottom: 16px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  overflow: hidden;
}
.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
}
.group-title {
  display: flex;
  align-items: center;
  gap: 8px;
}
.count {
  color: #909399;
  font-size: 13px;
}
.group-actions {
  display: flex;
  gap: 8px;
}
</style>
