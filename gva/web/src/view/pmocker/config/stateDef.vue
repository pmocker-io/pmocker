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
      title="状态流转由各模块状态机管理，此处为已发布状态定义的只读视图；状态编辑放入后续迭代。"
      style="margin-bottom: 12px"
    />

    <div class="gva-table-box">
      <el-table :data="tableData" row-key="ID" size="small">
        <el-table-column label="实体类型" prop="entityType" min-width="140" />
        <el-table-column label="状态值" prop="status" min-width="120" />
        <el-table-column label="显示名" prop="label" min-width="120" />
        <el-table-column label="Tag类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="row.tagType || 'info'">{{ row.tagType || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="排序" prop="sort" width="90" />
      </el-table>
      <el-empty v-if="tableData.length === 0" description="暂无已发布状态定义" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { listStateDefsPublic } from '@/api/pmocker/config'
import { listEntityTypes } from '@/api/pmocker/config'

defineOptions({ name: 'PmockerConfigStateDef' })

const searchInfo = reactive({ entityType: '' })
const entityTypes = ref([])
const tableData = ref([])

const loadEntityTypes = async () => {
  const res = await listEntityTypes({ includeDraft: false })
  if (res.code === 0) {
    entityTypes.value = res.data || []
  }
}

const loadData = async () => {
  const params = searchInfo.entityType ? { entityType: searchInfo.entityType } : {}
  const res = await listStateDefsPublic(params)
  if (res.code === 0) {
    tableData.value = res.data || []
  }
}

onMounted(() => {
  loadEntityTypes()
  loadData()
})
</script>
