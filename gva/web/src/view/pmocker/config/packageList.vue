<template>
  <div class="config-list">
    <el-page-header content="配置包管理">
      <template #extra>
        <el-button type="primary" @click="openCreate">
          <svg-icon icon="lucide:plus" /> 新建配置包
        </el-button>
        <el-button @click="loadData">刷新</el-button>
      </template>
    </el-page-header>

    <div class="flex items-center gap-4 mt-4 mb-4">
      <el-radio-group v-model="includeDraft" size="small" @change="loadData">
        <el-radio-button :value="true">全部</el-radio-button>
        <el-radio-button :value="false">仅已发布</el-radio-button>
      </el-radio-group>
      <el-button size="small" @click="handleExport">导出YAML</el-button>
    </div>

    <el-table :data="list" border stripe v-loading="loading">
      <el-table-column prop="code" label="编码" width="130" />
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column prop="entityType" label="实体类型" width="120" />
      <el-table-column prop="module" label="模块" width="100" />
      <el-table-column label="版本" width="80">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.version }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-width="260">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEditor(row)">编辑</el-button>
          <el-button link size="small" @click="handleCopy(row)">复制</el-button>
          <el-button v-if="row.status === 'draft' || row.status === 'reviewing' || row.status === 'published'" link type="success" size="small" @click="handlePublish(row)">发布</el-button>
          <el-button v-if="row.status === 'published'" link type="warning" size="small" @click="handleArchive(row)">归档</el-button>
          <el-button v-if="row.status === 'archived'" link type="primary" size="small" @click="handleRestore(row)">恢复</el-button>
          <el-button v-if="row.status === 'draft'" link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listPackages, createPackage, copyPackage, publishPackage, transitionPackage, deletePackage, exportConfig } from '@/api/pmocker/config'

defineOptions({ name: 'PmockerConfigPackageList' })

const router = useRouter()
const list = ref([])
const loading = ref(false)
const includeDraft = ref(true)

const loadData = async () => {
  loading.value = true
  const res = await listPackages({ includeDraft: includeDraft.value })
  loading.value = false
  if (res.code === 0) list.value = res.data || []
}

const openCreate = async () => {
  const { value } = await ElMessageBox.prompt('配置包编码（如 requirement/risk/schedule/eps）', '新建配置包', {
    inputPlaceholder: '请输入配置包编码',
    inputValidator: (v) => (v && v.trim() ? true : '编码不能为空')
  })
  if (!value) return
  const res = await createPackage({ code: value, name: value, entityType: value, module: value, status: 'draft' })
  if (res.code === 0) {
    ElMessage.success('已创建，请点击编辑配置详情')
    loadData()
  }
}

const openEditor = (row) => {
  router.push({ name: 'pmockerConfigPackageEditor', query: { id: row.ID } })
}

const handlePublish = async (row) => {
  await ElMessageBox.confirm(`确认发布「${row.name}」？发布将自动同步字段/状态/种子到数据库。`, '发布确认', { type: 'warning' })
  const res = await publishPackage(row.ID)
  if (res.code === 0) { ElMessage.success('发布成功'); loadData() }
}

const handleCopy = async (row) => {
  const res = await copyPackage(row.ID)
  if (res.code === 0) { ElMessage.success('已复制为草稿'); loadData() }
}

const handleArchive = async (row) => {
  await ElMessageBox.confirm(`确认归档「${row.name}」？`, '提示', { type: 'warning' })
  const res = await transitionPackage(row.ID, 'archived')
  if (res.code === 0) { ElMessage.success('已归档'); loadData() }
}

const handleRestore = async (row) => {
  const res = await transitionPackage(row.ID, 'restore')
  if (res.code === 0) { ElMessage.success('已恢复为草稿'); loadData() }
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除「${row.name}」？仅草稿可删除。`, '提示', { type: 'warning' })
  const res = await deletePackage(row.ID)
  if (res.code === 0) { ElMessage.success('已删除'); loadData() }
}

const handleExport = async () => {
  const res = await exportConfig()
  if (res.code === 0) ElMessage.success(res.msg || '导出成功')
}

const statusLabel = (s) => ({ draft: '草稿', reviewing: '评审中', published: '已发布', archived: '已归档' }[s] || s)
const statusTag = (s) => ({ draft: 'info', reviewing: 'warning', published: 'success', archived: '' }[s] || 'info')

onMounted(loadData)
</script>
