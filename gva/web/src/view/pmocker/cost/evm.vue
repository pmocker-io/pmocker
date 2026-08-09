<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true">
        <el-form-item>
          <el-button type="primary" @click="loadEVM">
            <svg-icon icon="lucide:trending-up" /> 执行挣值分析
          </el-button>
        </el-form-item>
      </el-form>
    </div>
    <div v-if="evmData" class="gva-table-box">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-card>
            <template #header><span class="font-medium">PV 计划价值</span></template>
            <div class="text-2xl font-bold">¥{{ evmData.PV?.toFixed(2) || '0.00' }}</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <template #header><span class="font-medium">EV 挣值</span></template>
            <div class="text-2xl font-bold">¥{{ evmData.EV?.toFixed(2) || '0.00' }}</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <template #header><span class="font-medium">AC 实际成本</span></template>
            <div class="text-2xl font-bold">¥{{ evmData.AC?.toFixed(2) || '0.00' }}</div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card>
            <template #header><span class="font-medium">BAC 完工预算</span></template>
            <div class="text-2xl font-bold">¥{{ evmData.BAC?.toFixed(2) || '0.00' }}</div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" class="mt-4">
        <el-col :span="8">
          <el-card>
            <template #header><span class="font-medium">SV 进度偏差</span></template>
            <div class="text-xl font-bold" :class="evmData.SV >= 0 ? 'text-success' : 'text-error'">
              ¥{{ evmData.SV?.toFixed(2) || '0.00' }}
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card>
            <template #header><span class="font-medium">CV 成本偏差</span></template>
            <div class="text-xl font-bold" :class="evmData.CV >= 0 ? 'text-success' : 'text-error'">
              ¥{{ evmData.CV?.toFixed(2) || '0.00' }}
            </div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card>
            <template #header><span class="font-medium">VAC 完工偏差</span></template>
            <div class="text-xl font-bold" :class="evmData.VAC >= 0 ? 'text-success' : 'text-error'">
              ¥{{ evmData.VAC?.toFixed(2) || '0.00' }}
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" class="mt-4">
        <el-col :span="8">
          <el-card>
            <template #header><span class="font-medium">SPI 进度绩效</span></template>
            <div class="text-xl font-bold">{{ evmData.SPI?.toFixed(2) || '0.00' }}</div>
            <div class="text-muted-foreground text-sm">{{ spiLabel(evmData.SPI) }}</div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card>
            <template #header><span class="font-medium">CPI 成本绩效</span></template>
            <div class="text-xl font-bold">{{ evmData.CPI?.toFixed(2) || '0.00' }}</div>
            <div class="text-muted-foreground text-sm">{{ cpiLabel(evmData.CPI) }}</div>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card>
            <template #header><span class="font-medium">EAC 完工估算</span></template>
            <div class="text-xl font-bold">¥{{ evmData.EAC?.toFixed(2) || '0.00' }}</div>
          </el-card>
        </el-col>
      </el-row>
    </div>
    <el-empty v-else description="点击上方按钮执行挣值分析" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { analyzeCostEVM } from '@/api/pmocker/cost'

defineOptions({ name: 'PmockerCostEVM' })

const evmData = ref(null)

const spiLabel = (val) => {
  if (val >= 1) return '进度超前'
  if (val >= 0.9) return '进度正常'
  return '进度落后'
}

const cpiLabel = (val) => {
  if (val >= 1) return '成本节约'
  if (val >= 0.9) return '成本正常'
  return '成本超支'
}

const loadEVM = async () => {
  const res = await analyzeCostEVM({ projectId: projectStore.projectId })
  if (res.code === 0) {
    evmData.value = res.data
  }
}
</script>
