#!/usr/bin/env bash
# generate-artifacts-list.sh scans for build artifacts produced by build.sh and
# emits a JSON manifest describing which artifacts exist. The manifest is written
# to bin/artifacts-list.json.
#
# Scanned artifacts:
#   - binary:     bin/tickraft
#   - docker:     tickraft-ce:<version>
#   - helm chart: charts/tickraft
#   - sbom:       bin/tickraft.sbom.json
#   - signature:  bin/tickraft.sig
set -euo pipefail

# Resolve repository root (script lives in scripts/).
cd "$(dirname "$0")/.."

BINARY_PATH="bin/tickraft"
DOCKER_IMAGE_NAME="tickraft-ce"
HELM_CHART_PATH="charts/tickraft"
SBOM_PATH="bin/tickraft.sbom.json"
SIGNATURE_PATH="bin/tickraft.sig"
OUTPUT_PATH="bin/artifacts-list.json"

# file_exists echoes "true" or "false".
file_exists() {
    if [ -e "$1" ]; then
        printf 'true'
    else
        printf 'false'
    fi
}

# file_size echoes the byte size of a file, or 0 when absent (portable across
# macOS and Linux without relying on stat flag differences).
file_size() {
    if [ ! -f "$1" ]; then
        printf '0'
        return
    fi
    wc -c < "$1" | tr -d '[:space:]'
}

# extract_binary_version returns the version reported by the binary, or empty
# string when the binary is missing or the version cannot be parsed.
extract_binary_version() {
    if [ ! -x "${BINARY_PATH}" ]; then
        return 0
    fi
    local output
    output="$("${BINARY_PATH}" version 2>/dev/null || true)"
    if [ -z "${output}" ]; then
        return 0
    fi
    printf '%s\n' "${output}" \
        | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+[^ ]*' \
        | head -n 1 \
        | sed 's/^v//' \
        || true
}

# docker_image_exists echoes "true" when the image:tag is present locally.
docker_image_exists() {
    local tag="$1"
    if [ -z "${tag}" ]; then
        printf 'false'
        return
    fi
    if ! command -v docker >/dev/null 2>&1; then
        printf 'false'
        return
    fi
    if docker image inspect "${DOCKER_IMAGE_NAME}:${tag}" >/dev/null 2>&1; then
        printf 'true'
    else
        printf 'false'
    fi
}

BINARY_EXISTS="$(file_exists "${BINARY_PATH}")"
BINARY_SIZE="$(file_size "${BINARY_PATH}")"
BINARY_VERSION="$(extract_binary_version)"
DOCKER_TAG="${BINARY_VERSION}"
DOCKER_EXISTS="$(docker_image_exists "${DOCKER_TAG}")"
HELM_EXISTS="$(file_exists "${HELM_CHART_PATH}")"
SBOM_EXISTS="$(file_exists "${SBOM_PATH}")"
SIGNATURE_EXISTS="$(file_exists "${SIGNATURE_PATH}")"

mkdir -p bin

# Emit JSON. Prefer jq when available for correctness, otherwise fall back to a
# hand-rolled template (all values are controlled, so no escaping is required).
if command -v jq >/dev/null 2>&1; then
    jq -n \
        --arg bin_path "${BINARY_PATH}" \
        --argjson bin_exists "${BINARY_EXISTS}" \
        --argjson bin_size "${BINARY_SIZE}" \
        --arg img_name "${DOCKER_IMAGE_NAME}" \
        --arg img_tag "${DOCKER_TAG}" \
        --argjson img_exists "${DOCKER_EXISTS}" \
        --arg helm_path "${HELM_CHART_PATH}" \
        --argjson helm_exists "${HELM_EXISTS}" \
        --arg sbom_path "${SBOM_PATH}" \
        --argjson sbom_exists "${SBOM_EXISTS}" \
        --arg sig_path "${SIGNATURE_PATH}" \
        --argjson sig_exists "${SIGNATURE_EXISTS}" \
        '{
            binary:     {path: $bin_path, exists: $bin_exists, size: $bin_size},
            docker_image: {name: $img_name, tag: $img_tag, exists: $img_exists},
            helm_chart: {path: $helm_path, exists: $helm_exists},
            sbom:       {path: $sbom_path, exists: $sbom_exists},
            signature:  {path: $sig_path, exists: $sig_exists}
        }' > "${OUTPUT_PATH}"
else
    printf '{"binary":{"path":"%s","exists":%s,"size":%s},"docker_image":{"name":"%s","tag":"%s","exists":%s},"helm_chart":{"path":"%s","exists":%s},"sbom":{"path":"%s","exists":%s},"signature":{"path":"%s","exists":%s}}\n' \
        "${BINARY_PATH}" "${BINARY_EXISTS}" "${BINARY_SIZE}" \
        "${DOCKER_IMAGE_NAME}" "${DOCKER_TAG}" "${DOCKER_EXISTS}" \
        "${HELM_CHART_PATH}" "${HELM_EXISTS}" \
        "${SBOM_PATH}" "${SBOM_EXISTS}" \
        "${SIGNATURE_PATH}" "${SIGNATURE_EXISTS}" \
        > "${OUTPUT_PATH}"
fi

echo "Artifacts manifest written to ${OUTPUT_PATH}"
