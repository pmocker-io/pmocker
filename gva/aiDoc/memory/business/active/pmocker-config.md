# 初始配置管理模块（pmocker_config 插件）

## 基本信息

- 提出日期：2026-08-09
- 当前状态：`active`
- 需求类型：新插件 / 新功能模块
- 优先级：高（M13，v1.0 发布后首个里程碑）
- 需求文件：`aiDoc/memory/business/active/pmocker-config.md`

## 用户原始意图摘要

以 Web 页面方式管理 PM 系统的完整配置。**一条配置记录（配置包）= 实体类型 + 实体的字段 + 初始值(种子数据，含项目) + 状态定义 + 流转规则(含退回)**，聚合为一个整体；类似 gva 自动化代码管理的包管理模式：配置包列表 → 点击 → 可编辑该配置。每条配置记录带**状态管理**（草稿/评审/发布/归档）和**版本管理**（快照+回滚）。EPS 中项目可新增/修改。完整种子数据 = 项目(EPS树，含状态及种子) × 各模块 × 模块字段 × 字段种子 + 模块 × 状态定义 × 流转规则。

## 影响范围

- 后端：`gva/server/plugin/pmocker_config/` 插件**推翻重建**（聚合配置包模型）；新增 `pm_config_packages`/`pm_config_versions` 表；seed_yaml 存储；发布时自动同步到运行表；EPS 项目编辑修复
- 前端：`gva/web/src/view/pmocker/config/` **推翻重建**（配置包列表→点击→编辑详情）；EPS 页面项目新增/修改修复
- 文档：重写 spec + plan；README 里程碑更新
- 插件 / 模块：pmocker_config 插件（server + web + pmocker 元数据）

## 涉及对象

- 模块：**每个模块一个配置包**（需求/进度/风险/问题/变更/交付物/范围/成本/团队），聚合实体/字段/状态/流转/项目种子；**独立 EPS 配置包**（entity_type=eps_node，描述树层级：非底层=抽象容器，叶子=基本单元项目）
- 接口：统一 `/pmocker/config/*`（packages/package/:id/versions/export）
- 页面：配置包列表页 + 配置编辑页（seed_yaml 表单/向导）+ EPS 项目编辑修复
- 配置：seed_yaml（YAML 真源）→ 发布时自动同步 DB

## 已确认约束

- **聚合配置包模型**：一条配置包 = 实体类型 + 字段 + 种子数据(含项目) + 状态定义 + 流转规则，聚合为一个整体，对标 gva autoCode 包管理
- **每模块一包**：需求/进度/风险/问题/变更/交付物/范围/成本/团队各一配置包 + 独立 EPS 配置包
- **EPS 树**：独立 EPS 配置包描述层级（集团→事业部→项目集→项目），非底层=抽象容器，叶子=基本单元项目；业务模块种子通过 project_id 引用 EPS 项目
- **存储**：配置包 seed_yaml 用 **YAML**（与 schema/seed/menu.yaml、loader、.pmi 镜像全链路对齐）
- **同步机制**：**发布时同步**——配置包发布时校验 seed_yaml 并自动写入运行表（pm_entity_types/pm_field_defs/pm_entities 等），编辑保存不写 DB，发布才生效
- **状态机**：配置包 draft → reviewing → published → archived，简化单人流转；默认值恒 published 不再适用（配置包按状态机流转）
- **版本管理**：发布生成不可变快照（pm_config_versions），支持查看历史/回滚
- **推翻现有 M13**：现有 6 类对象各自 CRUD 的 pmocker_config 实现推翻重建；保留可复用：published 过滤、元表 status 列、状态机概念
- 组织架构/岗位/角色/用户/权限：入口跳转 gva superAdmin

## 当前进展

- 2026-08-09：方向修正确认——从「6 类对象各自 CRUD」改为「聚合配置包模型」，推翻重建
- 2026-08-09：EPS 项目编辑存在 bug（前端传参 name/title 不匹配），待修复

## 后续待办

- [ ] 重写 spec（`docs/superpowers/specs/2026-08-09-m13-config-manager-design.md`）
- [ ] 重写 plan（`docs/superpowers/plans/2026-08-09-m13-config-manager.md`）
- [ ] 重建实现：pm_config_packages/versions 表 + seed_yaml + 发布同步 + 配置包编辑前端 + EPS 项目编辑修复
- [ ] 端到端验证 + 版本回滚验证
- [ ] 用户基于配置包更新种子数据 → v1.1 升级

## 更新规则

- 本文件只承载「初始配置管理模块」这一个功能点
- 同模块其他功能点（如导出工具的增强）新建独立文件
