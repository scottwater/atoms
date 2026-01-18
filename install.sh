#!/bin/bash
set -euo pipefail

REPO="scottwater/atoms"
BINARY="atom"
INSTALL_DIR="${HOME}/.local/bin"

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       echo "unsupported" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)             echo "unsupported" ;;
  esac
}

main() {
  local os=$(detect_os)
  local arch=$(detect_arch)

  if [[ "$os" == "unsupported" ]] || [[ "$arch" == "unsupported" ]]; then
    echo "Error: Unsupported OS or architecture"
    exit 1
  fi

  echo "Detecting system: ${os}/${arch}"

  local latest_url="https://api.github.com/repos/${REPO}/releases/latest"
  local version=$(curl -sSL "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

  if [[ -z "$version" ]]; then
    echo "Error: Could not determine latest version"
    exit 1
  fi

  echo "Latest version: ${version}"

  local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY}_${os}_${arch}.tar.gz"
  local tmp_dir=$(mktemp -d)
  trap "rm -rf ${tmp_dir}" EXIT

  echo "Downloading ${download_url}..."
  curl -sSL "$download_url" -o "${tmp_dir}/${BINARY}.tar.gz"

  echo "Extracting..."
  tar -xzf "${tmp_dir}/${BINARY}.tar.gz" -C "${tmp_dir}"

  mkdir -p "${INSTALL_DIR}"
  mv "${tmp_dir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"

  echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

  if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
    echo ""
    echo "Note: ${INSTALL_DIR} is not in your PATH."
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\${HOME}/.local/bin:\${PATH}\""
  fi

  echo ""
  echo "Run 'atom --version' to verify installation."
}

main
