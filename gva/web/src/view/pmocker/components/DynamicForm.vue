<template>
  <div v-loading="loading">
    <!-- 核心字段区 -->
    <el-divider v-if="coreFields.length > 0" content-position="left">基本信息</el-divider>
    <el-row :gutter="16">
      <el-col v-for="field in coreFields" :key="field.field_key" :span="getColSpan(field)">
        <el-form-item :label="field.field_label" :prop="'attrs.' + field.field_key">
          <!-- 人员字段：用户选择器 -->
          <el-select
            v-if="isUserField(field.field_key) || field.data_type === 'user'"
            v-model="attrs[field.field_key]"
            filterable
            clearable
            placeholder="请选择用户"
            style="width: 100%"
            @change="onUserFieldChange(field.field_key, $event)"
          >
            <el-option
              v-for="u in userList"
              :key="u.ID"
              :label="u.nickName + ' (' + u.userName + ')'"
              :value="u.ID"
            />
          </el-select>
          <!-- 枚举字段 -->
          <el-select
            v-else-if="field.data_type === 'enum'"
            v-model="attrs[field.field_key]"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="opt in parseOptions(field.options_json)"
              :key="opt"
              :label="opt"
              :value="opt"
            />
          </el-select>
          <!-- 布尔字段 -->
          <el-switch
            v-else-if="field.data_type === 'bool'"
            v-model="attrs[field.field_key]"
          />
          <!-- 日期字段 -->
          <el-date-picker
            v-else-if="field.data_type === 'date'"
            v-model="attrs[field.field_key]"
            type="date"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
          <!-- 日期时间字段 -->
          <el-date-picker
            v-else-if="field.data_type === 'datetime'"
            v-model="attrs[field.field_key]"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
          <!-- 数字字段 -->
          <el-input-number
            v-else-if="['int', 'decimal'].includes(field.data_type)"
            v-model="attrs[field.field_key]"
            :precision="field.data_type === 'decimal' ? 2 : 0"
            controls-position="right"
            style="width: 100%"
          />
          <!-- 多行文本 -->
          <el-input
            v-else-if="['text', 'json'].includes(field.data_type)"
            v-model="attrs[field.field_key]"
            type="textarea"
            :rows="field.data_type === 'json' ? 4 : 3"
          />
          <!-- 普通字符串 -->
          <el-input
            v-else
            v-model="attrs[field.field_key]"
          />
        </el-form-item>
      </el-col>
    </el-row>

    <!-- 扩展属性区 -->
    <el-collapse v-if="extendedFields.length > 0">
      <el-collapse-item :title="`扩展属性 (${extendedFields.length})`" name="extended">
        <el-row :gutter="16">
          <el-col v-for="field in extendedFields" :key="field.field_key" :span="getColSpan(field)">
            <el-form-item :label="field.field_label" :prop="'attrs.' + field.field_key">
              <el-select
                v-if="isUserField(field.field_key) || field.data_type === 'user'"
                v-model="attrs[field.field_key]"
                filterable
                clearable
                placeholder="请选择用户"
                style="width: 100%"
                @change="onUserFieldChange(field.field_key, $event)"
              >
                <el-option
                  v-for="u in userList"
                  :key="u.ID"
                  :label="u.nickName + ' (' + u.userName + ')'"
                  :value="u.ID"
                />
              </el-select>
              <el-select
                v-else-if="field.data_type === 'enum'"
                v-model="attrs[field.field_key]"
                clearable
                style="width: 100%"
              >
                <el-option
                  v-for="opt in parseOptions(field.options_json)"
                  :key="opt"
                  :label="opt"
                  :value="opt"
                />
              </el-select>
              <el-switch
                v-else-if="field.data_type === 'bool'"
                v-model="attrs[field.field_key]"
              />
              <el-date-picker
                v-else-if="field.data_type === 'date'"
                v-model="attrs[field.field_key]"
                type="date"
                value-format="YYYY-MM-DD"
                style="width: 100%"
              />
              <el-date-picker
                v-else-if="field.data_type === 'datetime'"
                v-model="attrs[field.field_key]"
                type="datetime"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
              <el-input-number
                v-else-if="['int', 'decimal'].includes(field.data_type)"
                v-model="attrs[field.field_key]"
                :precision="field.data_type === 'decimal' ? 2 : 0"
                controls-position="right"
                style="width: 100%"
              />
              <el-input
                v-else-if="['text', 'json'].includes(field.data_type)"
                v-model="attrs[field.field_key]"
                type="textarea"
                :rows="field.data_type === 'json' ? 4 : 3"
              />
              <el-input
                v-else
                v-model="attrs[field.field_key]"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, computed } from 'vue'
