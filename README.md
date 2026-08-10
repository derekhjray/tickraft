# Tickraft

Lightweight all-in-one scheduling, ops, and monitoring tool — Open Source Edition (AGPLv3 + Commercial dual license)

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Build Status](https://github.com/tickraft/tickraft/actions/workflows/ci.yml/badge.svg)](https://github.com/tickraft/tickraft/actions)

**English** | [简体中文](./README.zh-CN.md)

Repository: https://github.com/tickraft/tickraft

## Features

Tickraft is a lightweight, all-in-one scheduling, ops, and monitoring tool that provides an end-to-end pipeline from task scheduling to alert notification:

- **Scheduler**: supports cron / interval / event / one-shot scheduling types, with built-in HTTP / TCP / ICMP / local script executors to flexibly adapt to a variety of ops scenarios.
- **Collector**: dual-mode coverage with active probing (ICMP / TCP / HTTP) and a Webhook passive listener, meeting observability data collection needs.
- **Prism**: evaluates events based on flexible alert rule configuration and delivers real-time notifications via Webhook channels.
- **Asset management**: monitored asset registration, state-machine-driven lifecycle management, and timeout detection that automatically flags anomalies.
- **Dashboard**: a basic overview dashboard that visualizes key scheduling, collection, and alert metrics.
- **Auth & authorization**: a built-in admin user, JWT token authentication, and API key management to secure API access.

## Installation

Tickraft offers three deployment options; choose as needed.

### Download a Binary from Release

Download the latest version for your platform from [GitHub Releases](https://github.com/tickraft/tickraft/releases):

```bash
# Take linux-amd64 as an example
tar -xzf tickraft-ce-*-linux-amd64.tar.gz
./tickraft start --config config.yaml
```

### Build from Source

```bash
git clone https://github.com/tickraft/tickraft.git
cd tickraft/tickraft
make build
./bin/tickraft start --config config.yaml
```

Prerequisites: Go 1.26+ / Node.js 18+ / pnpm 8+.

### Docker Deployment

```bash
docker run -d \
  --name tickraft \
  -p 6153:6153 \
  -v $(pwd)/data:/var/lib/tickraft \
  -v $(pwd)/config.yaml:/etc/tickraft/config.yaml \
  ghcr.io/tickraft/tickraft-ce:latest
```

> Note: The official image is still in preparation; Docker deployment support is coming soon. You can build an image yourself from the [`Dockerfile`](Dockerfile) at the repository root.

### Running

The open-source edition runs in standalone mode, starting the API service, worker engine, alerter, and operator in a single process:

```bash
# Start with a config file
./bin/tickraft start --config config.yaml

# Local development (built-in defaults: SQLite + a development JWT secret)
./bin/tickraft start
```

See the config template at [`configs/config.example.yaml`](configs/config.example.yaml); copy it and adjust as needed:

```bash
cp configs/config.example.yaml config.yaml
```

The config file supports environment variable interpolation (`${VAR}` / `${VAR:-default}`), making it easy to inject sensitive values such as the DSN or JWT secret from the environment. You can validate the config before starting:

```bash
./bin/tickraft config validate -c config.yaml
```

### Development

Frontend dev server:

```bash
cd web && npm install && npm run dev
```

Backend dev:

```bash
go run ./cmd/tickraft start
```

## Architecture Overview

Tickraft consists of three independent modules that communicate via an event bus:

- **scheduler** — task metadata management, scheduled triggering, sharding/dispatch, and dependency orchestration
- **executor** — task execution, categorized by observer/mutator; `Role` is retained only as executor metadata
- **collector** — passive reception of data reports, with validation, aggregation, and persistence

Common base layer:

- **pkg/config** — YAML config loading and environment variable interpolation
- **pkg/pool** — unified goroutine pool management (`pool.Lambda`, `NewWorkerPool`, `NewIOPool`)
- **pkg/alert** — alert evaluation and notification engine (webhook channel)
- **pkg/event** — strongly typed event bus

For the full documentation index, see [docs/README.md](docs/README.md). The architecture design is in [docs/architecture.md](docs/architecture.md), and the deployment guide is in [docs/deployment.md](docs/deployment.md). A Chinese reference translation is available under [docs/zh-CN/](docs/zh-CN/README.md).

## Contributing

We welcome community contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for development environment setup, code standards, and the PR process. Contributors must sign the [CLA](CLA.md) before their first PR.

### Development Environment

- Go 1.26+
- Node.js 18+
- pnpm 8+

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -m 'feat: add my feature'` (following [Conventional Commits](https://www.conventionalcommits.org/))
4. Push the branch: `git push origin feature/my-feature`
5. Open a Pull Request

### Code Standards

- Backend: follow [Effective Go](https://golang.org/doc/effective_go) and the project's `.golangci.yml`
- Frontend: follow the ESLint + Prettier configuration
- Commit messages: follow Conventional Commits

## Version History

See [CHANGELOG.md](CHANGELOG.md).

## Maintainers

Tickraft is maintained by the Auzeka Labs team.

Commercial license inquiries: licensing@tickraft.com

## License

This repository is licensed under a **AGPLv3 + Commercial dual license** model. Users may choose either license:

- **AGPLv3 (default)**: for open-source users. Derivative works and network services (SaaS) must open-source all code under AGPLv3. See [LICENSES/AGPLv3.txt](LICENSES/AGPLv3.txt).
- **Commercial license**: for commercial users (enterprise private deployment, SaaS providers). Signing the commercial license agreement exempts you from all AGPLv3 obligations, allowing closed-source distribution and SaaS offerings. See [LICENSES/COMMERCIAL.txt](LICENSES/COMMERCIAL.txt).

For license selection guidance and the full statement, see [LICENSE](LICENSE). For commercial license inquiries, contact licensing@tickraft.com.
