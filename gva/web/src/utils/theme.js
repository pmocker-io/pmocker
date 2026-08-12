// 主题工具：读取运行时 CSS 变量，供 ECharts 等无法直接用原子类的场景取主题色
// 用法：cssVar('--el-color-success') -> '#67c23a'（跟随当前明暗主题）

/**
 * 读取当前生效的 CSS 变量值（跟随明暗/换肤）。
 * @param {string} name CSS 变量名，如 '--el-color-success'
 * @param {string} [fallback] 读取失败时的兜底值
 * @returns {string} 变量计算值（trim 后）
 */
export const cssVar = (name, fallback = '') => {
  if (typeof window === 'undefined') return fallback
  const v = getComputedStyle(document.documentElement).getPropertyValue(name)
  return v ? v.trim() : fallback
}
