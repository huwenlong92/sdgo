# 更新日志

## 未发布

### 新增

- 新增 `sdgo update` 和 `sdgo upgrade`，用于更新已安装的 CLI。
- `sdgo new` 支持 Git URL 模板源，可通过本机 Git 认证访问私有仓库。
- 新增 `--branch`，用于克隆模板的指定分支或 tag。
- 新增内置模板：
  - `sdkitgo`
  - `sdkitgo-admin-vue`
  - `sdkitgo-portal-vue`
- 新增 `sdgo template list`、`sdgo template info <template>` 和 `sdgo templates`。
- 新增通过 `package.json` 识别 Node/前端模板。
- 新增 `sdgo serve [target]`，用于运行 sdkitgo 的 serve 目标。
- `sdgo run` 和 `sdgo serve` 新增 `--no-watch`。
- 启动时输出实际运行命令和监听模式。

### 变更

- `sdgo new` 支持通过 `--template` 选择具名模板。
- `sdgo new` 未传 `--module` 时，Go module 默认使用项目名。
- `sdgo run <target>` 作为兼容短写，等价于 `go run ./cmd/<project> serve <target>`。
- 文件监听默认从当前项目目录开始，并跳过生成产物、运行时目录、依赖目录、VCS 目录和缓存目录。
- Go 测试文件（`*_test.go`）不再触发热重载重启。
- 模板身份改为从模板项目中识别，不再硬编码 `sdkitgo`。

### 修复

- 移除了源码和文档中的本机固定模板路径。
- 优化私有模板仓库或分支/tag 不存在时的 Git clone 失败提示。
