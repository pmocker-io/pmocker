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
              <svg-icon :icon="data.type === 'project' ? 'lucide:folder' : 'lucide:building'" />
              <span>{{ data.name }}</span>
              <el-tag size="small" :type="data.type === 'project' ? 'primary' : 'info'">
                {{ typeLabel(data.type) }}
              </el-tag>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type">
            <el-option label="组织" value="org" />
            <el-option label="项目" value="project" />
            <el-option label="子项目" value="subproject" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
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

defineOptions({ name: 'PmockerEpsTree' })

const treeData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, parentId: null, name: '', type: 'org', description: '' })
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const typeLabel = (type) => ({ org: '组织', project: '项目', subproject: '子项目' }[type] || type)

const loadTree = async () => {
  const res = await getEPSNodes({})
  if (res.code === 0) {
    treeData.value = res.data || []
  }
}

const openAddDialog = (parent) => {
  dialogType.value = 'add'
  dialogTitle.value = parent ? '新增子节点' : '新增根节点'
  Object.assign(form, { ID: null, parentId: parent?.ID || null, name: '', type: 'org', description: '' })
  dialogVisible.value = true
}

const openEditDialog = (data) => {
  dialogType.value = 'edit'
  dialogTitle.value = '编辑节点'
  Object.assign(form, data)
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const api = dialogType.value === 'add' ? createEPSNode : updateEPSNode
    const res = await api(form)
    if (res.code === 0) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      loadTree()
    }
  })
}

const handleDelete = (data) => {
  ElMessageBox.confirm(`确认删除「${data.name}」及其子节点吗？`, '提示', { type: 'warning' })
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
