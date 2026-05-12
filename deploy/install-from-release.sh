#!/usr/bin/env bash
set -euo pipefail

GITLAB_BASE_URL="${GITLAB_BASE_URL:-https://gitlab.com}"
GITLAB_PROJECT_ID="${GITLAB_PROJECT_ID:-}"
GITHUB_REPO="${GITHUB_REPO:-}"
GITHUB_BASE_URL="${GITHUB_BASE_URL:-https://github.com}"
VERSION="${VERSION:-}"
TARGET="${TARGET:-}"
RELEASE_URL="${RELEASE_URL:-}"
CHECKSUM_URL="${CHECKSUM_URL:-}"
WORK_DIR="${WORK_DIR:-/tmp/proidentity-release-install}"
SETUP_ARGS=()

usage() {
  cat <<'USAGE'
Usage: install-from-release.sh [download options] -- [setup options]

Downloads a published ProIdentity Mail binary release and runs the full server
setup without compiling on the target server.

Download options:
  --release-url URL          Direct URL to a proidentity-mail_*.tar.gz archive
  --checksum-url URL         Optional URL to SHA256SUMS
  --github-repo OWNER/REPO   GitHub repository for release downloads
  --github-base-url URL      Default: https://github.com
  --gitlab-project-id ID     GitLab numeric project ID for Generic Package Registry downloads
  --gitlab-base-url URL      Default: https://gitlab.com
  --version VERSION          Release tag/version, for example v0.2.0
  --target TARGET            x64, x86, arm, or arm64. Auto-detected when omitted.
  --work-dir PATH            Default: /tmp/proidentity-release-install
  -h, --help

Any argument after -- is passed to proidentity-production-setup.sh.
The release installer is a wrapper: it downloads/extracts the release archive,
then runs proidentity-production-setup.sh, which installs OS packages, creates
service users, configures MariaDB/Postfix/Dovecot/Rspamd/ClamAV/Nginx, writes
systemd units, starts services, and generates missing secrets.

Example:
  bash install-from-release.sh \
    --github-repo Pro-IT-Services/ProIdentity-Mail \
    --version v0.2.0 \
    -- \
    --public-ipv4 203.0.113.10 \
    --mail-hostname mail.example.com \
    --admin-hostname madmin.example.com \
    --webmail-hostname webmail.example.com \
    --tls-mode letsencrypt-dns-cloudflare
USAGE
}

require_root() {
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "install-from-release.sh must be run as root because the bootstrap installs packages and writes system configuration." >&2
    echo "Run it with sudo, for example: sudo /tmp/install-proidentity-mail.sh ..." >&2
    exit 1
  fi
}

arg_value() {
  local current="$1"
  local next="${2:-}"
  if [[ "${current}" == *=* ]]; then
    printf '%s' "${current#*=}"
    return 0
  fi
  if [[ -z "${next}" || "${next}" == --* ]]; then
    echo "Missing value for ${current%%=*}" >&2
    exit 2
  fi
  printf '%s' "${next}"
}

detect_target() {
  local machine
  machine="$(uname -m)"
  case "${machine}" in
    x86_64|amd64) printf 'x64' ;;
    i386|i686) printf 'x86' ;;
    aarch64|arm64) printf 'arm64' ;;
    armv6l|armv7l|armhf|arm) printf 'arm' ;;
    *)
      echo "Unsupported CPU architecture: ${machine}. Use --target x64|x86|arm|arm64." >&2
      exit 2
      ;;
  esac
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --release-url|--release-url=*)
      RELEASE_URL="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --checksum-url|--checksum-url=*)
      CHECKSUM_URL="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --github-repo|--github-repo=*)
      GITHUB_REPO="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --github-base-url|--github-base-url=*)
      GITHUB_BASE_URL="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --gitlab-project-id|--gitlab-project-id=*)
      GITLAB_PROJECT_ID="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --gitlab-base-url|--gitlab-base-url=*)
      GITLAB_BASE_URL="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --version|--version=*)
      VERSION="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --target|--target=*)
      TARGET="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --work-dir|--work-dir=*)
      WORK_DIR="$(arg_value "$1" "${2:-}")"
      [[ "$1" == *=* ]] || shift
      ;;
    --)
      shift
      SETUP_ARGS+=("$@")
      break
      ;;
    *)
      SETUP_ARGS+=("$1")
      ;;
  esac
  shift
