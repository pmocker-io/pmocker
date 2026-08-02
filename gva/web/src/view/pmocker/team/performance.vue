<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增绩效评估
        </el-button>
      </div>
      <el-table :data="tableData" row-key="id">
        <el-table-column label="评估周期" width="120">
          <template #default="{ row }">{{ row.attrs?.review_period }}</template>
        </el-table-column>
        <el-table-column label="评估类型" width="100">
          <template #default="{ row }">{{ row.attrs?.review_type }}</template>
        </el-table-column>
        <el-table-column label="评级" width="100">
          <template #default="{ row }">
            <el-tag :type="ratingType(row.attrs?.rating)">{{ row.attrs?.rating }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="评分" width="80">
          <template #default="{ row }">{{ row.attrs?.score }}</template>
        </el-table-column>
        <el-table-column label="评估日期" width="120">
          <template #default="{ row }">{{ row.attrs?.review_date }}</template>
        </el-table-column>
        <el-table-column label="下次评估" width="120">
          <template #default="{ row }">{{ row.attrs?.next_review_date }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" prop="status">
          <template #default="{ row }">
            <el-tag :type="reviewStatusType(row.attrs?.status || row.status)">{{ reviewStatusLabel(row.attrs?.status || row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">编辑</el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="800px" @closed="resetForm">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="名称" prop="title" :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]">
          <el-input v-model="form.title" placeholder="请输入评估名称" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="performance_review" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listPerformance, createPerformance, updatePerformance, deletePerformance } from '@/api/pmocker/team'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerTeamPerformance' })

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const saving = ref(false)
const editingId = ref(null)

const form = reactive({ title: '', status: 'draft', attrs: {} })

const reviewStatusMap = {
  draft: { label: '草稿', type: 'info' },
  in_review: { label: '评审中', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  appealed: { label: '申诉中', type: 'danger' }
}
const reviewStatusLabel = (s) => reviewStatusMap[s]?.label || s
const reviewStatusType = (s) => reviewStatusMap[s]?.type || 'info'

const ratingMap = {
  excellent: 'success', good: 'success', satisfactory: '',
  needs_improvement: 'warning', unsatisfactory: 'danger'
}
const ratingType = (s) => ratingMap[s] || 'info'

const getTableData = async () => {
  const res = await listPerformance({ page: page.value, pageSize: pageSize.value })
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const openDialog = (row) => {
  if (row) {
    editingId.value = row.id
    dialogTitle.value = '编辑绩效评估'
    Object.assign(form, { title: row.title, status: row.status, attrs: { ...row.attrs } })
  } else {
    editingId.value = null
    dialogTitle.value = '新增绩效评估'
    Object.assign(form, { title: '', status: 'draft', attrs: {} })
  }
  dialogVisible.value = true
}

const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, { title: '', status: 'draft', attrs: {} })
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingId.value) {
      await updatePerformance({ id: editingId.value, title: form.title, status: form.status, entity_type: 'performance_review', attrs: form.attrs })
    } else {
      await createPerformance({ title: form.title, attrs: form.attrs })
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    getTableData()
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该绩效评估?', '提示', { type: 'warning' }).then(async () => {
    await deletePerformance({ id: row.id })
    ElMessage.success('删除成功')
    getTableData()
  }).catch(() => {})
}

getTableData()
</script>
