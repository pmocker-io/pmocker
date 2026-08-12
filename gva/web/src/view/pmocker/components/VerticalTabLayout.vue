<template>
  <div class="vtab-layout">
    <div class="vtab-tabs" :style="{ width: tabWidth + 'px' }">
      <div
        v-for="t in tabs"
        :key="t.name"
        class="vtab-item"
        :class="{ active: activeTab === t.name }"
        @click="onTabClick(t.name)"
      >
        <span class="vtab-label">{{ t.label }}</span>
        <span v-if="t.count !== null && t.count !== undefined" class="vtab-badge">{{ t.count }}</span>
      </div>
    </div>
    <div class="vtab-content">
      <div v-if="$slots.toolbar" class="vtab-toolbar">
        <slot name="toolbar" />
      </div>
      <div class="vtab-body">
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ name: 'VerticalTabLayout' })

const props = defineProps({
  tabs: {
    type: Array,
    required: true,
    default: () => []
  },
  activeTab: {
    type: String,
    default: ''
  },
  tabWidth: {
    type: [Number, String],
    default: 140
  }
})

const emit = defineEmits(['update:activeTab', 'tab-change'])

const onTabClick = (name) => {
  if (props.activeTab === name) return
  emit('update:activeTab', name)
  emit('tab-change', name)
}
</script>

<style scoped>
.vtab-layout {
  display: flex;
  min-height: 500px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  overflow: hidden;
}
.vtab-tabs {
  flex-shrink: 0;
  background: var(--el-fill-color-light);
  border-right: 1px solid var(--el-border-color-lighter);
}
.vtab-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 44px;
  line-height: 44px;
  cursor: pointer;
  border-bottom: 1px solid var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
  transition: all 0.2s;
  user-select: none;
}
.vtab-item:hover {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
.vtab-item.active {
  background: var(--el-color-primary);
  color: var(--el-color-white);
  font-weight: 500;
}
.vtab-item.active:hover {
  background: var(--el-color-primary);
  color: var(--el-color-white);
}
.vtab-label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.vtab-badge {
  display: inline-block;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  line-height: 18px;
  text-align: center;
  font-size: 11px;
  border-radius: 9px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
  margin-left: 8px;
  flex-shrink: 0;
}
.vtab-item.active .vtab-badge {
  background: rgba(255, 255, 255, 0.3);
  color: var(--el-color-white);
}
.vtab-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.vtab-toolbar {
  padding: 12px;
  border-bottom: 1px solid #ebeef5;
  background: var(--el-color-white);
}
.vtab-body {
  flex: 1;
  padding: 12px;
  overflow: auto;
}
</style>
