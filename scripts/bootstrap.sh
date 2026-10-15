#!/bin/sh
# Bootstrap a macOS/Linux developer machine (D-0018). POSIX sh: it must run before make exists.
# Safe to re-run: installed things are only checked. Steps needing sudo are done only in an
# interactive terminal; otherwise the exact commands are printed for the owner (P-0007).
set -eu

MISE_VERSION=2026.10.11
# sha256 of mise-v$MISE_VERSION-<os>-<arch>.tar.gz, copied from the release's SHASUMS256.txt.
MISE_SHA256_linux_x64=a7d0aad3d6782d42675fe0a2732b39bfc356f6cf2b554437f84c597814a2fd85
MISE_SHA256_linux_arm64=7534d514495449e267997410583833080ed0b99a0053f3b26a10691794a597b5
MISE_SHA256_macos_x64=6eaffe02521c700b420ba6930dfbc528951f27c1c9896c827ab6aab7248c75f1
MISE_SHA256_macos_arm64=64592e2978459d91b9dc690352c97d59a8c2b327f586fd201468fa56afe66311

ROOT=$(cd "$(dirname "$0")/.." && pwd)
BIN_DIR="$HOME/.local/bin"
problems=""

say() { printf 'bootstrap: %s\n' "$*"; }
problem() { problems="$problems  - $*
"; }
have() { command -v "$1" >/dev/null 2>&1; }
can_sudo() { have sudo && { [ -t 0 ] || sudo -n true 2>/dev/null; }; }

case "$(uname -s)" in
  Darwin) os=macos ;;
  Linux) os=linux ;;
  *) say "unsupported OS '$(uname -s)': macOS and Linux only (P-0009)"; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=x64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) say "unsupported CPU '$(uname -m)'"; exit 1 ;;
esac
mkdir -p "$BIN_DIR"
case ":$PATH:" in *":$BIN_DIR:"*) ;; *) PATH="$BIN_DIR:$PATH" ;; esac

# 1. make: the OS make, or a user-space copy named make-proj (never "make", P-0009).
make_cmd=""
if have make; then
  make_cmd=make
elif have make-proj; then
  make_cmd=make-proj
elif [ "$os" = macos ]; then
  problem "make is missing: run 'xcode-select --install', then re-run ./scripts/bootstrap.sh"
else
  if can_sudo && have apt-get; then
    say "installing make with apt-get (sudo)"; sudo apt-get install -y make && make_cmd=make
  elif can_sudo && have dnf; then
    say "installing make with dnf (sudo)"; sudo dnf install -y make && make_cmd=make
  elif have apt-get && have dpkg; then
    say "no sudo: installing a user-space copy of GNU Make as $BIN_DIR/make-proj"
    tmp=$(mktemp -d)
    if (cd "$tmp" && apt-get download make >/dev/null 2>&1 && dpkg -x make_*.deb pkg); then
      install -m 755 "$tmp/pkg/usr/bin/make" "$BIN_DIR/make-proj" && make_cmd=make-proj
    fi
    rm -rf "$tmp"
  fi
  [ -n "$make_cmd" ] || problem "make is missing: run 'sudo apt-get install -y make' (Debian/Ubuntu) or 'sudo dnf install -y make' (Fedora/RHEL)"
fi

# 2. mise: pinned version, checksum-verified, installed per user without sudo.
install_mise() {
  eval "want=\$MISE_SHA256_${os}_${arch}"
  have curl || { problem "curl is missing: install it, then re-run"; return 1; }
  tmp=$(mktemp -d)
  url="https://github.com/jdx/mise/releases/download/v$MISE_VERSION/mise-v$MISE_VERSION-$os-$arch.tar.gz"
  say "downloading mise $MISE_VERSION"
  if ! curl -fsSL -o "$tmp/mise.tgz" "$url"; then
    rm -rf "$tmp"; problem "could not download $url"; return 1
  fi
  if have sha256sum; then got=$(sha256sum "$tmp/mise.tgz" | cut -d' ' -f1); else got=$(shasum -a 256 "$tmp/mise.tgz" | cut -d' ' -f1); fi
  if [ "$got" != "$want" ]; then
    rm -rf "$tmp"; problem "mise download checksum mismatch (got $got, want $want): not installed"; return 1
  fi
  if ! { tar -xzf "$tmp/mise.tgz" -C "$tmp" && install -m 755 "$tmp/mise/bin/mise" "$BIN_DIR/mise"; }; then
    rm -rf "$tmp"; problem "could not unpack mise into $BIN_DIR"; return 1
  fi
  rm -rf "$tmp"
}
mise_ok=1
have mise || install_mise || mise_ok=0

# 3. Pinned tools from .tool-versions (Go, Node, OpenTofu, buf, golangci-lint, OpenSpec).
if [ "$mise_ok" = 1 ]; then
  say "installing pinned tools from .tool-versions"
  (cd "$ROOT" && mise install) || { problem "'mise install' failed: see its output above"; mise_ok=0; }
fi

# 4. Docker needs admin rights: check only, never install silently.
if ! have docker; then
  problem "Docker is missing: install Docker Desktop (macOS) or Docker Engine + Compose plugin (Linux): https://docs.docker.com/engine/install/"
elif ! docker info >/dev/null 2>&1; then
  problem "Docker is installed but not running (or your user cannot reach it): start Docker / add your user to the docker group"
elif ! docker compose version >/dev/null 2>&1; then
  problem "Docker Compose v2 is missing: install the docker compose plugin"
fi

# 5. Verify with the same check everyone uses.
if [ -n "$make_cmd" ] && [ "$mise_ok" = 1 ]; then
  (cd "$ROOT" && mise exec -- "$make_cmd" doctor) || problem "'$make_cmd doctor' reported mismatches (see above)"
fi

if [ -n "$problems" ]; then
  printf '\nbootstrap: not ready. Fix these, then re-run ./scripts/bootstrap.sh:\n%s' "$problems"
  exit 1
fi
[ -n "${MISE_SHELL:-}" ] || say "tip: add 'eval \"\$(mise activate zsh)\"' (or bash) to your shell rc to use the pinned tools outside make"
say "ready. Next: $make_cmd check"
