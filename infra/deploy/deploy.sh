#!/bin/sh
# Deploy one environment over SSH. Runs on the CI runner through `make deploy`.
# Usage: deploy.sh <dev|prod> <host> <user> <env-file> <secrets-file> <known-hosts> [ssh-key]
#   env-file      non-secret release values: IMAGE_<NAME>=ghcr.io/...@sha256:..., ENV, DOMAIN, ACME_EMAIL
#   secrets-file  NAME=value lines. Each NAME becomes ~/harness-engineering/secrets/<name in lower case>, mode 0600, and never
#                 goes into an environment variable or the Compose env file
#   known-hosts   host key lines for <host> (from the DEPLOY_KNOWN_HOSTS variable). Strict host key checking:
#                 an unknown or changed key fails the deploy; nothing is trusted on first use
# Remote steps: copy Compose inputs and secrets to ~/harness-engineering → pull digests → up -d --no-build --wait → health → smoke.
set -eu

fail() {
  echo "deploy: $*" >&2
  exit 1
}

if [ "$#" -lt 6 ] || [ "$#" -gt 7 ]; then
  echo "usage: deploy.sh <dev|prod> <host> <user> <env-file> <secrets-file> <known-hosts> [ssh-key]" >&2
  exit 2
fi
env_name=$1
host=$2
user=$3
env_file=$4
secrets_file=$5
known_hosts=$6
ssh_key=${7:-}

case $env_name in
  dev | prod) ;;
  *) echo "deploy: environment must be dev or prod, got '$env_name'" >&2; exit 2 ;;
esac
for f in "$env_file" "$secrets_file" "$known_hosts"; do
  [ -f "$f" ] || fail "file not found: $f"
done
case $known_hosts in
  /*) ;;
  *) known_hosts="$(pwd)/$known_hosts" ;;
esac
# The pinned key must exist for this host, or the connection would fail later with a less clear message.
ssh-keygen -F "$host" -f "$known_hosts" >/dev/null 2>&1 \
  || fail "DEPLOY_KNOWN_HOSTS has no entry for $host: record its host key first (infra/README.md)"

target="$user@$host"
ssh_run() {
  if [ -n "$ssh_key" ]; then
    ssh -i "$ssh_key" -o IdentitiesOnly=yes -o BatchMode=yes -o ConnectTimeout=20 \
      -o StrictHostKeyChecking=yes -o GlobalKnownHostsFile=/dev/null \
      -o "UserKnownHostsFile=$known_hosts" "$target" "$@"
  else
    ssh -o BatchMode=yes -o ConnectTimeout=20 \
      -o StrictHostKeyChecking=yes -o GlobalKnownHostsFile=/dev/null \
      -o "UserKnownHostsFile=$known_hosts" "$target" "$@"
  fi
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
umask 077

# Runtime secrets: one 0600 file per NAME. Values never reach a log; errors name the line number only.
mkdir "$tmp/secrets"
count=0
line_no=0
while IFS= read -r line || [ -n "$line" ]; do
  line_no=$((line_no + 1))
  case $line in
    '' | '#'*) continue ;;
    *=*) ;;
    *) fail "secrets file line $line_no is not NAME=value" ;;
  esac
  name=${line%%=*}
  case $name in
    [A-Z]*) ;;
    *) fail "secrets file line $line_no: name must start with A-Z" ;;
  esac
  case $name in
    *[!A-Z0-9_]*) fail "secrets file line $line_no: name may contain only A-Z, 0-9 and _" ;;
  esac
  file=$(printf '%s' "$name" | tr '[:upper:]' '[:lower:]')
  printf '%s' "${line#*=}" > "$tmp/secrets/$file"
  count=$((count + 1))
done < "$secrets_file"
[ "$count" -gt 0 ] || fail "secrets file has no NAME=value entries"

# Non-secret environment for Compose: release values plus the secrets folder location on the host.
cp "$env_file" "$tmp/.env"
printf 'HARNESS_ENGINEERING_SECRETS_DIR=./secrets\n' >> "$tmp/.env"

# Compose inputs: base, deploy overlay, the environment's non-secret config, and every deployable fragment.
# Fragments are listed explicitly, so stale files on the host are never merged.
sources="infra/compose infra/traefik infra/postgres"
if [ -d "infra/env/$env_name" ]; then
  sources="$sources infra/env/$env_name"
fi
files="-f infra/compose/compose.yaml -f infra/compose/compose.deploy.yaml"
for f in services/*/compose.yaml apps/*/compose.yaml; do
  if [ -f "$f" ]; then
    sources="$sources $f"
    files="$files -f $f"
  fi
done

# Bundle first, then send: a tar failure stops the deploy (a pipe would hide it).
# shellcheck disable=SC2086 # sources is a list of paths
# infra/env/<env>/secrets is never archived: runtime secrets travel only through the secrets file (above).
tar -cf "$tmp/bundle.tar" --exclude='infra/env/*/secrets' $sources || fail "could not archive the Compose inputs"
tar -rf "$tmp/bundle.tar" -C "$tmp" .env secrets || fail "could not add the env file and secrets"

