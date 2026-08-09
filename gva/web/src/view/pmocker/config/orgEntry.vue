<template>
  <div>
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="组织与权限由系统管理模块统一维护，PMocker 配置页提供快捷入口，点击卡片跳转到对应管理页面。"
      style="margin-bottom: 12px"
    />
    <el-row :gutter="12">
      <el-col v-for="entry in entries" :key="entry.path" :span="6">
        <el-card shadow="hover" class="org-card" @click="goEntry(entry)">
          <div class="flex items-center gap-3">
            <div class="org-icon">
              <svg-icon :icon="entry.icon" />
            </div>
            <div>
              <div class="font-bold">{{ entry.title }}</div>
              <div class="text-gray-500 text-sm mt-1">{{ entry.desc }}</div>
            </div>
          </div>
          <div class="mt-3 flex justify-end">
            <el-link type="primary" :underline="false">
              前往
              <svg-icon icon="lucide:arrow-right" />
            </el-link>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

defineOptions({ name: 'PmockerConfigOrgEntry' })

const router = useRouter()

const entries = [
  { title: '用户管理', desc: '系统用户账号与基本信息', icon: 'lucide:users', path: '/superAdmin/user' },
  { title: '部门管理', desc: '组织架构与部门层级', icon: 'lucide:building-2', path: '/superAdmin/department' },
  { title: '岗位管理', desc: '岗位设置与职责说明', icon: 'lucide:badge', path: '/superAdmin/position' },
  { title: '角色管理', desc: '角色权限与数据权限', icon: 'lucide:shield', path: '/superAdmin/authority' }
]

const goEntry = (entry) => {
  router.push({ path: entry.path })
}
</script>

<style scoped>
.org-card {
  cursor: pointer;
  margin-bottom: 12px;
  transition: transform 0.15s;
}
.org-card:hover {
  transform: translateY(-2px);
}
.org-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border-radius: 8px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}
</style>
