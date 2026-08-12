<template>
  <div>
    <ProjectSelector @change="onProjectChange" />
    <el-tabs v-model="activeTab" class="raci-tabs">
      <el-tab-pane label="WBS 树" name="tree">
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
      </el-tab-pane>

      <el-tab-pane :label="`RACI 矩阵${roles.length ? '(' + roles.length + ')' : ''}`" name="raci">
        <div class="gva-table-box">
          <div class="gva-btn-list">
            <el-button @click="loadRaci">刷新</el-button>
            <el-button :type="editMode ? 'warning' : 'primary'" @click="editMode = !editMode">
              {{ editMode ? '退出编辑' : '编辑矩阵' }}
            </el-button>
            <span class="raci-legend">
              <el-tag type="danger" size="small" effect="dark">R</el-tag>执行
              <el-tag type="warning" size="small" effect="dark">A</el-tag>负责
              <el-tag type="primary" size="small" effect="dark">C</el-tag>咨询
              <el-tag type="info" size="small" effect="dark">I</el-tag>知会
            </span>
          </div>
          <el-table v-if="roles.length" :data="flatItems" row-key="ID" border size="small" :max-height="640">
            <el-table-column label="WBS 节点" min-width="260" fixed>
              <template #default="{ row }">
                <div class="raci-node" :style="{ paddingLeft: (row._level || 0) * 18 + 'px' }">
                  <svg-icon :icon="row._hasChildren ? 'lucide:folder' : 'lucide:file'" />
                  <span class="raci-node-title">{{ row.title }}</span>
                  <el-tag v-if="getAttr(row, 'wbs_code')" size="small" type="info">{{ getAttr(row, 'wbs_code') }}</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              v-for="role in roles"
              :key="roleKey(role)"
              min-width="110"
              align="center"
            >
              <template #header>
                <div class="raci-role-header">
                  <div>{{ role.attrs?.name || role.title }}</div>
                  <div v-if="role.attrs?.code" class="raci-role-code">{{ role.attrs.code }}</div>
                </div>
              </template>
              <template #default="{ row }">
                <div v-if="!editMode" class="raci-cell">
                  <template v-if="cellLetters(row, role).length">
                    <el-tag
                      v-for="l in cellLetters(row, role)"
                      :key="l"
                      :type="raciTagType(l)"
                      size="small"
                      effect="dark"
                    >{{ l }}</el-tag>
                  </template>
                  <span v-else class="raci-empty">—</span>
                </div>
                <div v-else class="raci-edit-btns">
                  <el-button
                    v-for="l in raciLetters"
                    :key="l"
                    :type="getRaciState(row, role)[l] ? raciTagType(l) : 'info'"
                    :plain="!getRaciState(row, role)[l]"
                    size="small"
                    circle
                    @click="toggleRaci(row, role, l)"
                  >{{ l }}</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="暂无团队角色，请先在「团队管理-角色定义」中创建角色" />
        </div>
      </el-tab-pane>
    </el-tabs>

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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getScopeWBS, createScopeItem, updateScopeItem, deleteScopeItem } from '@/api/pmocker/scope'
import { listRole } from '@/api/pmocker/team'
import DynamicForm from '../components/DynamicForm.vue'
import ProjectSelector from '../components/ProjectSelector.vue'
import { useProjectStore } from '@/pinia'

defineOptions({ name: 'PmockerScopeWBS' })

const projectStore = useProjectStore()
const onProjectChange = () => { loadRaci() }

const activeTab = ref('tree')
const treeData = ref([])
const roles = ref([])
const editMode = ref(false)
const raciLetters = ['R', 'A', 'C', 'I']

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
  const res = await getScopeWBS({ projectId: projectStore.projectId })
  if (res.code === 0) {
    treeData.value = res.data || []
  }
}

const loadRoles = async () => {
  const res = await listRole({ page: 1, pageSize: 100 })
  if (res.code === 0) {
    roles.value = res.data.list || []
  }
}

const loadRaci = () => {
  loadWBS()
  loadRoles()
}

