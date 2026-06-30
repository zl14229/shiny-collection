# ✨ Shiny Collection - 异色宝可梦狩猎记录系统

记录你在各代 Pokémon 游戏中刷取异色宝可梦（Shiny Pokémon）的完整工具。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端框架 | Gin (Go) |
| ORM | GORM |
| 数据库 | SQLite (本地文件) |
| 前端框架 | Vue 3 + TypeScript |
| 构建工具 | Vite |
| UI 组件库 | Element Plus |
| 状态管理 | Pinia |

## 快速开始

### 后端

```bash
cd backend
go run cmd/server/main.go
```

服务启动在 `http://localhost:8080`，自动建表并填充种子数据。

### 前端

```bash
cd frontend
npm run dev
```

开发服务器启动在 `http://localhost:5173`，自动代理 `/api` 到后端。

## 主要功能

- **📋 狩猎记录管理**：增删改查每一条狩猎记录
- **🎮 多游戏支持**：覆盖所有主系列 Pokémon 游戏版本
- **🔍 多种狩猎方式**：Masuda 孵蛋、大量出现、闪符三明治等 20+ 种方式
- **📊 仪表盘统计**：总闪数、方法分布、月度趋势、各游戏统计
- **🏷️ 标签系统**：为记录添加自定义标签
- **✨ 详细信息**：性格、性别、精灵球、证章等信息记录

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/records | 狩猎记录列表(分页+筛选) |
| POST | /api/v1/records | 新增记录 |
| GET | /api/v1/records/:id | 记录详情 |
| PUT | /api/v1/records/:id | 更新记录 |
| DELETE | /api/v1/records/:id | 删除记录 |
| GET | /api/v1/pokemon | 宝可梦列表(支持搜索) |
| GET | /api/v1/games | 游戏版本列表 |
| GET | /api/v1/methods | 狩猎方式列表 |
| GET | /api/v1/stats/overview | 统计概览 |
| GET | /api/v1/stats/by-game | 按游戏统计 |

## 项目结构

```
shiny-collection/
├── backend/              # Go 后端
│   ├── cmd/server/       # 入口
│   ├── internal/         # 内部包
│   │   ├── config/       # 配置
│   │   ├── handler/      # HTTP 处理器
│   │   ├── middleware/   # 中间件
│   │   ├── model/        # 数据模型
│   │   ├── repository/   # 数据访问
│   │   ├── router/       # 路由
│   │   └── service/      # 业务逻辑
│   ├── pkg/              # 公共包
│   └── seed/             # 种子数据
├── frontend/             # Vue 3 前端
│   └── src/
│       ├── api/          # API 接口
│       ├── layouts/      # 布局
│       ├── router/       # 路由
│       ├── stores/       # 状态管理
│       ├── types/        # TypeScript 类型
│       ├── utils/        # 工具函数
│       └── views/        # 页面
└── .env.example          # 环境变量示例
```

## 🗄️ 宝可梦数据库

系统自带 **全部 1025 只宝可梦**（全国图鉴编号、中英文名、属性）的种子数据，

首次启动后端时自动填充。包含以下世代：

| 世代 | 地区 | 编号范围 |
|------|------|----------|
| 1 | 关都 | #001–#151 |
| 2 | 城都 | #152–#251 |
| 3 | 丰缘 | #252–#386 |
| 4 | 神奥 | #387–#493 |
| 5 | 合众 | #494–#649 |
| 6 | 卡洛斯 | #650–#721 |
| 7 | 阿罗拉 | #722–#809 |
| 8 | 伽勒尔 & 洗翠 | #810–#905 |
| 9 | 帕底亚 | #906–#1025 |

### 在线更新（可选）

```bash
cd shiny-collection/backend
go run cmd/importer/main.go
```

尝试从神奇宝贝百科在线抓取最新数据，如网络不可用则自动使用内置完整数据。
