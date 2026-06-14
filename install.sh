#!/bin/sh
set -e

# Repository configuration
REPO="khrees/veilo"
BINARY_NAME="veilo"

# Local installation directory (default /usr/local/bin)
INSTALL_DIR="/usr/local/bin"

# Help message
show_help() {
  echo "Veilo CLI Installer"
  echo "Usage: install.sh [options]"
  echo ""
  echo "Options:"
  echo "  -d, --dir <dir>    Change directory to install binary (default: $INSTALL_DIR)"
  echo "  -h, --help         Show this help message"
  echo ""
}

# Parse command line options
while [ "$#" -gt 0 ]; do
  case "$1" in
    -d|--dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    -h|--help)
      show_help
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      show_help
      exit 1
      ;;
  esac
done

# Detect OS
OS_UNAME="$(uname -s)"
case "${OS_UNAME}" in
  Linux*)   OS="linux" ;;
  Darwin*)  OS="darwin" ;;
  CYGWIN*|MINGW*|MSYS*) OS="windows" ;;
  *)
    echo "Unsupported OS: ${OS_UNAME}"
    exit 1
    ;;
esac

# Detect Architecture
ARCH_UNAME="$(uname -m)"
case "${ARCH_UNAME}" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: ${ARCH_UNAME}"
    exit 1
    ;;
esac

# Build artifact name matching release naming convention
# e.g., veilo-linux-amd64
ARTIFACT="veilo-${OS}-${ARCH}"
if [ "${OS}" = "windows" ]; then
  ARTIFACT="${ARTIFACT}.exe"
fi

# Construct download URL
# Direct redirect download of the latest release from GitHub
DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ARTIFACT}"

# Download binary to temporary location
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Downloading ${BINARY_NAME} for ${OS}/${ARCH}..."
echo "Source: ${DOWNLOAD_URL}"

if command -v curl >/dev/null 2>&1; then
  curl -sSfL -o "${TMP_DIR}/${BINARY_NAME}" "${DOWNLOAD_URL}"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "${TMP_DIR}/${BINARY_NAME}" "${DOWNLOAD_URL}"
else
  echo "Error: Either 'curl' or 'wget' is required to download the binary."
  exit 1
fi

# Ensure install directory exists
if [ ! -d "${INSTALL_DIR}" ]; then
  echo "Directory ${INSTALL_DIR} does not exist. Creating it..."
  if [ -w "$(dirname "${INSTALL_DIR}")" ] || [ -w "${INSTALL_DIR}" ]; then
    mkdir -p "${INSTALL_DIR}"
  else
    sudo mkdir -p "${INSTALL_DIR}"
  fi
fi

# Make binary executable
chmod +x "${TMP_DIR}/${BINARY_NAME}"

# Install the binary
echo "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
if [ -w "${INSTALL_DIR}" ]; then
  mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
  echo "Write permission to ${INSTALL_DIR} required. Prompting for sudo..."
  sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "Successfully installed ${BINARY_NAME}!"
echo "Run '${BINARY_NAME} --help' to verify the installation."
