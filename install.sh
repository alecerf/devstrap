#!/bin/sh
# install.sh — install devstrap from GitHub releases.
#
# Usage:
#   curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | sh
#   curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | INSTALL_DIR=~/.local sh
#
# Environment variables:
#   INSTALL_DIR  — base directory (binary goes into $INSTALL_DIR/bin). Default: /usr/local
#   VERSION      — specific version to install (e.g. "1.0.0"). Default: latest

set -eu

REPO="alecerf/devstrap"
GITHUB_API="${GITHUB_API:-https://api.github.com/repos/${REPO}}"
GITHUB_DL="${GITHUB_DL:-https://github.com/${REPO}/releases/download}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local}"
BIN_DIR="${INSTALL_DIR}/bin"

log() { printf '%s\n' "$@"; }
fail() { log "ERROR: $1" >&2; exit 1; }

need_cmd() {
    command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "${ARCH}" in
        aarch64|arm64) ARCH="arm64" ;;
        x86_64|amd64)  ARCH="amd64" ;;
        *)             fail "unsupported architecture: ${ARCH}" ;;
    esac

    case "${OS}" in
        darwin) ;;
        linux)  ;;
        *)      fail "unsupported OS: ${OS}" ;;
    esac
}

fetch_latest_version() {
    curl -sSfL "${GITHUB_API}/releases/latest" |
        grep '"tag_name"' |
        sed -E 's/.*"tag_name":[[:space:]]*"v?([^"]+)".*/\1/'
}

download_and_verify() {
    VERSION="$1"
    ARCHIVE="devstrap_${VERSION}_${OS}_${ARCH}.tar.gz"
    CHECKSUMS="checksums.txt"
    TAG="v${VERSION}"

    TMPDIR="$(mktemp -d)"
    trap 'rm -rf "${TMPDIR}"' EXIT

    log "Downloading devstrap v${VERSION}..."
    curl -sSfL -o "${TMPDIR}/${ARCHIVE}" "${GITHUB_DL}/${TAG}/${ARCHIVE}" ||
        fail "failed to download ${ARCHIVE}"

    curl -sSfL -o "${TMPDIR}/${CHECKSUMS}" "${GITHUB_DL}/${TAG}/${CHECKSUMS}" ||
        fail "failed to download ${CHECKSUMS}"

    log "Verifying checksum..."
    EXPECTED="$(grep "${ARCHIVE}" "${TMPDIR}/${CHECKSUMS}" | awk '{print $1}')"
    [ -z "${EXPECTED}" ] && fail "checksum not found for ${ARCHIVE}"

    ACTUAL="$(${SHA_CMD} "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
    [ "${ACTUAL}" != "${EXPECTED}" ] && fail "checksum mismatch: expected ${EXPECTED}, got ${ACTUAL}"

    log "Extracting..."
    tar -xzf "${TMPDIR}/${ARCHIVE}" -C "${TMPDIR}"

    mkdir -p "${BIN_DIR}"
    install -m 755 "${TMPDIR}/devstrap" "${BIN_DIR}/devstrap"

    log "Installed devstrap v${VERSION} to ${BIN_DIR}/devstrap"
}

main() {
    need_cmd curl
    need_cmd tar

    if command -v shasum >/dev/null 2>&1; then
        SHA_CMD="shasum -a 256"
    elif command -v sha256sum >/dev/null 2>&1; then
        SHA_CMD="sha256sum"
    else
        fail "required command not found: shasum or sha256sum"
    fi

    need_cmd install

    detect_platform

    VERSION="${VERSION:-$(fetch_latest_version)}"
    [ -z "${VERSION}" ] && fail "could not determine latest version"

    download_and_verify "${VERSION}"
}

main
