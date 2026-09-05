# 小橘TV 前端 Monorepo

```text
web/
├── apps/
│   ├── client/   # 用户端（端口 5173）
│   └── admin/    # 管理端（端口 5174）
└── packages/
    └── shared/   # 共享类型与 API 工具
```

## 技术栈

- React + TypeScript + Vite
- React Router + Zustand（已依赖）
- shadcn/ui + Tailwind 可在后续阶段接入

## 启动

```bash
cd web
npm install

# 用户端
npm run dev:client

# 管理端
npm run dev:admin

# 构建
npm run build
```

开发服务器将 `/api` 代理到 `VITE_DEV_PROXY_TARGET`（默认 `http://127.0.0.1:8080`）。

API 前缀约定：

- 用户端：`/api/client/v1`
- 管理端：`/api/admin/v1`

## API 地址配置

client / admin 共用 `web/` 下的 env 文件（见 `.env.example`）。

| 场景 | 做法 |
| ---- | ---- |
| 开发切换后端 | 改 `web/.env.development` 的 `VITE_DEV_PROXY_TARGET`，或设 `VITE_API_BASE_URL` 直连 |
| 打包前写死生产 API | 改 `web/.env.production` 的 `VITE_API_BASE_URL` 后 `npm run build` / `make pack` |
| 打包后改 API | 编辑产物 `config.js` 的 `apiBaseUrl`，刷新即可 |

优先级：`config.js` > `VITE_API_BASE_URL` > 同源 `/api`。

只填 **origin**（如 `https://api.example.com`），不要带 `/api`。留空则同源。非法或带 path 的值会在控制台警告并回退。

发布包对应文件：`web/client/config.js`、`web/admin/config.js`（请限制写权限）。本地覆盖可用 `.env.*.local`。
