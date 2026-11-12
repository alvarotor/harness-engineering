#!/bin/sh
# Generate a backend service from templates/go-service: scripts/new-service.sh <name> [root]
# Changes nothing when the name is invalid or the folder exists. Normally run as `make new-service name=<name>`.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/.." && pwd)

refuse() {
  echo "new-service: $*" >&2
  exit 1
}

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "usage: $0 <name> [root]" >&2
  exit 2
fi
name=$1
root=${2:-$repo}

# Same rule as the spec: ^[a-z][a-z0-9]*(-[a-z0-9]+)*$ (letters, digits, single hyphens between segments).
LC_ALL=C
export LC_ALL
kebab="name '$name' must be kebab-case: a letter first, then a-z, 0-9, with single '-' between segments"
case $name in
  [a-z]*) ;;
  *) refuse "$kebab" ;;
esac
case $name in
  *[!a-z0-9-]*|*--*|*-) refuse "$kebab" ;;
esac
dest="$root/services/$name"
[ ! -e "$dest" ] || refuse "$dest already exists; choose another name"

# Generate in a temporary folder and move it into place, so a failure leaves no half-built service.
mkdir -p "$root/services"
tmp=$(mktemp -d "$root/services/.new-$name.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
cp -R "$repo/templates/go-service/." "$tmp/"
rm -rf "$tmp/coverage.out" "$tmp/coverage" "$tmp/bin"

envname=$(printf '%s' "$name" | tr 'a-z-' 'A-Z_')
for f in $(find "$tmp" -type f); do
  sed -e "s/__ENVNAME__/$envname/g" -e "s/__NAME__/$name/g" "$f" > "$f.tmp"
  mv "$f.tmp" "$f"
done

chmod 755 "$tmp"
[ ! -e "$dest" ] || refuse "$dest appeared while generating; nothing was kept"
mv "$tmp" "$dest"

echo "created $dest from templates/go-service"
cat <<EOF
Next steps:
  1. Add the catalog entry (status: planned) through an approved OpenSpec change (lane L).
     The first product service needs Q1 answered (docs/standards/new-product.md).
  2. make -C $dest test
EOF
