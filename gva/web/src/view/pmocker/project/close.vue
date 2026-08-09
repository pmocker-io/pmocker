<template>
  <div class="close-page">
    <el-page-header content="结项报告" @back="$router.back()" />
    <el-select v-model="projectId" placeholder="选择项目" filterable style="margin: 12px 200px 12px 0" @change="onProjectChange">
      <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
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
            <el-descriptions-item label="总工时">{{ (report.resourceStat?.totalHours || 0).toFixed(1) }}</el-descriptions-item>
            <el-descriptions-item label="人工成本">{{ (report.resourceStat?.totalCost || 0).toFixed(2) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" style="margin-top: 12px" v-if="report">
      <template #header>成本统计</template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="预算">{{ (report.costStat?.budget || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实际">{{ (report.costStat?.actual || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="偏差">
          <span :class="(report.costStat?.variance || 0) > 0 ? 'red' : 'green'">{{ (report.costStat?.variance || 0) > 0 ? '+' : '' }}{{ (report.costStat?.variance || 0).toFixed(2) }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCloseReport, archiveProject } from '@/api/pmocker/archive'
import { useProjectStore } from '@/pinia'
import service from '@/utils/request'

const projectStore = useProjectStore()
// 优先从全局项目上下文初始化（由 EPS/工作台"进入项目"时设置）
const projectId = ref(projectStore.projectId || '')
const projects = ref([])
const report = ref(null)

const flattenTree = (nodes) => {
  let result = []
  for (const node of nodes) {
    result.push(node)
    if (node.children && node.children.length > 0) {
      result = result.concat(flattenTree(node.children))
    }
  }
  return result
}

const loadProjects = async () => {
  const res = await service({ url: '/pmocker/eps/tree', method: 'get' })
  if (res.code === 0) {
    const treeData = res.data || []
    // 只展示项目节点（排除 group/division 组织节点）
    projects.value = flattenTree(treeData).filter(p => p.type !== 'group' && p.type !== 'division')
  }
}

// 切换项目时同步到全局 store，并加载数据
const onProjectChange = (val) => {
  const proj = projects.value.find(p => p.id === val)
  projectStore.setProject(val, proj ? proj.name : '')
  loadData()
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

onMounted(async () => {
  await loadProjects()
  // 若已有项目上下文（从 EPS/工作台进入），直接加载结项报告
  if (projectId.value) {
    loadData()
  }
})
</script>

<style scoped>
.close-page { padding: 16px; }
.red { color: #F56C6C; }
.green { color: #67C23A; }
</style>
