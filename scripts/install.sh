#!/bin/sh
set -e

# tonrelay installer - downloads and installs the tonrelay CLI binary
# Usage: curl -sSL https://raw.githubusercontent.com/TONresistor/tonrelay/main/scripts/install.sh | sudo sh
# Pin version: VERSION=v0.1.0 curl -sSL ... | sudo sh

REPO="TONresistor/tonrelay"
BINARY="tonrelay"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

log()  { printf '%s\n' "$1"; }
err()  { printf 'Error: %s\n' "$1" >&2; exit 1; }
warn() { printf 'Warning: %s\n' "$1" >&2; }

# download URL to stdout (curl preferred, wget fallback)
fetch() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$1"
    else
        err "curl or wget required"
    fi
}

# download URL to file
fetch_to() {
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$2" "$1"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$2" "$1"
    else
        err "curl or wget required"
    fi
}

main() {
    log "Installing tonrelay..."

    # Check root
    if [ "$(id -u)" -ne 0 ]; then
        err "this installer requires root privileges. Run with: sudo sh"
    fi

    # Detect existing installation
    if command -v "$BINARY" >/dev/null 2>&1; then
        CURRENT=$("$BINARY" version 2>/dev/null | head -1 || echo "unknown")
        log "Existing installation found: $CURRENT"
    fi

    # Detect OS
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    if [ "$OS" != "linux" ]; then
        err "tonrelay only supports Linux (got: $OS)"
    fi

    # Detect architecture
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) err "unsupported architecture: $ARCH" ;;
    esac

    # Temp files with cleanup trap
    TMPFILE=$(mktemp)
    TMPJSON=$(mktemp)
    TMPCHECKSUMS=$(mktemp)
    trap 'rm -f "$TMPFILE" "$TMPJSON" "$TMPCHECKSUMS"' EXIT INT TERM

    # Resolve version
    if [ -n "$VERSION" ]; then
        TAG="$VERSION"
        log "Pinned version: $TAG"
    else
        log "Fetching latest release..."
        RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
        if ! fetch "$RELEASE_URL" > "$TMPJSON" 2>/dev/null; then
            err "failed to fetch release info from $RELEASE_URL"
        fi
        # Detect rate limiting
        if grep -qi "rate limit" "$TMPJSON"; then
            err "GitHub API rate limit exceeded. Try again later or set VERSION=v0.x.x"
        fi
        TAG=$(grep '"tag_name"' "$TMPJSON" | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
        if [ -z "$TAG" ]; then
            err "could not determine latest version from $RELEASE_URL"
        fi
    fi

    # Validate tag format
    case "$TAG" in
        v[0-9]*) ;;
        *) err "invalid version tag format: $TAG (expected v[0-9]...)" ;;
    esac

    log "Version: $TAG"

    # Download binary
    ASSET_NAME="${BINARY}-linux-${ARCH}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET_NAME}"

    log "Downloading ${ASSET_NAME}..."
    if ! fetch_to "$DOWNLOAD_URL" "$TMPFILE"; then
        err "failed to download binary from $DOWNLOAD_URL"
    fi

    # Download and verify checksum
    CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"
    log "Verifying checksum..."
    if ! fetch_to "$CHECKSUMS_URL" "$TMPCHECKSUMS"; then
        err "failed to download checksums from $CHECKSUMS_URL"
    fi

    EXPECTED=$(grep "  ${ASSET_NAME}$" "$TMPCHECKSUMS" | awk '{print $1}')
    if [ -z "$EXPECTED" ]; then
        err "no checksum found for ${ASSET_NAME} in checksums.txt"
    fi

    ACTUAL=$(sha256sum "$TMPFILE" | awk '{print $1}')
    if [ "$ACTUAL" != "$EXPECTED" ]; then
        err "checksum mismatch for ${ASSET_NAME} (expected: $EXPECTED, got: $ACTUAL)"
    fi
    log "Checksum verified."

    # Install binary
    mkdir -p "$INSTALL_DIR"
    install -m 755 "$TMPFILE" "${INSTALL_DIR}/${BINARY}"

    # Check PATH
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *) warn "${INSTALL_DIR} is not in your PATH" ;;
    esac

    log ""
    log "tonrelay ${TAG} installed to ${INSTALL_DIR}/${BINARY}"
    log ""
    log "Next steps:"
    log "  sudo tonrelay install    # Download tunnel-node, create service"
    log "  sudo tonrelay start      # Start the relay"
    log "  tonrelay status          # View relay status"
}

main
