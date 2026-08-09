<template>
  <div class="workbench">
    <el-page-header content="项目工作台" />

    <el-tabs v-model="activeTab" style="margin-top: 12px" @tab-change="loadData">
      <el-tab-pane label="我创建的" name="created" />
      <el-tab-pane label="我负责的" name="lead" />
      <el-tab-pane label="我参与的" name="involved" />
      <el-tab-pane label="我关注的" name="focused" />
    </el-tabs>

    <el-radio-group v-model="statusFilter" size="small" style="margin-bottom: 12px" @change="loadData">
      <el-radio-button label="">全部</el-radio-button>
      <el-radio-button label="initiating">立项中</el-radio-button>
      <el-radio-button label="active">进行中</el-radio-button>
      <el-radio-button label="archived">已归档</el-radio-button>
      <el-radio-button label="paused">已暂停</el-radio-button>
    </el-radio-group>

    <el-row :gutter="12">
      <el-col v-for="card in cards" :key="card.projectId" :span="8" style="margin-bottom: 12px">
        <el-card shadow="hover" :body-style="{ padding: '16px' }" class="proj-card" @click="enterProject(card)">
          <div class="card-head">
            <span class="dot" :class="card.health" />
            <span class="proj-name">{{ card.projectName }}</span>
            <el-tag size="small" :type="priorityTag(card.priority)" style="margin-left: auto">{{ priorityLabel(card.priority) }}</el-tag>
          </div>
          <el-progress :percentage="Math.round(card.progress)" :color="healthColor(card.health)" style="margin: 8px 0" />
          <div class="card-row">
            <span>成本偏差：</span>
            <b :class="card.costVariance > 0 ? 'red' : 'green'">{{ card.costVariance > 0 ? '+' : '' }}{{ (card.costVariance || 0).toFixed(2) }}</b>
          </div>
          <div class="card-row">
            <span>风险数：</span><b>{{ card.riskCount }}</b>
            <span style="margin-left: 16px">负责人：</span><b>{{ card.leaderName }}</b>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getMyProjects, getFocusedProjects } from '@/api/pmocker/projectWorkbench'
import { useProjectStore } from '@/pinia'

const router = useRouter()
const projectStore = useProjectStore()
const activeTab = ref('created')
const statusFilter = ref('')
const cards = ref([])

// 点击项目卡片：设置项目上下文并跳转项目仪表盘
const enterProject = (card) => {
  projectStore.setProject(card.projectId, card.projectName)
  router.push({ name: 'pmockerDashboard' })
}

const loadData = async () => {
  if (activeTab.value === 'focused') {
    const res = await getFocusedProjects()
    if (res.code === 0) cards.value = res.data || []
  } else {
    const params = statusFilter.value ? { status: statusFilter.value } : {}
    const res = await getMyProjects(params)
    if (res.code === 0) {
      // 非过滤状态时返回 {created, lead, involved}，按 tab 取对应分类
      if (statusFilter.value) {
        cards.value = res.data || []
      } else {
        const grouped = res.data || {}
        cards.value = grouped[activeTab.value] || []
      }
    }
  }
}

const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')
const healthColor = (h) => ({ green: '#67C23A', yellow: '#E6A23C', red: '#F56C6C' }[h] || '#67C23A')

onMounted(() => { loadData() })
</script>

<style scoped>
.workbench { padding: 16px; }
.proj-card { cursor: pointer; transition: transform 0.15s; }
.proj-card:hover { transform: translateY(-2px); }
.card-head { display: flex; align-items: center; }
.proj-name { font-weight: bold; margin-left: 8px; }
.dot { width: 12px; height: 12px; border-radius: 50%; display: inline-block; }
.dot.green { background: #67C23A; }
.dot.yellow { background: #E6A23C; }
.dot.red { background: #F56C6C; }
.card-row { font-size: 13px; margin: 4px 0; }
.red { color: #F56C6C; }
.green { color: #67C23A; }
</style>
