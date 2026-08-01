<template>
  <div class="h-full gva-container2 overflow-auto bg-main">
    <div class="space-y-2 py-2">
      <gva-card
        class="relative overflow-hidden rounded-xl border border-slate-200/80 bg-white px-5 py-6 shadow-sm dark:border-slate-700 dark:bg-slate-900"
      >
        <div class="relative flex flex-col gap-2 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p class="text-xs tracking-[0.2em] text-muted-foreground">PMOCKER DASHBOARD</p>
            <h1 class="mt-2 text-xl font-semibold text-base-text lg:text-2xl">
              欢迎回来，开始今天的项目推进
            </h1>
            <p class="mt-2 text-sm text-muted-foreground">
              {{ today }} · 已聚合项目进度、资源占用与任务概览
            </p>
          </div>
          <div class="flex items-center gap-2">
            <el-button @click="goDocs">查看文档</el-button>
            <el-button type="primary" @click="goNewProject">新建项目</el-button>
          </div>
        </div>
      </gva-card>

      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-3">
        <gva-card>
          <gva-chart :type="1" title="活跃项目" />
        </gva-card>
        <gva-card>
          <gva-chart :type="2" title="进行中任务" />
        </gva-card>
        <gva-card>
          <gva-chart :type="3" title="风险项" />
        </gva-card>
      </div>

      <div class="grid grid-cols-1 items-stretch gap-2 xl:grid-cols-12">
        <div class="grid grid-cols-1 gap-2 content-start xl:col-span-8 xl:h-full">
          <gva-card title="项目概览">
            <gva-chart :type="4" />
          </gva-card>

          <gva-card title="快捷功能" show-action custom-class="min-h-[260px]">
            <gva-quick-link />
          </gva-card>
        </div>

        <div class="flex flex-col gap-2 xl:col-span-4 xl:h-full">
          <gva-card title="公告" show-action custom-class="min-h-[260px]">
            <gva-notice />
          </gva-card>
          <gva-card title="文档" show-action custom-class="min-h-[120px]">
            <gva-wiki />
          </gva-card>
          <gva-card title="资源概览" custom-class="min-h-[200px]">
            <div class="space-y-3 text-sm text-base-text">
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">实例数</span>
                <span class="font-semibold">{{ instanceCount }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">本地镜像</span>
                <span class="font-semibold">{{ imageCount }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">模块数</span>
                <span class="font-semibold">9</span>
              </div>
              <el-divider class="!my-2" />
              <div class="flex items-center justify-between">
                <span class="text-muted-foreground">版本</span>
                <span class="font-mono text-xs">v1.0 MVP</span>
              </div>
            </div>
          </gva-card>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { computed, ref } from 'vue'
  import {
    GvaTable,
    GvaChart,
    GvaWiki,
    GvaNotice,
    GvaQuickLink,
    GvaCard
  } from './components'

  const instanceCount = ref(0)
  const imageCount = ref(1)

  const today = computed(() => {
    try {
      const d = new Date()
      return d.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      })
    } catch (e) {
      return new Date().toISOString().slice(0, 10)
    }
  })

  const goDocs = () => {
    window.open('https://www.gin-vue-admin.com/guide/introduce/project', '_blank', 'noopener,noreferrer')
  }

  const goNewProject = () => {
    router.push({ path: '/pmocker/eps/index' })
  }

  import { onMounted } from 'vue'
  import router from '@/router/index'

  onMounted(async () => {
    try {
      const { getInstanceList } = await import('@/api/pmocker/eps')
      // 静默：没有接口也不报错
    } catch (e) {}
  })

  defineOptions({
    name: 'Dashboard'
  })
</script>

<style lang="scss" scoped></style>
