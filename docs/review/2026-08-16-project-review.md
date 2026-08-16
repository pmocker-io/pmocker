# PMocker 项目审阅报告

> 审阅日期：2026-08-16
> 审阅范围：需求文档 r6、README、aiDoc 记忆层、14 份里程碑 plan、149 条 commit 史、pkg/cli/gva 三层约 1300 个文件的实际代码核对
> 版本基线：v1.0（M1-M13 已交付）

---

## 一、项目初衷与三层定位

PMocker 的原始愿景是一个概念新颖的「**项目管理系统的 Docker**」：把不同方法论（PMBOK/PRINCE2/ISO 21502/敏捷）的 PMS 封装为可分享、可版本化、可**差异升级**的 `.pmi` 镜像，让个人/团队「一条命令启动一套完整项目管理系统」。

核心洞察在于三层抽象（`需求文档.MD` §1.3）：
- **PMI 镜像**（≈ Docker 镜像）：模块清单 + schema + 菜单 + 权限 + seed + 主题，内容寻址
- **PMSystem**（≈ 容器）：独立 gva 进程 + 独立数据卷
- **项目**（≈ 容器内数据）：EAV 驱动的业务数据

技术路线的关键决策是 **EAV 元数据驱动 + gva 插件化**：10 个模块全部作为 gin-vue-admin 标准插件实现，用 EAV 统一数据结构支撑 schema diff 升级。这让「镜像」与「运行时系统」解耦——镜像只是 schema/seed 的载体，业务代码固定，数据流动可变。

## 二、需求演变脉络（Git 历史洞察）

仓库共 **149 个提交，全部集中在 2026-07（9 个）到 2026-08（140 个）**——这是一个约 6 周内从零搭建到 v1.0 的极快节奏。按 commit 主题聚类可清晰看到演变轨迹：

| 阶段 | 主题 | 核心提交 |
|------|------|---------|
| **M1-M8 平台期**（7月底） | 骨架/EAV/插件/镜像/前端/CLI/集成/MCP | 从搭建到 8 内置工具 + 78 动态工具 + 双进程 MCP |
| **M9-M11 深化期**（8月初） | 字段补全、DynamicForm、甘特图/检入检出/变更diff/跨模块联动 | 从「能 CRUD」到「像样子的 PM 系统」 |
| **M12 业务闭环**（8月初） | 基线/偏差/事件引擎/完成度/仪表盘/PMO/结项/任务中心/工作台 | 从 CRUD 到真正的项目管理逻辑（SPI/CPI、健康度 RAG、4类任务聚合） |
| **M13 元层**（8月中） | 初始配置管理：聚合配置包 + 层级可视化编辑器 + 发布同步 + 版本回滚 | **最关键的转向** |

**演变洞察**：项目从一个「基础设施/平台型」产品（Docker 类比、镜像格式、OCI）逐步演进为「单系统产品 + 运行时配置工具」。M13 尤其意味深长——它把「定义系统的那份配置（schema/seed/状态/流转）」从**代码层**（plugin 内 YAML）提升为**产品层**（Web 可视化编辑 + 版本快照 + 发布同步 DB），即用户（非开发者）也能编排一套 PMS。这实际上是走向 v1.1 Hub 的必经之路，也是把「自定义项目管理方法论」从愿景落地为可交付能力的临门一脚。

## 三、需求计划 vs 实现现状核对

### 3.1 里程碑合规表（M1-M13）

| 里程碑 | 计划要求 | 现状 | 判定 |
|------|---------|------|------|
| M1 骨架 | go.work + gva subtree + 空 CLI | ✅ 完成 | 达标 |
| M2 内核 | EAV 引擎 + 三维权限 + 工作流引擎 | ⚠️ EAV/工作流引擎在 gva 侧实现且可运行；**三维权限只实现 ①gva RBAC + ③状态权限，②EPS 节点权限未实现**（`gva/server/service/pmocker/rbac.go:15-17` 自注"M3 阶段补充"） | 部分 |
| M3 模块 | 10 插件（server+web+元数据） | ✅ 11 业务插件齐全，schema 字段与文档基本对齐（team 58 字段符合）；CPM/EVM/Check-in-Out/风险矩阵/影响分析/利用率算法均在 | 达标 |
| M4 镜像 | OCI 打包 + diff/upgrade | ⚠️ OCI 读写与 schema diff 真实可用；**diff 只算 schema，plugins/seed diff 声明未实现**（`pkg/pmocker/diff/diff.go`）；**assets 层（前端预编译层）从未构建**，实际只有 3/4 类层（`images/pmbok6-hybrid/build.go:67-85`） | 部分 |
| M5 CLI 闭环 | 12 命令 + 默认镜像 | ⚠️ 12 命令全部存在且 run/ps/stop/start/rm/logs/inspect 可用；**`commit`/`export` 只是重新打标签复制原始镜像，不抓取实例当前状态**（`cli/cmd/commit.go:42-47`）；**`upgrade` 是空壳，只打印迁移计划，不连接 DB 执行迁移**（`cli/cmd/upgrade.go:47-52`，自注"M4 打印操作，M5 连接数据库执行"） | 部分 |
| M6-M8 | 集成修复 + MCP | ✅ 实例模式自动初始化、静态资源、MCP 双进程、token 签发均工作 | 达标 |
| M9 字段补全 | 10 模块 schema/seed + team | ✅ | 达标 |
| M10 表单适配 | DynamicForm + Schema API | ✅ 全模块复用 DynamicForm | 达标 |
| M11 可视化 | 甘特/检入检出/变更diff/联动 | ⚠️ 主要视图存在，但甘特图实为 **el-table + DynamicForm 而非真图表**；CPM 只支持 FS 依赖（自注简化）；真实 echarts 在 critical/evm/risk-matrix/curve | 部分 |
| M12 业务闭环 | 基线/偏差/事件/完成度/仪表盘/PMO/结项/任务中心 | ✅ 19 个 Service 全实现（baseline/variance/hooks/progress/dashboard/pmo/archive/task_center/workbench） | 达标 |
| M13 配置管理 | 聚合配置包 + 层级编辑器 + 发布同步 + 版本回滚 | ✅ 模型三次纠正（6类CRUD→每模块一包→聚合配置包），9 个边界单测，端到端验证通过 | 达标 |

