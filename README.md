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
go install github.com/huwenlong92/sdgo/cmd/sdgo@latest
```

## Quick Start

Create a project from the local `sdkitgo` template:

```bash
sdgo new demo
cd demo
sdgo dev
```

`sdgo new` copies the `sdkitgo` template project and rewrites project identity:

- `go.mod` module path.
- Go imports from `sdkitgo/...` to the new module path.
- `cmd/sdkitgo` to `cmd/<project>`.
- runtime names in Go, YAML, Makefile, and Dockerfile files.
- local cache/runtime folders are skipped.

If no local `sdkitgo` template is found, `sdgo new` falls back to `git@gitee.com:sd0/sdkitgo.git`.

## Common Commands

```bash
sdgo new demo
sdgo new demo --source ../sdkitgo
sdgo new demo --source git@gitee.com:sd0/sdkitgo.git
sdgo new demo --source git@gitee.com:sd0/sdkitgo.git --branch dev
sdgo new admin-demo --template sdkitgo-admin-vue
sdgo new portal-demo --template sdkitgo-portal-vue
sdgo new api-demo --module mycorp/api-demo
sdgo template list
sdgo template info sdkitgo
```

```bash
cd demo
sdgo dev
sdgo run
sdgo serve api
sdgo serve worker
sdgo run api
sdgo run worker
sdgo serve api --no-watch
sdgo run --cmd "go run ./cmd/demo serve custom"
```

```bash
sdgo gen module user
```

```bash
sdgo update
sdgo upgrade v0.2.0
```

## Documentation

- [Usage](docs/usage.md): command usage and examples.
- [Design](docs/design.md): tool design, generation rules, and maintenance rules.
- [Commands](docs/commands.md): command reference.
- [更新日志](CHANGELOG.md): 重要变更记录。
