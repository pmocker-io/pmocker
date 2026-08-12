<template>
  <div class="variance-page">
    <el-card shadow="never">
      <el-form :inline="true" size="small">
        <el-form-item label="项目ID">
          <el-input v-model="projectId" style="width: 120px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadAll">计算偏差</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-row :gutter="12" class="mt-3">
      <el-col :span="8">
        <el-card shadow="never" header="SPI 进度绩效">
          <div ref="spiRef" style="height: 240px" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" header="CPI 成本绩效">
          <div ref="cpiRef" style="height: 240px" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" header="偏差（EV-PV/AC）">
          <div ref="barRef" style="height: 240px" />
        </el-card>
      </el-col>
    </el-row>
    <el-card v-if="alerts.length" shadow="never" class="mt-3" header="预警列表">
      <el-table :data="alerts" size="small" border>
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column prop="severity" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : 'warning'" size="small">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="detail" label="详情" />
      </el-table>
    </el-card>
  </div>
</template>
<script setup>
import { ref, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { calcVariance, getAlerts } from '@/api/pmocker/variance'
import { cssVar } from '@/utils/theme'
const projectId = ref('')
const spiRef = ref(null)
const cpiRef = ref(null)
const barRef = ref(null)
const alerts = ref([])
let charts = []
const renderGauge = (dom, value, name) => {
  if (!dom) return
  const inst = echarts.init(dom)
  charts.push(inst)
  inst.setOption({
    series: [{
      type: 'gauge', min: 0, max: 2, splitNumber: 8,
      progress: { show: true, width: 18 },
      axisLine: { lineStyle: { width: 18 } },
      pointer: { width: 5 },
      detail: { valueAnimation: true, formatter: '{value}', fontSize: 20, offsetCenter: [0, '70%'] },
      data: [{ value: Number(value.toFixed(2)), name }]
    }]
  })
}
const renderBar = (dom, sv, cv) => {
  if (!dom) return
  const success = cssVar('--el-color-success', '#67c23a')
  const danger = cssVar('--el-color-danger', '#f56c6c')
  const inst = echarts.init(dom)
  charts.push(inst)
  inst.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['进度偏差 SV', '成本偏差 CV'] },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar', barWidth: 40,
      data: [
        { value: Number(sv.toFixed(2)), itemStyle: { color: sv >= 0 ? success : danger } },
        { value: Number(cv.toFixed(2)), itemStyle: { color: cv >= 0 ? success : danger } }
      ]
    }]
  })
}
const loadAll = async () => {
  if (!projectId.value) return ElMessage.warning('请输入项目ID')
  charts.forEach(c => c.dispose()); charts = []
  const [vr, ar] = await Promise.all([
    calcVariance({ projectId: projectId.value }),
    getAlerts({ projectId: projectId.value })
  ])
  if (vr.code === 0 && vr.data) {
    await nextTick()
    renderGauge(spiRef.value, vr.data.spi, 'SPI')
    renderGauge(cpiRef.value, vr.data.cpi, 'CPI')
    renderBar(barRef.value, vr.data.sv, vr.data.cv)
  }
  if (ar.code === 0) alerts.value = ar.data || []
}
onBeforeUnmount(() => { charts.forEach(c => c.dispose()); charts = [] })
</script>
