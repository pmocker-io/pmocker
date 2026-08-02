<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增培训记录
        </el-button>
      </div>
      <el-table :data="tableData" row-key="id">
        <el-table-column label="课程名称" min-width="150">
          <template #default="{ row }">{{ row.attrs?.course_name || row.title }}</template>
        </el-table-column>
        <el-table-column label="培训类型" width="100">
          <template #default="{ row }">{{ row.attrs?.training_type }}</template>
        </el-table-column>
        <el-table-column label="开始日期" width="120">
          <template #default="{ row }">{{ row.attrs?.start_date }}</template>
        </el-table-column>
        <el-table-column label="结束日期" width="120">
          <template #default="{ row }">{{ row.attrs?.end_date }}</template>
        </el-table-column>
        <el-table-column label="时长(h)" width="80">
          <template #default="{ row }">{{ row.attrs?.duration_hours }}</template>
        </el-table-column>
        <el-table-column label="获证" width="80">
          <template #default="{ row }">
            <el-tag :type="row.attrs?.certification_obtained ? 'success' : 'info'">{{ row.attrs?.certification_obtained ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="效果评分" width="100">
          <template #default="{ row }">{{ row.attrs?.effectiveness_score }}</template>
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
          <el-input v-model="form.title" placeholder="请输入培训记录名称" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="training_record" />
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
import { listTraining, createTraining, updateTraining, deleteTraining } from '@/api/pmocker/team'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerTeamTraining' })

const tableData = ref([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const saving = ref(false)
const editingId = ref(null)

const form = reactive({ title: '', status: 'planned', attrs: {} })

const getTableData = async () => {
  const res = await listTraining({ page: page.value, pageSize: pageSize.value })
  if (res.code === 0) {
    tableData.value = res.data.list || []
    total.value = res.data.total || 0
  }
}

const openDialog = (row) => {
  if (row) {
    editingId.value = row.id
    dialogTitle.value = '编辑培训记录'
    Object.assign(form, { title: row.title, status: row.status, attrs: { ...row.attrs } })
  } else {
    editingId.value = null
    dialogTitle.value = '新增培训记录'
    Object.assign(form, { title: '', status: 'planned', attrs: {} })
  }
  dialogVisible.value = true
}

const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, { title: '', status: 'planned', attrs: {} })
}

const handleSave = async () => {
  saving.value = true
  try {
    if (editingId.value) {
      await updateTraining({ id: editingId.value, title: form.title, status: form.status, entity_type: 'training_record', attrs: form.attrs })
    } else {
      // status 由后端 service 默认值处理（training→planned）
      await createTraining({ title: form.title, attrs: form.attrs })
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
  ElMessageBox.confirm('确认删除该培训记录?', '提示', { type: 'warning' }).then(async () => {
    await deleteTraining({ id: row.id })
    ElMessage.success('删除成功')
    getTableData()
  }).catch(() => {})
}

getTableData()
</script>
