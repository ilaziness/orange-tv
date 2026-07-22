# agents.md — Orange TV 前端 Monorepo AI 编码指南

> 本文件仅保留每次任务都需要的关键信息。详细内容按需查阅相关 `README.md` 与源码，避免重复加载。

## 项目概述

Orange TV 前端为 monorepo，使用 workspaces 管理。

| 类别 | 技术 |
| ------ | ------ |
| 框架 | React 19 + TypeScript 7 |
| 构建 | Vite 8 |
| 路由 | React Router 8 |
| 状态 | Zustand 5 |
| 样式 | Tailwind CSS 4 + CSS 变量 |
| 组件 | shadcn/ui (base-nova) |
| 校验 | Zod（管理端表单） |
| Linter | Oxlint |
| 包管理 | npm workspaces |

目录：

```text
web/
├── apps/client   # 用户端（端口 5173）
├── apps/admin    # 管理端（端口 5174）
└── packages/shared   # 共享类型与 API 工具
```

开发服务器将 `/api` 代理到 `http://127.0.0.1:8080`。

API 前缀：

- 用户端：`/api/client/v1`
- 管理端：`/api/admin/v1`

## npm run 命令说明

所有命令在 `web/` 目录下执行。

### 根目录（Monorepo 级别）

| 命令 | 说明 |
| ------ | ------ |
| `npm run dev:client` | 启动用户端开发服务器（端口 5173） |
| `npm run dev:admin` | 启动管理端开发服务器（端口 5174） |
| `npm run build` | 构建 shared → client → admin（全量） |
| `npm run build:client` | 仅构建用户端 |
| `npm run build:admin` | 仅构建管理端 |
| `npm run lint` | 对 client + admin 执行 Oxlint 检查 |
| `npm run lint:client` | 仅检查用户端 |
| `npm run lint:admin` | 仅检查管理端 |
| `npm run lint:fix` | 对 client + admin 自动修复 lint 问题 |
| `npm run lint:fix:client` | 仅自动修复用户端 |
| `npm run lint:fix:admin` | 仅自动修复管理端 |
| `npm run typecheck` | 对 shared + client + admin 执行 TypeScript 类型检查 |
| `npm run typecheck:client` | 仅类型检查用户端 |
| `npm run typecheck:admin` | 仅类型检查管理端 |

### 子包级别

在各子包目录（`apps/client`、`apps/admin`、`packages/shared`）下可直接运行：

| 命令 | 适用包 | 说明 |
| ------ | ------ | ------ |
| `npm run dev` | client / admin | 启动 Vite 开发服务器 |
| `npm run build` | client / admin | `tsc -b` + `vite build` |
| `npm run build` | shared | `tsc --noEmit` 类型检查（无产物输出） |
| `npm run lint` | client / admin | Oxlint 检查 |
| `npm run lint:fix` | client / admin | Oxlint 自动修复 |
| `npm run typecheck` | client / admin / shared | `tsc -b --noEmit`（shared 为 `tsc --noEmit`） |
| `npm run preview` | client / admin | Vite 本地预览构建产物 |

## 核心规则（始终遵守）

1. **格式化与检查**：`npm run lint` 和 `npm run typecheck` 必须通过；提交前执行 `npm run build`
2. **TypeScript**：禁用隐式 `any`，函数与组件参数必须带类型；优先从 `@orange-tv/shared` 复用类型
3. **组件**：函数组件 + Hooks；优先使用 shadcn/ui 官方组件与项目已有 `@/components/ui`，不重复造基础组件。`components/ui/` 下组件文件使用小写，业务组件与页面使用 PascalCase
4. **状态管理**：全局状态用 Zustand（`src/store/*`），局部状态用 `useState` / `useReducer`
5. **API 调用**：统一通过 `@orange-tv/shared` 的 `apiGet` / `apiPost` / `apiPut` / `apiDelete`；不要在页面内直接写 `fetch`
6. **路径别名**：`@` 指向当前应用 `src`，`@orange-tv/shared` 指向共享包入口；不要使用相对路径跨层引用
7. **样式**：Tailwind CSS 4 工具类优先；主题变量通过 CSS 变量管理，禁止硬编码色值
8. **路由**：React Router v8，路由集中配置在 `src/routes.tsx`；新页面放在 `src/pages/<模块>/` 并按模块分子包
9. **依赖管理**：新增依赖在每个应用或 `packages/shared` 的 `package.json` 中声明；使用 `npm install -w <workspace>`，不要手动改 lock 文件
10. **修改后验证**：`npm run lint && npm run typecheck && npm run build`
11. **外部输入边界验证**：处理 HTTP 响应、表单输入、URL 参数、localStorage 等外部数据时验证边界（空值、类型、范围、越界）；无效输入应给出明确用户提示并阻止继续处理

## 按需查阅

| 主题 | 文件 | 何时阅读 |
| ------ | ------ | ---------- |
| 启动与脚本 | [README.md](README.md) | 首次运行或新增脚本 |
| 后端规范 | [../AGENTS.md](../AGENTS.md) | 涉及 API、errcode、数据库 |
| 用户端配置 | [apps/client/package.json](apps/client/package.json) | 用户端依赖与脚本 |
| 管理端配置 | [apps/admin/package.json](apps/admin/package.json) | 管理端依赖与脚本 |
| 共享包入口 | [packages/shared/src/index.ts](packages/shared/src/index.ts) | API 工具与类型定义 |
