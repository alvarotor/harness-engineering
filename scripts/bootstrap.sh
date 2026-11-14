#!/bin/sh
# Report missing developer tools and host-specific setup guidance. This script does not install tools.
set -eu

say() { printf 'bootstrap: %s\n' "$*"; }
have() { command -v "$1" >/dev/null 2>&1; }
missing=0

# Termux runs on the Linux kernel, so identify its userspace before selecting Linux guidance.
case "${PREFIX:-}" in
  */com.termux/files/usr) os=termux ;;
  *)
    if [ -n "${TERMUX_VERSION:-}" ]; then
      os=termux
    else
      case "$(uname -s)" in
        Darwin) os=macos ;;
        Linux) os=linux ;;
        *) say "unsupported OS '$(uname -s)': supported hosts are macOS, Linux, and Termux"; exit 1 ;;
      esac
    fi
    ;;
esac

say "checking tools for $os"

guidance() {
  tool=$1
  if [ "$tool" = make ]; then
    case "$os" in
      macos) say "install make with 'xcode-select --install'" ;;
      linux) say "install make with 'sudo apt-get install -y make' (Debian/Ubuntu) or 'sudo dnf install -y make' (Fedora/RHEL)" ;;
      termux) say "install make with 'pkg install make'" ;;
    esac
  elif [ "$os" = termux ] && [ "$tool" = go ]; then
    say "install Go with 'pkg install golang'"
  elif [ "$os" = termux ] && [ "$tool" = node ]; then
    say "install Node.js with 'pkg install nodejs-lts'"
  else
    say "install $tool using the instructions for $os; see CONTRIBUTING.md#prerequisites"
  fi
}

check_tool() {
  tool=$1
  if have "$tool"; then
    say "ok $tool"
  else
    say "missing $tool"
    guidance "$tool"
    missing=1
  fi
}

for tool in make go node buf golangci-lint openspec; do
  check_tool "$tool"
done

if have docker && docker compose version >/dev/null 2>&1; then
  say "ok docker compose"
else
  say "missing docker compose"
  if [ "$os" = termux ]; then
    say "Docker Compose availability is not guaranteed in Termux; targets that require it cannot run without it"
  else
    say "install Docker Desktop (macOS) or Docker Engine with the Compose plugin (Linux)"
  fi
  missing=1
fi

if [ "$missing" -ne 0 ]; then
  say "install the reported tools yourself, or ask an AI agent to install them (it will ask for your approval first); then run 'make doctor'"
  exit 1
fi

say "all required tools are present. Next: make check"
