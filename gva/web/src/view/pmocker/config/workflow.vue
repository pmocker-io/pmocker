<template>
  <div>
    <el-card shadow="never">
      <div class="flex items-center gap-3">
        <svg-icon icon="lucide:workflow" class="text-2xl" />
        <div>
          <div class="font-bold">工作流管理</div>
          <div class="text-gray-500 text-sm mt-1">
            工作流管理规划中。PMocker 工作流由各模块的 YAML 定义（pmocker/workflows）在启动时自动装载，
            驱动实体状态机流转（如需求评审、变更审批）。
          </div>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="mt-4">
      <el-empty description="工作流管理规划中，暂无独立的管理界面" />
    </el-card>
  </div>
</template>

<script setup>
defineOptions({ name: 'PmockerConfigWorkflow' })
</script>
