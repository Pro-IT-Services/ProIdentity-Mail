#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-${VERSION:-${CI_COMMIT_TAG:-dev}}}"
DIST_DIR="${DIST_DIR:-${ROOT_DIR}/dist}"
WORK_DIR="${WORK_DIR:-${ROOT_DIR}/build/release}"
COMMIT_SHA="${CI_COMMIT_SHA:-$(git -C "${ROOT_DIR}" rev-parse --short=12 HEAD 2>/dev/null || printf 'unknown')}"
BUILT_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

if [[ -z "${VERSION}" ]]; then
  echo "version is required as first argument or VERSION/CI_COMMIT_TAG" >&2
  exit 2
fi

mkdir -p "${DIST_DIR}" "${WORK_DIR}"
rm -rf "${WORK_DIR:?}"/*

targets=(
  "linux amd64 '' x64"
  "linux 386 '' x86"
  "linux arm 7 arm"
  "linux arm64 '' arm64"
)

commands=(webadmin webmail groupware mailctl)

copy_release_files() {
  local stage="$1"
  install -m 0755 "${ROOT_DIR}/deploy/proidentity-production-setup.sh" "${stage}/proidentity-production-setup.sh"
  install -m 0755 "${ROOT_DIR}/deploy/devmail/apply-mail-config.sh" "${stage}/apply-mail-config"
  install -m 0644 "${ROOT_DIR}/README.md" "${stage}/README.md"
  install -m 0644 "${ROOT_DIR}/LICENSE" "${stage}/LICENSE"
  install -m 0644 "${ROOT_DIR}/COMMERCIAL-LICENSE.md" "${stage}/COMMERCIAL-LICENSE.md"
  cat > "${stage}/RELEASE_MANIFEST.txt" <<EOF
name=proidentity-mail
version=${VERSION}
commit=${COMMIT_SHA}
built_at=${BUILT_AT}
target=${target_label}
EOF
}

write_checksums() {
  local checksum_file="${DIST_DIR}/SHA256SUMS"
  : > "${checksum_file}"
  for archive in "${DIST_DIR}"/*.tar.gz; do
    [[ -e "${archive}" ]] || continue
    if command -v sha256sum >/dev/null 2>&1; then
      (cd "${DIST_DIR}" && sha256sum "$(basename "${archive}")") >> "${checksum_file}"
    else
      (cd "${DIST_DIR}" && shasum -a 256 "$(basename "${archive}")") >> "${checksum_file}"
    fi
  done
}

for target in "${targets[@]}"; do
  read -r goos goarch goarm target_label <<< "${target}"
  stage="${WORK_DIR}/proidentity-mail_${VERSION}_${goos}_${target_label}"
  mkdir -p "${stage}"

  export CGO_ENABLED=0
  export GOOS="${goos}"
  export GOARCH="${goarch}"
  if [[ -n "${goarm}" && "${goarm}" != "''" ]]; then
    export GOARM="${goarm}"
  else
    unset GOARM || true
  fi

  for command_name in "${commands[@]}"; do
    echo "building ${command_name} for ${goos}/${goarch}${GOARM:+/v${GOARM}}"
    (cd "${ROOT_DIR}" && go build -trimpath -buildvcs=false -ldflags "-s -w" -o "${stage}/${command_name}" "./cmd/${command_name}")
  done

  copy_release_files "${stage}"

  archive_name="proidentity-mail_${VERSION}_${goos}_${target_label}.tar.gz"
  tar -C "${stage}" -czf "${DIST_DIR}/${archive_name}" .
done

write_checksums

echo "release artifacts written to ${DIST_DIR}"
