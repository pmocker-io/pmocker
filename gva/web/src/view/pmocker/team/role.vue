<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增角色
        </el-button>
      </div>
      <el-table :data="tableData" row-key="id">
        <el-table-column label="编号" width="100">
          <template #default="{ row }">{{ row.attrs?.code }}</template>
        </el-table-column>
        <el-table-column label="角色名称" min-width="120">
          <template #default="{ row }">{{ row.attrs?.name || row.title }}</template>
        </el-table-column>
        <el-table-column label="权限级别" width="120">
          <template #default="{ row }">{{ row.attrs?.authority_level }}</template>
        </el-table-column>
        <el-table-column label="RACI默认" width="100">
          <template #default="{ row }">{{ row.attrs?.raci_default }}</template>
        </el-table-column>
        <el-table-column label="最低经验(年)" width="110">
          <template #default="{ row }">{{ row.attrs?.min_experience_years }}</template>
        </el-table-column>
        <el-table-column label="编制上限" width="100">
          <template #default="{ row }">{{ row.attrs?.max_headcount }}</template>
        </el-table-column>
        <el-table-column label="启用" width="80">
          <template #default="{ row }">
            <el-tag :type="row.attrs?.is_active ? 'success' : 'info'">{{ row.attrs?.is_active ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">编辑</el-button>
            <el-button v-if="row.status === 'draft'" type="success" link @click="handleTransition(row, 'active', transitionRole)">启用</el-button>
            <el-button v-if="row.status === 'active'" type="danger" link @click="handleTransition(row, 'inactive', transitionRole)">停用</el-button>
            <el-button v-if="row.status === 'inactive'" type="success" link @click="handleTransition(row, 'active', transitionRole)">启用</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="800px" @closed="resetForm">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="名称" prop="title" :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]">
          <el-input v-model="form.title" placeholder="请输入角色名称" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="team_role" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listRole, createRole, updateRole, deleteRole, transitionRole } from '@/api/pmocker/team'
import { useProjectStore } from '@/pinia'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerTeamRole' })
const projectStore = useProjectStore()

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const saving = ref(false)
const editingId = ref(null)

const form = reactive({ title: '', status: 'active', attrs: {} })

const getTableData = async () => {
  const res = await listRole({ projectId: projectStore.projectId, page: page.value, pageSize: pageSize.value })
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const openDialog = (row) => {
  if (row) {
    editingId.value = row.id
    dialogTitle.value = '编辑角色'
    Object.assign(form, { title: row.title, status: row.status, attrs: { ...row.attrs } })
  } else {
    editingId.value = null
    dialogTitle.value = '新增角色'
    Object.assign(form, { title: '', status: 'active', attrs: {} })
  }
  dialogVisible.value = true
}

const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, { title: '', status: 'active', attrs: {} })
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingId.value) {
      await updateRole({ id: editingId.value, title: form.title, status: form.status, entity_type: 'team_role', attrs: form.attrs })
    } else {
      // status 由后端 service 默认值处理（role→active）
      await createRole({ title: form.title, attrs: form.attrs })
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    getTableData()
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const handleTransition = async (row, status, apiFn) => {
  try {
    await ElMessageBox.confirm('确认执行此操作？', '提示', { type: 'warning' })
    await apiFn({ id: row.id, status })
    ElMessage.success('操作成功')
    getTableData()
  } catch (e) { if (e !== 'cancel') ElMessage.error('操作失败') }
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该角色?', '提示', { type: 'warning' }).then(async () => {
    await deleteRole({ id: row.id })
    ElMessage.success('删除成功')
    getTableData()
  }).catch(() => {})
}

getTableData()
</script>
