#!/usr/bin/env bash

set -euo pipefail

# render_formula writes a concrete Homebrew formula using the release version
# and the GoReleaser checksums file for the four supported platform archives.
render_formula() {
  local version="$1"
  local checksums_file="$2"
  local template_file="$3"
  local output_file="$4"

  local darwin_amd64_sha
  local darwin_arm64_sha
  local linux_amd64_sha
  local linux_arm64_sha

  darwin_amd64_sha="$(checksum_for "$checksums_file" "faz_${version}_darwin_amd64.tar.gz")"
  darwin_arm64_sha="$(checksum_for "$checksums_file" "faz_${version}_darwin_arm64.tar.gz")"
  linux_amd64_sha="$(checksum_for "$checksums_file" "faz_${version}_linux_amd64.tar.gz")"
  linux_arm64_sha="$(checksum_for "$checksums_file" "faz_${version}_linux_arm64.tar.gz")"

  sed \
    -e "s/__VERSION__/${version}/g" \
    -e "s/__DARWIN_AMD64_SHA256__/${darwin_amd64_sha}/g" \
    -e "s/__DARWIN_ARM64_SHA256__/${darwin_arm64_sha}/g" \
    -e "s/__LINUX_AMD64_SHA256__/${linux_amd64_sha}/g" \
    -e "s/__LINUX_ARM64_SHA256__/${linux_arm64_sha}/g" \
    "$template_file" >"$output_file"
}

# checksum_for returns the checksum for one release archive and fails if the
# archive is not present in the GoReleaser checksums file.
checksum_for() {
  local checksums_file="$1"
  local archive_name="$2"
  local checksum

  checksum="$(awk -v archive="$archive_name" '$2 == archive { print $1 }' "$checksums_file")"
  if [[ -z "$checksum" ]]; then
    echo "missing checksum for ${archive_name}" >&2
    exit 1
  fi

  printf '%s\n' "$checksum"
}

main() {
  if [[ "$#" -ne 4 ]]; then
    echo "usage: $0 <version> <checksums-file> <template-file> <output-file>" >&2
    exit 1
  fi

  render_formula "$1" "$2" "$3" "$4"
}

main "$@"
