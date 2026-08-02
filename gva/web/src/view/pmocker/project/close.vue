<template>
  <div class="close-page">
    <el-page-header content="结项报告" @back="$router.back()" />
    <el-select v-model="projectId" placeholder="选择项目" filterable style="margin: 12px 200px 12px 0" @change="loadData">
      <el-option v-for="p in projects" :key="p.ID" :label="p.title" :value="p.ID" />
    </el-select>
    <el-button type="danger" :disabled="!report || report.archivedAt" @click="doArchive">执行归档</el-button>

    <el-descriptions v-if="report" :column="3" border title="项目基本信息" style="margin-top: 16px">
      <el-descriptions-item label="项目名称">{{ report.projectName }}</el-descriptions-item>
      <el-descriptions-item label="开始日期">{{ report.startDate }}</el-descriptions-item>
      <el-descriptions-item label="结束日期">{{ report.endDate }}</el-descriptions-item>
      <el-descriptions-item label="归档状态">
        <el-tag :type="report.archivedAt ? 'success' : 'info'">{{ report.archivedAt ? '已归档' : '未归档' }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="归档时间" v-if="report.archivedAt">{{ report.archivedAt }}</el-descriptions-item>
    </el-descriptions>

    <el-row :gutter="12" style="margin-top: 16px" v-if="report">
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>任务统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.taskStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已完成">{{ report.taskStat.done }}</el-descriptions-item>
            <el-descriptions-item label="未完成">{{ report.taskStat.open }}</el-descriptions-item>
            <el-descriptions-item label="逾期">{{ report.taskStat.overdue }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>问题统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.issueStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已关闭">{{ report.issueStat.closed }}</el-descriptions-item>
            <el-descriptions-item label="未关闭">{{ report.issueStat.open }}</el-descriptions-item>
            <el-descriptions-item label="逾期">{{ report.issueStat.overdue }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>风险统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.riskStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已关闭">{{ report.riskStat.closed }}</el-descriptions-item>
            <el-descriptions-item label="未关闭">{{ report.riskStat.open }}</el-descriptions-item>
            <el-descriptions-item label="逾期">{{ report.riskStat.overdue }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="12" style="margin-top: 12px" v-if="report">
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>需求统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.reqStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已实现">{{ report.reqStat.done }}</el-descriptions-item>
            <el-descriptions-item label="未实现">{{ report.reqStat.open }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>变更统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.changeStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已批准">{{ report.changeStat.done }}</el-descriptions-item>
            <el-descriptions-item label="未批准">{{ report.changeStat.open }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>资源统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="人数">{{ report.resourceStat.memberCount }}</el-descriptions-item>
            <el-descriptions-item label="总工时">{{ report.resourceStat.totalHours.toFixed(1) }}</el-descriptions-item>
            <el-descriptions-item label="人工成本">{{ report.resourceStat.totalCost.toFixed(2) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" style="margin-top: 12px" v-if="report">
      <template #header>成本统计</template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="预算">{{ report.costStat.budget.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实际">{{ report.costStat.actual.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="偏差">
          <span :class="report.costStat.variance > 0 ? 'red' : 'green'">{{ report.costStat.variance > 0 ? '+' : '' }}{{ report.costStat.variance.toFixed(2) }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCloseReport, archiveProject } from '@/api/pmocker/archive'
import service from '@/utils/request'

const projectId = ref('')
const projects = ref([])
const report = ref(null)

const loadProjects = async () => {
  const res = await service({ url: '/pmocker/eps/tree', method: 'get' })
  if (res.code === 0) projects.value = res.data || []
}

const loadData = async () => {
  if (!projectId.value) return
  const res = await getCloseReport({ projectId: projectId.value })
  if (res.code === 0) report.value = res.data
}

const doArchive = async () => {
  try {
    await ElMessageBox.confirm('确认归档此项目？归档后所有数据将变为只读。', '确认归档', { type: 'warning' })
  } catch { return }
  const res = await archiveProject({ projectId: projectId.value })
  if (res.code === 0) {
    ElMessage.success('项目已归档')
    loadData()
  }
}

onMounted(() => { loadProjects() })
</script>

<style scoped>
.close-page { padding: 16px; }
.red { color: #F56C6C; }
.green { color: #67C23A; }
</style>
