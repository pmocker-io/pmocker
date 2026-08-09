<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="交付物ID">
          <el-input v-model="searchInfo.deliverableId" placeholder="留空查询全部" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadTrace">查询</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <el-table :data="traceData" border row-key="ID">
        <el-table-column label="交付物" prop="deliverableTitle" min-width="200" />
        <el-table-column label="版本" prop="version" width="80" />
        <el-table-column label="关联需求" prop="requirementTitle" width="200" />
        <el-table-column label="关联范围项" prop="scopeItemTitle" width="200" />
        <el-table-column label="状态" prop="status" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.status }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { getDeliverableTraceReport } from '@/api/pmocker/deliverable'

defineOptions({ name: 'PmockerDeliverableTrace' })

const searchInfo = ref({})
const traceData = ref([])

const loadTrace = async () => {
  const res = await getDeliverableTraceReport({ projectId: projectStore.projectId, ...searchInfo.value })
  if (res.code === 0) {
    traceData.value = res.data.list || []
  }
}

loadTrace()
</script>
