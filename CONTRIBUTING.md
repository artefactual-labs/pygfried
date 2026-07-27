# Contributing

## Prerequisites

- [uv](https://docs.astral.sh/uv/getting-started/installation/)
- [Go](https://go.dev/dl/)

Recent versions are fine. uv and Go can provision or enforce the versions
required by this repo.

## Setup

Configure the project's environment:

```bash
uv sync
```

You can skip the project build:

```bash
uv sync --no-install-project
```

That matters in this repo because installing the project builds the Go-backed
extension. After changing Go sources or dependencies, rebuild the extension
with `uv sync --reinstall-package pygfried` before running the Python tests.
Note that `uv build` only refreshes artifacts under `build/` and `dist/`; it
does not refresh the in-place extension used by `uv run pytest`.

## Common commands

Lint:

```bash
uv run ruff check .
uv run ruff format --check .
```

Tests in Python:

```bash
uv run pytest
```

Tests in Go:

```bash
go test -race .
```

Build:

```bash
uv build --wheel
```

Check artifacts:

```bash
uv run python -m twine check --strict dist/*
```

Smoke test an installed package or built wheel the way a user would:

```bash
./.github/scripts/smoke-install.sh
./.github/scripts/smoke-install.sh dist/pygfried-*.whl
./.github/scripts/smoke-install.sh dist/pygfried-*.whl 3.14
```

The smoke test defaults to the interpreter version in `.python-version`,
installs `pygfried` via `uvx`, runs `identify("README.md", detailed=True)`,
prints the result, and asserts a few stable fields. This is most useful as a
quick pre-release or post-release sanity check in addition to the normal test
suite. Pass an explicit Python version when you need to test a wheel that is
specific to a Python minor version.

## Dependency management

When upgrading Python, Go, or CI inputs, update the files that define support
and the files that lock or exercise those versions together.

### Python

- `.python-version` describes the minimum supported interpreter version.
- [`pyproject.toml`](./pyproject.toml) reflects Python support in
  `requires-python` and the Python classifiers.
- [`uv.lock`](./uv.lock) records the resolved Python dependencies for this repo.
- Development dependencies live in [`pyproject.toml`](./pyproject.toml) and are
  locked in [`uv.lock`](./uv.lock).
- When changing Python support, update `.python-version`,
  [`pyproject.toml`](./pyproject.toml), [`uv.lock`](./uv.lock), and the Python
  build matrix in [`.github/workflows/_build.yml`](./.github/workflows/_build.yml)
  together.

### Go

- [`go.mod`](./go.mod) defines the Go toolchain version.
- [`go.mod`](./go.mod) also defines the Go module dependencies.
- [`go.sum`](./go.sum) records dependency checksums.

#### Upgrade Siegfried

The Siegfried dependency and the bundled Archivematica signature database must
be upgraded together. The default database comes from Siegfried's `pkg/static`
package at compile time. Pygfried separately embeds
`internal/signatures/archivematica.sig`, copied from
`cmd/roy/data/archivematica.sig` in the same module version.

Use this sequence:

```bash
go get github.com/richardlehane/siegfried@vX.Y.Z
go mod tidy
go run ./internal/signatures/cmd/update
go test -race .
uv sync --reinstall-package pygfried
uv run pytest
uv build --wheel
./.github/scripts/smoke-install.sh dist/pygfried-*.whl
```

The update command reads the version selected by `go.mod`, downloads that exact
module, copies its complete Archivematica database, and records its source
version and SHA-256 checksum in `internal/signatures/archivematica.json`. Do not
edit either generated file manually.

The Go tests fail with the update command to run if the version or checksum is
out of sync. Review dependency-update pull requests for this failure: updating
`go.mod` alone is intentionally insufficient.

To verify the checked-in files directly against the selected dependency without
changing them, run:

```bash
go run ./internal/signatures/cmd/update -check
```

Before merging an upgrade, verify both normal PRONOM behavior and the focused
Archivematica parity cases in the Go and Python suites. The race test is part of
the compatibility check because scanner instances are documented as safe for
concurrent use.

### GitHub Actions

- [`.github/workflows/ci.yml`](./.github/workflows/ci.yml) runs CI for `main`
  and pull requests.
- [`.github/workflows/_build.yml`](./.github/workflows/_build.yml) defines the
  shared Python build matrix.
- [`.github/workflows/release.yml`](./.github/workflows/release.yml) defines the
  manual release flow.
- If you change supported Python versions, update the shared build matrix there
  too.

## Release

Run the `Release` workflow from the CLI:

```bash
gh workflow run Release \
  --repo artefactual-labs/pygfried \
  --ref main \
  -f version=0.13.0 \
  -f ref=main
```

Inputs:

- `version`: the stable release version in `X.Y.Z` format.
- `ref`: the git ref or commit SHA to release from. The default is `main`.

Notes:

- Do not include the `v` prefix in `version`. For example, use `0.13.0`, not `v0.13.0`.
- The package published to PyPI uses the exact `version` value, such as `0.13.0`.
- The workflow creates the git tag and GitHub release as `vX.Y.Z`, such as `v0.13.0`.
- `gh workflow run --ref main` selects the branch that provides the workflow definition.
- `-f ref=main` selects the branch, tag, or commit to release.

The workflow builds and tests the release artifacts, publishes them to PyPI, and
only then creates the matching `vX.Y.Z` git tag and GitHub release.
