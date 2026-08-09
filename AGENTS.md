# PMocker AGENTS.md

## 目的

本文档是 opencode 在本仓库的自动加载入口，职责是：**把所有 pmocker / gva 代码操作约束到 `gva/AGENT.MD` 与 `gva/aiDoc/` 规范**，并把高频约束摘要在此，避免每次会话重复全量读取。

> 规则唯一真源：`gva/AGENT.MD`（冲突以它为准）。细节规范在 `gva/aiDoc/`。

## 加载顺序

1. `gva/AGENT.MD`（唯一真源）
2. `gva/aiDoc/README.md`（索引）
3. 按任务用 `ai-doc` skill 按需加载对应子目录细则：
   - `aiDoc/relations/`（仓库结构、技术栈、开发流程、commit 规范）
   - `aiDoc/modules/`（后端分层、插件开发、模块职责）
   - `aiDoc/frontend-backend/`（前后端契约、前端规则、工具复用、主题类名）
   - `aiDoc/examples/`（各层书写标准，新建文件先看对应示例）
   - `aiDoc/memory/`（长期记忆 / 业务记忆维护）
4. 涉及新增/修改功能点时，遵循 `gva/AGENT.MD` 的记忆规则维护 `aiDoc/memory/business/`

## 高频约束（摘要，细节以 aiDoc 为准）

### 后端（Go / Gin / GORM）
- 分层：`Router → API → Service → Model`，禁止跨层调用；`enter.go` 作组装入口
- Service 首参 `ctx context.Context`，用 `global.GVA_DB.WithContext(ctx)`；不依赖 `gin.Context`；分页统一 `info.LimitOffset()`（MaxPageSize=100）
- 数据权限由 GORM 全局回调自动处理，Service 不手写 `dept_id`/`created_by` 过滤、不手动盖章
- 全量写用 `Omit("dept_id","created_by")` 保护归属列；更新护栏禁止漏 where 的 update/delete
- 每个对外 API 必须写完整 Swagger，`@Success` 落具体类型（`response.PageResult{list=[]Model}` / `[]Model` / `Model`），不用空 `object`；`@Security ApiKeyAuth` 只写私有分组
- 统一响应 `{code,data,msg}`；统一分页 `{page,pageSize,total,list}`
- 测试复用 `server/internal/testutil`（`NewMemoryDB`/`InitMemoryCache`/`InitNopLogger`/`NewRedisOrSkip`），不另起内联样板

### 插件（gin-vue-admin 插件化）
- 后端插件 `server/plugin/<name>/`，前端 `web/src/plugin/<name>/`
- 插件私有路由组中间件链必须与主系统 PrivateGroup 对齐：`JWTAuth → MustChangePwdGuard → CasbinHandler → DataScope`
- `plugin.go` 实现 v2 接口 `interfaces.Plugin`，`init()` 中 `interfaces.Register(Plugin)` 自注册

### 前端（Vue 3 / Element Plus / UnoCSS）
- HTTP 请求统一 `@/utils/request`；全局状态统一 Pinia；文件名 `kebab-case`、组件名 `PascalCase`
- v-model 一律 `defineModel()`（禁 `props.modelValue + emits('update:modelValue')` 老样板）
- 样式优先 UnoCSS 原子类；只有 `:deep()`/伪类/复杂选择器才写 `<style scoped>`；避免内联 `style`
- 图标统一 `<svg-icon>`；业务/系统图标用 lucide 等，不手写裸 `<svg><path>`
- 复用优先：`src/utils/`（request/date/format/dictionary/stringFun/bus/btnAuth 等），严禁重复造轮子
- 跨栈边界变更时同步更新 `aiDoc/frontend-backend/`

### 其他
- 浏览器点测遵循 `aiDoc/frontend-backend/page-click-testing.md`（token 从 `.local/gva-test-token` 静默读取，不写入任何提交文件）
- **改了前端代码后要让实例生效**，遵循 `aiDoc/frontend-backend/instance-frontend-sync.md`：`pmocker run --force --rebuild` 重建实例，或手动同步数据卷 `<volumeID>/dist`；`pmocker stop`+`start` 不会刷新数据卷内 dist，会导致新页面空白/404
- 不直接读取 `node_modules/` 代码
- commit 规范：`type(scope): description`（type: feat/fix/docs/style/refactor/test/chore），描述用中文，如 `feat(pmocker): xxx`

## AI 工具使用提示

- 代码语义检索优先用 **codegraph MCP**（`codegraph_search`/`codegraph_context`/`codegraph_callers`/`codegraph_impact`），再配合 Read/Grep 确认细节
- 创造性工作（新功能/组件/流程）先走 `brainstorming` skill；多步骤任务先 `writing-plans`；改 bug 先 `systematic-debugging`；写代码前先 `test-driven-development`
- 本仓库业务技术背景见 `README.md`（PMocker：项目管理系统的 Docker，EAV + 10 模块插件 + MCP + .pmi 镜像）