echo "deploy: $env_name -> $target ($count secret files)"
# Remote umask 022 and a traversable ~/harness-engineering: the non-root Traefik reads its config from a bind mount. Secrets keep
# their 0600 and 0700 modes from the archive.
ssh_run 'umask 022 && mkdir -p ~/harness-engineering && chmod 755 ~/harness-engineering && cd ~/harness-engineering && rm -rf infra services apps secrets .env'
ssh_run 'umask 022 && cd ~/harness-engineering && tar -xf -' < "$tmp/bundle.tar"

# Compose runs on the host with the socket group from the host (the socket proxy joins it, see compose.yaml).
# $1 is the Compose subcommand as one shell-quoted string (the remote shell parses it).
remote_compose() {
  # shellcheck disable=SC2086 # files is a list of -f options
  ssh_run "cd ~/harness-engineering && DOCKER_GID=\$(stat -L -c %g /var/run/docker.sock) docker compose --project-directory . --env-file .env $files $1"
}

remote_compose 'pull --ignore-buildable'
# Migrations go here once a service owns a database, before the stack starts:
#   remote_compose 'run --rm --no-deps <service> migrate'
remote_compose 'up -d --no-build --remove-orphans --wait'

# Every service must be running and healthy. The one-shot ACME volume helper must have exited with 0.
status=$(remote_compose "ps --all --format '{{.Service}}|{{.State}}|{{.Health}}|{{.ExitCode}}'")
bad=$(printf '%s\n' "$status" | awk -F'|' '
  $1 == "letsencrypt-init" { if ($2 != "exited" || $4 != "0") print "deploy: one-shot did not finish cleanly: " $0; next }
  $2 != "running" || ($3 != "" && $3 != "healthy") { print "deploy: not healthy: " $0 }')
if [ -n "$bad" ]; then
  printf '%s\n' "$bad" >&2
  exit 1
fi
echo "deploy: all services running and healthy"

# DOMAIN comes from the release env file. ACME may take minutes on the first deploy, so retry. Any status below
# 500 counts: Traefik answered over TLS (404 means no route for / yet).
domain=$(sed -n 's/^DOMAIN=//p' "$env_file" | tail -n 1)
if [ -n "$domain" ]; then
  tries=0
  until code=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 10 "https://$domain/") && [ "$code" -lt 500 ]; do
    tries=$((tries + 1))
    if [ "$tries" -ge 10 ]; then
      fail "smoke test failed for https://$domain/ (last status: ${code:-none})"
    fi
    sleep 15
  done
  echo "deploy: smoke test https://$domain/ answered $code"
else
  echo "deploy: DOMAIN not set, HTTPS smoke test skipped"
fi
