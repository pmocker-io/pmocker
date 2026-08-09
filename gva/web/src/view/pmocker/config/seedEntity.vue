<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="实体类型">
          <el-select v-model="searchInfo.entityType" clearable filterable placeholder="请选择实体类型" style="width: 240px" @change="getTableData">
            <el-option v-for="et in entityTypes" :key="et.typeCode" :label="et.name + ' (' + et.typeCode + ')'" :value="et.typeCode" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="getTableData">查询</el-button>
          <el-button @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <el-table :data="tableData" row-key="id" size="small">
        <el-table-column label="ID" prop="id" width="80" />
        <el-table-column label="标题" prop="title" min-width="200" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="statusType(row.status)">{{ row.status || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负责人" width="120">
          <template #default="{ row }">{{ row.ownerName || row.createdByName || '-' }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">查看</el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        class="gva-pagination"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="getTableData"
        @current-change="getTableData"
      />
    </div>

    <el-dialog v-model="detailVisible" title="业务种子详情" width="520px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="标题">{{ detail.title }}</el-descriptions-item>
        <el-descriptions-item label="实体类型">{{ detail.entity_type }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status }}</el-descriptions-item>
        <el-descriptions-item label="负责人">{{ detail.ownerName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建人">{{ detail.createdByName || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="属性">
          <pre class="attrs-pre">{{ JSON.stringify(detail.attrs || {}, null, 2) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="editVisible" title="编辑业务种子" width="480px" @closed="resetEditForm">
      <el-form ref="editFormRef" :model="editForm" label-width="90px">
        <el-form-item label="标题" prop="title" :rules="[{ required: true, message: '请输入标题', trigger: 'blur' }]">
          <el-input v-model="editForm.title" />
        </el-form-item>
        <el-form-item label="状态">
          <el-input v-model="editForm.status" placeholder="状态机当前态（如 active / archived）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatDate } from '@/utils/format'
import { listEntities, updateEntity, deleteEntity } from '@/api/pmocker/schema'
import { listEntityTypes } from '@/api/pmocker/config'
import { useProjectStore } from '@/pinia'
import ProjectSelector from '../components/ProjectSelector.vue'

defineOptions({ name: 'PmockerConfigSeedEntity' })

const projectStore = useProjectStore()

const searchInfo = reactive({ entityType: '' })
const entityTypes = ref([])
const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const detailVisible = ref(false)
const detail = ref({})

const editVisible = ref(false)
const editFormRef = ref(null)
const editForm = reactive({ id: null, title: '', status: '' })

const statusType = (s) => {
  const map = {
    active: 'success',
    completed: 'success',
    archived: 'info',
    draft: 'info',
    paused: 'warning',
    initiating: 'warning'
  }
  return map[s] || 'info'
}

const onProjectChange = () => { getTableData() }

const loadEntityTypes = async () => {
  const res = await listEntityTypes({ includeDraft: false })
  if (res.code === 0) {
    entityTypes.value = res.data || []
  }
}

const getTableData = async () => {
  const params = {
    entityType: searchInfo.entityType,
    projectId: projectStore.projectId,
    page: page.value,
    pageSize: pageSize.value
  }
  if (!params.entityType) {
    tableData.value = []
    total.value = 0
    return
  }
  const res = await listEntities(params)
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const onReset = () => {
  searchInfo.entityType = ''
  page.value = 1
  getTableData()
}

const openDetail = (row) => {
  detail.value = row
  detailVisible.value = true
}

const openEdit = (row) => {
  editForm.id = row.id
  editForm.title = row.title
  editForm.status = row.status || ''
  editVisible.value = true
}

const resetEditForm = () => {
  editFormRef.value?.resetFields()
  Object.assign(editForm, { id: null, title: '', status: '' })
}

const handleSave = async () => {
  if (!editFormRef.value) return
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await updateEntity({ id: editForm.id, title: editForm.title, status: editForm.status })
    if (res.code === 0) {
      ElMessage.success('保存成功')
      editVisible.value = false
      getTableData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确认删除业务种子「${row.title}」吗？`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteEntity(row.id)
      if (res.code === 0) {
        ElMessage.success('删除成功')
        getTableData()
      }
    })
    .catch(() => {})
}

onMounted(loadEntityTypes)
</script>

<style scoped>
.attrs-pre {
  margin: 0;
  max-height: 240px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.6;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 8px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
