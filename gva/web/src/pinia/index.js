import { createPinia } from 'pinia'
import { useAppStore } from '@/pinia/modules/app'
import { useThemeStore } from '@/pinia/modules/theme'
import { useUserStore } from '@/pinia/modules/user'
import { useDictionaryStore } from '@/pinia/modules/dictionary'
import { useProjectStore } from '@/pinia/modules/project'

const store = createPinia()

export { store, useAppStore, useThemeStore, useUserStore, useDictionaryStore, useProjectStore }
