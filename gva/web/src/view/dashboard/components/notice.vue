<template>
  <div class="space-y-3">
    <el-scrollbar max-height="260px">
      <div class="space-y-2 pr-1">
        <div
          v-for="(item, index) in notices"
          :key="index"
          class="group rounded-lg border border-black/10 bg-white/70 p-3 transition-all duration-200 hover:-translate-y-0.5 hover:border-black/20 hover:shadow-sm dark:border-white/10 dark:bg-white/[0.02] dark:hover:border-white/20"
        >
          <div class="flex items-start gap-3">
            <span class="mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full" :class="item.dotClass" />
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <span class="rounded-md px-2 py-0.5 text-[11px] font-semibold leading-4" :class="item.tagClass">
                  {{ item.typeTitle }}
                </span>
                <span class="shrink-0 text-[11px] text-muted-foreground">{{ item.time }}</span>
              </div>
              <el-tooltip effect="light" :content="item.title" placement="top">
                <p class="mt-1.5 line-clamp-2 text-sm text-base-text">
                  {{ item.title }}
                </p>
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup>
  const notices = [
    {
      typeTitle: '发布',
      time: '今天',
      title: 'PMocker v1.0 MVP 已发布，包含 9 大项目管理模块：范围、进度、成本、风险、问题、需求、EPS、交付物、变更。',
      dotClass: 'bg-cyan-500',
      tagClass: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-200'
    },
    {
      typeTitle: '功能',
      time: '2天前',
      title: 'CLI 新增 pmocker run 命令，支持从 PMI 镜像一键启动实例，自动初始化数据库和前端。',
      dotClass: 'bg-emerald-500',
      tagClass: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-200'
    },
    {
      typeTitle: '提示',
      time: '3天前',
      title: 'PMOCKER_HOME 环境变量可自定义数据目录，解决 Windows 下 ~/.pmocker 目录写权限问题。',
      dotClass: 'bg-amber-500',
      tagClass: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-200'
    },
    {
      typeTitle: '文档',
      time: '5天前',
      title: 'EAV 动态字段、权限模型（RBAC + EPS 继承 + 状态权限）、工作流引擎设计文档已归档到 docs/superpowers/。',
      dotClass: 'bg-violet-500',
      tagClass: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-200'
    }
  ]
</script>

<style scoped lang="scss"></style>
