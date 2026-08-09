<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增成本项
        </el-button>
        <el-button type="success" @click="handleCreateBaseline">
          <svg-icon icon="lucide:git-branch" /> 创建基线
        </el-button>
      </div>
      <el-table :data="tableData" row-key="id">
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

    <el-card shadow="never" style="margin-top: 16px">
      <template #header>
        <div style="display: flex; align-items: center; justify-content: space-between">
          <span>成员成本分摊（基于团队投入度核算）</span>
          <span style="font-size: 12px; color: #909399">
            月度合计：¥{{ formatNum(memberCostTotal) }}　|　公式：时薪 × 投入度% × {{ MONTHLY_HOURS }}h/月
          </span>
        </div>
      </template>
      <el-table :data="memberList" row-key="id">
        <el-table-column label="姓名" min-width="120">
          <template #default="{ row }">{{ row.attrs?.full_name || row.title }}</template>
        </el-table-column>
        <el-table-column label="角色" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.attrs?.role }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时薪" width="110">
          <template #default="{ row }">¥{{ formatNum(row.attrs?.hourly_rate) }}</template>
        </el-table-column>
        <el-table-column label="投入度" width="100">
          <template #default="{ row }">{{ row.attrs?.allocation_percent }}%</template>
        </el-table-column>
        <el-table-column label="估算工时(h/月)" width="140">
          <template #default="{ row }">{{ formatNum(estimatedHours(row)) }}</template>
        </el-table-column>
        <el-table-column label="成本贡献" width="140">
          <template #default="{ row }">¥{{ formatNum(memberCost(row)) }}</template>
        </el-table-column>
        <el-table-column label="占比" min-width="180">
          <template #default="{ row }">
            <el-progress :percentage="memberCostPercent(row)" :stroke-width="14" :text-inside="true" />
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!memberList.length" description="暂无团队成员数据" />
    </el-card>

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
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getCostItems,
  createCostItem,
  updateCostItem,
  deleteCostItem,
  createCostBaseline
} from '@/api/pmocker/cost'
import { listMember } from '@/api/pmocker/team'
import DynamicForm from '../components/DynamicForm.vue'
import ProjectSelector from '../components/ProjectSelector.vue'
import { useProjectStore } from '@/pinia'

defineOptions({ name: 'PmockerCostBudget' })

const projectStore = useProjectStore()
const onProjectChange = () => { loadData(); loadMembers() }

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ id: null, title: '', status: 'planned', attrs: {} })

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

// ---- 成员成本分摊（投入度→成本核算联动）----
// 月度基准工时：8 小时/天 × 22 工作日
const MONTHLY_HOURS = 160
const memberList = ref([])

// 估算工时 = 投入度% × 月度基准工时
const estimatedHours = (row) => {
  const alloc = Number(row.attrs?.allocation_percent) || 0
  return (alloc / 100) * MONTHLY_HOURS
}

// 成本贡献 = 时薪 × 估算工时
const memberCost = (row) => {
  const rate = Number(row.attrs?.hourly_rate) || 0
  return rate * estimatedHours(row)
}

const memberCostTotal = computed(() =>
  memberList.value.reduce((sum, row) => sum + memberCost(row), 0)
)

const memberCostPercent = (row) => {
  const total = memberCostTotal.value
  if (!total) return 0
  return Number(((memberCost(row) / total) * 100).toFixed(1))
}

const loadMembers = async () => {
  const res = await listMember({ projectId: projectStore.projectId, page: 1, pageSize: 1000 })
  if (res.code === 0) {
    memberList.value = res.data.list || []
  }
}

const loadData = async () => {
  const res = await getCostItems({ projectId: projectStore.projectId })
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const openDialog = (row) => {
  dialogType.value = row ? 'edit' : 'add'
  dialogTitle.value = row ? '编辑成本项' : '新增成本项'
  if (row) {
    Object.assign(form, {
      id: row.id,
      title: row.title || '',
      status: row.status || 'planned',
      attrs: { ...(row.attrs || {}) }
    })
  } else {
    Object.assign(form, { id: null, title: '', status: 'planned', attrs: {} })
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
        id: form.id,
        entity_type: 'cost_item',
        ...payload
      })
    } else {
      res = await createCostItem({ ...payload, projectId: projectStore.projectId })
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
      const res = await deleteCostItem({ id: row.id })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadData()
      }
    })
    .catch(() => {})
}

const handleCreateBaseline = async () => {
  const res = await createCostBaseline({ projectId: projectStore.projectId })
  if (res.code === 0) {
    ElMessage.success('基线创建成功')
    loadData()
  }
}

loadData()
loadMembers()
</script>
