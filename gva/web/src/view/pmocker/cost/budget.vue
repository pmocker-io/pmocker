<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增成本项
        </el-button>
        <el-button type="success" @click="handleCreateBaseline">
          <svg-icon icon="lucide:git-branch" /> 创建基线
        </el-button>
      </div>
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="科目" prop="title" min-width="200" />
        <el-table-column label="计划价值(PV)" width="140">
          <template #default="{ row }">¥{{ formatNum(row.attrs?.planned_value) }}</template>
        </el-table-column>
        <el-table-column label="挣值(EV)" width="130">
          <template #default="{ row }">¥{{ formatNum(row.attrs?.earned_value) }}</template>
        </el-table-column>
        <el-table-column label="实际成本(AC)" width="140">
          <template #default="{ row }">¥{{ formatNum(row.attrs?.actual_cost) }}</template>
        </el-table-column>
        <el-table-column label="CPI" width="100">
          <template #default="{ row }">
            <el-tag :type="cpiTagType(row.attrs?.cpi)">{{ formatNum(row.attrs?.cpi) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" prop="status" width="110">
          <template #default="{ row }">
            <el-tag :type="costStatusType(row.status)">{{ costStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">编辑</el-button>
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
        <DynamicForm v-model="form" entity-type="cost_item" />
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
  getCostItems,
  createCostItem,
  updateCostItem,
  deleteCostItem,
  createCostBaseline
} from '@/api/pmocker/cost'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerCostBudget' })

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, title: '', status: 'planned', attrs: {} })

const formatNum = (val) => {
  if (val === null || val === undefined || val === '') return '0.00'
  const num = Number(val)
  return isNaN(num) ? '0.00' : num.toFixed(2)
}

const cpiTagType = (cpi) => {
  if (cpi === null || cpi === undefined || cpi === '') return 'info'
  const num = Number(cpi)
  if (isNaN(num)) return 'info'
  if (num >= 1) return 'success'
  return 'danger'
}

const costStatusType = (status) => {
  const map = { planned: 'info', in_progress: 'warning', completed: 'success', closed: 'info' }
  return map[status] || 'info'
}

const costStatusLabel = (status) => {
  const map = { planned: '计划中', in_progress: '进行中', completed: '已完成', closed: '已关闭' }
  return map[status] || status
}

const loadData = async () => {
  const res = await getCostItems({})
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const openDialog = (row) => {
  dialogType.value = row ? 'edit' : 'add'
  dialogTitle.value = row ? '编辑成本项' : '新增成本项'
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
      res = await updateCostItem({
        id: form.ID,
        entity_type: 'cost_item',
        ...payload
      })
    } else {
      res = await createCostItem(payload)
    }
    if (res.code === 0) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      loadData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该成本项吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteCostItem({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadData()
      }
    })
    .catch(() => {})
}

const handleCreateBaseline = async () => {
  const res = await createCostBaseline({})
  if (res.code === 0) {
    ElMessage.success('基线创建成功')
    loadData()
  }
}

loadData()
</script>
