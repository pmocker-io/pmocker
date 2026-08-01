<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog">
          <svg-icon icon="lucide:plus" /> 新增成本项
        </el-button>
        <el-button type="success" @click="handleCreateBaseline">
          <svg-icon icon="lucide:git-branch" /> 创建基线
        </el-button>
      </div>
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="科目" prop="title" min-width="200" />
        <el-table-column label="类型" prop="type" width="120">
          <template #default="{ row }">
            <el-tag>{{ costTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="预算金额" prop="budget" width="150">
          <template #default="{ row }">¥{{ row.budget?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="实际花费" prop="actual" width="150">
          <template #default="{ row }">¥{{ row.actual?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="偏差" width="120">
          <template #default="{ row }">
            <el-tag :type="row.actual <= row.budget ? 'success' : 'danger'">
              ¥{{ ((row.budget || 0) - (row.actual || 0)).toFixed(2) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" title="新增成本项" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="科目" prop="title">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type">
            <el-option label="人工" value="labor" />
            <el-option label="材料" value="material" />
            <el-option label="设备" value="equipment" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="预算金额" prop="budget">
          <el-input-number v-model="form.budget" :min="0" :precision="2" />
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
import { getCostItems, createCostItem, createCostBaseline } from '@/api/pmocker/cost'

defineOptions({ name: 'PmockerCostBudget' })

const tableData = ref([])
const dialogVisible = ref(false)
const formRef = ref(null)

const form = reactive({ title: '', type: 'labor', budget: 0 })
const rules = {
  title: [{ required: true, message: '请输入科目', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  budget: [{ required: true, message: '请输入金额', trigger: 'blur' }]
}

const costTypeLabel = (type) => {
  const map = { labor: '人工', material: '材料', equipment: '设备', other: '其他' }
  return map[type] || type
}

const loadData = async () => {
  const res = await getCostItems({})
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const openDialog = () => {
  Object.assign(form, { title: '', type: 'labor', budget: 0 })
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await createCostItem(form)
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
      const res = await createCostItem({ action: 'delete', ID: row.ID })
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
