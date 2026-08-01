<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="变更ID">
          <el-input v-model="searchInfo.changeId" placeholder="变更ID" clearable />
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="searchInfo.dateRange"
            type="daterange"
            value-format="YYYY-MM-DD"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
      <el-timeline>
        <el-timeline-item
          v-for="log in logList"
          :key="log.ID"
          :timestamp="formatDate(log.CreatedAt)"
          :type="logType(log.action)"
          placement="top"
        >
          <el-card>
            <div class="flex items-center justify-between">
              <span class="font-medium">{{ log.title }}</span>
              <el-tag size="small" :type="logType(log.action)">{{ logLabel(log.action) }}</el-tag>
            </div>
            <p class="text-muted-foreground mt-2">{{ log.description }}</p>
            <p class="text-muted-foreground text-xs mt-1">操作人: {{ log.operator }}</p>
          </el-card>
        </el-timeline-item>
      </el-timeline>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { getChangeLogs } from '@/api/pmocker/change'

defineOptions({ name: 'PmockerChangeLog' })

const searchInfo = ref({})
const logList = ref([])

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : ''

const logType = (action) => {
  const map = {
    submit: 'primary', analyze: 'info', ccb_review: 'warning',
    approve: 'success', reject: 'danger', implement: 'primary',
    verify: 'info', close: 'success'
  }
  return map[action] || 'info'
}

const logLabel = (action) => {
  const map = {
    submit: '提交', analyze: '分析', ccb_review: 'CCB评审',
    approve: '批准', reject: '驳回', implement: '实施',
    verify: '验证', close: '关闭'
  }
  return map[action] || action
}

const loadData = async () => {
  const res = await getChangeLogs(searchInfo.value)
  if (res.code === 0) {
    logList.value = res.data.list || []
  }
}

loadData()
</script>
