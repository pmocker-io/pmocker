<template>
  <div v-loading="loading">
    <!-- 核心字段区 -->
    <el-divider v-if="coreFields.length > 0" content-position="left">基本信息</el-divider>
    <el-row :gutter="16">
      <el-col v-for="field in coreFields" :key="field.field_key" :span="getColSpan(field)">
        <el-form-item :label="field.field_label" :prop="'attrs.' + field.field_key">
          <component
            :is="getComponent(field)"
            v-model="attrs[field.field_key]"
            v-bind="getComponentProps(field)"
            style="width: 100%"
          >
            <template v-if="field.data_type === 'enum'">
              <el-option
                v-for="opt in parseOptions(field.options_json)"
                :key="opt"
                :label="opt"
                :value="opt"
              />
            </template>
          </component>
        </el-form-item>
      </el-col>
    </el-row>

    <!-- 扩展属性区 -->
    <el-collapse v-if="extendedFields.length > 0">
      <el-collapse-item :title="`扩展属性 (${extendedFields.length})`" name="extended">
        <el-row :gutter="16">
          <el-col v-for="field in extendedFields" :key="field.field_key" :span="getColSpan(field)">
            <el-form-item :label="field.field_label" :prop="'attrs.' + field.field_key">
              <component
                :is="getComponent(field)"
                v-model="attrs[field.field_key]"
                v-bind="getComponentProps(field)"
                style="width: 100%"
              >
                <template v-if="field.data_type === 'enum'">
                  <el-option
                    v-for="opt in parseOptions(field.options_json)"
                    :key="opt"
                    :label="opt"
                    :value="opt"
                  />
                </template>
              </component>
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
import { getCoreFieldKeys } from './coreFields'

const props = defineProps({
  entityType: { type: String, required: true },
  modelValue: { type: Object, default: () => ({}) }
})

const emit = defineEmits(['update:modelValue'])
const loading = ref(false)
const allFields = ref([])

// attrs 引用，直接操作 modelValue.attrs
const attrs = computed({
  get: () => props.modelValue.attrs || {},
  set: (val) => emit('update:modelValue', { ...props.modelValue, attrs: val })
})

// 确保 attrs 初始化
const ensureAttrs = () => {
  if (!props.modelValue.attrs) {
    emit('update:modelValue', { ...props.modelValue, attrs: {} })
  }
}

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

// 加载 schema
const loadSchema = async () => {
  if (!props.entityType) return
  loading.value = true
  try {
    const res = await getSchema(props.entityType)
    if (res.code === 0) {
      allFields.value = res.data.fields || []
      // 为每个字段设置默认值
      ensureAttrs()
      const currentAttrs = props.modelValue.attrs || {}
      allFields.value.forEach(f => {
        if (currentAttrs[f.field_key] === undefined) {
          if (f.default_value) {
            currentAttrs[f.field_key] = parseDefaultValue(f)
          } else if (f.data_type === 'bool') {
            currentAttrs[f.field_key] = false
          } else if (['int', 'decimal'].includes(f.data_type)) {
            currentAttrs[f.field_key] = 0
          } else {
            currentAttrs[f.field_key] = null
          }
        }
      })
      emit('update:modelValue', { ...props.modelValue, attrs: currentAttrs })
    }
  } catch (e) {
    console.error('loadSchema error:', e)
  } finally {
    loading.value = false
  }
}

// 组件类型映射
const getComponent = (field) => {
  const map = {
    string: 'el-input',
    text: 'el-input',
    int: 'el-input-number',
    decimal: 'el-input-number',
    date: 'el-date-picker',
    datetime: 'el-date-picker',
    bool: 'el-switch',
    enum: 'el-select',
    ref: 'el-input',
    json: 'el-input'
  }
  return map[field.data_type] || 'el-input'
}

// 组件属性映射
const getComponentProps = (field) => {
  const props = {}
  switch (field.data_type) {
    case 'text':
      props.type = 'textarea'
      props.rows = 3
      break
    case 'int':
      props.controlsPosition = 'right'
      break
    case 'decimal':
      props.precision = 2
      props.controlsPosition = 'right'
      break
    case 'date':
      props.type = 'date'
      props.valueFormat = 'YYYY-MM-DD'
      break
    case 'datetime':
      props.type = 'datetime'
      props.valueFormat = 'YYYY-MM-DD HH:mm:ss'
      break
    case 'json':
      props.type = 'textarea'
      props.rows = 4
      break
  }
  return props
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
  loadSchema()
})

watch(() => props.entityType, () => {
  loadSchema()
})
</script>
