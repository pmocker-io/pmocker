<template>
  <div class="relation-picker">
    <el-space wrap>
      <el-select v-model="selectedType" placeholder="关联类型" style="width: 140px" size="small">
        <el-option label="分解为" value="decomposes" />
        <el-option label="关联到" value="relates_to" />
        <el-option label="触发" value="triggers" />
        <el-option label="影响" value="impacts" />
        <el-option label="交付" value="delivers" />
        <el-option label="变更" value="changes" />
      </el-select>
      <el-input-number v-model="targetId" :min="1" placeholder="目标实体ID" style="width: 140px" size="small" controls-position="right" />
      <el-button type="primary" size="small" @click="addRelation">
        <svg-icon icon="lucide:link" /> 添加
      </el-button>
    </el-space>
    <el-divider content-position="left" class="my-2.5">已有关联</el-divider>
    <el-table :data="relations" size="small" border class="w-full">
      <el-table-column prop="relationType" label="类型" width="110" />
      <el-table-column prop="srcId" label="源ID" width="70" />
      <el-table-column prop="dstId" label="目标ID" width="90" />
      <el-table-column label="操作" width="80">
        <template #default="{ row }">
          <el-button type="danger" size="small" link @click="removeRelation(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { createRelation, deleteRelation, listRelations } from '@/api/pmocker/relation'

const props = defineProps({
  entityId: { type: Number, required: true }
})

const selectedType = ref('')
const targetId = ref(null)
const relations = ref([])

const load = async () => {
  const res = await listRelations({ entityId: props.entityId, direction: 'both' })
  if (res.code === 0) relations.value = res.data || []
}

const addRelation = async () => {
  if (!selectedType.value || !targetId.value) return
  await createRelation({
    srcId: props.entityId,
    dstId: targetId.value,
    relationType: selectedType.value
  })
  selectedType.value = ''
  targetId.value = null
  load()
}

const removeRelation = async (row) => {
  await deleteRelation({ id: row.id })
  load()
}

watch(() => props.entityId, load, { immediate: true })
</script>

<style scoped>
.relation-picker { padding: 10px 0; }
</style>
