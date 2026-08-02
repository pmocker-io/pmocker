<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openAddDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增根节点
        </el-button>
        <el-button @click="loadWBS">刷新</el-button>
      </div>
      <el-tree
        :data="treeData"
        node-key="ID"
        default-expand-all
        :expand-on-click-node="false"
      >
        <template #default="{ node, data }">
          <div class="flex items-center justify-between w-full">
            <span class="flex items-center gap-2">
              <svg-icon :icon="nodeIcon(data)" />
              {{ data.title }}
              <el-tag v-if="getAttr(data, 'wbs_code')" size="small" type="info">{{ getAttr(data, 'wbs_code') }}</el-tag>
              <el-tag v-if="getAttr(data, 'is_work_package')" size="small" type="success">工作包</el-tag>
              <el-tag v-if="getAttr(data, 'acceptance_status')" size="small" :type="acceptanceType(getAttr(data, 'acceptance_status'))">{{ getAttr(data, 'acceptance_status') }}</el-tag>
              <el-tag v-if="data.isBaseline" size="small" type="success">基线</el-tag>
            </span>
            <span class="flex gap-1">
              <el-button type="primary" link @click.stop="openAddDialog(data)">添加子项</el-button>
              <el-button type="warning" link @click.stop="openEditDialog(data)">编辑</el-button>
              <el-button type="danger" link @click.stop="handleDelete(data)">删除</el-button>
            </span>
          </div>
        </template>
      </el-tree>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="800px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="名称" prop="title" :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]">
          <el-input v-model="form.title" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="scope_item" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getScopeWBS, createScopeItem, updateScopeItem, deleteScopeItem } from '@/api/pmocker/scope'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerScopeWBS' })

const treeData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')
const parentNode = ref(null)

const form = reactive({
  ID: null,
  parentId: null,
  title: '',
  status: 'draft',
  attrs: {}
})

const getAttr = (row, key) => (row.attrs && row.attrs[key] !== undefined ? row.attrs[key] : (row[key] || ''))

const nodeIcon = (data) => {
  if (data.children && data.children.length > 0) return 'lucide:folder'
  return 'lucide:file'
}

const acceptanceType = (status) => {
  const map = { pending: 'info', accepted: 'success', rejected: 'danger' }
  return map[status] || 'info'
}

const loadWBS = async () => {
  const res = await getScopeWBS({})
  if (res.code === 0) {
    treeData.value = res.data || []
  }
}

const resetForm = (parentId) => {
  Object.assign(form, { ID: null, parentId: parentId || null, title: '', status: 'draft', attrs: {} })
}

const openAddDialog = (parent) => {
  dialogType.value = 'add'
  parentNode.value = parent
  dialogTitle.value = parent ? '新增子项' : '新增根节点'
  resetForm(parent?.ID)
  dialogVisible.value = true
}

const openEditDialog = (data) => {
  dialogType.value = 'edit'
  dialogTitle.value = '编辑范围项'
  resetForm()
  form.ID = data.ID
  form.parentId = data.parentId || null
  form.title = data.title
  form.status = data.status || 'draft'
  form.attrs = data.attrs ? { ...data.attrs } : {}
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const api = dialogType.value === 'add' ? createScopeItem : updateScopeItem
    const res = await api(form)
    if (res.code === 0) {
      ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
      dialogVisible.value = false
      loadWBS()
    }
  })
}

const handleDelete = (data) => {
  ElMessageBox.confirm(`确认删除「${data.title}」及其子项吗？`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteScopeItem({ ID: data.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadWBS()
      }
    })
    .catch(() => {})
}

loadWBS()
</script>
