<template>
  <el-dialog v-model="visible" title="登记工时" width="500px" @close="handleClose">
    <el-form :model="form" label-width="80px">
      <el-form-item label="日期" required>
        <el-date-picker v-model="form.date" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" class="w-full" />
      </el-form-item>
      <el-form-item label="工时" required>
        <el-input-number v-model="form.hours" :min="0.5" :max="24" :step="0.5" precision="1" class="w-full" />
      </el-form-item>
      <el-form-item label="时薪">
        <el-input-number v-model="form.hourlyRate" :precision="2" :min="0" class="w-full" />
      </el-form-item>
      <el-form-item label="工作内容">
        <el-input v-model="form.description" type="textarea" :rows="3" maxlength="500" show-word-limit />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSave">保存草稿</el-button>
      <el-button type="success" @click="handleSubmit">提交审批</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, watch } from 'vue'
import { createTimeEntry, submitTimeEntry } from '@/api/pmocker/timeEntry'

const props = defineProps({
  projectId: { type: Number, required: true },
  taskId: { type: Number, required: true },
  memberId: { type: Number, required: true },
  userId: { type: Number, required: true },
  defaultHourlyRate: { type: Number, default: 100 }
})
const visible = defineModel({ type: Boolean, default: false })
const emit = defineEmits(['success'])

const form = reactive({
  date: '', hours: 8, hourlyRate: props.defaultHourlyRate, description: ''
})

watch(visible, (v) => {
  if (v) {
    form.hourlyRate = props.defaultHourlyRate
  }
})

const handleClose = () => {
  form.date = ''
  form.hours = 8
  form.description = ''
}

const handleSave = async () => {
  await createTimeEntry({
    projectId: props.projectId,
    taskId: props.taskId,
    memberId: props.memberId,
    userId: props.userId,
    date: form.date,
    hours: form.hours,
    hourlyRate: form.hourlyRate,
    description: form.description
  })
  emit('success')
  visible.value = false
}

const handleSubmit = async () => {
  const res = await createTimeEntry({
    projectId: props.projectId,
    taskId: props.taskId,
    memberId: props.memberId,
    userId: props.userId,
    date: form.date,
    hours: form.hours,
    hourlyRate: form.hourlyRate,
    description: form.description
  })
  if (res.code === 0 && res.data) {
    await submitTimeEntry({ id: res.data.ID || res.data })
  }
  emit('success')
  visible.value = false
}
</script>