### 3.2 需求文档宣称 vs 实际的关键偏差（已验证）

1. **"v1.0 MVP 已达成，M1-M13 全部完成"（README）表述偏乐观**——CLI 的镜像闭环（commit/export/upgrade）是文档/CLI 壳而非端到端功能，"镜像可差异升级"这一**最核心的差异化卖点尚未真正可用**。
2. **三维权限名不副实**：实际是「gva RBAC ∧ 实体状态权限」两维。
3. **团队模块的绩效评估流存在必现运行时故障**：`performance_review_flow.yaml:8` 声明 auto 节点 `handler: pmocker.team.score_calc`，但全仓只注册了 3 个 handler（schedule/cost/risk，grep 确认），工作流引擎遇到未注册 handler 直接报错（`workflow.go:209`）→ 绩效评估流程走到评分节点必挂。
4. **`--db/--db-dsn/--volume` 三个 run 参数声明但从未读取**（`cli/cmd/run.go:133-135`），config 恒为 sqlite + 自动卷。
5. **MCP 端口 8899 在 CLI 中硬编码**（`manager.go:133`），多实例并存时 MCP 必然冲突；且与 standalone_manager/http_client 的默认值 8889 不一致（实例通过 config.yaml 显式写 8899 才碰巧可用，属潜在地雷）。

## 四、架构与代码质量评估

### 4.1 优点（值得肯定的地方）

- **架构一致性极高**：12 个插件全部遵循 `schema.yaml → seed.yaml → workflow → menu → api.yaml` 元数据 + EAV 运行时的同一范式；路由中间件链与主系统 PrivateGroup 对齐（`JWTAuth → MustChangePwdGuard → CasbinHandler → DataScope`）；api.yaml 与 router 路径计数逐一吻合（deliverable 15/15、team 24/24 等）。
- **EAV + 专用表混合设计务实**：把树（WBS/EPS）、任务依赖、基线快照、检入检出锁放在专用强结构表，避免 EAV 性能灾难，是正确的工程取舍。
- **插件零硬编码自动发现真正落地**：`build.go` 目录扫描 + `gen_pmocker_register.go` go:generate 生成空白导入，新增插件仅需一条命令，与文档承诺一致。
- **文档纪律极佳**：需求文档 r1→r6 有修订记录、aiDoc 记忆层（长期/业务记忆 + 索引）、每里程碑独立 plan、commit 规范统一且带里程碑标注。这在同类项目里非常少见。
- **测试纪律在后期显著改善**：约 118 个 pmocker 专属测试；M13 配置包发布/回滚补了 9 个边界单测；MCP 有 ~41 个用例；EAV/progress/各模块算法均有单测。
- **无历史包袱**：旧 6 类 CRUD 残留已清干净（grep 无 `6class/legacy` 命中），说明重构是有纪律的。
- 后期前端统一 gva 风格（颜色语义 token + defineModel + UnoCSS 原子类）并沉淀了风格检查脚本，属于持续收口的工程化行为。

### 4.2 质量短板

