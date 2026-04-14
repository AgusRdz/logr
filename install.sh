#!/bin/sh
set -e

REPO="AgusRdz/logr"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux*)  OS="linux" ;;
  Darwin*) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
  *) echo "unsupported OS: $OS" >&2; exit 1 ;;
esac

# Set default install dir
if [ -z "$LOGR_INSTALL_DIR" ]; then
  if [ "$OS" = "windows" ]; then
    INSTALL_DIR="$(cygpath "$LOCALAPPDATA/Programs/logr" 2>/dev/null || echo "$HOME/AppData/Local/Programs/logr")"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
else
  INSTALL_DIR="$LOGR_INSTALL_DIR"
fi

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

EXT=""
[ "$OS" = "windows" ] && EXT=".exe"

BINARY="logr-${OS}-${ARCH}${EXT}"

# Resolve version
if [ -z "$LOGR_VERSION" ]; then
  LOGR_VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//')
fi

[ -z "$LOGR_VERSION" ] && echo "error: failed to determine latest version" >&2 && exit 1

BASE_URL="https://github.com/${REPO}/releases/download/${LOGR_VERSION}"

echo "installing logr ${LOGR_VERSION} (${OS}/${ARCH})..."

# Download to temp dir and verify checksum before installing
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "${BASE_URL}/${BINARY}" -o "${TMPDIR}/logr${EXT}"
curl -fsSL "${BASE_URL}/checksums.txt" -o "${TMPDIR}/checksums.txt"

EXPECTED=$(grep "${BINARY}" "${TMPDIR}/checksums.txt" | awk '{print $1}')
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "${TMPDIR}/logr${EXT}" | awk '{print $1}')
else
  ACTUAL=$(shasum -a 256 "${TMPDIR}/logr${EXT}" | awk '{print $1}')
fi

if [ -z "$EXPECTED" ] || [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "error: checksum mismatch - aborting install" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
cp "${TMPDIR}/logr${EXT}" "${INSTALL_DIR}/logr${EXT}"
chmod +x "${INSTALL_DIR}/logr${EXT}"

echo "installed logr to ${INSTALL_DIR}/logr${EXT}"
echo ""

# Add to PATH if not already present
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    if [ "$OS" = "windows" ]; then
      WIN_DIR=$(cygpath -w "$INSTALL_DIR" 2>/dev/null || echo "$INSTALL_DIR")
      powershell.exe -NoProfile -Command \
        "\$p = [Environment]::GetEnvironmentVariable('Path', 'User'); \
         \$d = '${WIN_DIR}'.TrimEnd('\\'); \
         if ((\$p -split ';' | ForEach-Object { \$_.TrimEnd('\\') }) -notcontains \$d) { \
           [Environment]::SetEnvironmentVariable('Path', \"\$d;\$p\", 'User') }"
      export PATH="${INSTALL_DIR}:$PATH"
    else
      SHELL_NAME="$(basename "${SHELL:-}")"
      case "$SHELL_NAME" in
        zsh)  SHELL_RC="$HOME/.zshrc" ;;
        bash) SHELL_RC="$HOME/.bashrc" ;;
        *)    SHELL_RC="" ;;
      esac
      PATH_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""
      if [ -n "$SHELL_RC" ] && ! grep -qF "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
        printf '\n# logr\n%s\n' "$PATH_LINE" >> "$SHELL_RC"
        echo "added ${INSTALL_DIR} to PATH in ${SHELL_RC}"
        echo "reload your shell: source ${SHELL_RC}"
      fi
    fi
    ;;
esac

echo ""
echo "quick start:"
echo "  cat app.log | logr"
echo "  logr app.log --level error --since 30m"
echo "  logr --follow app.log --level error"
