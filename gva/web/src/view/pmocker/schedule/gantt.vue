<template>
  <div>
    <div class="gva-btn-list">
      <el-button type="primary" @click="openTaskDialog(null)">
        <svg-icon icon="lucide:plus" /> 新增任务
      </el-button>
      <el-button-group class="ml-2">
        <el-button :type="viewMode==='day'?'primary':''" @click="setViewMode('day')">日</el-button>
        <el-button :type="viewMode==='week'?'primary':''" @click="setViewMode('week')">周</el-button>
        <el-button :type="viewMode==='month'?'primary':''" @click="setViewMode('month')">月</el-button>
      </el-button-group>
    </div>
    <div class="gva-table-box">
      <div v-if="tableData.length" ref="chartRef" style="width: 100%; height: 500px" />
      <el-empty v-else description="暂无任务数据" />
    </div>
    <div class="gva-table-box mt-4">
      <el-table :data="tableData" row-key="ID">
        <el-table-column label="任务名称" prop="title" min-width="200" />
        <el-table-column label="开始日期" width="120">
          <template #default="{ row }">{{ formatDate(row.attrs?.start_date) }}</template>
        </el-table-column>
        <el-table-column label="结束日期" width="120">
          <template #default="{ row }">{{ formatDate(row.attrs?.end_date) }}</template>
        </el-table-column>
        <el-table-column label="工期" width="80">
          <template #default="{ row }">{{ row.attrs?.duration ?? '' }}</template>
        </el-table-column>
        <el-table-column label="进度" width="150">
          <template #default="{ row }">
            <el-progress :percentage="row.attrs?.progress || 0" />
          </template>
        </el-table-column>
        <el-table-column label="关键路径" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.attrs?.is_critical_path" type="danger">关键</el-tag>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="总浮动" width="100">
          <template #default="{ row }">{{ row.attrs?.total_float ?? '' }}</template>
        </el-table-column>
        <el-table-column label="依赖类型" width="120">
          <template #default="{ row }">{{ dependencyLabel(row.attrs?.dependency_type) }}</template>
        </el-table-column>
        <el-table-column label="状态" prop="status" width="110">
          <template #default="{ row }">
            <el-tag :type="taskStatusType(row.status)">{{ taskStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openTaskDialog(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="800px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="名称" prop="title" :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]">
          <el-input v-model="form.title" />
        </el-form-item>
        <DynamicForm v-model="form" entity-type="task" />
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as echarts from 'echarts'
import {
  getScheduleTasks,
  createScheduleTask,
  updateTask,
  deleteTask
} from '@/api/pmocker/schedule'
import DynamicForm from '../components/DynamicForm.vue'

defineOptions({ name: 'PmockerScheduleGantt' })

const tableData = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref(null)
const dialogType = ref('add')

const form = reactive({ ID: null, title: '', status: 'planned', attrs: {} })

const chartRef = ref(null)
let chartInstance = null
const viewMode = ref('day')

const MIN_INTERVAL_MAP = {
  day: 24 * 3600 * 1000,
  week: 7 * 24 * 3600 * 1000,
  month: 30 * 24 * 3600 * 1000
}

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleDateString('zh-CN') : ''

const taskStatusType = (status) => {
  const map = { planned: 'info', pending: 'info', running: 'warning', in_progress: 'warning', done: 'success', completed: 'success' }
  return map[status] || 'info'
}

const taskStatusLabel = (status) => {
  const map = { planned: '计划中', pending: '待开始', running: '进行中', in_progress: '进行中', done: '已完成', completed: '已完成' }
  return map[status] || status
}

const dependencyLabel = (type) => {
  const map = { FS: '完成-开始', SS: '开始-开始', FF: '完成-完成', SF: '开始-完成' }
  return map[type] || type || '—'
}

// 里程碑判断：task_type 无 milestone 枚举，按 duration=0 判定
const isMilestoneTask = (t) => {
  const d = Number(t?.attrs?.duration)
  return d === 0
}

// 解析前置任务（predecessors 为 json 字段，兼容 [id,...] / [{id,type},...] 多种格式）
const parsePredecessors = (raw) => {
  if (!raw) return []
  let arr = raw
  if (typeof raw === 'string') {
    try { arr = JSON.parse(raw) } catch (e) { return [] }
  }
  if (!Array.isArray(arr)) return []
  const result = []
  arr.forEach(item => {
    let id = null
    if (typeof item === 'number') id = item
    else if (typeof item === 'string') id = Number(item)
    else if (item && typeof item === 'object') {
      const pid = item.id || item.task_id || item.predecessor_id || item.ID
      id = pid != null ? Number(pid) : null
    }
    if (id != null && !isNaN(id)) result.push({ id })
  })
  return result
}

const loadData = async () => {
  const res = await getScheduleTasks({})
  if (res.code === 0) {
    tableData.value = res.data.list || []
    await nextTick()
    renderGantt()
  }
}

const renderGantt = () => {
  if (chartInstance) { chartInstance.dispose(); chartInstance = null }
  if (!chartRef.value) return

  const allTasks = tableData.value
  // 仅渲染有 start_date/end_date 的任务，其余跳过
  const validTasks = allTasks.filter(t => t && t.attrs && t.attrs.start_date && t.attrs.end_date)
  if (!validTasks.length) return

  chartInstance = echarts.init(chartRef.value)

  // 任务 ID -> 在 validTasks 中的索引
  const taskIndexMap = {}
  validTasks.forEach((t, idx) => { taskIndexMap[Number(t.ID)] = idx })

  // 计算时间范围，留 10% 边距
  const times = []
  validTasks.forEach(t => {
    times.push(new Date(t.attrs.start_date).getTime())
    times.push(new Date(t.attrs.end_date).getTime())
  })
  let minTime = Math.min(...times)
  let maxTime = Math.max(...times)
  if (minTime === maxTime) {
    minTime -= 24 * 3600 * 1000
    maxTime += 24 * 3600 * 1000
  }
  const pad = (maxTime - minTime) * 0.1

  const yCategories = validTasks.map(t => t.title || '(未命名)')

  // 任务条 custom series
  // value: [taskName, startTime, endTime, progress, isCritical, idx, isMilestone]
  const barSeries = {
    type: 'custom',
    name: '任务条',
    renderItem: (params, api) => {
      const isMilestone = api.value(6)
      if (isMilestone) return // 里程碑由 scatter 绘制
      const categoryName = api.value(0)
      const start = api.coord([api.value(1), categoryName])
      const end = api.coord([api.value(2), categoryName])
      const progress = api.value(3)
      const isCritical = api.value(4)
      const title = api.value(5)
      const barHeight = api.size([0, 1])[1] * 0.5
      const rectShape = echarts.graphic.clipRectByRect(
        { x: start[0], y: start[1] - barHeight / 2, width: Math.max(end[0] - start[0], 1), height: barHeight },
        { x: params.coordSys.x, y: params.coordSys.y, width: params.coordSys.width, height: params.coordSys.height }
      )
      if (!rectShape) return
      const clampedProgress = Math.min(Math.max(progress, 0), 100)
      const progressWidth = rectShape.width * (clampedProgress / 100)
      const fillColor = isCritical ? '#ffe6e6' : '#e6f7ff'
      const strokeColor = isCritical ? '#ff4d4f' : '#1890ff'
      return {
        type: 'group',
        children: [
          { type: 'rect', transition: ['shape'], shape: rectShape, style: { fill: fillColor, stroke: strokeColor, lineWidth: 1 } },
          { type: 'rect', shape: { x: rectShape.x, y: rectShape.y, width: progressWidth, height: rectShape.height }, style: { fill: strokeColor, opacity: 0.6 } },
          { type: 'text', style: { text: title, x: rectShape.x + 4, y: rectShape.y + rectShape.height / 2, fill: '#333', fontSize: 11, textVerticalAlign: 'middle' } }
        ]
      }
    },
    encode: { x: [1, 2], y: 0 },
    data: validTasks.map((t, idx) => {
      const progress = Number(t.attrs.progress) || 0
      const isCritical = t.attrs.is_critical_path ? 1 : 0
      const isMilestone = isMilestoneTask(t) ? 1 : 0
      return {
        name: t.title,
        value: [
          t.title || '(未命名)',
          new Date(t.attrs.start_date).getTime(),
          new Date(t.attrs.end_date).getTime(),
          progress,
          isCritical,
          idx,
          isMilestone
        ]
      }
    })
  }

  // 里程碑 scatter 菱形
  const milestoneSeries = {
    type: 'scatter',
    name: '里程碑',
    symbol: 'diamond',
    symbolSize: 18,
    itemStyle: { color: '#faad14', borderColor: '#d48806', borderWidth: 1.5 },
    encode: { x: [1, 2], y: 0 },
    data: validTasks.filter(t => isMilestoneTask(t)).map(t => {
      const idx = taskIndexMap[Number(t.ID)]
      const time = new Date(t.attrs.end_date || t.attrs.start_date).getTime()
      const progress = Number(t.attrs.progress) || 0
      const isCritical = t.attrs.is_critical_path ? 1 : 0
      return {
        name: t.title,
        value: [t.title || '(未命名)', time, time, progress, isCritical, idx, 1]
      }
    })
  }

  // 依赖连线 custom series（基于 predecessors json 字段）
  // value: [fromName, toName, fromEnd, toStart]
  const depLinks = []
  validTasks.forEach((t, toIdx) => {
    const predecessors = parsePredecessors(t.attrs.predecessors)
    predecessors.forEach(p => {
      const fromIdx = taskIndexMap[p.id]
      if (fromIdx === undefined || fromIdx === null) return
      const fromTask = validTasks[fromIdx]
      if (!fromTask) return
      const fromEnd = new Date(fromTask.attrs.end_date).getTime()
      const toStart = new Date(t.attrs.start_date).getTime()
      depLinks.push({
        value: [fromTask.title || '(未命名)', t.title || '(未命名)', fromEnd, toStart]
      })
    })
  })
  const depSeries = {
    type: 'custom',
    name: '依赖关系',
    renderItem: (params, api) => {
      const fromName = api.value(0)
      const toName = api.value(1)
      const fromEnd = api.value(2)
      const toStart = api.value(3)
      const fromPoint = api.coord([fromEnd, fromName])
      const toPoint = api.coord([toStart, toName])
      const stroke = '#888'
      const midX = fromPoint[0] + (toPoint[0] - fromPoint[0]) / 2
      return {
        type: 'group',
        children: [
          { type: 'line', shape: { x1: fromPoint[0], y1: fromPoint[1], x2: midX, y2: fromPoint[1] }, style: { stroke, lineWidth: 1 } },
          { type: 'line', shape: { x1: midX, y1: fromPoint[1], x2: midX, y2: toPoint[1] }, style: { stroke, lineWidth: 1 } },
          { type: 'line', shape: { x1: midX, y1: toPoint[1], x2: toPoint[0] - 6, y2: toPoint[1] }, style: { stroke, lineWidth: 1 } },
          { type: 'polygon', shape: { points: [[toPoint[0], toPoint[1]], [toPoint[0] - 6, toPoint[1] - 3], [toPoint[0] - 6, toPoint[1] + 3]] }, style: { fill: stroke } }
        ]
      }
    },
    encode: { x: [2, 3], y: [0, 1] },
    data: depLinks,
    z: 0
  }

  chartInstance.setOption({
    title: { text: '甘特图', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: {
      trigger: 'item',
      formatter: (p) => {
        if (p.seriesName === '依赖关系') return ''
        const v = p.value
        if (!v || v.length < 7) return ''
        const idx = v[5]
        const task = validTasks[idx]
        if (!task) return ''
        const start = new Date(v[1]).toLocaleDateString('zh-CN')
        const end = new Date(v[2]).toLocaleDateString('zh-CN')
        const progress = v[3]
        const isCritical = v[4]
        const isMilestone = v[6]
        return [
          `<div style="font-weight:600;margin-bottom:4px">${task.title || ''}</div>`,
          `<div>类型：${isMilestone ? '里程碑' : '任务'}</div>`,
          `<div>开始：${start}</div>`,
          `<div>结束：${end}</div>`,
          `<div>进度：${progress}%</div>`,
          `<div>关键路径：${isCritical ? '是' : '否'}</div>`,
          `<div>状态：${taskStatusLabel(task.status)}</div>`
        ].join('')
      }
    },
    grid: { left: 220, right: 40, top: 50, bottom: 60 },
    xAxis: {
      type: 'time',
      min: minTime - pad,
      max: maxTime + pad,
      minInterval: MIN_INTERVAL_MAP[viewMode.value],
      axisLabel: { fontSize: 11 }
    },
    yAxis: {
      type: 'category',
      data: yCategories,
      inverse: true,
      axisLabel: { fontSize: 11 }
    },
    dataZoom: [
      { type: 'slider', xAxisIndex: 0, filterMode: 'weakFilter', bottom: 10, height: 18 },
      { type: 'inside', xAxisIndex: 0, filterMode: 'weakFilter' }
    ],
    series: [depSeries, barSeries, milestoneSeries]
  })
}

const setViewMode = (mode) => {
  viewMode.value = mode
  if (chartInstance) {
    chartInstance.setOption({ xAxis: { minInterval: MIN_INTERVAL_MAP[mode] } })
  }
}

const handleResize = () => {
  if (chartInstance) chartInstance.resize()
}

const openTaskDialog = (row) => {
  dialogType.value = row ? 'edit' : 'add'
  dialogTitle.value = row ? '编辑任务' : '新增任务'
  if (row) {
    Object.assign(form, {
      ID: row.ID,
      title: row.title || '',
      status: row.status || 'planned',
      attrs: { ...(row.attrs || {}) }
    })
  } else {
    Object.assign(form, { ID: null, title: '', status: 'planned', attrs: {} })
  }
  dialogVisible.value = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const payload = {
      title: form.title,
      status: form.status,
      attrs: form.attrs
    }
    let res
    if (dialogType.value === 'edit') {
      res = await updateTask({
        id: form.ID,
        entity_type: 'task',
        ...payload
      })
    } else {
      res = await createScheduleTask(payload)
    }
    if (res.code === 0) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      loadData()
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确认删除该任务吗？', '提示', { type: 'warning' })
    .then(async () => {
      const res = await deleteTask({ ID: row.ID })
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadData()
      }
    })
    .catch(() => {})
}

onMounted(() => {
  loadData()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (chartInstance) { chartInstance.dispose(); chartInstance = null }
})
</script>
