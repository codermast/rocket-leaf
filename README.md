# 🍃 Rocket-Leaf

<p align="center">
  <img src="docs/images/logo.png" alt="Rocket-Leaf Logo" width="128">
</p>

<p align="center">
  <strong>一款轻量、美观、跨平台的 RocketMQ 客户端</strong>
</p>

<p align="center">
  <a href="#功能特性">功能特性</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#技术栈">技术栈</a> •
  <a href="#路线图">路线图</a> •
  <a href="#贡献指南">贡献指南</a>
</p>

---

## 📖 项目简介

**Rocket-Leaf** 是一款基于现代技术栈打造的 RocketMQ 跨平台图形化客户端，旨在为开发者和运维人员提供一个**轻量、美观、高效**的 RocketMQ 管理工具。

与传统的 Web 控制台相比，Rocket-Leaf 作为桌面应用，具有以下优势：

- 🚀 **原生体验** - 无需部署 Web 服务，开箱即用
- 💻 **跨平台** - 支持 Windows、macOS、Linux
- 🔒 **安全可靠** - 本地运行，无需暴露管理端口
- ⚡ **响应迅速** - 原生性能，操作丝滑流畅

## ✨ 功能特性

### 🔗 连接管理
- 多连接配置管理（添加/编辑/删除）
- 快速切换不同 RocketMQ 集群
- 连接状态实时检测
- 连接配置本地 JSON 持久化（重启不丢失）

### 📋 Topic 管理
- Topic 列表展示与搜索过滤
- Topic 详情查看（分区、权限、消费者等）
- Topic 创建与删除
- 消息统计与分析

### 👥 消费者组管理
- 消费者组列表展示
- 消费进度查看与重置
- 消费者客户端详情
- 订阅关系可视化

### 📬 消息管理
- 多维度消息查询（Topic / Key / MessageId）
- 消息详情查看（Header、Body、属性）
- 消息发送测试
- 消息轨迹追踪

### 📊 监控统计
- Broker 集群状态监控
- 生产/消费 TPS 实时图表
- 消息堆积告警
- 系统资源使用情况

## 🛠 技术栈

| 分类          | 技术                                         | 版本         |
| ------------- | -------------------------------------------- | ------------ |
| **桌面框架**  | [Wails](https://wails.io)                    | v3.0.0-alpha |
| **前端框架**  | [Vue.js](https://vuejs.org)                  | ^3.2.45      |
| **UI 组件库** | [Naive UI](https://www.naiveui.com)          | 最新版       |
| **开发语言**  | [TypeScript](https://www.typescriptlang.org) | ^4.9.3       |
| **构建工具**  | [Vite](https://vitejs.dev)                   | ^5.0.0       |
| **后端语言**  | [Go](https://golang.org)                     | 1.25+        |

## 🚀 快速开始

### 环境要求

- **Go** 1.25 或更高版本
- **Node.js** 18.0 或更高版本
- **Wails CLI** v3

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### 克隆项目

```bash
git clone https://github.com/codermast/rocket-leaf.git
cd rocket-leaf
```

### 安装依赖

```bash
# 安装前端依赖
cd frontend && npm install && cd ..

# 下载 Go 依赖
go mod download
```

### 开发模式

```bash
wails3 dev
```

### 连接配置存储位置

连接管理数据会持久化到本地 JSON 文件：

- macOS: `~/Library/Application Support/rocket-leaf/connections.json`
- Linux: `~/.config/rocket-leaf/connections.json`
- Windows: `%AppData%\\rocket-leaf\\connections.json`

### 构建发布版本

```bash
wails3 build
```

构建产物位于 `build/bin/` 目录。

## 📁 项目结构

```
rocket-leaf/
├── build/                  # 构建配置与资源
│   ├── appicon.png         # 应用图标
│   ├── darwin/             # macOS 构建配置
│   ├── windows/            # Windows 构建配置
│   └── linux/              # Linux 构建配置
├── frontend/               # Vue 前端代码
│   ├── src/
│   │   ├── components/     # 组件
│   │   ├── views/          # 页面视图
│   │   ├── stores/         # 状态管理
│   │   ├── services/       # 服务层
│   │   └── App.vue         # 根组件
│   ├── package.json
│   └── vite.config.ts
├── internal/               # Go 内部包
│   ├── rocketmq/           # RocketMQ 操作
│   └── storage/            # 数据持久化
├── main.go                 # 应用入口
├── go.mod
└── README.md
```

## 🗺 路线图

详见 [ROADMAP.md](docs/ROADMAP.md)

| 阶段    | 目标         | 状态     |
| ------- | ------------ | -------- |
| Phase 1 | 基础框架搭建 | 🚧 进行中 |
| Phase 2 | 连接管理     | 📋 规划中 |
| Phase 3 | Topic 管理   | 📋 规划中 |
| Phase 4 | 消费者组管理 | 📋 规划中 |
| Phase 5 | 消息管理     | 📋 规划中 |
| Phase 6 | 监控告警     | 📋 规划中 |

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

### 开发规范

- 前端代码遵循 [Vue Style Guide](https://vuejs.org/style-guide/)
- Go 代码遵循 [Effective Go](https://golang.org/doc/effective_go)
- 提交信息遵循 [Conventional Commits](https://www.conventionalcommits.org/)

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/codermast">CoderMast</a>
</p>
