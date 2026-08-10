#!/usr/bin/env bash
set -euo pipefail
echo "Building tickraft..."
cd "$(dirname "$0")/.."
VERSION="${VERSION:-dev}"
GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo none)"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
BUILD_TAGS="${BUILD_TAGS:-}"
LDFLAGS="-X github.com/tickraft/tickraft/cmd/tickraft.version=${VERSION} -X github.com/tickraft/tickraft/cmd/tickraft.gitCommit=${GIT_COMMIT} -X github.com/tickraft/tickraft/cmd/tickraft.buildTime=${BUILD_TIME} -X github.com/tickraft/tickraft/cmd/tickraft.buildTags=${BUILD_TAGS}"
# Build frontend
cd web && npm ci && npm run build && cd ..
# Stage frontend dist into pkg/web/dist for go:embed
rm -rf pkg/web/dist
cp -r web/dist pkg/web/dist
# Build binary
go build -ldflags "${LDFLAGS}" -o bin/tickraft ./cmd/tickraft
echo "Build complete: bin/tickraft"
