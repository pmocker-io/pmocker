<template>
  <div class="baseline-page">
    <el-card shadow="never">
      <el-form :inline="true" size="small">
        <el-form-item label="项目ID">
          <el-input v-model="projectId" style="width: 120px" />
        </el-form-item>
        <el-form-item label="基线类型">
          <el-select v-model="baselineType" style="width: 140px" clearable>
            <el-option label="计划基线" value="schedule" />
            <el-option label="成本基线" value="cost" />
            <el-option label="范围基线" value="scope" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">查询</el-button>
          <el-button type="success" @click="createBaseline">生成基线</el-button>
        </el-form-item>
      </el-form>
      <el-table :data="baselines" size="small" border highlight-current-row @row-click="selectBaseline">
        <el-table-column prop="ID" label="ID" width="80" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">{{ typeText(row.type) }}</template>
        </el-table-column>
        <el-table-column prop="CreatedAt" label="创建时间" width="180" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click.stop="compare(row)">对比</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-card v-if="diffs.length" shadow="never" style="margin-top: 12px" header="基线对比差异">
      <el-table :data="diffs" size="small" border>
        <el-table-column prop="entityTitle" label="实体" width="160" />
        <el-table-column prop="entityType" label="类型" width="100" />
        <el-table-column prop="fieldKey" label="字段" width="140" />
        <el-table-column prop="baselineVal" label="基线值" />
        <el-table-column prop="currentVal" label="当前值" />
        <el-table-column prop="change" label="变化" width="100">
          <template #default="{ row }">
            <el-tag :type="tagType(row.change)" size="small">{{ row.change }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { listBaselines, createBaseline as createBaselineApi, compareBaseline } from '@/api/pmocker/baseline'
const projectId = ref('')
const baselineType = ref('')
const baselines = ref([])
const diffs = ref([])
const typeText = (t) => ({ schedule: '计划基线', cost: '成本基线', scope: '范围基线' }[t] || t)
const tagType = (c) => ({ added: 'success', removed: 'danger', modified: 'warning' }[c] || 'info')
const loadList = async () => {
  if (!projectId.value) return ElMessage.warning('请输入项目ID')
  const res = await listBaselines({ projectId: projectId.value, type: baselineType.value })
  if (res.code === 0) baselines.value = res.data || []
}
const createBaseline = async () => {
  if (!projectId.value || !baselineType.value) return ElMessage.warning('请输入项目ID并选择类型')
  const res = await createBaselineApi({ projectId: Number(projectId.value), type: baselineType.value })
  if (res.code === 0) { ElMessage.success('基线已生成'); loadList() }
}
const selectBaseline = (row) => { compare(row) }
const compare = async (row) => {
  const res = await compareBaseline({ baselineId: row.ID })
  if (res.code === 0) { diffs.value = res.data || []; if (!diffs.value.length) ElMessage.info('无差异') }
}
</script>
