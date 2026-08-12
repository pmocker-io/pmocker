#!/usr/bin/env node
/**
 * pmocker 前端风格一致性检查脚本
 *
 * 依据 gva/aiDoc/frontend-backend/frontend-rules.md：
 *   - v-model 一律 defineModel()（禁 props.modelValue + emit('update:modelValue') 老样板）
 *   - 硬编码颜色必须用语义 token（禁 color:#xxx / background:#xxx / border:#xxx）
 *   - 静态内联 style 应换 UnoCSS 原子类（宽度/间距等）
 *   - 图标统一 <svg-icon>，禁 el-icon 与裸 <svg>
 *
 * 用法：
 *   node scripts/check-pmocker-style.mjs          # 只检查 view/pmocker
 *   node scripts/check-pmocker-style.mjs --all    # 检查整个 src/view
 *
 * 退出码：0=通过，1=发现问题
 */
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'

const ROOT = new URL('..', import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, '$1')
const PMOCKER_DIR = join(ROOT, 'src/view/pmocker')
const ALL = process.argv.includes('--all')

const targets = ALL ? [join(ROOT, 'src/view')] : [PMOCKER_DIR]

// 内联 style 的合理例外：ECharts 图表固定高度、动态绑定(:style)不在此列（:style 单独匹配）
const REASONABLE_INLINE = [
  /height:\s*\d+px/,           // echarts 图表容器高度
  /width:\s*\d+px/,            // 固定宽度（少量）
  /margin-left:\s*auto/,       // 组件级微调
]

function collectVueFiles(dir) {
  const out = []
  for (const name of readdirSync(dir)) {
    const p = join(dir, name)
    const st = statSync(p)
    if (st.isDirectory()) out.push(...collectVueFiles(p))
    else if (name.endsWith('.vue')) out.push(p)
  }
  return out
}

function checkFile(file) {
  const rel = file.replace(ROOT + '\\', '').replace(ROOT + '/', '')
  const src = readFileSync(file, 'utf-8')
  const lines = src.split('\n')
  const problems = []

  const at = (idx) => `L${idx + 1}`
  const ctx = (idx) => lines[idx].trim().slice(0, 100)

  // 1. 硬编码颜色
  lines.forEach((ln, i) => {
    const m = ln.match(/(?:color|background|background-color|border[^:]*):\s*#[0-9A-Fa-f]{3,8}/)
    if (m) problems.push(`[硬编码颜色] ${at(i)}: ${ctx(i)}  → 改用语义 token（text-/bg-/border-success/warning/error/primary/muted-foreground）`)
  })

  // 2. 老式 v-model 样板
  lines.forEach((ln, i) => {
    if (ln.includes("emit('update:modelValue')") || ln.includes('emit("update:modelValue")')) {
      problems.push(`[老式v-model] ${at(i)}: ${ctx(i)}  → 改 defineModel()（见 frontend-rules.md）`)
    }
  })

  // 3. el-icon / 裸 svg
  lines.forEach((ln, i) => {
    if (ln.includes('<el-icon')) problems.push(`[el-icon] ${at(i)}: ${ctx(i)}  → 用 <svg-icon>`)
    if (/<svg(?!-icon|\s)/.test(ln)) problems.push(`[裸svg] ${at(i)}: ${ctx(i)}  → 用 <svg-icon>`)
  })

  // 4. 静态内联 style（排除 :style 动态绑定）
  const staticStyleRe = /<[^>]*\sstyle="([^"]*)"/g
  let m
  while ((m = staticStyleRe.exec(src)) !== null) {
    const inline = m[1]
    if (REASONABLE_INLINE.some((r) => r.test(inline))) continue
    const lineIdx = src.slice(0, m.index).split('\n').length - 1
    problems.push(`[内联style] ${at(lineIdx)}: ${ctx(lineIdx)}  → 静态样式换 UnoCSS 原子类（w-full/mt-4/...）`)
  }

  return { rel, problems }
}

let total = 0
let files = 0
for (const dir of targets) {
  if (!exists(dir)) continue
  for (const file of collectVueFiles(dir)) {
    const { rel, problems } = checkFile(file)
    files++
    if (problems.length) {
      console.log(`\n${rel}`)
      for (const p of problems) console.log(`  ${p}`)
      total += problems.length
    }
  }
}

function exists(p) {
  try { statSync(p); return true } catch { return false }
}

console.log(`\n检查完成：${files} 个 .vue，问题 ${total} 处`)
process.exit(total > 0 ? 1 : 0)
