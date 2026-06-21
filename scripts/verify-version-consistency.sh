#!/usr/bin/env bash
# verify-version-consistency.sh verifies that the version reported by the
# tickraft binary matches the version baked into the Docker image OCI
# label and the Helm Chart appVersion (when those sources are available).
#
# Missing sources (no Docker image, no Helm chart, no OCI label) are skipped
# and only the available sources are compared against the binary version.
set -euo pipefail

# Resolve repository root (script lives in scripts/).
cd "$(dirname "$0")/.."

BINARY_PATH="bin/tickraft"
DOCKER_IMAGE_NAME="tickraft-ce"
HELM_CHART_PATH="charts/tickraft/Chart.yaml"

# extract_binary_version parses the version string emitted by `tickraft version`.
# Expected output: "tickraft v1.0.0 (commit: ..., built: ..., tags: ..., ...)"
extract_binary_version() {
    local output
    if ! output="$("${BINARY_PATH}" version 2>/dev/null)"; then
        echo "Error: failed to execute '${BINARY_PATH} version'" >&2
        return 1
    fi
    # Match "v<semver>" allowing pre-release suffixes (e.g. v1.0.0-rc1).
    local version
    version="$(printf '%s\n' "${output}" \
        | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+[^ ]*' \
        | head -n 1 \
        | sed 's/^v//')"
    if [ -z "${version}" ]; then
        echo "Error: could not parse version from binary output: ${output}" >&2
        return 1
    fi
    printf '%s\n' "${version}"
}

# extract_docker_version returns the OCI version label for the image, or empty
# string when the image or label is unavailable.
extract_docker_version() {
    local image_tag="${DOCKER_IMAGE_NAME}:${BINARY_VERSION}"
    if ! command -v docker >/dev/null 2>&1; then
        return 0
    fi
    if ! docker image inspect "${image_tag}" >/dev/null 2>&1; then
        return 0
    fi
    # Use the index form so keys containing dots are resolved correctly.
    docker inspect "${image_tag}" \
        --format '{{index .Config.Labels "org.opencontainers.image.version"}}' \
        2>/dev/null || true
}

# extract_helm_version returns the appVersion from Chart.yaml, or empty string
# when the chart is absent.
extract_helm_version() {
    if [ ! -f "${HELM_CHART_PATH}" ]; then
        return 0
    fi
    grep '^appVersion:' "${HELM_CHART_PATH}" \
        | awk '{print $2}' \
        | tr -d '"' \
        | tr -d "'"
}

if [ ! -x "${BINARY_PATH}" ]; then
    echo "Error: binary not found or not executable at ${BINARY_PATH}" >&2
    exit 1
fi

BINARY_VERSION="$(extract_binary_version)" || exit 1

DOCKER_VERSION="$(extract_docker_version)"
HELM_VERSION="$(extract_helm_version)"

# Compare every available source against the binary version (the canonical
# source). Missing sources are skipped.
mismatch=0
if [ -n "${DOCKER_VERSION}" ] && [ "${DOCKER_VERSION}" != "${BINARY_VERSION}" ]; then
    mismatch=1
fi
if [ -n "${HELM_VERSION}" ] && [ "${HELM_VERSION}" != "${BINARY_VERSION}" ]; then
    mismatch=1
fi

if [ "${mismatch}" -eq 1 ]; then
    docker_display="${DOCKER_VERSION:-N/A}"
    helm_display="${HELM_VERSION:-N/A}"
    echo "Version mismatch: binary=${BINARY_VERSION}, docker=${docker_display}, helm=${helm_display}" >&2
    exit 1
fi

echo "Version consistency check passed: ${BINARY_VERSION}"
