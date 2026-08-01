<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openTaskDialog">
          <svg-icon icon="lucide:plus" /> 新增任务
        </el-button>
      </div>
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="任务名称" prop="title" min-width="200" />
        <el-table-column label="开始日期" prop="startDate" width="120">
          <template #default="{ row }">{{ formatDate(row.startDate) }}</template>
        </el-table-column>
        <el-table-column label="结束日期" prop="endDate" width="120">
          <template #default="{ row }">{{ formatDate(row.endDate) }}</template>
        </el-table-column>
        <el-table-column label="工期" prop="duration" width="80" />
        <el-table-column label="进度" prop="progress" width="150">
          <template #default="{ row }">
            <el-progress :percentage="row.progress || 0" />
          </template>
        </el-table-column>
        <el-table-column label="状态" prop="status" width="100">
          <template #default="{ row }">
            <el-tag :type="taskStatusType(row.status)">{{ taskStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openTaskDialog(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="任务名称" prop="title">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="开始日期" prop="startDate">
          <el-date-picker v-model="form.startDate" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="结束日期" prop="endDate">
          <el-date-picker v-model="form.endDate" type="date" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="进度" prop="progress">
          <el-slider v-model="form.progress" :max="100" />
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
import { getScheduleTasks, createScheduleTask } from '@/api/pmocker/schedule'

defineOptions({ name: 'PmockerScheduleGantt' })

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, title: '', startDate: '', endDate: '', progress: 0 })

const rules = {
  title: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  startDate: [{ required: true, message: '请选择开始日期', trigger: 'change' }],
  endDate: [{ required: true, message: '请选择结束日期', trigger: 'change' }]
}

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleDateString('zh-CN') : ''

const taskStatusType = (status) => {
  const map = { pending: 'info', running: 'warning', done: 'success' }
  return map[status] || 'info'
}

const taskStatusLabel = (status) => {
  const map = { pending: '待开始', running: '进行中', done: '已完成' }
  return map[status] || status
}

const loadData = async () => {
  const res = await getScheduleTasks({})
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const openTaskDialog = (row) => {
  dialogType.value = row ? 'edit' : 'add'
  dialogTitle.value = row ? '编辑任务' : '新增任务'
  if (row) {
    Object.assign(form, row)
  } else {
    Object.assign(form, { ID: null, title: '', startDate: '', endDate: '', progress: 0 })
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await createScheduleTask(form)
    if (res.code === 0) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      loadData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该任务吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await createScheduleTask({ action: 'delete', ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadData()
      }
    })
    .catch(() => {})
}

loadData()
</script>
