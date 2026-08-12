<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true">
        <el-form-item>
          <el-button type="primary" @click="analyzeCPM">
            <svg-icon icon="lucide:git-fork" /> 执行 CPM 分析
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <div v-if="cpmResult" ref="chartRef" class="w-full h-[500px]" />
      <el-empty v-else description="点击上方按钮执行 CPM 分析" />
    </div>
    <div v-if="cpmResult" class="gva-table-box mt-4">
      <el-table :data="cpmResult.criticalPath || []" border>
        <el-table-column label="任务" prop="title" />
        <el-table-column label="最早开始" prop="earlyStart" width="120" />
        <el-table-column label="最早完成" prop="earlyFinish" width="120" />
        <el-table-column label="最迟开始" prop="lateStart" width="120" />
        <el-table-column label="最迟完成" prop="lateFinish" width="120" />
        <el-table-column label="总浮动" prop="totalFloat" width="100" />
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { analyzeScheduleCPM } from '@/api/pmocker/schedule'
import { cssVar } from '@/utils/theme'

defineOptions({ name: 'PmockerScheduleCritical' })

const chartRef = ref(null)
const cpmResult = ref(null)
let chartInstance = null

const analyzeCPM = async () => {
  const res = await analyzeScheduleCPM({ projectId: projectStore.projectId })
  if (res.code === 0) {
    cpmResult.value = res.data
    await nextTick()
    renderChart(res.data)
  }
}

const renderChart = (data) => {
  if (chartInstance) chartInstance.dispose()
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)

  const nodes = (data.tasks || []).map(t => ({
    id: String(t.ID),
    name: t.title,
    x: t.earlyStart * 100,
    y: 0
  }))

  const links = (data.dependencies || []).map(d => ({
    source: String(d.from),
    target: String(d.to),
    label: { show: true }
  }))

  chartInstance.setOption({
    title: { text: '关键路径图', left: 'center' },
    tooltip: {},
    series: [{
      type: 'graph',
      layout: 'none',
      symbolSize: 50,
      roam: true,
      label: { show: true },
      edgeSymbol: ['none', 'arrow'],
      data: nodes,
      links: links,
      lineStyle: { color: cssVar('--el-text-color-secondary', '#333'), curveness: 0.3 }
    }]
  })
}

onBeforeUnmount(() => {
  if (chartInstance) chartInstance.dispose()
})
</script>
