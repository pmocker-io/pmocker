# 实例模式前端资源同步（dist 更新流程）

> 适用场景：修改了 `gva/web` 前端代码后，需要让运行中的 PMocker 实例（PMSystem）加载最新前端。

## 核心机制（重要，避免踩坑）

PMocker 实例的前端资源加载链路：

```
gva/web 源码 → npm run build → gva/web/dist
                                    │
                        (pmocker run 时) copyFrontendDist
                                    ▼
              .pmocker-data/bin/dist  →  复制到  数据卷 <volumeID>/dist
                                    │
                        gva-server 进程 cwd = 数据卷，服务 ./dist
                                    ▼
                          浏览器访问 http://localhost:8080
```

关键点：

1. **`pmocker run`（新建实例）** 会执行 `copyFrontendDist`，把 `.pmocker-data/bin/dist` 复制到新数据卷
2. **`pmocker stop` + `pmocker start`（重启已停止实例）** **不会**重新复制 dist——`Start` 只重启进程，数据卷内的 dist 保持旧版
3. **gva-server 的 cwd 是数据卷目录**（不是 `.pmocker-data/bin/`），所以它服务的是**数据卷内的 dist**，不是 bin 同级 dist

## 症状

改了前端代码 → 重建 dist → `stop`+`start` 重启实例 → 浏览器仍显示旧页面。

典型表现：
- 新页面/新组件访问**空白**或 **404**
- 浏览器 console 报 `[asyncRouter] 未找到组件: view/xxx/yyy.vue，已使用占位组件代替`
- 页面引用的 JS chunk 名与 `gva/web/dist` 最新产物不一致（如浏览器加载 `user.xxx.js` 而 dist 里是 `user.yyy.js`）

## 正确更新流程

### 方式一：强制重建（推荐，最干净）

```bash
cd gva/web && npm run build        # 1. 构建最新前端
# 2. 停止并重建实例（--force 删除同名，--rebuild 重建二进制+dist）
pmocker run -i images/pmbok6-hybrid/pmbok6-hybrid.pmi -p 8080 -n <name> --force --rebuild
```

> `--rebuild` 会重建 gva-server/gva-mcp 二进制和前端 dist；`--force` 删除同名实例（含旧数据卷）。

### 方式二：手动同步数据卷（保留数据）

```bash
cd gva/web && npm run build
# 找到实例数据卷（pmocker inspect <name> 查看 volumeID）
# 覆盖数据卷内的 dist
Remove-Item -Recurse -Force .pmocker-data/volumes/<volumeID>/dist
Copy-Item gva/web/dist .pmocker-data/volumes/<volumeID>/dist -Recurse
# 重启实例
pmocker stop <name> && pmocker start <name>
```

> 同时建议同步 `.pmocker-data/bin/dist` 保持一致（下次 run 用它做源）。

### 方式三：开发模式（无需实例）

```bash
make run-gva-web    # Vite dev server，热更新，不走实例
```

## 排障速查

| 现象 | 原因 | 处理 |
|------|------|------|
| 新页面空白/404 | 数据卷 dist 是旧版 | 按「正确更新流程」同步 dist |
| console 报"未找到组件" | 路由动态 import 的组件不在已加载 dist 中 | 同上 |
| 页面 JS chunk 名对不上 | 浏览器缓存或 dist 未更新 | 硬刷新（Ctrl+Shift+R）或清缓存；确认数据卷 dist 时间戳 |
| lucide 图标 CDN 超时 | `api.unisvg.com`/`iconify` 外网不可达 | 网络问题，与代码无关；可检查图标改用本地 symbol |
