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
        node-key="id"
        default-expand-all
        draggable
        :expand-on-click-node="false"
        @node-drop="handleDrop"
      >
        <template #default="{ node, data }">
          <div class="flex items-center justify-between w-full">
            <span class="flex items-center gap-2">
              <svg-icon :icon="nodeIcon(data)" />
              <span>{{ data.name }}</span>
              <el-tag size="small" :type="typeTagType(data)">{{ typeLabel(data.type) }}</el-tag>
              <el-tag v-if="data.code" size="small" type="info">{{ data.code }}</el-tag>
            </span>
            <span class="flex gap-1">
              <el-button v-if="isProject(data)" type="primary" link @click.stop="enterProject(data)">进入项目</el-button>
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
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getEPSTree, createEPSNode, updateEPSNode, deleteEPSNode, moveEPSNode, findEPSNode } from '@/api/pmocker/eps'
import { useProjectStore } from '@/pinia'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerEpsTree' })

const router = useRouter()
const projectStore = useProjectStore()

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

// 判断是否项目节点（非组织节点 group/division 即为项目）
const isProject = (data) => data.type !== 'group' && data.type !== 'division'

const typeLabel = (type) => ({ org: '组织', group: '集团', division: '部门', project: '项目', subproject: '子项目' }[type] || (type ? type : '项目'))

const typeTagType = (data) => isProject(data) ? 'primary' : 'info'

const nodeIcon = (data) => isProject(data) ? 'lucide:folder' : 'lucide:building'

// 进入项目：设置全局项目上下文并跳转项目仪表盘
const enterProject = (data) => {
  projectStore.setProject(data.id, data.name)
  router.push({ name: 'pmockerDashboard' })
}

const loadTree = async () => {
  const res = await getEPSTree()
  if (res.code === 0) {
    treeData.value = res.data || []
  }
}

const resetForm = (parentId) => {
  Object.assign(form, { ID: null, parentId: parentId || null, title: '', status: 'draft', attrs: {} })
}

const openAddDialog = (parent) => {
  dialogType.value = 'add'
  dialogTitle.value = parent ? `在「${parent.name}」下新增子节点` : '新增根节点'
  resetForm(parent?.id)
  dialogVisible.value = true
}

const openEditDialog = async (data) => {
  dialogType.value = 'edit'
  dialogTitle.value = '编辑节点'
  resetForm()
  // 编辑时调用 findEPSNode 获取完整节点详情（含 attrs）
  const res = await findEPSNode({ ID: data.id })
  if (res.code === 0 && res.data) {
    const node = res.data
    form.ID = node.id || node.ID
    form.parentId = node.parentId || null
    form.title = node.title || node.name || ''
    form.status = node.status || 'draft'
    form.attrs = node.attrs ? { ...node.attrs } : {}
  } else {
    // 降级：直接用 tree 数据填充
    form.ID = data.id
    form.title = data.name || ''
    form.attrs = { code: data.code, type: data.type }
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    let res
    if (dialogType.value === 'add') {
      // 后端 CreateNode 期望 {name, parentId, attrs, status}
      res = await createEPSNode({
        name: form.title,
        parentId: form.parentId || 0,
        attrs: { ...form.attrs },
        status: form.status || 'active'
      })
    } else {
      // 后端 UpdateNode 期望 Entity{ID, entityType, title, attrs, status}
      res = await updateEPSNode({
        ID: form.ID,
        entityType: 'eps_node',
        title: form.title,
        attrs: { ...form.attrs },
        status: form.status || 'active'
      })
    }
    if (res.code === 0) {
      ElMessage.success(dialogType.value === 'add' ? '添加成功' : '更新成功')
      dialogVisible.value = false
      loadTree()
    }
  })
}

const handleDelete = (data) => {
  ElMessageBox.confirm(`确认删除「${data.name}」及其子节点吗？`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteEPSNode({ ID: data.id })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadTree()
      }
    })
    .catch(() => {})
}

const handleDrop = async (draggingNode, dropNode, dropType) => {
  const res = await moveEPSNode({
    nodeID: draggingNode.data.id,
    targetID: dropNode.data.id,
    position: dropType
  })
  if (res.code !== 0) {
    ElMessage.error('移动失败')
    loadTree()
  }
}

loadTree()
</script>
