# Tickraft

轻量化一体化调度运维监控工具 — 开源版（AGPLv3 + 商业授权双协议）

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Build Status](https://github.com/tickraft/tickraft/actions/workflows/ci.yml/badge.svg)](https://github.com/tickraft/tickraft/actions)

[English](./README.md) | **简体中文**

仓库地址：https://github.com/tickraft/tickraft

## 功能特性

Tickraft 是一个轻量化、一体化调度运维监控工具，提供从任务调度到告警通知的完整链路能力：

- **调度引擎（scheduler）**：支持 cron / interval / event / one-shot 调度类型，内置 HTTP / TCP / ICMP / 本地脚本执行器，灵活适配多种运维场景。
- **采集引擎（collector）**：主动探测（ICMP / TCP / HTTP）与 Webhook 被动监听双模覆盖，满足可观测性数据采集需求。
- **告警引擎（prism）**：基于灵活告警规则配置进行事件评估，通过 Webhook 渠道实时通知。
- **资产管理**：监控资产注册、状态机驱动生命周期管理、超时检测自动标记异常。
- **仪表盘**：基础总览大盘，可视化呈现调度、采集与告警关键指标。
- **认证与授权**：内置 admin 用户、JWT 令牌鉴权、API 密钥管理，保障接口访问安全。

## 安装指南

Tickraft 提供三种部署方式，可按需选择。

### 从 Release 下载二进制

从 [GitHub Releases](https://github.com/tickraft/tickraft/releases) 下载对应平台最新版本：

```bash
# 以 linux-amd64 为例
tar -xzf tickraft-ce-*-linux-amd64.tar.gz
./tickraft start --config config.yaml
```

### 从源码构建

```bash
git clone https://github.com/tickraft/tickraft.git
cd tickraft/tickraft
make build
./bin/tickraft start --config config.yaml
```

前置要求：Go 1.26+ / Node.js 18+ / pnpm 8+。

### Docker 部署

```bash
docker run -d \
  --name tickraft \
  -p 6153:6153 \
  -v $(pwd)/data:/var/lib/tickraft \
  -v $(pwd)/config.yaml:/etc/tickraft/config.yaml \
  ghcr.io/tickraft/tickraft-ce:latest
```

> 注：官方镜像尚在筹备中，Docker 部署即将支持。可基于仓库根目录 [`Dockerfile`](Dockerfile) 自行构建镜像。

### 运行

开源版以 standalone 模式运行，单个进程启动 API 服务、worker 引擎、alerter 与 operator：

```bash
# 使用配置文件启动
./bin/tickraft start --config config.yaml

# 本地开发（使用内置默认配置，SQLite + 开发用 JWT 密钥）
./bin/tickraft start
```

配置文件模板参见 [`configs/config.example.yaml`](configs/config.example.yaml)，拷贝后按需修改：

```bash
cp configs/config.example.yaml config.yaml
```

配置文件支持环境变量插值（`${VAR}` / `${VAR:-default}`），便于将 DSN、JWT 密钥等敏感信息从环境注入。启动前可校验配置：

```bash
./bin/tickraft config validate -c config.yaml
```

### 开发

前端开发服务器：

```bash
cd web && npm install && npm run dev
```

后端开发：

```bash
go run ./cmd/tickraft start
```

## 架构概览

Tickraft 由三个独立模块通过 event bus 通信组成：

- **scheduler**（调度引擎）— 任务元数据管理、定时触发、分片分发、依赖编排
- **executor**（执行引擎）— 任务执行，按 observer/mutator 分类，`Role` 仅作为执行器元数据保留
- **collector**（采集引擎）— 被动接收数据上报，校验、聚合、持久化

公共基础层：

- **pkg/config** — YAML 配置加载与环境变量插值
- **pkg/pool** — 统一 goroutine 池管理（`pool.Lambda`、`NewWorkerPool`、`NewIOPool`）
- **pkg/alert** — 告警评估与通知引擎（webhook 渠道）
- **pkg/event** — 强类型事件总线

完整文档索引参见 [docs/README.md](docs/README.md)，架构设计参见 [docs/architecture.md](docs/architecture.md)，部署指南参见 [docs/deployment.md](docs/deployment.md)。中文参考翻译位于 [docs/zh-CN/](docs/zh-CN/README.md)。

## 贡献指南

我们欢迎社区贡献！请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解开发环境搭建、代码规范与 PR 流程。贡献者在首次 PR 前需签署 [CLA](CLA.md)。

### 开发环境

- Go 1.26+
- Node.js 18+
- pnpm 8+

### 开发流程

1. Fork 仓库
2. 创建特性分支：`git checkout -b feature/my-feature`
3. 提交变更：`git commit -m 'feat: add my feature'`（遵循 [Conventional Commits](https://www.conventionalcommits.org/)）
4. 推送分支：`git push origin feature/my-feature`
5. 提交 Pull Request

### 代码规范

- 后端：遵循 [Effective Go](https://golang.org/doc/effective_go) 与项目 `.golangci.yml`
- 前端：遵循 ESLint + Prettier 配置
- 提交信息：遵循 Conventional Commits

## 版本历史

详见 [CHANGELOG.md](CHANGELOG.md)。

## 维护者

Tickraft 由 Auzeka Labs 团队维护。

商业授权咨询：licensing@tickraft.com

## 许可证

本仓库采用 **AGPLv3 + 商业授权双协议** 模式，使用者可在两种协议中任选其一：

- **AGPLv3（默认）**：适用于开源使用者。衍生作品与网络服务（SaaS）必须以 AGPLv3 开源全部代码。详见 [LICENSES/AGPLv3.txt](LICENSES/AGPLv3.txt)。
- **商业授权**：适用于商业使用者（企业私有化部署、SaaS 服务商）。签署商业授权协议后豁免 AGPLv3 全部义务，可闭源分发与 SaaS 化。详见 [LICENSES/COMMERCIAL.txt](LICENSES/COMMERCIAL.txt)。

协议选择指引与完整声明详见 [LICENSE](LICENSE)。商业授权咨询请联系 licensing@tickraft.com。
