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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="title">
          <el-input v-model="form.title" placeholder="请输入范围项名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="工作量" prop="effort">
          <el-input-number v-model="form.effort" :min="0" :step="0.5" />
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
import { getScopeWBS, createScopeItem, buildScopeWBS } from '@/api/pmocker/scope'

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
  description: '',
  effort: 0
})

const rules = {
  title: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

const nodeIcon = (data) => {
  if (data.children && data.children.length > 0) return 'lucide:folder'
  return 'lucide:file'
}

const loadWBS = async () => {
  const res = await getScopeWBS({})
  if (res.code === 0) {
    treeData.value = res.data || []
  }
}

const openAddDialog = (parent) => {
  dialogType.value = 'add'
  parentNode.value = parent
  dialogTitle.value = parent ? '新增子项' : '新增根节点'
  Object.assign(form, { ID: null, parentId: parent?.ID || null, title: '', description: '', effort: 0 })
  dialogVisible.value = true
}

const openEditDialog = (data) => {
  dialogType.value = 'edit'
  dialogTitle.value = '编辑范围项'
  Object.assign(form, data)
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await createScopeItem(form)
    if (res.code === 0) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      loadWBS()
    }
  })
}

const handleDelete = (data) => {
  ElMessageBox.confirm(`确认删除「${data.title}」及其子项吗？`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await buildScopeWBS({ action: 'delete', ID: data.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadWBS()
      }
    })
    .catch(() => {})
}

loadWBS()
</script>
