<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="风险名称">
          <el-input v-model="searchInfo.keyword" placeholder="请输入风险名称" />
        </el-form-item>
        <el-form-item label="类别">
          <el-select v-model="searchInfo.category" placeholder="请选择类别" clearable>
            <el-option label="技术" value="technical" />
            <el-option label="管理" value="management" />
            <el-option label="商业" value="commercial" />
            <el-option label="外部" value="external" />
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
        <el-button type="primary" @click="openDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增风险
        </el-button>
      </div>

      <!-- 按状态分组展示 -->
      <div v-for="group in groupedData" :key="group.status" class="status-group">
        <div class="group-header">
          <div class="group-title">
            <el-tag :type="group.tagType">{{ group.label }}</el-tag>
            <span class="count">({{ group.items.length }})</span>
          </div>
          <div class="group-actions">
            <el-button v-for="action in group.actions" :key="action.label"
              :type="action.type" size="small"
              :disabled="!selectedMap[group.status] || selectedMap[group.status].length === 0"
              @click="handleBatchAction(group, action)">
              {{ action.label }}
            </el-button>
            <el-button type="warning" size="small"
              :disabled="!selectedMap[group.status] || selectedMap[group.status].length === 0"
              @click="handleBatchAssess(group)">
              AI评估
            </el-button>
          </div>
        </div>
        <el-table :data="group.items" row-key="id" size="small"
          @selection-change="(val) => onSelectionChange(group.status, val)">
          <el-table-column type="selection" width="40" />
          <el-table-column label="ID" prop="id" width="70" />
          <el-table-column label="风险名称" prop="title" min-width="200" />
          <el-table-column label="类别" width="100">
            <template #default="{ row }">{{ categoryLabel(row.attrs?.category) }}</template>
          </el-table-column>
          <el-table-column label="概率" width="80">
            <template #default="{ row }">{{ row.attrs?.probability ?? '' }}</template>
          </el-table-column>
          <el-table-column label="影响" width="80">
            <template #default="{ row }">{{ row.attrs?.impact ?? '' }}</template>
          </el-table-column>
          <el-table-column label="风险值" width="100">
            <template #default="{ row }">
              <el-tag :type="riskScoreType(row.attrs?.probability, row.attrs?.impact)">
                {{ riskScore(row.attrs?.probability, row.attrs?.impact) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="风险等级" width="110">
            <template #default="{ row }">
              <el-tag :type="riskLevelType(row.attrs?.risk_level)">{{ riskLevelLabel(row.attrs?.risk_level) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="EMV" width="120">
            <template #default="{ row }">¥{{ formatNum(row.attrs?.expected_monetary_value) }}</template>
          </el-table-column>
          <el-table-column label="策略" width="100">
            <template #default="{ row }">{{ strategyLabel(row.attrs?.response_strategy) }}</template>
          </el-table-column>
          <el-table-column label="机会策略" width="120">
            <template #default="{ row }">{{ strategyLabel(row.attrs?.opportunity_strategy) }}</template>
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
        <DynamicForm v-model="form" entity-type="risk" />
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
  getRiskList,
  createRisk,
  updateRisk,
  deleteRisk,
  assessRisk
} from '@/api/pmocker/risk'
import DynamicForm from '../components/DynamicForm.vue'
import ProjectSelector from '../components/ProjectSelector.vue'
import { useProjectStore } from '@/pinia'
import { groupByStatus } from '../components/statusTransitions.js'

defineOptions({ name: 'PmockerRiskRegister' })

const projectStore = useProjectStore()
const onProjectChange = () => { getTableData() }

const searchInfo = ref({})
const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const selectedMap = ref({})

const form = reactive({ id: null, title: '', status: 'identified', attrs: {} })

const formatNum = (val) => {
  if (val === null || val === undefined || val === '') return '0.00'
  const num = Number(val)
  return isNaN(num) ? '0.00' : num.toFixed(2)
}

const riskScore = (probability, impact) => {
  const p = Number(probability) || 0
  const i = Number(impact) || 0
  return p * i
}

const riskScoreType = (probability, impact) => {
  const score = riskScore(probability, impact)
  if (score >= 15) return 'danger'
  if (score >= 8) return 'warning'
  if (score > 0) return 'info'
  return 'info'
}

const riskLevelType = (level) => {
  const map = { low: 'info', medium: 'warning', high: 'danger', critical: 'danger' }
  return map[level] || 'info'
}

const riskLevelLabel = (level) => {
  const map = { low: '低', medium: '中', high: '高', critical: '严重' }
  return map[level] || level || '—'
}

const categoryLabel = (category) => {
  const map = { technical: '技术', management: '管理', commercial: '商业', external: '外部' }
  return map[category] || category || '—'
}

const strategyLabel = (strategy) => {
  const map = { avoid: '规避', transfer: '转移', mitigate: '减轻', accept: '接受', exploit: '开拓', share: '分享', enhance: '提高' }
  return map[strategy] || strategy || '—'
}

// 按状态分组
const groupedData = computed(() => groupByStatus(tableData.value, 'risk'))

// API 函数映射
const apiMap = { updateRisk }

const getTableData = async () => {
  const params = { page: 1, pageSize: 999, projectId: projectStore.projectId, ...searchInfo.value }
  const res = await getRiskList(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const onSubmit = () => { getTableData() }
const onReset = () => { searchInfo.value = {}; getTableData() }

const onSelectionChange = (status, selection) => {
  selectedMap.value = { ...selectedMap.value, [status]: selection }
}

// 批量状态流转
const handleBatchAction = async (group, action) => {
  const selected = selectedMap.value[group.status] || []
  if (selected.length === 0) return
  try {
    await ElMessageBox.confirm(`确认将选中的 ${selected.length} 条记录执行「${action.label}」操作？`, '提示', { type: 'warning' })
    const apiFn = apiMap[action.apiFn]
    const promises = selected.map(row => {
      return apiFn({ id: row.id, title: row.title, status: action.target, entity_type: 'risk', attrs: row.attrs || {} })
    })
    await Promise.all(promises)
    ElMessage.success(`成功${action.label} ${selected.length} 条记录`)
    selectedMap.value[group.status] = []
    getTableData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

// 批量AI评估
const handleBatchAssess = async (group) => {
  const selected = selectedMap.value[group.status] || []
  if (selected.length === 0) return
  try {
    await ElMessageBox.confirm(`确认对选中的 ${selected.length} 条风险执行AI评估？`, '提示', { type: 'warning' })
    const promises = selected.map(row => assessRisk({ id: row.id }))
    await Promise.all(promises)
    ElMessage.success(`成功评估 ${selected.length} 条风险`)
    selectedMap.value[group.status] = []
    getTableData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('评估失败')
  }
}

const openDialog = (row) => {
  dialogType.value = row ? 'edit' : 'add'
  dialogTitle.value = row ? '编辑风险' : '新增风险'
  if (row) {
    Object.assign(form, {
      id: row.id,
      title: row.title || '',
      status: row.status || 'identified',
      attrs: { ...(row.attrs || {}) }
    })
  } else {
    Object.assign(form, { id: null, title: '', status: 'identified', attrs: {} })
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
      res = await updateRisk({
        id: form.id,
        entity_type: 'risk',
        ...payload
      })
    } else {
      res = await createRisk({ ...payload, projectId: projectStore.projectId })
    }
    if (res.code === 0) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      getTableData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该风险吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteRisk({ id: row.id })
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
