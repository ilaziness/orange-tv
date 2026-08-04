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

开发服务器将 `/api` 代理到 `http://127.0.0.1:8080`。

API 前缀约定：

- 用户端：`/api/client/v1`
- 管理端：`/api/admin/v1`
