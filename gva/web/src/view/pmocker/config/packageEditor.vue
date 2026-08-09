<template>
  <div class="config-editor">
    <el-page-header :content="`配置包编辑：${pkg.name || ''}`" @back="goBack">
      <template #extra>
        <el-button @click="saveDraft">保存草稿</el-button>
        <el-button type="success" @click="handlePublish">发布</el-button>
      </template>
    </el-page-header>

    <el-row :gutter="16" class="mt-4">
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>基本信息</template>
          <el-form :model="pkg" label-width="90px">
            <el-form-item label="编码">
              <el-input v-model="pkg.code" :disabled="true" />
            </el-form-item>
            <el-form-item label="名称">
              <el-input v-model="pkg.name" />
            </el-form-item>
            <el-form-item label="实体类型">
              <el-input v-model="pkg.entityType" :disabled="true" />
            </el-form-item>
            <el-form-item label="模块">
              <el-input v-model="pkg.module" :disabled="true" />
            </el-form-item>
            <el-form-item label="状态">
              <el-tag :type="statusTag(pkg.status)" size="small">{{ statusLabel(pkg.status) }}</el-tag>
              <el-tag size="small" type="info" style="margin-left: 8px">v{{ pkg.version }}</el-tag>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card shadow="never" class="mt-4">
          <template #header>
            <span>种子数据（seed_yaml）</span>
            <el-tooltip content="聚合：实体字段(fields) + 状态定义(states) + 流转规则(transitions) + 项目种子(projects)。发布时自动同步到数据库。" placement="top">
              <svg-icon icon="lucide:info" class="ml-2" />
            </el-tooltip>
          </template>
          <el-input
            v-model="seedYaml"
            type="textarea"
            :rows="22"
            class="font-mono"
            placeholder="输入 YAML 种子数据"
          />
          <div class="mt-2 text-sm text-gray-400">
            YAML 结构：entity_type / module / name / fields / states / transitions / projects（EPS 包含树层级 children）
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="never">
          <template #header>版本历史</template>
          <el-table :data="versions" size="small" border>
            <el-table-column prop="version" label="版本" width="60" />
            <el-table-column label="类型" width="70">
              <template #default="{ row }">
                <el-tag size="small" :type="row.flag === 1 ? 'warning' : 'success'">{{ row.flag === 1 ? '回滚' : '发布' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="时间" prop="CreatedAt" width="110">
              <template #default="{ row }">{{ formatDate(row.CreatedAt) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="handleRollback(row)">回滚</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPackage, updatePackageSeed, publishPackage, listPackageVersions, rollbackPackage } from '@/api/pmocker/config'
import { formatDate } from '@/utils/format'

defineOptions({ name: 'PmockerConfigPackageEditor' })

const router = useRouter()
const route = useRoute()
const pkgId = Number(route.query.id)
const pkg = ref({ code: '', name: '', entityType: '', module: '', status: '', version: 0 })
const seedYaml = ref('')
const versions = ref([])

const loadDetail = async () => {
  if (!pkgId) { ElMessage.error('缺少配置包ID'); return }
  const res = await getPackage(pkgId)
  if (res.code === 0) {
    pkg.value = res.data
    seedYaml.value = res.data.seedYaml || ''
  }
  const vres = await listPackageVersions(pkgId)
  if (vres.code === 0) versions.value = vres.data || []
}

const saveDraft = async () => {
  const res = await updatePackageSeed(pkgId, { seedYaml: seedYaml.value })
  if (res.code === 0) { ElMessage.success('草稿已保存'); loadDetail() }
}

const handlePublish = async () => {
  await ElMessageBox.confirm('确认发布？发布将自动同步字段/状态/种子到数据库并生成版本快照。', '发布确认', { type: 'warning' })
  const res = await publishPackage(pkgId)
  if (res.code === 0) { ElMessage.success('发布成功'); loadDetail() }
}

const handleRollback = async (row) => {
  await ElMessageBox.confirm(`确认回滚到版本 v${row.version}？将恢复该版本种子并重新发布。`, '回滚确认', { type: 'warning' })
  const res = await rollbackPackage(pkgId, row.ID)
  if (res.code === 0) { ElMessage.success('已回滚'); loadDetail() }
}

const goBack = () => router.push({ name: 'pmockerConfigPackageList' })
const statusLabel = (s) => ({ draft: '草稿', reviewing: '评审中', published: '已发布', archived: '已归档' }[s] || s)
const statusTag = (s) => ({ draft: 'info', reviewing: 'warning', published: 'success', archived: '' }[s] || 'info')

onMounted(loadDetail)
</script>
