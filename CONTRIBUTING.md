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
extension. If you only need the development environment first, you can skip
that build and postpone it until a later step such as `uv build`.

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
go test -v .
```

Build:

```bash
uv build --wheel
```

Check artifacts:

```bash
uv run python -m twine check --strict dist/*
```

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
