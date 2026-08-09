<template>
  <div>
    <ProjectSelector @change="loadData" />
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog">
          <svg-icon icon="lucide:flag" /> 新增里程碑
        </el-button>
      </div>
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="里程碑" prop="title" min-width="200" />
        <el-table-column label="计划日期" prop="planDate" width="150">
          <template #default="{ row }">{{ formatDate(row.planDate) }}</template>
        </el-table-column>
        <el-table-column label="实际日期" prop="actualDate" width="150">
          <template #default="{ row }">{{ row.actualDate ? formatDate(row.actualDate) : '未完成' }}</template>
        </el-table-column>
        <el-table-column label="状态" prop="status" width="120">
          <template #default="{ row }">
            <el-tag :type="milestoneStatusType(row.status)">{{ milestoneStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" title="新增里程碑" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="title">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="计划日期" prop="planDate">
          <el-date-picker v-model="form.planDate" type="date" value-format="YYYY-MM-DD" />
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
import { getScheduleMilestones, createScheduleMilestone } from '@/api/pmocker/schedule'
import { useProjectStore } from '@/pinia'
import ProjectSelector from '../components/ProjectSelector.vue'

defineOptions({ name: 'PmockerScheduleMilestone' })

const projectStore = useProjectStore()

const tableData = ref([])
const dialogVisible = ref(false)
const formRef = ref(null)

const form = reactive({ title: '', planDate: '' })
const rules = {
  title: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  planDate: [{ required: true, message: '请选择日期', trigger: 'change' }]
}

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleDateString('zh-CN') : ''

const milestoneStatusType = (status) => {
  const map = { pending: 'info', achieved: 'success', delayed: 'danger' }
  return map[status] || 'info'
}

const milestoneStatusLabel = (status) => {
  const map = { pending: '待达成', achieved: '已达成', delayed: '已延期' }
  return map[status] || status
}

const loadData = async () => {
  const res = await getScheduleMilestones({ projectId: projectStore.projectId })
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const openDialog = () => {
  Object.assign(form, { title: '', planDate: '' })
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await createScheduleMilestone({ ...form, projectId: projectStore.projectId })
    if (res.code === 0) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      loadData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该里程碑吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await createScheduleMilestone({ action: 'delete', id: row.id })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadData()
      }
    })
    .catch(() => {})
}

loadData()
</script>
