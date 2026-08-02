<template>
  <div class="task-center">
    <el-page-header content="个人任务中心" />

    <el-row :gutter="12" style="margin-top: 12px">
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>任务总数</span><b>{{ stats.total }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>完成率</span><b style="color:#67C23A">{{ stats.doneRate.toFixed(1) }}%</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>逾期数</span><b style="color:#F56C6C">{{ stats.overdueCount }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>进行中</span><b style="color:#409EFF">{{ stats.doingCount }}</b></div></el-card></el-col>
    </el-row>

    <el-tabs v-model="activeTab" style="margin-top: 16px" @tab-change="loadTasks">
      <el-tab-pane label="我的待办" name="todo" />
      <el-tab-pane label="进行中" name="doing" />
      <el-tab-pane label="已完成" name="done" />
      <el-tab-pane label="已逾期" name="overdue" />
      <el-tab-pane label="我关注的" name="focused" />
    </el-tabs>

    <el-table :data="tasks" border size="small">
      <el-table-column prop="title" label="任务名称" min-width="180" />
      <el-table-column label="来源" width="110">
        <template #default="{ row }">
          <el-tag size="small" :type="sourceTag(row.sourceType)">{{ sourceLabel(row.sourceType) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="projectName" label="所属项目" width="160" />
      <el-table-column label="优先级" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="priorityTag(row.priority)">{{ priorityLabel(row.priority) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="endDate" label="截止日期" width="120" />
      <el-table-column label="进度" width="140">
        <template #default="{ row }">
          <el-progress :percentage="row.progress" :status="row.overdue ? 'exception' : ''" />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getMyTasks, getFocusedTasks, getTaskStats } from '@/api/pmocker/taskCenter'

const activeTab = ref('todo')
const tasks = ref([])
const stats = reactive({ total: 0, done: 0, doneRate: 0, overdueCount: 0, todoCount: 0, doingCount: 0 })

const loadStats = async () => {
  const res = await getTaskStats()
  if (res.code === 0) Object.assign(stats, res.data)
}

const loadTasks = async () => {
  if (activeTab.value === 'focused') {
    const res = await getFocusedTasks()
    if (res.code === 0) tasks.value = res.data || []
  } else {
    const res = await getMyTasks({ status: activeTab.value })
    if (res.code === 0) tasks.value = res.data || []
  }
}

const sourceLabel = (s) => ({ project_task: '项目任务', issue_task: '问题任务', change_task: '变更任务', deliverable_task: '交付物任务' }[s] || s)
const sourceTag = (s) => ({ project_task: '', issue_task: 'warning', change_task: 'danger', deliverable_task: 'success' }[s] || '')
const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')

onMounted(() => { loadStats(); loadTasks() })
</script>

<style scoped>
.task-center { padding: 16px; }
.stat { display: flex; justify-content: space-between; align-items: center; }
.stat b { font-size: 24px; }
</style>
