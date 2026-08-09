<template>
  <div class="task-center">
    <el-page-header content="个人任务中心" />

    <el-row :gutter="12" style="margin-top: 12px">
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>任务总数</span><b>{{ stats.total }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>完成率</span><b style="color:#67C23A">{{ (stats.doneRate || 0).toFixed(1) }}%</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>逾期数</span><b style="color:#F56C6C">{{ stats.overdueCount }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>进行中</span><b style="color:#409EFF">{{ stats.doingCount }}</b></div></el-card></el-col>
    </el-row>

    <div style="margin-top: 16px">
      <VerticalTabLayout
        v-model:activeTab="activeTab"
        :tabs="tabs"
        @tab-change="switchTab"
      >
        <!-- 表格上方的工具栏/按钮区（预留给后续扩展：如新增、批量分配、导出等） -->
        <template #toolbar>
          <div class="task-toolbar">
            <span class="tab-title">{{ currentTabLabel }}</span>
          </div>
        </template>

        <!-- 表格主内容区 -->
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
      </VerticalTabLayout>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { getMyTasks, getFocusedTasks, getTaskStats } from '@/api/pmocker/taskCenter'
import VerticalTabLayout from '../components/VerticalTabLayout.vue'

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

const currentTabLabel = computed(() => {
  const cur = tabs.value.find(t => t.name === activeTab.value)
  return cur ? cur.label : ''
})

const switchTab = async (name) => {
  await loadTasks()
  const cur = tabs.value.find(t => t.name === name)
  if (cur) {
    cur.count = tasks.value.length
    tabs.value = [...tabs.value]
  }
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
  const results = await Promise.all(tabs.value.map(t => fetchTabCount(t.name)))
  tabs.value = tabs.value.map((t, i) => ({ ...t, count: results[i] }))
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
.task-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.tab-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}
</style>
