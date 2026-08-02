<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openTaskDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增任务
        </el-button>
      </div>
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="任务名称" prop="title" min-width="200" />
        <el-table-column label="开始日期" width="120">
          <template #default="{ row }">{{ formatDate(row.attrs?.start_date) }}</template>
        </el-table-column>
        <el-table-column label="结束日期" width="120">
          <template #default="{ row }">{{ formatDate(row.attrs?.end_date) }}</template>
        </el-table-column>
        <el-table-column label="工期" width="80">
          <template #default="{ row }">{{ row.attrs?.duration ?? '' }}</template>
        </el-table-column>
        <el-table-column label="进度" width="150">
          <template #default="{ row }">
            <el-progress :percentage="row.attrs?.progress || 0" />
          </template>
        </el-table-column>
        <el-table-column label="关键路径" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.attrs?.is_critical_path" type="danger">关键</el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="总浮动" width="100">
          <template #default="{ row }">{{ row.attrs?.total_float ?? '' }}</template>
        </el-table-column>
        <el-table-column label="依赖类型" width="120">
          <template #default="{ row }">{{ dependencyLabel(row.attrs?.dependency_type) }}</template>
        </el-table-column>
        <el-table-column label="状态" prop="status" width="110">
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="800px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="名称" prop="title" :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]">
          <el-input v-model="form.title" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="task" />
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
  getScheduleTasks,
  createScheduleTask,
  updateTask,
  deleteTask
} from '@/api/pmocker/schedule'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerScheduleGantt' })

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, title: '', status: 'planned', attrs: {} })

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleDateString('zh-CN') : ''

const taskStatusType = (status) => {
  const map = { planned: 'info', pending: 'info', running: 'warning', in_progress: 'warning', done: 'success', completed: 'success' }
  return map[status] || 'info'
}

const taskStatusLabel = (status) => {
  const map = { planned: '计划中', pending: '待开始', running: '进行中', in_progress: '进行中', done: '已完成', completed: '已完成' }
  return map[status] || status
}

const dependencyLabel = (type) => {
  const map = { FS: '完成-开始', SS: '开始-开始', FF: '完成-完成', SF: '开始-完成' }
  return map[type] || type || '—'
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
    Object.assign(form, {
      ID: row.ID,
      title: row.title || '',
      status: row.status || 'planned',
      attrs: { ...(row.attrs || {}) }
    })
  } else {
    Object.assign(form, { ID: null, title: '', status: 'planned', attrs: {} })
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const payload = {
      title: form.title,
      status: form.status,
      attrs: form.attrs
    }
    let res
    if (dialogType.value === 'edit') {
      res = await updateTask({
        id: form.ID,
        entity_type: 'task',
        ...payload
      })
    } else {
      res = await createScheduleTask(payload)
    }
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
      const res = await deleteTask({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadData()
      }
    })
    .catch(() => {})
}

loadData()
</script>
