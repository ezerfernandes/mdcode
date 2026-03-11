---
name: dev-workflow
description: Build, test, lint, coverage, doc generation, and snapshot tasks for the mdcode Go project. Use when performing development tasks like running tests, building binaries, linting code, generating docs, or creating release snapshots.
---

# Dev Workflow

This skill covers the standard development tasks for the mdcode project (a Go CLI tool).

## Project Layout

- `go.mod` — Go module definition (`github.com/ezerfernandes/mdcode`)
- `main.go` — entrypoint
- `internal/cmd/` — CLI command implementations (root, extract, update, run, exec, dump, list, filter, walk, help, options)
- `.golangci.yml` — golangci-lint configuration
- `.goreleaser.yaml` — goreleaser release configuration
- `README.md` — project documentation (also used as an mdcode test document)
- `examples/` — example markdown documents
- `docs/` — additional documentation / tutorials

## Tasks

All commands should be run from the project root.

### Lint

Run the static analyzer:

```sh
golangci-lint run
```

### Test

Run tests with race detection and coverage:

```sh
go test -count 1 -race -coverprofile=build/coverage.txt ./...
```

### Coverage

View the test coverage report (requires a prior test run):

```sh
go tool cover -html=build/coverage.txt
```

### Build

Build the executable binary:

```sh
go build -ldflags="-w -s" -o build/mdcode .
```

### Snapshot

Create an executable binary with a snapshot version using goreleaser:

```sh
goreleaser build --snapshot --clean --single-target -o build/mdcode
```

### Doc Generation

Regenerate documentation (CLI reference, example codes):

```sh
go generate
```

### Doc Test

Regenerate docs and run all tests (Go + Node):

```sh
go generate
go test ./...
node --test
```

### Clean

Delete the build directory:

```sh
rm -rf build
```

## Notes

- The `build/` directory is used for build artifacts and coverage output.
- The `xc` task runner can also be used: tasks are defined in the Development section of `README.md`.
- Before committing, always run lint and test.
