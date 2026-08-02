import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useStorage } from '@vueuse/core'

// 当前项目上下文 store：所有 pmocker 业务模块共享选中的项目 ID
export const useProjectStore = defineStore('project', () => {
  // 持久化到 localStorage，避免刷新丢失
  const projectId = useStorage('pmocker_project_id', 0)
  const projectName = useStorage('pmocker_project_name', '')

  const hasProject = computed(() => projectId.value > 0)

  const setProject = (id, name) => {
    projectId.value = Number(id) || 0
    projectName.value = name || ''
  }

  const clearProject = () => {
    projectId.value = 0
    projectName.value = ''
  }

  return {
    projectId,
    projectName,
    hasProject,
    setProject,
    clearProject
  }
})