done

require_root

TARGET="${TARGET:-$(detect_target)}"
case "${TARGET}" in
  x64|x86|arm|arm64) ;;
  *)
    echo "Unsupported target ${TARGET}; expected x64, x86, arm, or arm64" >&2
    exit 2
    ;;
esac

if [[ -z "${RELEASE_URL}" ]]; then
  if [[ -z "${VERSION}" ]]; then
    echo "Provide --version when using --github-repo or --gitlab-project-id." >&2
    exit 2
  fi
  archive_name="proidentity-mail_${VERSION}_linux_${TARGET}.tar.gz"
  if [[ -n "${GITHUB_REPO}" ]]; then
    release_base="${GITHUB_BASE_URL%/}/${GITHUB_REPO}/releases/download/${VERSION}"
    RELEASE_URL="${release_base}/${archive_name}"
    CHECKSUM_URL="${CHECKSUM_URL:-${release_base}/SHA256SUMS}"
  elif [[ -n "${GITLAB_PROJECT_ID}" ]]; then
    package_base="${GITLAB_BASE_URL%/}/api/v4/projects/${GITLAB_PROJECT_ID}/packages/generic/proidentity-mail/${VERSION}"
    RELEASE_URL="${package_base}/${archive_name}"
    CHECKSUM_URL="${CHECKSUM_URL:-${package_base}/SHA256SUMS}"
  else
    echo "Provide --release-url, --github-repo, or --gitlab-project-id." >&2
    exit 2
  fi
else
  archive_without_query="${RELEASE_URL%%\?*}"
  archive_name="$(basename "${archive_without_query}")"
fi

auth_header=()
if [[ -n "${GITLAB_TOKEN:-}" ]]; then
  auth_header=(-H "PRIVATE-TOKEN: ${GITLAB_TOKEN}")
fi
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  auth_header+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
fi

rm -rf "${WORK_DIR}"
install -d -m 0700 "${WORK_DIR}"

echo "Downloading ${RELEASE_URL}"
curl -fsSL --retry 3 "${auth_header[@]}" "${RELEASE_URL}" -o "${WORK_DIR}/${archive_name}"

if [[ -n "${CHECKSUM_URL}" ]]; then
  if curl -fsSL --retry 3 "${auth_header[@]}" "${CHECKSUM_URL}" -o "${WORK_DIR}/SHA256SUMS"; then
    if grep -F " ${archive_name}" "${WORK_DIR}/SHA256SUMS" > "${WORK_DIR}/SHA256SUMS.selected"; then
      (cd "${WORK_DIR}" && sha256sum -c SHA256SUMS.selected)
    else
      echo "warning: checksum file did not contain ${archive_name}; skipping checksum verification" >&2
    fi
  else
    echo "warning: could not download checksum file; skipping checksum verification" >&2
  fi
fi

install -d -m 0700 "${WORK_DIR}/artifact"
tar -xzf "${WORK_DIR}/${archive_name}" -C "${WORK_DIR}/artifact"

for required in webadmin webmail groupware mailctl apply-mail-config proidentity-production-setup.sh; do
  if [[ ! -x "${WORK_DIR}/artifact/${required}" ]]; then
    echo "release archive is missing executable ${required}" >&2
    exit 1
  fi
done

echo "Running ProIdentity Mail setup from binary release"
echo "The extracted proidentity-production-setup.sh will install packages and configure the full server."
exec bash "${WORK_DIR}/artifact/proidentity-production-setup.sh" --artifact-dir "${WORK_DIR}/artifact" "${SETUP_ARGS[@]}"
