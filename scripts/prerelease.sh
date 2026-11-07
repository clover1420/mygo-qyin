#!/usr/bin/env bash
# Trims a draft release to what people download: each platform's installer
# and portable archive. The signed update manifests stay, so installed apps
# can still update; everything else mygo build uploaded is removed.
# A missing installer or archive fails the job before the draft is published.
set -euo pipefail

tag=${1:-${GITHUB_REF_NAME:-}}
repo=${GITHUB_REPOSITORY:-}
if [[ -z $tag ]]; then
  echo "usage: prerelease.sh <tag>" >&2
  exit 2
fi
ver=${tag#v}

assets() {
  if [[ -n ${PRERELEASE_ASSETS_FILE:-} ]]; then
    cat "$PRERELEASE_ASSETS_FILE"
  else
    gh release view "$tag" --repo "$repo" --json assets --jq '.assets[].name'
  fi
}

keep() {
  case "$1" in
    update-*.json) return 0 ;;
    yunhe-setup-*.exe) return 0 ;;
    yunhe-*.dmg) return 0 ;;
    "yunhe_${ver}_"*.deb) return 0 ;;
    "yunhe-$ver-"*.tar.gz) return 0 ;;
    *) return 1 ;;
  esac
}

required=(
  "yunhe-setup-$ver-amd64.exe" "yunhe-setup-$ver-arm64.exe"
  "yunhe-$ver-amd64.dmg" "yunhe-$ver-arm64.dmg"
  "yunhe_${ver}_amd64.deb" "yunhe_${ver}_arm64.deb"
  "yunhe-$ver-windows-amd64.tar.gz" "yunhe-$ver-windows-arm64.tar.gz"
  "yunhe-$ver-darwin-amd64.tar.gz" "yunhe-$ver-darwin-arm64.tar.gz"
  "yunhe-$ver-linux-amd64.tar.gz" "yunhe-$ver-linux-arm64.tar.gz"
)

mapfile -t found < <(assets)
missing=()
for want in "${required[@]}"; do
  if ! printf '%s\n' "${found[@]}" | grep -Fxq -- "$want"; then
    missing+=("$want")
  fi
done
if ((${#missing[@]})); then
  printf 'missing release assets:\n' >&2
  printf '  %s\n' "${missing[@]}" >&2
  exit 1
fi

remove=()
for name in "${found[@]}"; do
  if ! keep "$name"; then
    remove+=("$name")
  fi
done

if ((${#remove[@]})) && [[ -z ${PRERELEASE_DRY_RUN:-} ]]; then
  gh release delete-assets "$tag" "${remove[@]}" --repo "$repo"
fi

printf 'kept %d assets, removed %d\n' "$((${#found[@]} - ${#remove[@]}))" "${#remove[@]}"
