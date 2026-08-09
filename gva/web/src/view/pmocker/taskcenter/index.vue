<template>
  <div class="task-center">
    <el-page-header content="个人任务中心" />

    <el-row :gutter="12" style="margin-top: 12px">
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>任务总数</span><b>{{ stats.total }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>完成率</span><b style="color:#67C23A">{{ (stats.doneRate || 0).toFixed(1) }}%</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>逾期数</span><b style="color:#F56C6C">{{ stats.overdueCount }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>进行中</span><b style="color:#409EFF">{{ stats.doingCount }}</b></div></el-card></el-col>
    </el-row>

    <div class="task-body">
      <div class="task-tabs">
        <div
          v-for="t in tabs"
          :key="t.name"
          class="tab-item"
          :class="{ active: activeTab === t.name }"
          @click="switchTab(t.name)"
        >
          {{ t.label }}
          <span v-if="t.count !== null" class="tab-badge">{{ t.count }}</span>
        </div>
      </div>
      <div class="task-content">
        <el-table :data="tasks" border size="small" v-loading="loading">
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
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getMyTasks, getFocusedTasks, getTaskStats } from '@/api/pmocker/taskCenter'

const activeTab = ref('todo')
const tasks = ref([])
const loading = ref(false)
const stats = reactive({ total: 0, done: 0, doneRate: 0, overdueCount: 0, todoCount: 0, doingCount: 0 })

const tabs = ref([
  { name: 'todo', label: '我的待办', count: null },
  { name: 'doing', label: '进行中', count: null },
  { name: 'done', label: '已完成', count: null },
  { name: 'overdue', label: '已逾期', count: null },
  { name: 'focused', label: '我关注的', count: null }
])

const switchTab = (name) => {
  activeTab.value = name
  loadTasks()
}

const loadStats = async () => {
  const res = await getTaskStats()
  if (res.code === 0) Object.assign(stats, res.data)
}

const fetchTabCount = async (tabName) => {
  try {
    let res
    if (tabName === 'focused') {
      res = await getFocusedTasks()
    } else {
      res = await getMyTasks({ status: tabName })
    }
    return res.code === 0 ? (res.data?.length ?? 0) : 0
  } catch {
    return 0
  }
}

const preloadCounts = async () => {
  const others = tabs.value.filter(t => t.name !== activeTab.value)
  const results = await Promise.all(others.map(t => fetchTabCount(t.name)))
  results.forEach((cnt, i) => { others[i].count = cnt })
}

const loadTasks = async () => {
  loading.value = true
  try {
    let res
    if (activeTab.value === 'focused') {
      res = await getFocusedTasks()
    } else {
      res = await getMyTasks({ status: activeTab.value })
    }
    if (res.code === 0) {
      tasks.value = res.data || []
      const cur = tabs.value.find(t => t.name === activeTab.value)
      if (cur) cur.count = tasks.value.length
    }
  } finally {
    loading.value = false
  }
}

const sourceLabel = (s) => ({ project_task: '项目任务', issue_task: '问题任务', change_task: '变更任务', deliverable_task: '交付物任务' }[s] || s)
const sourceTag = (s) => ({ project_task: '', issue_task: 'warning', change_task: 'danger', deliverable_task: 'success' }[s] || '')
const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')

onMounted(() => { loadStats(); preloadCounts(); loadTasks() })
</script>

<style scoped>
.task-center { padding: 16px; }
.stat { display: flex; justify-content: space-between; align-items: center; }
.stat b { font-size: 24px; }

.task-body {
  display: flex;
  margin-top: 16px;
  min-height: 500px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  overflow: hidden;
}
.task-tabs {
  width: 140px;
  flex-shrink: 0;
  background: #fafafa;
  border-right: 1px solid #e4e7ed;
}
.tab-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 44px;
  line-height: 44px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
  color: #606266;
  transition: all 0.2s;
  user-select: none;
}
.tab-item:hover {
  background: #ecf5ff;
  color: #409eff;
}
.tab-item.active {
  background: #409eff;
  color: #fff;
  font-weight: 500;
}
.tab-item.active:hover {
  background: #409eff;
  color: #fff;
}
.tab-badge {
  display: inline-block;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  line-height: 18px;
  text-align: center;
  font-size: 11px;
  border-radius: 9px;
  background: #f0f0f0;
  color: #909399;
}
.tab-item.active .tab-badge {
  background: rgba(255, 255, 255, 0.3);
  color: #fff;
}
.task-content {
  flex: 1;
  padding: 12px;
  overflow: auto;
}
</style>
