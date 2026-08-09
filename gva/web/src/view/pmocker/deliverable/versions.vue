<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="交付物">
          <el-input v-model="searchInfo.deliverableId" placeholder="交付物ID" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="reloadAll">查询</el-button>
          <el-button type="success" @click="openVersionDialog">
            <svg-icon icon="lucide:git-branch" /> 创建版本
          </el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 检入检出锁定状态面板 -->
    <el-card v-if="searchInfo.deliverableId" class="lock-panel" shadow="never">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <span class="text-base-text font-medium">锁定状态：</span>
          <el-tag v-if="lockStatus === 'available'" type="success" size="default">可用</el-tag>
          <el-tag v-else type="danger" size="default">已检出</el-tag>
          <span v-if="lockStatus === 'checked_out'" class="text-muted-foreground">
            检出人：用户ID {{ checkedOutBy }}{{ isLockedByMe ? '（我）' : '' }}
            <template v-if="checkedOutAt">｜检出时间：{{ formatDate(checkedOutAt) }}</template>
          </span>
        </div>
        <div class="flex gap-2">
          <el-button
            v-if="lockStatus === 'available'"
            type="warning"
            @click="handleCheckOut"
          >
            <svg-icon icon="lucide:lock" /> 检出
          </el-button>
          <el-tooltip
            v-if="lockStatus === 'checked_out' && !isLockedByMe"
            :content="`已被用户 ${checkedOutBy} 检出，仅检出人或管理员可检入`"
            placement="top"
          >
            <span>
              <el-button type="primary" disabled>
                <svg-icon icon="lucide:lock-open" /> 检入
              </el-button>
            </span>
          </el-tooltip>
          <el-button
            v-if="lockStatus === 'checked_out' && isLockedByMe"
            type="primary"
            @click="openCheckInDialog"
          >
            <svg-icon icon="lucide:lock-open" /> 检入
          </el-button>
        </div>
      </div>
    </el-card>

    <div class="gva-table-box">
      <el-timeline v-if="versions.length > 0">
        <el-timeline-item
          v-for="v in versions"
          :key="v.ID"
          :timestamp="formatDate(v.CreatedAt)"
          :type="v.isBaseline ? 'success' : 'primary'"
          placement="top"
        >
          <el-card>
            <div class="flex items-center justify-between">
              <span class="text-base-text font-medium">v{{ v.version }}</span>
              <div class="flex gap-2">
                <el-tag v-if="v.isBaseline" type="success" size="small">基线</el-tag>
                <el-button v-if="!v.isBaseline" type="primary" link @click="handleBaseline(v)">设为基线</el-button>
              </div>
            </div>
            <p v-if="v.description" class="text-muted-foreground mt-2">{{ v.description }}</p>
          </el-card>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else description="暂无版本" />
    </div>

    <el-dialog v-model="dialogVisible" title="创建新版本" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="交付物ID" prop="deliverableId">
          <el-input-number v-model="form.deliverableId" :min="1" :disabled="!!searchInfo.deliverableId" />
        </el-form-item>
        <el-form-item label="版本号" prop="version">
          <el-input v-model="form.version" placeholder="如 1.0.0" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateVersion">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="checkInDialogVisible" title="检入交付物" width="500px">
      <el-form ref="checkInFormRef" :model="checkInForm" label-width="100px">
        <el-form-item label="版本说明">
          <el-input
            v-model="checkInForm.versionNote"
            type="textarea"
            :rows="3"
            placeholder="填写版本说明后将自动生成新版本记录；留空则仅解锁"
          />
        </el-form-item>
        <el-form-item label="附件引用">
          <el-input v-model="checkInForm.fileRef" placeholder="可选，附件路径" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="checkInDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCheckIn">确认检入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getDeliverableVersions,
  createDeliverableVersion,
  createDeliverableBaseline,
  findDeliverable,
  checkOutDeliverable,
  checkInDeliverable
} from '@/api/pmocker/deliverable'
import { useUserStore } from '@/pinia/modules/user'
import { useProjectStore } from '@/pinia'

defineOptions({ name: 'PmockerDeliverableVersions' })

const userStore = useUserStore()
const projectStore = useProjectStore()
const currentUserId = computed(() => Number(userStore.userInfo?.ID) || 0)

const searchInfo = ref({})
const versions = ref([])
const currentDeliverable = ref(null)
const dialogVisible = ref(false)
const formRef = ref(null)

const form = reactive({ deliverableId: null, version: '', description: '' })
const rules = {
  deliverableId: [{ required: true, message: '请输入交付物ID', trigger: 'blur' }],
  version: [{ required: true, message: '请输入版本号', trigger: 'blur' }]
}

const checkInDialogVisible = ref(false)
const checkInFormRef = ref(null)
const checkInForm = reactive({ id: null, versionNote: '', fileRef: '' })

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : ''

const lockStatus = computed(() => {
  const v = currentDeliverable.value?.attrs?.lock_status
  return v === 'checked_out' ? 'checked_out' : 'available'
})
const checkedOutBy = computed(() => {
  const v = currentDeliverable.value?.attrs?.checked_out_by
  return v === null || v === undefined || v === '' ? null : Number(v)
})
const checkedOutAt = computed(() => currentDeliverable.value?.attrs?.checked_out_at || '')
const isLockedByMe = computed(() => lockStatus.value === 'checked_out' && checkedOutBy.value === currentUserId.value)

const loadVersions = async () => {
  const res = await getDeliverableVersions({ projectId: projectStore.projectId, ...searchInfo.value })
  if (res.code === 0) {
    versions.value = res.data.list || []
  }
}

const loadDeliverable = async () => {
  const id = Number(searchInfo.value.deliverableId)
  if (!id) {
    currentDeliverable.value = null
    return
  }
  const res = await findDeliverable({ id })
  if (res.code === 0) {
    currentDeliverable.value = res.data
  }
}

const reloadAll = async () => {
  await loadDeliverable()
  await loadVersions()
}

// 交付物ID 变化时重新加载锁定状态与版本
watch(() => searchInfo.value.deliverableId, () => {
  if (searchInfo.value.deliverableId) {
    reloadAll()
  }
})

const openVersionDialog = () => {
  form.deliverableId = searchInfo.value.deliverableId ? Number(searchInfo.value.deliverableId) : null
  form.version = ''
  form.description = ''
  dialogVisible.value = true
}

const handleCreateVersion = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const res = await createDeliverableVersion(form)
    if (res.code === 0) {
      ElMessage.success('版本创建成功')
      dialogVisible.value = false
      loadVersions()
    }
  })
}

const handleBaseline = async (v) => {
  const res = await createDeliverableBaseline({ versionId: v.ID })
  if (res.code === 0) {
    ElMessage.success('已设为基线')
    loadVersions()
  }
}

const handleCheckOut = async () => {
  const id = Number(searchInfo.value.deliverableId)
  if (!id) {
    ElMessage.warning('请先填写交付物ID')
    return
  }
  const res = await checkOutDeliverable({ id })
  if (res.code === 0) {
    ElMessage.success('检出成功')
    await reloadAll()
  }
}

const openCheckInDialog = () => {
  checkInForm.id = Number(searchInfo.value.deliverableId)
  checkInForm.versionNote = ''
  checkInForm.fileRef = ''
  checkInDialogVisible.value = true
}

const handleCheckIn = async () => {
  const res = await checkInDeliverable(checkInForm)
  if (res.code === 0) {
    ElMessage.success('检入成功')
    checkInDialogVisible.value = false
    await reloadAll()
  }
}

reloadAll()
</script>

<style scoped>
.lock-panel {
  margin-bottom: 12px;
}
</style>
