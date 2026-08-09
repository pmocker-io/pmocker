<template>
  <div class="config-index">
    <VerticalTabLayout
      :active-tab="activeTab"
      :tabs="tabs"
      @tab-change="switchTab"
    >
      <EntityType v-if="activeTab === 'entityType'" />
      <FieldDef v-else-if="activeTab === 'fieldDef'" />
      <Dictionary v-else-if="activeTab === 'dictionary'" />
      <StateDef v-else-if="activeTab === 'stateDef'" />
      <Workflow v-else-if="activeTab === 'workflow'" />
      <SeedEntity v-else-if="activeTab === 'seedEntity'" />
      <OrgEntry v-else-if="activeTab === 'org'" />
    </VerticalTabLayout>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import VerticalTabLayout from '../components/VerticalTabLayout.vue'
import EntityType from './entityType.vue'
import FieldDef from './fieldDef.vue'
import Dictionary from './dictionary.vue'
import StateDef from './stateDef.vue'
import Workflow from './workflow.vue'
import SeedEntity from './seedEntity.vue'
import OrgEntry from './orgEntry.vue'

defineOptions({ name: 'PmockerConfigIndex' })

const activeTab = ref('entityType')

const tabs = [
  { name: 'entityType', label: '实体类型', count: null },
  { name: 'fieldDef', label: '字段定义', count: null },
  { name: 'dictionary', label: '字典', count: null },
  { name: 'stateDef', label: '状态流转', count: null },
  { name: 'workflow', label: '工作流', count: null },
  { name: 'seedEntity', label: '业务种子', count: null },
  { name: 'org', label: '组织权限', count: null }
]

const switchTab = (name) => {
  activeTab.value = name
}
</script>

<style scoped>
.config-index {
  padding: 16px;
}
</style>
