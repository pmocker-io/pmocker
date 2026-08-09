# 初始配置管理模块（pmocker_config 插件）

## 基本信息

- 提出日期：2026-08-09
- 当前状态：`active`
- 需求类型：新插件 / 新功能模块
- 优先级：高（M13，v1.0 发布后首个里程碑）
- 需求文件：`aiDoc/memory/business/active/pmocker-config.md`

## 用户原始意图摘要

把当前 PM 各模块涉及的种子数据（字段定义、初始值、字典、状态流转、工作流、业务种子数据）以 Web 页面方式管理（CRUD），快速配置各模块；配置项带状态机（草稿/评审/发布/归档）管理；支持编辑、复制复用、删除；配置完善后更新种子数据并支撑 v1.1 升级。

## 影响范围

- 后端：新增 `gva/server/plugin/pmocker_config/` 插件；现有 `pm_entity_types`/`pm_field_defs`/`pm_workflow_defs`/`pm_relation_types` 表加 `status` 列；新增 `pm_state_defs` 表；loader 灌入标记 published；getSchema/ListEntities 加 published 过滤
- 前端：新增 `gva/web/src/view/pmocker/config/` 子页 + `api/pmocker/config.js`；`statusTransitions.js` 改读 API；复用 VerticalTabLayout
- 文档：新增 spec（`docs/superpowers/specs/`）；README 里程碑更新
- 插件 / 模块：新插件 `pmocker_config`（server + web + pmocker 元数据）

## 涉及对象

- 模块：实体类型、字段定义、字典、状态流转、工作流定义、业务种子数据（6 类管理对象）
- 接口：统一 `/pmocker/config/*`（entityTypes/fields/dictionaries/stateDefs/workflows/seedEntities/export）
- 页面：单菜单「初始配置」多子页（VerticalTabLayout）+ 组织权限入口跳转 gva superAdmin
- 配置：导出 YAML 三件套（schema.yaml/seed.yaml/menu.yaml）到镜像源

## 已确认约束

- 架构：方案 A——EAV 元表直接 CRUD + 导出生成 YAML（复用现有表/loader/动态表单）
- 生效机制：直写 DB，**仅 published 配置生效**（状态机驱动）；动态表单/列表读取时按 published 过滤，提供 `?includeDraft=true` 供配置页预览
- 状态机：所有配置项统一 `draft → reviewing → published → archived`，archived 可恢复 draft；draft 可删除；简化单人流转（无多人审批）
- 复用语义：一键复制为 draft（从任意状态配置复制）
- 组织架构/岗位/角色/用户/权限：入口跳转 gva superAdmin，不重复建设
- 导出产物：YAML 三件套写 `images/pmbok6-hybrid/`，供 rebuild 镜像 → v1.1 upgrade
- 兼容：loader 灌入默认标记 published；statusTransitions.js 保留本地 fallback

## 当前进展

- 2026-08-09：完成 brainstorm，设计全部确认，待写 spec

## 后续待办

- [ ] 写 spec 文档到 `docs/superpowers/specs/2026-08-09-m13-config-manager-design.md` 并提交
- [ ] spec 用户审阅
- [ ] writing-plans 生成实施计划（M13-A 数据模型+状态机+CRUD 后端；M13-B 前端配置页；M13-C 导出+published 过滤+测试）

## 更新规则

- 本文件只承载「初始配置管理模块」这一个功能点
- 同模块其他功能点（如导出工具的增强）新建独立文件
