<template>
  <div>
    <div class="flex gap-4">
      <div class="w-1/3">
        <el-card>
          <template #header>
            <span class="font-medium">选择 EPS 节点</span>
          </template>
          <el-tree
            :data="treeData"
            node-key="ID"
            default-expand-all
            highlight-current
            @node-click="handleNodeClick"
          >
            <template #default="{ data }">
              <span class="flex items-center gap-1">
                <svg-icon icon="lucide:folder" />
                {{ data.name }}
              </span>
            </template>
          </el-tree>
        </el-card>
      </div>
      <div class="flex-1">
        <el-card>
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-medium">
                {{ selectedNode ? `${selectedNode.name} - 成员管理` : '请选择左侧节点' }}
              </span>
              <el-button v-if="selectedNode" type="primary" @click="openAddDialog">
                <svg-icon icon="lucide:user-plus" /> 添加成员
              </el-button>
            </div>
          </template>
          <el-table v-if="selectedNode" :data="memberList">
            <el-table-column label="用户名" prop="username" />
            <el-table-column label="角色" prop="role" width="150">
              <template #default="{ row }">
                <el-tag>{{ row.role }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="加入时间" prop="CreatedAt" width="180">
              <template #default="{ row }">{{ formatDate(row.CreatedAt) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button type="danger" link @click="handleRemove(row)">移除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="请选择左侧 EPS 节点" />
        </el-card>
      </div>
    </div>

    <el-dialog v-model="dialogVisible" title="添加成员" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="用户ID" prop="userId">
          <el-input-number v-model="form.userId" :min="1" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role">
            <el-option label="项目经理" value="manager" />
            <el-option label="成员" value="member" />
            <el-option label="观察者" value="observer" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAdd">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getEPSNodes, getEPSMembers, addEPSMember, removeEPSMember } from '@/api/pmocker/eps'

defineOptions({ name: 'PmockerEpsMembers' })

const treeData = ref([])
const selectedNode = ref(null)
const memberList = ref([])
const dialogVisible = ref(false)
const formRef = ref(null)

const form = reactive({ userId: 1, role: 'member' })
const rules = {
  userId: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : ''

const loadTree = async () => {
  const res = await getEPSNodes({})
  if (res.code === 0) {
    treeData.value = res.data || []
  }
}

const handleNodeClick = async (data) => {
  selectedNode.value = data
  const res = await getEPSMembers({ nodeId: data.ID })
  if (res.code === 0) {
    memberList.value = res.data.list || []
  }
}

const openAddDialog = () => {
  Object.assign(form, { userId: 1, role: 'member' })
  dialogVisible.value = true
}

const handleAdd = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await addEPSMember({ nodeId: selectedNode.value.ID, ...form })
    if (res.code === 0) {
      ElMessage.success('添加成功')
      dialogVisible.value = false
      handleNodeClick(selectedNode.value)
    }
  })
}

const handleRemove = (row) => {
  ElMessageBox.confirm(`确认移除「${row.username}」吗？`, '提示', { type: 'warning' })
    .then(async () => {
      const res = await removeEPSMember({ nodeId: selectedNode.value.ID, userId: row.userId })
      if (res.code === 0) {
        ElMessage.success('已移除')
        handleNodeClick(selectedNode.value)
      }
    })
    .catch(() => {})
}

loadTree()
</script>
