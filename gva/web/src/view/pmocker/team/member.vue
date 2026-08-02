<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增成员
        </el-button>
      </div>
      <el-table :data="tableData" row-key="id">
        <el-table-column label="编号" width="100">
          <template #default="{ row }">{{ row.attrs?.code }}</template>
        </el-table-column>
        <el-table-column label="姓名" min-width="120">
          <template #default="{ row }">{{ row.attrs?.full_name || row.title }}</template>
        </el-table-column>
        <el-table-column label="角色" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.attrs?.role }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="部门" width="120">
          <template #default="{ row }">{{ row.attrs?.department }}</template>
        </el-table-column>
        <el-table-column label="投入度" width="100">
          <template #default="{ row }">{{ row.attrs?.allocation_percent }}%</template>
        </el-table-column>
        <el-table-column label="技能等级" width="100">
          <template #default="{ row }">
            <el-tag :type="skillLevelType(row.attrs?.skill_level)">{{ row.attrs?.skill_level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" prop="status">
          <template #default="{ row }">
            <el-tag :type="memberStatusType(row.status)">{{ memberStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDialog(row)">编辑</el-button>
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
          <el-input v-model="form.title" placeholder="请输入成员名称" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="team_member" />
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
import { listMember, createMember, updateMember, deleteMember } from '@/api/pmocker/team'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerTeamMember' })

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const saving = ref(false)
const editingId = ref(null)

const form = reactive({ title: '', status: 'candidate', attrs: {} })

const memberStatusMap = {
  candidate: { label: '候选人', type: 'info' },
  onboarded: { label: '已入职', type: 'warning' },
  active: { label: '在职', type: 'success' },
  leaving: { label: '离职中', type: 'warning' },
  offboarded: { label: '已离职', type: 'info' }
}
const memberStatusLabel = (s) => memberStatusMap[s]?.label || s
const memberStatusType = (s) => memberStatusMap[s]?.type || 'info'

const skillLevelMap = { junior: 'info', mid: '', senior: 'success', expert: 'danger' }
const skillLevelType = (s) => skillLevelMap[s] || 'info'

const getTableData = async () => {
  const res = await listMember({ page: page.value, pageSize: pageSize.value })
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const openDialog = (row) => {
  if (row) {
    editingId.value = row.id
    dialogTitle.value = '编辑成员'
    Object.assign(form, { title: row.title, status: row.status, attrs: { ...row.attrs } })
  } else {
    editingId.value = null
    dialogTitle.value = '新增成员'
    Object.assign(form, { title: '', status: 'candidate', attrs: {} })
  }
  dialogVisible.value = true
}

const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, { title: '', status: 'candidate', attrs: {} })
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingId.value) {
      await updateMember({ id: editingId.value, title: form.title, status: form.status, entity_type: 'team_member', attrs: form.attrs })
    } else {
      await createMember({ title: form.title, attrs: form.attrs })
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

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该成员?', '提示', { type: 'warning' }).then(async () => {
    await deleteMember({ id: row.id })
    ElMessage.success('删除成功')
    getTableData()
  }).catch(() => {})
}

getTableData()
</script>
