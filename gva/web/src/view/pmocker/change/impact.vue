<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="变更ID">
          <el-input v-model="searchInfo.changeId" placeholder="变更ID" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadReport">
            <svg-icon icon="lucide:analytics" /> 生成报告
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    <div v-if="report" class="gva-table-box">
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
            <el-tabs v-model="activeTab">
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
    <el-empty v-else description="输入变更ID并生成影响报告" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { getChangeImpactReport } from '@/api/pmocker/change'

defineOptions({ name: 'PmockerChangeImpact' })

const searchInfo = ref({})
const report = ref(null)
const activeTab = ref('scope')

const loadReport = async () => {
  if (!searchInfo.value.changeId) return
  const res = await getChangeImpactReport(searchInfo.value)
  if (res.code === 0) {
    report.value = res.data
  }
}
</script>
