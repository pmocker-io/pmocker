<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="实体类型">
          <el-select v-model="searchInfo.entityType" clearable filterable placeholder="请选择实体类型" style="width: 240px" @change="loadData">
            <el-option v-for="et in entityTypes" :key="et.typeCode" :label="et.name + ' (' + et.typeCode + ')'" :value="et.typeCode" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="字段定义由 EAV Schema 驱动（元表 pm_field_defs），当前 MVP 提供查看与状态管理；字段新增/编辑 API 规划中。"
      style="margin-bottom: 12px"
    />

    <div class="gva-table-box">
      <el-table :data="tableData" row-key="id" size="small">
        <el-table-column label="字段Key" prop="field_key" min-width="140" />
        <el-table-column label="标签" prop="field_label" min-width="120" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.data_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="默认值" prop="default_value" min-width="120">
          <template #default="{ row }">{{ row.default_value || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">查看</el-button>
            <el-tooltip content="复制需字段管理 API 返回 id，当前接口暂不支持" :disabled="!!row.id">
              <span class="tooltip-wrap">
                <el-button type="primary" link size="small" :disabled="!row.id" @click="handleCopy(row)">复制</el-button>
              </span>
            </el-tooltip>
            <el-tooltip content="状态流转需字段管理 API 返回 id，当前接口暂不支持" :disabled="!!row.id">
              <span class="tooltip-wrap">
                <el-button type="warning" link size="small" :disabled="!row.id" @click="handleTransition(row, 'reviewing')">流转</el-button>
              </span>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="tableData.length === 0" description="请选择实体类型查看字段定义" />
    </div>

    <el-dialog v-model="detailVisible" title="字段定义详情" width="520px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="字段Key">{{ detail.field_key }}</el-descriptions-item>
        <el-descriptions-item label="标签">{{ detail.field_label }}</el-descriptions-item>
        <el-descriptions-item label="数据类型">{{ detail.data_type }}</el-descriptions-item>
        <el-descriptions-item label="默认值">{{ detail.default_value || '-' }}</el-descriptions-item>
        <el-descriptions-item label="所属实体类型">{{ detail.entity_type }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusLabel(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="选项JSON">{{ detail.options_json || '-' }}</el-descriptions-item>
        <el-descriptions-item label="校验规则">{{ detail.validators || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSchema } from '@/api/pmocker/schema'
import { listEntityTypes, transitionConfig, copyAsDraft } from '@/api/pmocker/config'

defineOptions({ name: 'PmockerConfigFieldDef' })

const searchInfo = reactive({ entityType: '' })
const entityTypes = ref([])
const tableData = ref([])
const detailVisible = ref(false)
const detail = ref({})

const STATUS_TABLE = {
  draft: { label: '草稿', type: 'info' },
  reviewing: { label: '评审中', type: 'warning' },
  published: { label: '已发布', type: 'success' },
  archived: { label: '已归档', type: 'info' }
}
const statusLabel = (s) => STATUS_TABLE[s]?.label || s || '-'
const statusType = (s) => STATUS_TABLE[s]?.type || 'info'

const loadEntityTypes = async () => {
  const res = await listEntityTypes({ includeDraft: true })
  if (res.code === 0) {
    entityTypes.value = res.data || []
  }
}

const loadData = async () => {
  if (!searchInfo.entityType) {
    tableData.value = []
    return
  }
  const res = await getSchema(searchInfo.entityType)
  if (res.code === 0) {
    tableData.value = (res.data && res.data.fields) || []
  }
}

const openDetail = (row) => {
  detail.value = row
  detailVisible.value = true
}

const handleCopy = async (row) => {
  if (!row.id) {
    ElMessage.warning('当前接口未返回字段 id，暂不支持复制')
    return
  }
  const res = await copyAsDraft({ table: 'pm_field_defs', id: row.id })
  if (res.code === 0) {
    ElMessage.success('复制成功')
    loadData()
  }
}

const handleTransition = async (row, to) => {
  if (!row.id) {
    ElMessage.warning('当前接口未返回字段 id，暂不支持状态流转')
    return
  }
  const res = await transitionConfig({ table: 'pm_field_defs', id: row.id, from: row.status, to })
  if (res.code === 0) {
    ElMessage.success('状态流转成功')
    loadData()
  }
}

onMounted(loadEntityTypes)
</script>

<style scoped>
.tooltip-wrap {
  display: inline-block;
}
</style>
