<template>
  <div>
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
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="ID" prop="ID" width="80" />
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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">编辑</el-button>
            <el-button type="warning" link @click="handleAssess(row)">评估</el-button>
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
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getRiskList,
  createRisk,
  updateRisk,
  deleteRisk,
  assessRisk
} from '@/api/pmocker/risk'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerRiskRegister' })

const searchInfo = ref({})
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, title: '', status: 'identified', attrs: {} })

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

const getTableData = async () => {
  const params = { page: page.value, pageSize: pageSize.value, ...searchInfo.value }
  const res = await getRiskList(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const onSubmit = () => { page.value = 1; getTableData() }
const onReset = () => { searchInfo.value = {}; page.value = 1; getTableData() }

const openDialog = (row) => {
  dialogType.value = row ? 'edit' : 'add'
  dialogTitle.value = row ? '编辑风险' : '新增风险'
  if (row) {
    Object.assign(form, {
      ID: row.ID,
      title: row.title || '',
      status: row.status || 'identified',
      attrs: { ...(row.attrs || {}) }
    })
  } else {
    Object.assign(form, { ID: null, title: '', status: 'identified', attrs: {} })
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
        id: form.ID,
        entity_type: 'risk',
        ...payload
      })
    } else {
      res = await createRisk(payload)
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
      const res = await deleteRisk({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

const handleAssess = async (row) => {
  const res = await assessRisk({ ID: row.ID })
  if (res.code === 0) {
    ElMessage.success('评估完成')
    getTableData()
  }
}

getTableData()
</script>
