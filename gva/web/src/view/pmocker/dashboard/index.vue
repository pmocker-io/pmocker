<template>
  <div class="dashboard-page">
    <el-page-header content="项目仪表盘" @back="$router.back()" />
    <el-select v-model="projectId" placeholder="选择项目" filterable style="margin: 12px 200px 12px 0" @change="onProjectChange">
      <el-option v-for="p in projects" :key="p.id" :label="p.name" :value="p.id" />
    </el-select>
    <el-button type="primary" @click="genSnapshot">生成月报快照</el-button>

    <el-row :gutter="16" style="margin-top: 12px">
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>项目进度</template>
          <div ref="progressChart" style="height: 220px" />
          <p style="text-align:center; margin: 8px 0 0">
            <el-tag :type="priorityTag(dash.priority)">{{ priorityLabel(dash.priority) }}</el-tag>
            <el-tag :type="healthTag(dash.health)" style="margin-left: 8px">{{ healthLabel(dash.health) }}</el-tag>
          </p>
        </el-card>
      </el-col>
      <el-col :span="9">
        <el-card shadow="hover">
          <template #header>成本 S 曲线（预算 vs 实际）</template>
          <div ref="costChart" style="height: 220px" />
        </el-card>
      </el-col>
      <el-col :span="9">
        <el-card shadow="hover">
          <template #header>问题统计</template>
          <div ref="issueChart" style="height: 220px" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>风险矩阵</template>
          <div ref="riskChart" style="height: 260px" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>资源利用率</template>
          <el-table :data="dash.resourceSummary.members" size="small" max-height="260">
            <el-table-column prop="memberName" label="成员" width="100" />
            <el-table-column prop="hourlyRate" label="时薪" width="80" />
            <el-table-column prop="loggedHours" label="已登记工时" width="110" />
            <el-table-column label="利用率">
              <template #default="{ row }">
                <el-progress :percentage="Math.min(row.utilization, 100)" :status="row.utilization > 100 ? 'exception' : ''" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" style="margin-top: 16px">
      <template #header>里程碑</template>
      <el-timeline>
        <el-timeline-item v-for="m in dash.milestones" :key="m.id" :timestamp="m.date" :type="m.status === 'done' ? 'success' : 'primary'">
          {{ m.title }} <el-tag size="small" style="margin-left: 8px">{{ m.status }}</el-tag>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { getDashboard, generateSnapshot } from '@/api/pmocker/dashboard'
import { useProjectStore } from '@/pinia'
import service from '@/utils/request'

const projectStore = useProjectStore()
// 优先从全局项目上下文初始化（由 EPS/工作台"进入项目"时设置）
const projectId = ref(projectStore.projectId || '')
const projects = ref([])
const dash = reactive({
  progress: 0, health: 'green', priority: 2,
  costSummary: {}, riskSummary: { bySeverity: {} }, issueSummary: { byStatus: {} },
  resourceSummary: { members: [] }, milestones: []
})
const progressChart = ref(null)
const costChart = ref(null)
const issueChart = ref(null)
const riskChart = ref(null)
let pChart, cChart, iChart, rChart

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
    const allNodes = flattenTree(treeData)
    projects.value = allNodes.filter(p => p.type === 'project' || !p.type)
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
  const res = await getDashboard({ projectId: projectId.value })
  if (res.code !== 0) return
  Object.assign(dash, res.data)
  await nextTick()
  renderCharts()
}

const renderCharts = () => {
  // 进度环形图
  if (pChart) pChart.dispose()
  pChart = echarts.init(progressChart.value)
  pChart.setOption({
    series: [{
      type: 'gauge', startAngle: 90, endAngle: -270, radius: '90%',
      progress: { show: true, width: 14 },
      axisLine: { lineStyle: { width: 14 } },
      pointer: { show: false },
      detail: { valueAnimation: true, fontSize: 24, formatter: '{value}%' },
      data: [{ value: Math.round(dash.progress) }]
    }]
  })
  // 成本 S 曲线
  if (cChart) cChart.dispose()
  cChart = echarts.init(costChart.value)
  const months = ['1月', '2月', '3月', '4月', '5月', '6月']
  const budget = months.map((_, i) => +(dash.costSummary.budget * (i + 1) / 6).toFixed(2))
  const actual = months.map((_, i) => +(dash.costSummary.actual * (i + 1) / 6).toFixed(2))
  cChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['预算', '实际'] },
    xAxis: { type: 'category', data: months },
    yAxis: { type: 'value' },
    series: [
      { name: '预算', type: 'line', smooth: true, data: budget },
      { name: '实际', type: 'line', smooth: true, data: actual, areaStyle: { opacity: 0.1 } }
    ]
  })
  // 问题统计柱状图
  if (iChart) iChart.dispose()
  iChart = echarts.init(issueChart.value)
  const issueStatuses = Object.keys(dash.issueSummary.byStatus || {})
  iChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: issueStatuses.length ? issueStatuses : ['open', 'in_progress', 'closed'] },
    yAxis: { type: 'value' },
    series: [{ type: 'bar', data: issueStatuses.length ? Object.values(dash.issueSummary.byStatus) : [dash.issueSummary.open, 0, dash.issueSummary.closed], itemStyle: { color: '#409EFF' } }]
  })
  // 风险矩阵散点图（概率 vs 影响）
  if (rChart) rChart.dispose()
  rChart = echarts.init(riskChart.value)
  rChart.setOption({
    tooltip: { formatter: p => `${p.data[2]}` },
    xAxis: { name: '影响', min: 0, max: 5, type: 'value' },
    yAxis: { name: '概率', min: 0, max: 5, type: 'value' },
    series: [{
      type: 'scatter',
      symbolSize: 20,
      data: (dash.riskSummary.bySeverity ? [] : []),
      itemStyle: { color: '#F56C6C' }
    }],
    visualMap: { show: false, pieces: [
      { gte: 0, lt: 4, color: '#67C23A' },
      { gte: 4, lt: 9, color: '#E6A23C' },
      { gte: 9, color: '#F56C6C' }
    ], dimension: 2, min: 0, max: 25 }
  })
}

const genSnapshot = async () => {
  if (!projectId.value) return
  const res = await generateSnapshot({ projectId: projectId.value, reportType: 'dashboard', period: '' })
  if (res.code === 0) ElMessage.success('月报快照已生成')
}

const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')
const healthLabel = (h) => ({ green: '健康', yellow: '关注', red: '预警' }[h] || '健康')
const healthTag = (h) => ({ green: 'success', yellow: 'warning', red: 'danger' }[h] || 'success')

onMounted(async () => {
  await loadProjects()
  // 若已有项目上下文（从 EPS/工作台进入），直接加载仪表盘数据
  if (projectId.value) {
    loadData()
  }
})
</script>

<style scoped>
.dashboard-page { padding: 16px; }
</style>
