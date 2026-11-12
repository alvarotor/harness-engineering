#!/bin/sh
# Rename the project slug everywhere: scripts/rename.sh <slug> [root]. Run it once in a new product repository,
# normally as `make rename slug=<slug>`. The current slug is the Compose project `name:` in infra/compose/compose.yaml.
# Replaces the kebab (my-shop), snake (my_shop) and UPPER_SNAKE (MY_SHOP) forms in every text file outside .git.
# shortcut: plain text replace, safe only from a distinctive slug (the template's); renaming a short slug again
# (e.g. "shop") would also hit words like "workshop". Upgrade to word-boundary matching if re-renames are needed.
set -eu

here=$(cd "$(dirname "$0")" && pwd)

refuse() {
  echo "rename: $*" >&2
  exit 1
}

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "usage: $0 <slug> [root]" >&2
  exit 2
fi
new=$1
root=${2:-$(cd "$here/.." && pwd)}

# Same rule as new-service.sh: ^[a-z][a-z0-9]*(-[a-z0-9]+)*$
LC_ALL=C
export LC_ALL
is_kebab() {
  case $1 in [a-z]*) ;; *) return 1 ;; esac
  case $1 in *[!a-z0-9-]*|*--*|*-) return 1 ;; esac
}
is_kebab "$new" || refuse "slug '$new' must be kebab-case: a letter first, then a-z, 0-9, with single '-' between segments"

compose="$root/infra/compose/compose.yaml"
[ -f "$compose" ] || refuse "$compose not found"
old=$(sed -n 's/^name: *//p' "$compose" | head -n 1)
is_kebab "$old" || refuse "the 'name:' in $compose ('$old') is not a kebab-case slug; fix it by hand"
[ "$old" != "$new" ] || refuse "the slug is already '$new'; nothing changed"

old_s=$(printf '%s' "$old" | tr '-' '_')
new_s=$(printf '%s' "$new" | tr '-' '_')
old_u=$(printf '%s' "$old" | tr 'a-z-' 'A-Z_')
new_u=$(printf '%s' "$new" | tr 'a-z-' 'A-Z_')

# List first, so the search never sees the temporary files. -I skips binary files.
list=$(mktemp)
trap 'rm -f "$list"' EXIT
rc=0
grep -rlI --exclude-dir=.git --exclude-dir=node_modules -e "$old" -e "$old_s" -e "$old_u" "$root" > "$list" || rc=$?
[ "$rc" -le 1 ] || refuse "searching $root failed; nothing changed"
# Rewriting through cat keeps each file's mode (scripts stay executable).
while IFS= read -r f; do
  sed -e "s/$old_u/$new_u/g" -e "s/$old_s/$new_s/g" -e "s/$old/$new/g" "$f" > "$f.rename-tmp"
  cat "$f.rename-tmp" > "$f"
  rm -f "$f.rename-tmp"
  echo "renamed: ${f#"$root"/}"
done < "$list"

echo "slug '$old' -> '$new'. Next: make check, then review 'git diff'."
