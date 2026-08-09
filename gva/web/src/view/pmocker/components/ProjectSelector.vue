<template>
  <div :class="wrapClass">
    <span v-if="variant === 'page'" class="label">当前项目：</span>
    <el-tree-select
      v-model="selectedId"
      :data="treeData"
      :props="treeProps"
      node-key="id"
      check-strictly
      default-expand-all
      clearable
      filterable
      :placeholder="variant === 'header' ? '选择项目' : '请选择项目'"
      :no-data-text="loading ? '加载中...' : '暂无项目'"
      @change="handleChange"
      :style="{ width: variant === 'header' ? '200px' : '260px' }"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { getEPSTree } from '@/api/pmocker/eps'
import { useProjectStore } from '@/pinia'

defineOptions({ name: 'PmockerProjectSelector' })
const props = defineProps({
  // 'page' = 模块列表页样式（带背景/内边距）；'header' = 全局顶栏紧凑样式
  variant: { type: String, default: 'page' }
})
const emit = defineEmits(['change'])

const projectStore = useProjectStore()
const selectedId = ref(projectStore.projectId || null)
const treeData = ref([])
const loading = ref(false)
const treeProps = { label: 'name', children: 'children' }

const wrapClass = computed(() =>
  props.variant === 'header'
    ? 'project-selector-header'
    : 'project-selector-wrap'
)

// 判断是否组织节点（不可作为业务项目上下文）
const isOrgNode = (n) => n.type === 'group' || n.type === 'division'

// 递归找第一个项目叶子节点（跳过 group/division 组织节点）
const findFirstProject = (nodes) => {
  for (const n of nodes) {
    if (!isOrgNode(n) && (!n.children || n.children.length === 0)) return n
    if (n.children && n.children.length > 0) {
      const leaf = findFirstProject(n.children)
      if (leaf) return leaf
    }
  }
  return null
}

// 标记组织节点为禁用（el-tree-select 不可选）
const markOrgDisabled = (nodes) => {
  for (const n of nodes) {
    if (isOrgNode(n)) n.disabled = true
    if (n.children && n.children.length > 0) markOrgDisabled(n.children)
  }
}

const findNodeName = (nodes, id) => {
  for (const n of nodes) {
    if (n.id === id) return n.name
    if (n.children) {
      const name = findNodeName(n.children, id)
      if (name) return name
    }
  }
  return ''
}

const loadTree = async () => {
  loading.value = true
  try {
    const res = await getEPSTree()
    if (res.code === 0) {
      treeData.value = res.data || []
      markOrgDisabled(treeData.value)
      // store 为空时默认选第一个项目节点（跳过组织节点）
      if (!selectedId.value && treeData.value.length > 0) {
        const first = findFirstProject(treeData.value)
        if (first) {
          selectedId.value = first.id
          projectStore.setProject(first.id, first.name)
          emit('change', first.id)
        }
      }
    }
  } finally {
    loading.value = false
  }
}

const handleChange = (val) => {
  if (val) {
    const name = findNodeName(treeData.value, val)
    projectStore.setProject(val, name)
  } else {
    projectStore.clearProject()
  }
  emit('change', val)
}

onMounted(() => { loadTree() })
</script>

<style scoped>
.project-selector-wrap {
  display: inline-flex;
  align-items: center;
  padding: 8px 16px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 12px;
}
.project-selector-header {
  display: inline-flex;
  align-items: center;
}
.label {
  font-size: 14px;
  color: #606266;
  margin-right: 8px;
  white-space: nowrap;
}
</style>
