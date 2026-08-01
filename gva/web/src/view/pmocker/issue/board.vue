<template>
  <div>
    <div class="flex gap-4 overflow-x-auto pb-4">
      <div
        v-for="col in columns"
        :key="col.status"
        class="flex-shrink-0 w-72"
      >
        <div class="flex items-center justify-between mb-3">
          <span class="font-medium text-base-text">{{ col.title }}</span>
          <el-tag size="small" round>{{ col.items.length }}</el-tag>
        </div>
        <draggable
          v-model="col.items"
          group="issues"
          item-key="ID"
          class="space-y-3 min-h-[200px] p-2 rounded bg-muted"
          @end="onDragEnd($event, col)"
        >
          <template #item="{ element }">
            <el-card shadow="hover" class="cursor-move">
              <div class="flex items-start justify-between">
                <span class="text-sm font-medium">{{ element.title }}</span>
                <el-tag size="small" :type="priorityType(element.priority)">
                  {{ priorityLabel(element.priority) }}
                </el-tag>
              </div>
              <p v-if="element.description" class="text-xs text-muted-foreground mt-2 line-clamp-2">
                {{ element.description }}
              </p>
              <div class="flex items-center justify-between mt-3">
                <span class="text-xs text-muted-foreground">#{{ element.ID }}</span>
                <span v-if="element.assigneeName" class="text-xs">{{ element.assigneeName }}</span>
              </div>
            </el-card>
          </template>
        </draggable>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import draggable from 'vuedraggable'
import { getIssueBoard, updateIssue } from '@/api/pmocker/issue'

defineOptions({ name: 'PmockerIssueBoard' })

const columns = ref([
  { title: '待处理', status: 'open', items: [] },
  { title: '处理中', status: 'in_progress', items: [] },
  { title: '已解决', status: 'resolved', items: [] },
  { title: '已关闭', status: 'closed', items: [] }
])

const priorityType = (p) => ({ urgent: 'danger', high: 'warning', medium: 'info', low: 'info' }[p] || 'info')
const priorityLabel = (p) => ({ urgent: '紧急', high: '高', medium: '中', low: '低' }[p] || p)

const loadBoard = async () => {
  const res = await getIssueBoard({})
  if (res.code === 0) {
    const groups = res.data.groups || {}
    columns.value.forEach(col => {
      col.items = groups[col.status] || []
    })
  }
}

const onDragEnd = async (evt, targetCol) => {
  const item = targetCol.items[evt.newIndex]
  if (item && item.status !== targetCol.status) {
    const oldStatus = item.status
    item.status = targetCol.status
    const res = await updateIssue({ ID: item.ID, status: targetCol.status })
    if (res.code !== 0) {
      item.status = oldStatus
      await loadBoard()
    }
  }
}

loadBoard()
</script>