// 扁平化 WBS 树作为矩阵行（保留节点引用以支持响应式编辑）
const flatItems = computed(() => {
  const out = []
  const walk = (nodes, level) => {
    if (!nodes) return
    for (const n of nodes) {
      n._level = level
      n._hasChildren = !!(n.children && n.children.length)
      out.push(n)
      walk(n.children, level + 1)
    }
  }
  walk(treeData.value, 0)
  return out
})

const roleId = (r) => (r && (r.id || r.ID)) || ''
const roleKey = (r) => roleId(r) || (r && r.title) || Math.random().toString(36)

const raciTagType = (l) => ({ R: 'danger', A: 'warning', C: 'primary', I: 'info' }[l] || 'info')

// 计算 scope_item 在某角色下的 RACI 命中状态
// R/A 为 ref 单值，C/I 为 json 数组
const getRaciState = (item, role) => {
  const rid = String(roleId(role))
  const a = (item && item.attrs) || {}
  const consulted = Array.isArray(a.raci_consulted) ? a.raci_consulted : (a.raci_consulted ? [a.raci_consulted] : [])
  const informed = Array.isArray(a.raci_informed) ? a.raci_informed : (a.raci_informed ? [a.raci_informed] : [])
  return {
    R: !!rid && String(a.raci_responsible || '') === rid,
    A: !!rid && String(a.raci_accountable || '') === rid,
    C: !!rid && consulted.some((x) => String(x) === rid),
    I: !!rid && informed.some((x) => String(x) === rid)
  }
}

const cellLetters = (item, role) => {
  const s = getRaciState(item, role)
  const arr = []
  if (s.R) arr.push('R')
  if (s.A) arr.push('A')
  if (s.C) arr.push('C')
  if (s.I) arr.push('I')
  return arr
}

// 切换某 (scope_item, role) 的 RACI 标记：ref 字段覆盖写，json 字段增删元素
const toggleRaci = async (item, role, letter) => {
  const rid = roleId(role)
  if (!rid) return
  const a = item.attrs ? { ...item.attrs } : {}
  const consulted = Array.isArray(a.raci_consulted) ? [...a.raci_consulted] : []
  const informed = Array.isArray(a.raci_informed) ? [...a.raci_informed] : []
  const s = getRaciState(item, role)
  let newAttrs
  if (letter === 'R') {
    newAttrs = { ...a, raci_responsible: s.R ? '' : rid }
  } else if (letter === 'A') {
    newAttrs = { ...a, raci_accountable: s.A ? '' : rid }
  } else if (letter === 'C') {
    newAttrs = { ...a, raci_consulted: s.C ? consulted.filter((x) => String(x) !== String(rid)) : [...consulted, rid] }
  } else if (letter === 'I') {
    newAttrs = { ...a, raci_informed: s.I ? informed.filter((x) => String(x) !== String(rid)) : [...informed, rid] }
  } else {
    return
  }
  // 乐观更新本地节点 attrs
  item.attrs = newAttrs
  try {
    const res = await updateScopeItem({
      ID: item.ID,
      parentId: item.parentId,
      title: item.title,
      status: item.status,
      attrs: newAttrs
    })
    if (res.code !== 0) {
      ElMessage.error('RACI 保存失败，已刷新')
      await loadWBS()
    }
  } catch (e) {
    ElMessage.error('RACI 保存失败，已刷新')
    await loadWBS()
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
    const res = await api({ ...form, projectId: projectStore.projectId })
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

onMounted(() => {
  loadWBS()
  loadRoles()
})
</script>

<style scoped>
.raci-tabs :deep(.el-tabs__content) {
  padding-top: 8px;
}
.raci-legend {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-regular);
}
.raci-legend .el-tag {
  margin: 0 4px 0 8px;
}
.raci-node {
  display: flex;
  align-items: center;
  gap: 6px;
}
.raci-node-title {
  font-weight: 500;
}
.raci-cell {
  min-height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
}
.raci-empty {
  color: var(--el-text-color-placeholder);
}
.raci-edit-btns {
  display: flex;
  gap: 4px;
  justify-content: center;
}
.raci-edit-btns :deep(.el-button) {
  width: 26px;
  height: 26px;
  min-height: 26px;
  padding: 0;
  font-size: 12px;
}
.raci-role-header {
  text-align: center;
  line-height: 1.3;
}
.raci-role-code {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}
</style>
