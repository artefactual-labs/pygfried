#!/usr/bin/env bash

set -euo pipefail

readonly default_branch="${DEFAULT_BRANCH:-main}"
readonly package_name="${PACKAGE_NAME:-pygfried}"

fail() {
    echo "$1" >&2
    exit 1
}

write_output() {
    local key="$1"
    local value="$2"
    echo "${key}=${value}" >> "${GITHUB_OUTPUT}"
}

require_env() {
    local name="$1"
    if [[ -z "${!name:-}" ]]; then
        fail "Required environment variable ${name} is not set."
    fi
}

validate_version() {
    if [[ ! "${RAW_VERSION}" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
        fail "Version must use stable X.Y.Z format."
    fi

    package_version="${RAW_VERSION}"
    tag_name="v${package_version}"

    write_output "package_version" "${package_version}"
    write_output "tag_name" "${tag_name}"
}

validate_git_state() {
    git fetch origin "${default_branch}" --tags --force

    if ! git merge-base --is-ancestor HEAD "origin/${default_branch}"; then
        fail "Release ref must be reachable from origin/${default_branch}."
    fi

    release_sha="$(git rev-parse HEAD)"
    write_output "release_sha" "${release_sha}"

    if git rev-parse -q --verify "refs/tags/${tag_name}" >/dev/null; then
        local tag_sha
        tag_sha="$(git rev-list -n 1 "${tag_name}")"
        if [[ "${tag_sha}" != "${release_sha}" ]]; then
            fail "Tag ${tag_name} already exists at ${tag_sha}, expected ${release_sha}."
        fi
        write_output "tag_exists" "true"
    else
        write_output "tag_exists" "false"
    fi
}

validate_pypi_state() {
    local status_code

    status_code="$(
        curl \
            --silent \
            --show-error \
            --output /dev/null \
            --write-out '%{http_code}' \
            "https://pypi.org/pypi/${package_name}/${package_version}/json"
    )"

    case "${status_code}" in
        404)
            write_output "pypi_exists" "false"
            ;;
        200)
            if [[ "${tag_exists}" != "true" ]]; then
                fail "${package_name} ${package_version} already exists on PyPI without a matching git tag in this repository state."
            fi
            write_output "pypi_exists" "true"
            ;;
        *)
            fail "Unexpected PyPI response status: ${status_code}"
            ;;
    esac
}

require_env "RAW_VERSION"
require_env "GITHUB_OUTPUT"

package_version=""
tag_name=""
release_sha=""
tag_exists=""

validate_version
validate_git_state
validate_pypi_state
