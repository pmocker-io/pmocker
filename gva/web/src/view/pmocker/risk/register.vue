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
        <el-table-column label="类别" prop="category" width="100">
          <template #default="{ row }">{{ categoryLabel(row.category) }}</template>
        </el-table-column>
        <el-table-column label="概率" prop="probability" width="80" />
        <el-table-column label="影响" prop="impact" width="80" />
        <el-table-column label="风险值" width="100">
          <template #default="{ row }">
            <el-tag :type="levelType(row.probability * row.impact)">
              {{ row.probability * row.impact }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="策略" prop="strategy" width="100">
          <template #default="{ row }">{{ strategyLabel(row.strategy) }}</template>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="风险名称" prop="title">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="类别" prop="category">
          <el-select v-model="form.category">
            <el-option label="技术" value="technical" />
            <el-option label="管理" value="management" />
            <el-option label="商业" value="commercial" />
            <el-option label="外部" value="external" />
          </el-select>
        </el-form-item>
        <el-form-item label="概率" prop="probability">
          <el-slider v-model="form.probability" :min="1" :max="5" show-stops />
        </el-form-item>
        <el-form-item label="影响" prop="impact">
          <el-slider v-model="form.impact" :min="1" :max="5" show-stops />
        </el-form-item>
        <el-form-item label="策略" prop="strategy">
          <el-select v-model="form.strategy">
            <el-option label="规避" value="avoid" />
            <el-option label="转移" value="transfer" />
            <el-option label="减轻" value="mitigate" />
            <el-option label="接受" value="accept" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" />
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
import { getRiskList, createRisk, deleteRisk, assessRisk } from '@/api/pmocker/risk'

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

const form = reactive({
  ID: null, title: '', category: 'technical',
  probability: 1, impact: 1, strategy: 'mitigate', description: ''
})
const rules = {
  title: [{ required: true, message: '请输入风险名称', trigger: 'blur' }],
  category: [{ required: true, message: '请选择类别', trigger: 'change' }]
}

const levelType = (level) => {
  const map = { low: 'info', medium: 'warning', high: 'danger' }
  return map[level] || 'info'
}

const levelLabel = (level) => {
  const map = { low: '低', medium: '中', high: '高' }
  return map[level] || level
}

const categoryLabel = (category) => {
  const map = { technical: '技术', management: '管理', commercial: '商业', external: '外部' }
  return map[category] || category
}

const strategyLabel = (strategy) => {
  const map = { avoid: '规避', transfer: '转移', mitigate: '减轻', accept: '接受' }
  return map[strategy] || strategy
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
    Object.assign(form, row)
  } else {
    Object.assign(form, {
      ID: null, title: '', category: 'technical',
      probability: 1, impact: 1, strategy: 'mitigate', description: ''
    })
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await createRisk(form)
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
