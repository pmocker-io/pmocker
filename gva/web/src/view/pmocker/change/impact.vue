<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="变更请求">
          <el-select
            v-model="searchInfo.changeId"
            placeholder="请选择变更请求"
            filterable
            clearable
            style="width: 320px"
            @change="handleLoad"
          >
            <el-option
              v-for="item in changeOptions"
              :key="item.ID"
              :label="`#${item.ID} ${item.title}`"
              :value="item.ID"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLoad">
            <svg-icon icon="lucide:analytics" /> 加载
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <div v-if="!searchInfo.changeId" class="gva-table-box">
      <el-empty description="请选择变更请求后加载影响分析或变更对比" />
    </div>

    <div v-else class="gva-table-box">
      <el-tabs v-model="activeTab" @tab-change="onTabChange">
        <el-tab-pane label="影响分析" name="impact">
          <div v-if="report">
            <el-row :gutter="16">
              <el-col :span="8">
                <el-card>
                  <template #header><span class="font-medium">影响概览</span></template>
                  <el-descriptions :column="1" border>
                    <el-descriptions-item label="范围影响">{{ report.scopeImpact || 0 }} 项</el-descriptions-item>
                    <el-descriptions-item label="进度影响">{{ report.scheduleImpact || 0 }} 天</el-descriptions-item>
                    <el-descriptions-item label="成本影响">{{ report.costImpact || 0 }} 元</el-descriptions-item>
                    <el-descriptions-item label="风险影响">{{ report.riskImpact || 0 }} 项</el-descriptions-item>
                  </el-descriptions>
                </el-card>
              </el-col>
              <el-col :span="16">
                <el-card>
                  <template #header><span class="font-medium">受影响项</span></template>
                  <el-tabs v-model="affectedTab">
                    <el-tab-pane label="范围项" name="scope">
                      <el-table :data="report.scopeItems || []">
                        <el-table-column label="名称" prop="title" />
                        <el-table-column label="影响类型" prop="impactType" width="120" />
                      </el-table>
                    </el-tab-pane>
                    <el-tab-pane label="任务" name="task">
                      <el-table :data="report.tasks || []">
                        <el-table-column label="任务" prop="title" />
                        <el-table-column label="影响类型" prop="impactType" width="120" />
                      </el-table>
                    </el-tab-pane>
                    <el-tab-pane label="成本" name="cost">
                      <el-table :data="report.costs || []">
                        <el-table-column label="项" prop="title" />
                        <el-table-column label="影响类型" prop="impactType" width="120" />
                      </el-table>
                    </el-tab-pane>
                    <el-tab-pane label="风险" name="risk">
                      <el-table :data="report.risks || []">
                        <el-table-column label="风险" prop="title" />
                        <el-table-column label="影响类型" prop="impactType" width="120" />
                      </el-table>
                    </el-tab-pane>
                  </el-tabs>
                </el-card>
              </el-col>
            </el-row>
          </div>
          <el-empty v-else description="点击「加载」生成影响报告" />
        </el-tab-pane>

        <el-tab-pane label="变更对比" name="diff">
          <el-alert
            type="info"
            :closable="false"
            show-icon
            title="基线快照在提交评审（CCB）时自动生成；未提交评审的变更将显示全部为新增字段。"
            class="mb-3"
          />
          <el-table :data="diffList" :row-class-name="diffRowClass" border>
            <el-table-column label="字段名" prop="fieldLabel" min-width="140" />
            <el-table-column label="基线值" prop="oldValue" min-width="220">
              <template #default="{ row }">
                <span :class="{ 'diff-empty': !row.oldValue }">{{ row.oldValue || '（空）' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="当前值" prop="newValue" min-width="220">
              <template #default="{ row }">
                <span :class="{ 'diff-empty': !row.newValue }">{{ row.newValue || '（空）' }}</span>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!diffList.length" description="暂无 diff 数据" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getChangeImpactReport, getChangeDiff, getChangeList } from '@/api/pmocker/change'
import { useProjectStore } from '@/pinia'

defineOptions({ name: 'PmockerChangeImpact' })
const projectStore = useProjectStore()

const searchInfo = ref({ changeId: '' })
const report = ref(null)
const diffList = ref([])
const changeOptions = ref([])
const activeTab = ref('impact')
const affectedTab = ref('scope')

const loadChangeOptions = async () => {
  const res = await getChangeList({ offset: 0, limit: 100, projectId: projectStore.projectId })
  if (res.code === 0) {
    changeOptions.value = res.data.list || []
  }
}

const loadReport = async () => {
  if (!searchInfo.value.changeId) return
  const res = await getChangeImpactReport({ id: searchInfo.value.changeId })
  if (res.code === 0) {
    report.value = res.data
  }
}

const loadDiff = async () => {
  if (!searchInfo.value.changeId) return
  const res = await getChangeDiff({ id: searchInfo.value.changeId })
  if (res.code === 0) {
    diffList.value = res.data || []
  }
}

const handleLoad = () => {
  if (activeTab.value === 'impact') {
    loadReport()
  } else {
    loadDiff()
  }
}

const onTabChange = (tab) => {
  if (tab === 'impact' && searchInfo.value.changeId && !report.value) {
    loadReport()
  } else if (tab === 'diff' && searchInfo.value.changeId && !diffList.value.length) {
    loadDiff()
  }
}

// 行高亮：变化黄色 / 新增(基线空)绿色 / 删除(当前空)红色
const diffRowClass = ({ row }) => {
  if (!row.changed) return ''
  if (!row.oldValue) return 'diff-row-added'
  if (!row.newValue) return 'diff-row-removed'
  return 'diff-row-changed'
}

onMounted(() => {
  loadChangeOptions()
})
</script>

<style scoped>
:deep(.diff-row-changed td) {
  background-color: #fff3cd !important;
}
:deep(.diff-row-added td) {
  background-color: #d4edda !important;
}
:deep(.diff-row-removed td) {
  background-color: #f8d7da !important;
}
.diff-empty {
  color: #c0c4cc;
  font-style: italic;
}
.mb-3 {
  margin-bottom: 12px;
}
</style>
