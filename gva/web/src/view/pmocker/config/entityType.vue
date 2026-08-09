<template>
  <div>
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="实体类型（元表 pm_entity_types）由状态机统一管理：草稿 → 评审 → 发布 → 归档，复制可基于任意状态生成新草稿。"
      style="margin-bottom: 12px"
    />
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="handleAdd">
          <svg-icon icon="lucide:plus" /> 新增
        </el-button>
        <el-radio-group v-model="includeDraft" size="small" @change="loadData">
          <el-radio-button :label="false">仅发布</el-radio-button>
          <el-radio-button :label="true">含草稿</el-radio-button>
        </el-radio-group>
        <el-button @click="handleExport">
          <svg-icon icon="lucide:download" /> 导出YAML
        </el-button>
        <el-button @click="loadData">
          <svg-icon icon="lucide:refresh-cw" /> 刷新
        </el-button>
      </div>

      <el-table :data="tableData" row-key="ID" size="small">
        <el-table-column label="编码" prop="typeCode" min-width="140" />
        <el-table-column label="名称" prop="name" min-width="140" />
        <el-table-column label="模块" prop="moduleCode" width="120">
          <template #default="{ row }">{{ row.moduleCode || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 'draft'" type="primary" link size="small" @click="handleTransition(row, 'reviewing')">提交评审</el-button>
            <el-button v-if="row.status === 'reviewing'" type="success" link size="small" @click="handleTransition(row, 'published')">发布</el-button>
            <el-button v-if="row.status === 'reviewing'" type="warning" link size="small" @click="handleTransition(row, 'draft')">退回</el-button>
            <el-button v-if="row.status === 'published'" type="warning" link size="small" @click="handleTransition(row, 'archived')">归档</el-button>
            <el-button v-if="row.status === 'archived'" type="primary" link size="small" @click="handleTransition(row, 'draft')">恢复</el-button>
            <el-button type="primary" link size="small" @click="handleCopy(row)">复制</el-button>
            <el-button v-if="row.status === 'draft'" type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="tableData.length === 0" description="暂无实体类型" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listEntityTypes, createEntityType, transitionConfig, copyAsDraft, exportConfig } from '@/api/pmocker/config'

defineOptions({ name: 'PmockerConfigEntityType' })

const tableData = ref([])
const includeDraft = ref(true)

const STATUS_TABLE = {
  draft: { label: '草稿', type: 'info' },
  reviewing: { label: '评审中', type: 'warning' },
  published: { label: '已发布', type: 'success' },
  archived: { label: '已归档', type: 'info' }
}
const statusLabel = (s) => STATUS_TABLE[s]?.label || s || '-'
const statusType = (s) => STATUS_TABLE[s]?.type || 'info'

const TRANSITION_LABEL = {
  reviewing: '提交评审',
  published: '发布',
  archived: '归档',
  draft: '恢复'
}

const loadData = async () => {
  const res = await listEntityTypes({ includeDraft: includeDraft.value })
  if (res.code === 0) {
    tableData.value = res.data || []
  }
}

// 新增：两步 prompt（编码 → 名称），名称默认与编码一致
const handleAdd = async () => {
  let typeCode = ''
  try {
    const res = await ElMessageBox.prompt('请输入实体类型编码（如 issue / risk / eps_node）', '新增实体类型', {
      confirmButtonText: '下一步',
      cancelButtonText: '取消',
      inputValidator: (v) => (v && v.trim() ? true : '编码不能为空')
    })
    typeCode = res.value.trim()
  } catch (e) {
    return
  }
  let name = typeCode
  try {
    const res = await ElMessageBox.prompt('请输入实体类型名称（默认与编码一致）', '新增实体类型', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValue: typeCode,
      inputValidator: (v) => (v && v.trim() ? true : '名称不能为空')
    })
    name = res.value.trim()
  } catch (e) {
    return
  }
  const res = await createEntityType({ typeCode, name })
  if (res.code === 0) {
    ElMessage.success('创建成功')
    loadData()
  }
}

const handleTransition = async (row, to) => {
  const actionLabel = TRANSITION_LABEL[to] || to
  try {
    await ElMessageBox.confirm(`确认将「${row.name}」执行「${actionLabel}」操作吗？`, '提示', { type: 'warning' })
    const res = await transitionConfig({ table: 'pm_entity_types', id: row.ID, from: row.status, to })
    if (res.code === 0) {
      ElMessage.success(actionLabel + '成功')
      loadData()
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除草稿「${row.name}」吗？删除后不可恢复。`, '提示', { type: 'warning' })
    const res = await transitionConfig({ table: 'pm_entity_types', id: row.ID, from: row.status, to: 'delete' })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      loadData()
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const handleCopy = async (row) => {
  try {
    await ElMessageBox.confirm(`确认将「${row.name}」复制为草稿吗？（编码将追加 -copy）`, '提示', { type: 'warning' })
    const res = await copyAsDraft({ table: 'pm_entity_types', id: row.ID })
    if (res.code === 0) {
      ElMessage.success('复制成功')
      loadData()
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('复制失败')
  }
}

const handleExport = async () => {
  const res = await exportConfig()
  if (res.code === 0) {
    ElMessage.success('导出成功，YAML 已写入服务端镜像目录 images/pmbok6-hybrid')
  }
}

onMounted(loadData)
</script>
