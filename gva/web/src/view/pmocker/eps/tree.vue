<template>
  <div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" @click="openAddDialog(null)">
          <svg-icon icon="lucide:plus" /> 新增根节点
        </el-button>
        <el-button @click="loadTree">刷新</el-button>
      </div>
      <el-tree
        :data="treeData"
        node-key="ID"
        default-expand-all
        draggable
        :expand-on-click-node="false"
        @node-drop="handleDrop"
      >
        <template #default="{ node, data }">
          <div class="flex items-center justify-between w-full">
            <span class="flex items-center gap-2">
              <svg-icon :icon="nodeIcon(data)" />
              <span>{{ data.title || data.name }}</span>
              <el-tag size="small" :type="typeTagType(data)">
                {{ typeLabel(getAttr(data, 'type') || data.type) }}
              </el-tag>
              <el-tag v-if="getAttr(data, 'governance_type')" size="small" type="warning">{{ getAttr(data, 'governance_type') }}</el-tag>
              <el-tag v-if="getAttr(data, 'lifecycle_phase')" size="small" type="success">{{ getAttr(data, 'lifecycle_phase') }}</el-tag>
              <el-tag v-if="getAttr(data, 'health_status')" size="small" :type="healthType(getAttr(data, 'health_status'))">{{ getAttr(data, 'health_status') }}</el-tag>
            </span>
            <span class="flex gap-1">
              <el-button type="primary" link @click.stop="openAddDialog(data)">添加子节点</el-button>
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
        <DynamicForm v-model="form" entity-type="eps_node" />
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
import { getEPSNodes, createEPSNode, updateEPSNode, deleteEPSNode, moveEPSNode } from '@/api/pmocker/eps'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerEpsTree' })

const treeData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({
  ID: null,
  parentId: null,
  title: '',
  status: 'draft',
  attrs: {}
})

const getAttr = (row, key) => (row.attrs && row.attrs[key] !== undefined ? row.attrs[key] : (row[key] || ''))

const typeLabel = (type) => ({ org: '组织', project: '项目', subproject: '子项目' }[type] || type || '')

const typeTagType = (data) => {
  const type = getAttr(data, 'type') || data.type
  return type === 'project' ? 'primary' : 'info'
}

const healthType = (status) => {
  const map = { green: 'success', yellow: 'warning', red: 'danger' }
  return map[status] || 'info'
}

const nodeIcon = (data) => {
  const type = getAttr(data, 'type') || data.type
  return type === 'project' ? 'lucide:folder' : 'lucide:building'
}

const loadTree = async () => {
  const res = await getEPSNodes({})
  if (res.code === 0) {
    treeData.value = res.data || []
  }
}

const resetForm = (parentId) => {
  Object.assign(form, { ID: null, parentId: parentId || null, title: '', status: 'draft', attrs: {} })
}

const openAddDialog = (parent) => {
  dialogType.value = 'add'
  dialogTitle.value = parent ? '新增子节点' : '新增根节点'
  resetForm(parent?.ID)
  dialogVisible.value = true
}

const openEditDialog = (data) => {
  dialogType.value = 'edit'
  dialogTitle.value = '编辑节点'
  resetForm()
  form.ID = data.ID
  form.parentId = data.parentId || null
  form.title = data.title || data.name || ''
  form.status = data.status || 'draft'
  form.attrs = data.attrs ? { ...data.attrs } : {}
  // 兼容旧数据：把顶层字段合并到 attrs
  if (data.name && form.attrs.name === undefined) form.attrs.name = data.name
  if (data.type && form.attrs.type === undefined) form.attrs.type = data.type
  if (data.description && form.attrs.description === undefined) form.attrs.description = data.description
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const api = dialogType.value === 'add' ? createEPSNode : updateEPSNode
    const res = await api(form)
    if (res.code === 0) {
      ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
      dialogVisible.value = false
      loadTree()
    }
  })
}

const handleDelete = (data) => {
  const label = data.title || data.name
  ElMessageBox.confirm(`确认删除「${label}」及其子节点吗？`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteEPSNode({ ID: data.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadTree()
      }
    })
    .catch(() => {})
}

const handleDrop = async (draggingNode, dropNode, dropType) => {
  const res = await moveEPSNode({
    nodeID: draggingNode.data.ID,
    targetID: dropNode.data.ID,
    position: dropType
  })
  if (res.code !== 0) {
    ElMessage.error('移动失败')
    loadTree()
  }
}

loadTree()
</script>
