# sdgo

`sdgo` 是用于创建和开发 `sdkitgo` 项目的本地 CLI。

它有两类能力：

- 通用脚手架：通过模板快速创建 Go 或前端项目。
- sdkitgo 开发工具：运行 Go 项目、热重载。

## 安装

从 GitHub 安装：

```bash
go install github.com/huwenlong92/sdgo/cmd/sdgo@latest
```

本仓库开发安装：

```bash
go install ./cmd/sdgo
```

验证：

```bash
sdgo version
sdgo --help
sdgo template list
```

## 创建项目

创建默认 Go 后端项目：

```bash
sdgo new demo
cd demo
sdgo dev
```

默认模板是 `sdkitgo`。如果本地没有模板，会从内置源拉取：

```text
git@gitee.com:sd0/sdkitgo.git
```

指定模板：

```bash
sdgo new admin-demo --template sdkitgo-admin-vue
sdgo new portal-demo --template sdkitgo-portal-vue
```

查看内置模板：

```bash
sdgo template list
sdgo template info sdkitgo
```

指定本地或 Git 模板源：

```bash
sdgo new demo --source ../sdkitgo
sdgo new demo --source git@gitee.com:sd0/sdkitgo.git
sdgo new demo --source git@gitee.com:sd0/sdkitgo.git --branch dev
```

Go 模板可以指定 module path；不传时默认使用项目名：

```bash
sdgo new api-demo --module mycorp/api-demo
```

前端模板不使用 `--module`，只会改写 `package.json` 的 `name`。

## 运行项目

默认开发运行，带热重载：

```bash
sdgo dev
sdgo run
```

运行 sdkitgo serve 目标：

```bash
sdgo serve api
sdgo serve worker
```

兼容短写：

```bash
sdgo run api
sdgo run worker
```

`api`、`worker` 这类目标来自项目内的 `cmd/<target>/main.go`，新增入口后可以直接用同名目标运行。

完整自定义命令：

```bash
sdgo run --cmd "go run ./cmd/api -c configs/config.yaml"
sdgo run go run ./cmd/api -c configs/config.yaml
```

默认监听当前项目目录，并跳过依赖、构建产物、运行时目录、缓存目录和 `*_test.go`。

关闭监听，适合进程管理器或线上风格运行：

```bash
sdgo serve api --no-watch
```

手动限制监听目录：

```bash
sdgo dev --watch app,configs,command
```

## 升级 CLI

```bash
sdgo upgrade
sdgo upgrade v0.2.0
```

## 命令补全

`completion` 不会显示在 `sdgo --help` 里，但可以用来安装 shell 补全。

zsh：

```bash
mkdir -p ~/.zsh/completions
sdgo completion zsh > ~/.zsh/completions/_sdgo
```

然后确保 `~/.zshrc` 里有：

```bash
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit
compinit
```

重开终端后即可补全 `sdgo` 命令和参数。

## 内置模板

```text
sdkitgo             go    git@gitee.com:sd0/sdkitgo.git
sdkitgo-admin-vue   node  git@gitee.com:sd0/admin.sdkitgo.cn.git
sdkitgo-portal-vue  node  git@gitee.com:sd0/portal.sdkit.cn.git
```

## 文档

- [使用说明](docs/usage.md)
- [命令参考](docs/commands.md)
- [设计说明](docs/design.md)
- [更新日志](CHANGELOG.md)
