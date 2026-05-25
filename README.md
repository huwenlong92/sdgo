# sdgo

`sdgo` is the local CLI for creating and running projects from the `sdkitgo` template repository.

The split is intentional:

- `sdkitgo`: template project, used as the source skeleton.
- `sdgo`: scaffold and development tool, used in new project directories.

## Install

From this repository:

```bash
go install ./cmd/sdgo
```

Or from a module path:

```bash
go install github.com/huwenlong/sdgo/cmd/sdgo@latest
```

## Quick Start

Create a project from the local `sdkitgo` template:

```bash
sdgo new demo --module github.com/acme/demo
cd demo
sdgo dev
```

`sdgo new` copies the `sdkitgo` template project and rewrites project identity:

- `go.mod` module path.
- Go imports from `sdkitgo/...` to the new module path.
- `cmd/sdkitgo` to `cmd/<project>`.
- runtime names in Go, YAML, Makefile, and Dockerfile files.
- local cache/runtime folders are skipped.

## Common Commands

```bash
sdgo new demo
sdgo new demo --module github.com/acme/demo
sdgo new demo --source /Users/huwenlong/data/lab/sdkitgo
```

```bash
cd demo
sdgo dev
sdgo run
sdgo run --cmd "go run ./cmd/demo serve api"
sdgo run go run ./cmd/demo serve worker
```

```bash
sdgo gen module user
```

## Documentation

- [Usage](docs/usage.md): command usage and examples.
- [Design](docs/design.md): tool design, generation rules, and maintenance rules.
- [Commands](docs/commands.md): command reference.
