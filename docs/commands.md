# Commands

## sdgo version

Prints the CLI version.

## sdgo new <project>

Creates a runnable project from the sdkitgo template repository.

Options:

- `--module`: Go module path. Only applies to Go templates.
- `--source`: template source directory or Git URL.
- `--template`: template name used for default source lookup. Defaults to `sdkitgo`.
- `--branch`: Git branch or tag to clone when `--source` is a Git URL.
- `--force`: overwrite existing files.

## sdgo run

Runs the current Go project with the built-in hot reload runner.

Aliases:

- `sdgo dev`

Examples:

```bash
sdgo run
sdgo run api
sdgo run worker
sdgo run --cmd "go run ./cmd/demo serve custom"
```

With one short argument, `sdgo run <target>` is treated as `go run ./cmd/<project> serve <target>`.

Options:

- `--cmd`: command to execute.
- `--watch`: comma-separated watch paths.
- `--no-watch`: run without watching files.

## sdgo serve [target]

Runs an sdkitgo `serve` target with built-in hot reload.

Examples:

```bash
sdgo serve
sdgo serve api
sdgo serve worker
```

Options:

- `--watch`: comma-separated watch paths.
- `--no-watch`: run without watching files.

## sdgo template

Shows available project templates.

Aliases:

- `sdgo templates`

## sdgo template list

Lists built-in project templates.

## sdgo template info <template>

Shows details for a built-in project template, including type, environment override variable, and built-in source.

## sdgo upgrade [version]

Upgrades the `sdgo` CLI itself by running `go install`.

Examples:

```bash
sdgo upgrade
sdgo upgrade v0.2.0
sdgo upgrade latest
```

Options:

- `--target`: Go install target. Defaults to `github.com/huwenlong92/sdgo/cmd/sdgo`.

## sdgo completion [shell]

Generates shell completion scripts. This command is hidden from normal help output.

Supported shells:

- `bash`
- `zsh`
- `fish`
- `powershell`
