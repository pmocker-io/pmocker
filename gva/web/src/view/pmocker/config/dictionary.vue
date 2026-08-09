<template>
  <div>
    <el-card shadow="never">
      <div class="flex items-center gap-3">
        <svg-icon icon="lucide:book-open" class="text-2xl" />
        <div>
          <div class="font-bold">字典管理</div>
          <div class="text-gray-500 text-sm mt-1">
            字典数据（数据字典）由系统管理模块统一维护，请前往「系统设置 → 字典管理」页面进行增删改查。PMocker 配置暂不维护字典。
          </div>
        </div>
        <div class="ml-auto">
          <el-button type="primary" @click="goDict">
            <svg-icon icon="lucide:external-link" /> 前往字典管理
          </el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="mt-4">
      <template #header>
        <span>说明</span>
      </template>
      <ul class="text-gray-500 text-sm leading-7 list-disc pl-5">
        <li>字典通常用于下拉选项、状态文案等枚举展示。</li>
        <li>PMocker 各模块的动态字段（EAV）选项存于字段定义 options_json，与系统字典相互独立。</li>
        <li>需要系统级字典时，通过上方按钮跳转到系统管理维护。</li>
      </ul>
    </el-card>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

defineOptions({ name: 'PmockerConfigDictionary' })

const router = useRouter()

const goDict = () => {
  router.push({ path: '/superAdmin/dictionary' })
}
</script>