| 领域 | 问题 | 位置 |
|------|------|------|
| **工作流引擎** | approvers/quorum/SLA/on_complete **全部声明但不执行**；`NodeTask` 与 approval 同处理；**零测试** | `gva/server/service/pmocker/workflow.go` |
| **RBAC** | pkg 层 rbac 只有 3 个类型无求值器；`EntityStatePerm` 定义了没人用；维度②缺失 | `pkg/pmocker/rbac/types.go` |
| **OCI/镜像** | reader 不做 digest 校验（损坏档案静默读取）；`CreatedAt` annotation 构建器从不写入；`AddImage` 吞掉写文件错误 | `pkg/pmocker/oci/reader.go:54-57`, `builder.go`, `image/store.go` |
| **pkg 边界** | `pkg/pmocker/plugin/loader` import gva 的 `model/system`，但 `pkg/go.mod` 未 require → **共享库只在 go.work workspace 内才能编译**，违背「独立共享库」定位 | `loader.go:10-11` |
| **Loader** | `LoadSchema` 丢弃 states/workflows，`LoadSeed` 丢弃 roles（有注释说 M4 扩展，但没做） | `pkg/pmocker/plugin/loader/loader.go:55-79` |
| **测试盲区** | M12 的核心服务 baseline/variance/archive/dashboard/task_center/pmo 全部无测试；workflow 无测试；CLI builder/manager/auto_init 无测试 | — |
| **仓库卫生** | 仓库根残留 `images/index.json`（重复条目）、`instances/instances.db`、`volumes/`；**有未提交 WIP**：`pmocker_config` 的 `import_service.go`/`import_service_test.go` 未跟踪 + `export_service.go` 已修改 | `git status` |
| **构建孤儿** | 前端 dist 内容哈希孤儿文件清理靠 CLI 层拷贝时手动清，机制脆弱但已文档化 | `cli/internal/builder/builder.go:116-146` |

## 五、总体研判

**一句话结论**：这是一个**工程能力和文档纪律远超平均水平的「高完成度 MVP」**——业务功能（10 模块 CRUD + M12 业务闭环 + M13 配置管理）几乎全部落地，架构干净、无历史包袱、测试在关键路径上够用；但它**最独特的卖点——「Docker 式镜像生命周期」——恰恰是完成度最低的部分**，且存在 2 个必现运行时故障（绩效流 handler 缺失）、1 个名不副实的权限维度。

换句话说：**PMocker 目前是一个「很好的单机 PM 产品」+「半成品镜像平台」**，而产品定位却是后者。这是价值主张与现状之间最大的落差。

## 六、后续建议（按优先级）

### P0 · 止血（必修）

1. **修复团队绩效流必现故障**：为 `pmocker.team.score_calc` 注册 handler（或从 workflow 移除 auto 节点改为人工评分），并端到端跑通绩效评估流程。
2. **落实 commit/export/upgrade 三件套**：`commit`/`export` 应把实例数据卷的真实状态（sqlite dump + dist + 业务数据）打包进 `.pmi`；`upgrade` 应真正执行 diff→migration（可先做 sqlite-only 的加列/建表 + 备份回滚），这是"镜像可差异升级"的核心承诺。建议以 M14 立项，明确验收标准。
3. **提交并收尾未完成的 import_service WIP**，避免丢失。
4. **清理仓库根残留**（images/index.json、instances.db、volumes/），纳入 `.gitignore`。

### P1 · 补齐（建议在一个里程碑内完成）

5. **RBAC 维度②（EPS 节点权限）落地或显式降级文档**：要么实现 `eps_member` 节点权限继承判定，要么把需求文档从"三维权限"改为"二维 + 预留"，避免承诺与实现脱节。
6. **工作流引擎语义收口**：对 approvers/quorum/SLA/on_complete 二选一——实现之，或从 schema 移除并在文档标注"仅状态机 + auto 节点"。同时为 WorkflowService 补测试（当前零测试，风险最高）。
7. **修 MCP 多实例端口冲突**：按实例分配端口；统一 8899/8889 默认值。
8. **`run --db/--db-dsn/--volume` 三个死参数**：实现或移除，避免误导。

### P2 · 加固（质量/规模化）

9. **给 M12 核心服务补单测**（baseline/variance/archive/dashboard/task_center/pmo），这些是当前产品力最强的部分，却完全无测试保护。
10. **OCI 完整性**：reader 校验 digest 与 manifest 一致性；builder 写入 created 时间戳；补齐 assets 层或文档化延期。
11. **pkg 解耦**：消除 loader 对 gva 的依赖（把 `model/system` 的依赖改为注入的注册函数），恢复独立共享库定位。
12. **CI 落地**：目前 149 提交全是 main 单分支、无 CI。建议引入 GitHub Actions（go test + vue build + 冒烟），为 v1.1 Hub 和多人协作者铺路。

### P3 · 战略（下一版本方向）

13. **产品重心抉择**：M13 的配置编辑能力 + 待补全的镜像闭环，天然指向 v1.1「Hub = Git 仓库驱动 + pull/push/search」。建议把「镜像生命周期闭环」作为 v1.1 前置条件而不是并行项——否则推送出去的是无法真正重建/升级的镜像。
14. **方法论镜像验证**：用第二套方法论（如纯敏捷/ISO 21502）验证「配置包 + 镜像」能否真正产出不同的 PMS——这是对核心愿景的最佳真实验证。
15. **规模化准备**：默认 admin/123456、100 年 MCP token 等在单机可接受，v1.2 多用户前需引入密码策略、token 轮换、审计落库。

---

> 报告数据来源：需求文档 r6、README、aiDoc 记忆层、14 份里程碑 plan、149 条 commit 史、pkg/cli/gva 三层约 1300 个文件的实际代码核对（含 3 个并行深度探索 + 关键文件点验）。
