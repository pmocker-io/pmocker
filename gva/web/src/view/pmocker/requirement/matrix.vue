<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="项目">
          <el-input v-model="searchInfo.projectId" placeholder="项目ID" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadMatrix">查询</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <el-table :data="matrixData" border>
        <el-table-column label="需求" prop="requirementTitle" fixed width="200" />
        <el-table-column
          v-for="scope in scopeColumns"
          :key="scope.ID"
          :label="scope.title"
          :prop="scope.ID"
          width="120"
          align="center"
        >
          <template #default="{ row }">
            <el-tag v-if="getCell(row, scope.ID)" type="success" size="small">
              <svg-icon icon="lucide:check" />
            </el-tag>
            <span v-else class="text-muted-foreground">-</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { getRequirementTraceMatrix } from '@/api/pmocker/requirement'

defineOptions({ name: 'PmockerRequirementMatrix' })

const searchInfo = ref({})
const matrixData = ref([])
const scopeColumns = ref([])

const getCell = (row, scopeId) => {
  const cell = row.cells?.find(c => c.scopeId === scopeId)
  return cell?.linked
}

const loadMatrix = async () => {
  const res = await getRequirementTraceMatrix(searchInfo.value)
  if (res.code === 0) {
    matrixData.value = res.data.requirements || []
    scopeColumns.value = res.data.scopeItems || []
  }
}

loadMatrix()
</script>
