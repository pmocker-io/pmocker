<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="handleCreateBaseline">
          <svg-icon icon="lucide:git-branch" /> 创建基线
        </el-button>
      </div>
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="ID" prop="ID" width="80" />
        <el-table-column label="基线名称" prop="name" min-width="200" />
        <el-table-column label="版本" prop="version" width="100" />
        <el-table-column label="范围项数" prop="itemCount" width="100" />
        <el-table-column label="创建时间" prop="CreatedAt" width="180">
          <template #default="{ row }">
            {{ formatDate(row.CreatedAt) }}
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getScopeItems, createScopeBaseline } from '@/api/pmocker/scope'

defineOptions({ name: 'PmockerScopeBaseline' })

const tableData = ref([])

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
}

const loadData = async () => {
  const res = await getScopeItems({ type: 'baseline' })
  if (res.code === 0) {
    tableData.value = res.data.list || []
  }
}

const handleCreateBaseline = async () => {
  const res = await createScopeBaseline({})
  if (res.code === 0) {
    ElMessage.success('基线创建成功')
    loadData()
  }
}

loadData()
</script>
