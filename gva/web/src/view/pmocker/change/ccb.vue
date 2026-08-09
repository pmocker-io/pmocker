<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable @change="loadData">
            <el-option label="待CCB评审" value="ccb_review" />
            <el-option label="已批准" value="approved" />
            <el-option label="已驳回" value="rejected" />
            <el-option label="实施中" value="implementing" />
            <el-option label="验证中" value="verifying" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-row :gutter="16">
        <el-col :span="8">
          <el-card>
            <template #header><span class="font-medium">CCB 统计</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="待评审">{{ stats.pending || 0 }}</el-descriptions-item>
              <el-descriptions-item label="本月批准">{{ stats.approvedThisMonth || 0 }}</el-descriptions-item>
              <el-descriptions-item label="本月驳回">{{ stats.rejectedThisMonth || 0 }}</el-descriptions-item>
              <el-descriptions-item label="实施中">{{ stats.implementing || 0 }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>
        <el-col :span="16">
          <el-card>
            <template #header><span class="font-medium">待处理变更</span></template>
            <el-table :data="changeList" @row-click="handleSelectChange">
              <el-table-column label="标题" prop="title" min-width="150" />
              <el-table-column label="优先级" prop="priority" width="80">
                <template #default="{ row }">
                  <el-tag :type="priorityType(row.priority)" size="small">{{ row.priority }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="状态" prop="status" width="120">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.status }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>

      <el-card v-if="selectedChange" class="mt-4">
        <template #header>
          <span class="font-medium">{{ selectedChange.title }} - CCB 审批流程</span>
        </template>
        <el-steps :active="currentStep" align-center>
          <el-step title="提交" :description="formatDate(selectedChange.CreatedAt)" />
          <el-step title="影响分析" />
          <el-step title="CCB评审" />
          <el-step title="批准/驳回" />
          <el-step title="实施" />
          <el-step title="验证" />
          <el-step title="关闭" />
        </el-steps>

        <div class="flex gap-2 mt-6 justify-center">
          <el-button v-if="selectedChange.status === 'ccb_review'" type="success" @click="handleApprove">
            <svg-icon icon="lucide:check" /> 批准
          </el-button>
          <el-button v-if="selectedChange.status === 'ccb_review'" type="danger" @click="handleReject">
            <svg-icon icon="lucide:x" /> 驳回
          </el-button>
          <el-button v-if="selectedChange.status === 'approved'" type="primary" @click="handleImplement">
            <svg-icon icon="lucide:play" /> 开始实施
          </el-button>
          <el-button v-if="selectedChange.status === 'implementing'" type="warning" @click="handleVerify">
            <svg-icon icon="lucide:search-check" /> 提交验证
          </el-button>
          <el-button v-if="selectedChange.status === 'verifying'" type="success" @click="handleClose">
            <svg-icon icon="lucide:check-check" /> 关闭变更
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getChangeList, getChangeCCBStats, approveChange, rejectChange, implementChange, verifyChange, closeChange } from '@/api/pmocker/change'
import { useProjectStore } from '@/pinia'

defineOptions({ name: 'PmockerChangeCCB' })
const projectStore = useProjectStore()

const searchInfo = ref({})
const changeList = ref([])
const stats = ref({})
const selectedChange = ref(null)

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : ''
const priorityType = (p) => ({ urgent: 'danger', high: 'warning', medium: 'info', low: 'info' }[p] || 'info')

const statusStepMap = {
  submitted: 0, analyzing: 1, ccb_review: 2,
  approved: 3, rejected: 3,
  implementing: 4, verifying: 5, closed: 6
}

const currentStep = computed(() => {
  if (!selectedChange.value) return 0
  return statusStepMap[selectedChange.value.status] || 0
})

const loadData = async () => {
  const [listRes, statsRes] = await Promise.all([
    getChangeList({ status: searchInfo.value.status || 'ccb_review', page: 1, pageSize: 20, projectId: projectStore.projectId }),
    getChangeCCBStats({ projectId: projectStore.projectId })
  ])
  if (listRes.code === 0) {
    changeList.value = listRes.data.list || []
  }
  if (statsRes.code === 0) {
    stats.value = statsRes.data
  }
}

const handleSelectChange = (row) => {
  selectedChange.value = row
}

const handleApprove = async () => {
  const res = await approveChange({ ID: selectedChange.value.ID })
  if (res.code === 0) {
    ElMessage.success('已批准')
    loadData()
  }
}

const handleReject = () => {
  ElMessageBox.prompt('请输入驳回原因', '驳回', { type: 'warning' })
    .then(async ({ value }) => {
      const res = await rejectChange({ ID: selectedChange.value.ID, reason: value })
      if (res.code === 0) {
        ElMessage.success('已驳回')
        loadData()
      }
    })
    .catch(() => {})
}

const handleImplement = async () => {
  const res = await implementChange({ ID: selectedChange.value.ID })
  if (res.code === 0) {
    ElMessage.success('已开始实施')
    loadData()
  }
}

const handleVerify = async () => {
  const res = await verifyChange({ ID: selectedChange.value.ID })
  if (res.code === 0) {
    ElMessage.success('已提交验证')
    loadData()
  }
}

const handleClose = async () => {
  const res = await closeChange({ ID: selectedChange.value.ID })
  if (res.code === 0) {
    ElMessage.success('已关闭')
    loadData()
  }
}

loadData()
</script>
