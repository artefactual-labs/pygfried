#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

usage() {
    echo "Usage: $0 [package-spec] [python-version] [target-file] [expected-id]" >&2
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    usage
    exit 0
fi

if [[ "$#" -gt 4 ]]; then
    usage
    exit 1
fi

PACKAGE_SPEC_RAW="${1:-pygfried}"
PYTHON_VERSION="${2:-$(tr -d '[:space:]' < "${REPO_ROOT}/.python-version")}"
TARGET_FILE="${3:-README.md}"
EXPECTED_ID="${4:-fmt/1149}"

if [[ -e "${PACKAGE_SPEC_RAW}" ]]; then
    PACKAGE_SPEC="$(cd "$(dirname "${PACKAGE_SPEC_RAW}")" && pwd)/$(basename "${PACKAGE_SPEC_RAW}")"
else
    PACKAGE_SPEC="${PACKAGE_SPEC_RAW}"
fi

cd "${REPO_ROOT}"

if [[ ! -f "${TARGET_FILE}" ]]; then
    echo "Smoke test target does not exist: ${TARGET_FILE}" >&2
    exit 1
fi

echo "Smoke testing ${PACKAGE_SPEC} with Python ${PYTHON_VERSION} against ${TARGET_FILE}"

uvx \
    --isolated \
    --python "${PYTHON_VERSION}" \
    --with "${PACKAGE_SPEC}" \
    --with rich \
    python -c '
from pygfried import identify
from rich.pretty import pprint

target_file = "'"${TARGET_FILE}"'"
expected_id = "'"${EXPECTED_ID}"'"

result = identify(target_file, detailed=True)
pprint(result)

files = result["files"]
assert len(files) == 1, files

file0 = files[0]
assert file0["filename"] == target_file, file0
assert file0["errors"] == "", file0

matches = file0["matches"]
assert matches, file0
assert matches[0]["id"] == expected_id, matches
'
