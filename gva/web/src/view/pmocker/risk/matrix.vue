<template>
  <div>
    <div class="gva-table-box">
      <div ref="chartRef" class="w-full h-[500px]" />
      <div class="mt-4">
        <el-table :data="riskList" border>
          <el-table-column label="风险" prop="title" min-width="200" />
          <el-table-column label="概率" prop="probability" width="80" />
          <el-table-column label="影响" prop="impact" width="80" />
          <el-table-column label="风险值" width="100">
            <template #default="{ row }">
              <el-tag :type="levelType(row.probability * row.impact)">
                {{ row.probability * row.impact }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { getRiskMatrix } from '@/api/pmocker/risk'

defineOptions({ name: 'PmockerRiskMatrix' })

const chartRef = ref(null)
const riskList = ref([])
let chartInstance = null

const levelType = (value) => {
  if (value <= 3) return 'success'
  if (value <= 9) return 'warning'
  return 'danger'
}

const loadData = async () => {
  const res = await getRiskMatrix({ projectId: projectStore.projectId })
  if (res.code === 0) {
    riskList.value = res.data.risks || []
    await nextTick()
    renderChart(res.data)
  }
}

const renderChart = (data) => {
  if (chartInstance) chartInstance.dispose()
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)

  const xLabels = ['1', '2', '3', '4', '5']
  const yLabels = ['1', '2', '3', '4', '5']

  const heatmapData = []
  for (let i = 0; i < 5; i++) {
    for (let j = 0; j < 5; j++) {
      const value = (i + 1) * (j + 1)
      heatmapData.push([j, i, value])
    }
  }

  const risks = data.risks || []
  const scatterData = risks.map(r => ({
    value: [(r.probability || 1) - 1, (r.impact || 1) - 1, r.title]
  }))

  chartInstance.setOption({
    tooltip: {
      formatter: (params) => {
        if (params.seriesIndex === 1) {
          return params.value[2]
        }
        return `概率: ${yLabels[params.value[1]]}, 影响: ${xLabels[params.value[0]]}, 风险值: ${params.value[2]}`
      }
    },
    grid: { height: '70%', top: '10%' },
    xAxis: { type: 'category', data: xLabels, name: '影响', splitArea: { show: true } },
    yAxis: { type: 'category', data: yLabels, name: '概率', splitArea: { show: true } },
    visualMap: {
      min: 1, max: 25,
      calculable: true,
      orient: 'horizontal',
      left: 'center',
      bottom: '5%',
      inRange: { color: ['#67c23a', '#e6a23c', '#f56c6c'] }
    },
    series: [
      {
        name: '风险区域',
        type: 'heatmap',
        data: heatmapData,
        label: { show: true }
      },
      {
        name: '风险点',
        type: 'scatter',
        data: scatterData,
        symbolSize: 20,
        itemStyle: { color: '#000' }
      }
    ]
  })
}

onMounted(() => {
  loadData()
})

onBeforeUnmount(() => {
  if (chartInstance) chartInstance.dispose()
})
</script>
