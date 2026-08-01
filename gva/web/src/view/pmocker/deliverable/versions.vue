<template>
  <div>
    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo">
        <el-form-item label="交付物">
          <el-input v-model="searchInfo.deliverableId" placeholder="交付物ID" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadVersions">查询</el-button>
          <el-button type="success" @click="openVersionDialog">
            <svg-icon icon="lucide:git-branch" /> 创建版本
          </el-button>
        </el-form-item>
      </el-form>
    </div>
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
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { getDeliverableVersions, createDeliverableVersion, createDeliverableBaseline } from '@/api/pmocker/deliverable'

defineOptions({ name: 'PmockerDeliverableVersions' })

const searchInfo = ref({})
const versions = ref([])
const dialogVisible = ref(false)
const formRef = ref(null)

const form = reactive({ deliverableId: null, version: '', description: '' })
const rules = {
  deliverableId: [{ required: true, message: '请输入交付物ID', trigger: 'blur' }],
  version: [{ required: true, message: '请输入版本号', trigger: 'blur' }]
}

const formatDate = (dateStr) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : ''

const loadVersions = async () => {
  const res = await getDeliverableVersions(searchInfo.value)
  if (res.code === 0) {
    versions.value = res.data.list || []
  }
}

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

loadVersions()
</script>
