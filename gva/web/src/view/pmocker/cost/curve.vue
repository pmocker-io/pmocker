<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true">
        <el-form-item>
          <el-button type="primary" @click="loadData">
            <svg-icon icon="lucide:line-chart" /> 生成S曲线
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <div v-if="chartData" ref="chartRef" class="w-full h-[500px]" />
      <el-empty v-else description="点击上方按钮生成S曲线" />
    </div>
  </div>
</template>

<script setup>
import { ref, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getCostItems } from '@/api/pmocker/cost'

defineOptions({ name: 'PmockerCostCurve' })

const chartRef = ref(null)
const chartData = ref(null)
let chartInstance = null

const loadData = async () => {
  const res = await getCostItems({ projectId: projectStore.projectId })
  if (res.code === 0) {
    chartData.value = res.data
    await nextTick()
    renderChart(res.data)
  }
}

const renderChart = (data) => {
  if (chartInstance) chartInstance.dispose()
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)

  const periods = data.periods || []
  const pvData = periods.map(p => p.pv || 0)
  const evData = periods.map(p => p.ev || 0)
  const acData = periods.map(p => p.ac || 0)

  chartInstance.setOption({
    title: { text: 'S曲线 - 成本绩效', left: 'center' },
    tooltip: { trigger: 'axis' },
    legend: { data: ['PV 计划价值', 'EV 挣值', 'AC 实际成本'], bottom: 0 },
    xAxis: { type: 'category', data: periods.map(p => p.label) },
    yAxis: { type: 'value', name: '金额' },
    series: [
      { name: 'PV 计划价值', type: 'line', smooth: true, data: pvData },
      { name: 'EV 挣值', type: 'line', smooth: true, data: evData },
      { name: 'AC 实际成本', type: 'line', smooth: true, data: acData }
    ]
  })
}

onBeforeUnmount(() => {
  if (chartInstance) chartInstance.dispose()
})
</script>