import { getSchema } from '@/api/pmocker/schema'
import { getCoreFieldKeys, isUserField } from './coreFields'
import { getUserList } from '@/api/user'

const props = defineProps({
  entityType: { type: String, required: true },
  modelValue: { type: Object, default: () => ({}) }
})

const loading = ref(false)
const allFields = ref([])
const userList = ref([])

// attrs 直接引用 modelValue.attrs，利用 Vue 3 reactive 深层响应式
const attrs = computed(() => {
  if (!props.modelValue.attrs) {
    props.modelValue.attrs = {}
  }
  return props.modelValue.attrs
})

// 核心字段列表
const coreFields = computed(() => {
  const coreKeys = getCoreFieldKeys(props.entityType)
  return allFields.value.filter(f =>
    coreKeys.includes(f.field_key) &&
    f.field_key !== 'title' &&
    f.field_key !== 'status'
  )
})

// 扩展字段列表
const extendedFields = computed(() => {
  const coreKeys = getCoreFieldKeys(props.entityType)
  return allFields.value.filter(f =>
    !coreKeys.includes(f.field_key) &&
    f.field_key !== 'title' &&
    f.field_key !== 'status'
  )
})

// 加载用户列表
const loadUsers = async () => {
  try {
    const res = await getUserList({ page: 1, pageSize: 999 })
    if (res.code === 0) {
      userList.value = res.data.list || []
    }
  } catch (e) {
    console.error('loadUsers error:', e)
  }
}

// 人员字段变更时，同步写入 _name 后缀字段（供列表显示用户名）
const onUserFieldChange = (fieldKey, userId) => {
  if (!userId) {
    attrs.value[fieldKey + '_name'] = ''
    return
  }
  const user = userList.value.find(u => u.ID === userId)
  attrs.value[fieldKey + '_name'] = user ? user.nickName : ''
}

// 加载 schema
const loadSchema = async () => {
  if (!props.entityType) return
  loading.value = true
  try {
    const res = await getSchema(props.entityType)
    if (res.code === 0) {
      allFields.value = res.data.fields || []
      // 为每个字段设置默认值
      allFields.value.forEach(f => {
        if (attrs.value[f.field_key] === undefined) {
          if (f.default_value) {
            attrs.value[f.field_key] = parseDefaultValue(f)
          } else if (f.data_type === 'bool') {
            attrs.value[f.field_key] = false
          } else if (['int', 'decimal'].includes(f.data_type)) {
            attrs.value[f.field_key] = 0
          } else if (f.data_type === 'user') {
            attrs.value[f.field_key] = null
          } else {
            attrs.value[f.field_key] = null
          }
        }
        // 人员字段：初始化 _name
        if (isUserField(f.field_key) || f.data_type === 'user') {
          if (attrs.value[f.field_key] && !attrs.value[f.field_key + '_name']) {
            const user = userList.value.find(u => u.ID === attrs.value[f.field_key])
            attrs.value[f.field_key + '_name'] = user ? user.nickName : ''
          }
        }
      })
    }
  } catch (e) {
    console.error('loadSchema error:', e)
  } finally {
    loading.value = false
  }
}

// 栅格宽度：text/json 全宽，其余半宽
const getColSpan = (field) => {
  if (['text', 'json'].includes(field.data_type)) return 24
  return 12
}

// 解析 enum 选项
const parseOptions = (optionsJson) => {
  if (!optionsJson) return []
  try {
    return JSON.parse(optionsJson)
  } catch {
    return []
  }
}

// 解析默认值
const parseDefaultValue = (field) => {
  if (!field.default_value) return null
  switch (field.data_type) {
    case 'bool':
      return field.default_value === 'true' || field.default_value === true
    case 'int':
      return parseInt(field.default_value) || 0
    case 'decimal':
      return parseFloat(field.default_value) || 0
    default:
      return field.default_value
  }
}

onMounted(() => {
  loadUsers().then(() => loadSchema())
})

watch(() => props.entityType, () => {
  loadSchema()
})
</script>
