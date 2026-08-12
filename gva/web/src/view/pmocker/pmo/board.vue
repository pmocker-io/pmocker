<template>
  <div class="pmo-board">
    <el-page-header content="EPS PMO 看板" />
    <el-row :gutter="12" class="mt-3">
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>项目总数</span><b>{{ dash.totalProjects }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>健康（绿）</span><b class="text-success">{{ dash.healthDist.green || 0 }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>关注（黄）</span><b class="text-warning">{{ dash.healthDist.yellow || 0 }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>预警（红）</span><b class="text-error">{{ dash.healthDist.red || 0 }}</b></div></el-card></el-col>
    </el-row>

    <el-card shadow="hover" class="mt-4">
      <template #header>项目卡片网格</template>
      <el-row :gutter="12">
        <el-col v-for="card in dash.projectCards" :key="card.projectId" :span="8" class="mb-3">
          <el-card shadow="hover" :body-style="{ padding: '16px' }">
            <div class="card-head">
              <span class="dot" :class="card.health" />
              <span class="proj-name">{{ card.projectName }}</span>
              <el-tag size="small" :type="priorityTag(card.priority)" style="margin-left: auto">{{ priorityLabel(card.priority) }}</el-tag>
            </div>
            <el-progress :percentage="Math.round(card.progress || 0)" :color="healthColor(card.health)" class="my-2" />
            <div class="card-row">
              <span>成本偏差：</span>
              <b :class="(card.costVariance || 0) > 0 ? 'red' : 'green'">{{ (card.costVariance || 0) > 0 ? '+' : '' }}{{ (card.costVariance || 0).toFixed(2) }}</b>
            </div>
            <div class="card-row">
              <span>风险数：</span><b>{{ card.riskCount }}</b>
              <span class="ml-4">负责人：</span><b>{{ card.leaderName }}</b>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="hover" class="mt-4">
      <template #header>资源负荷汇总</template>
      <el-descriptions :column="4" border>
        <el-descriptions-item label="总人数">{{ dash.loadSummary.totalMembers || 0 }}</el-descriptions-item>
        <el-descriptions-item label="总工时">{{ (dash.loadSummary.totalHours || 0).toFixed(1) }}</el-descriptions-item>
        <el-descriptions-item label="平均负荷">{{ (dash.loadSummary.avgLoad || 0).toFixed(1) }}%</el-descriptions-item>
        <el-descriptions-item label="超负荷人数">{{ dash.loadSummary.overloadedCount || 0 }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { getPMODashboard } from '@/api/pmocker/pmo'

const dash = reactive({
  totalProjects: 0,
  healthDist: {},
  projectCards: [],
  loadSummary: {}
})

const loadData = async () => {
  const res = await getPMODashboard()
  if (res.code === 0) Object.assign(dash, res.data)
}

const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')
const healthColor = (h) => ({ green: 'var(--el-color-success)', yellow: 'var(--el-color-warning)', red: 'var(--el-color-danger)' }[h] || 'var(--el-color-success)')

onMounted(() => { loadData() })
</script>

<style scoped>
.pmo-board { padding: 16px; }
.stat { display: flex; justify-content: space-between; align-items: center; }
.stat b { font-size: 24px; }
.card-head { display: flex; align-items: center; }
.proj-name { font-weight: bold; margin-left: 8px; }
.dot { width: 12px; height: 12px; border-radius: 50%; display: inline-block; }
.dot.green { background: var(--el-color-success); }
.dot.yellow { background: var(--el-color-warning); }
.dot.red { background: var(--el-color-danger); }
.card-row { font-size: 13px; margin: 4px 0; }
.red { color: var(--el-color-danger); }
.green { color: var(--el-color-success); }
</style>
