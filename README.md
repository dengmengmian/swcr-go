# swcr-go：计算机软件著作权程序鉴别材料生成器（Go 版）

[![CI](https://github.com/dengmengmian/swcr-go/actions/workflows/ci.yml/badge.svg)](https://github.com/dengmengmian/swcr-go/actions/workflows/ci.yml)

`swcr-go` 是 [swcr](https://github.com/kenley2021/swcr)（作者 kenley，MIT 协议）的 Go 语言重写版。它与原版行为对齐：递归扫描源码目录，按后缀名筛出代码文件，跳过隐藏文件和被排除的路径，逐行剔除空行和注释行，把剩余的代码行写入一个 `.docx` 文档。页眉居中显示「软件名称 + 版本号」，右侧含页码；默认格式（宋体 10.5pt、10.5pt 固定行距、段前 0、段后 2.3pt）可保证每页恰好 50 行，满足中国版权保护中心对程序鉴别材料的格式要求。

## 与原版的关系

本项目是对 [kenley2021/swcr](https://github.com/kenley2021/swcr)（Python，MIT 协议）的 Go 重写，参数和行为完全对齐。以下为关键差异：

- **零第三方 docx 依赖**：直接用标准库 `archive/zip` + `encoding/xml` 构造 OOXML，不依赖 python-docx 或任何商业授权的 docx 库。
- **跨平台单二进制**：Go 编译为独立可执行文件，无需 Python 环境。
- **CLI 框架**：使用 `spf13/cobra`，与原版 `click` 体验一致。

**Go 版增强功能**（相比原版新增）：

- **块注释支持**：自动识别 `/* */`、`<!-- -->`、`""" """`、`''' '''` 等多行注释，支持自定义块注释对。
- **智能排除**：默认自动跳过 `node_modules`、`vendor`、`__pycache__`、`.venv`、`dist`、`build`、`target` 等构建/依赖目录，以及 `.pyc`、`.class`、`.o`、`.so` 等二进制文件。
- **预览模式（dry-run）**：`--dry-run` 打印文件数、代码行数、预估页数，不生成 `.docx`。
- **页数截取**：`--max-pages N` 限制输出页数，支持 `--page-mode first|last|front30back30` 三种策略，适配「前 30 页 + 后 30 页」等常见软著场景。
- **确定性排序**：输出文件按路径字母序排列。
- **`--version`**：打印版本信息和构建元数据。
- **Shell 补全**：内置 `swcr completion bash|zsh|fish|powershell`。

致谢原作者 **kenley** —— 本项目的参数设计、格式参数和核心算法均源于原项目。

## 安装

要求 Go 1.21+。

```bash
go install github.com/dengmengmian/swcr-go/cmd/swcr@latest
```

或者克隆仓库本地构建：

```bash
git clone https://github.com/dengmengmian/swcr-go.git
cd swcr-go
go build -o swcr ./cmd/swcr
```

## 用法

```
swcr [flags]
```

### 参数

| 参数 | 短名 | 默认值 | 说明 |
|------|------|--------|------|
| `--title` | `-t` | `软件著作权程序鉴别材料生成器V1.0` | 软件名称+版本号，用于生成页眉 |
| `--indir` | `-i` | 当前目录 | 源码所在文件夹，可指定多个 |
| `--ext` | `-e` | `py` | 源代码后缀（不含点），可指定多个 |
| `--comment-char` | `-c` | `#` `//` | 注释字符串，可指定多个 |
| `--font-name` | | `宋体` | 字体名称 |
| `--font-size` | | `10.5` | 字号（pt） |
| `--space-before` | | `0` | 段前间距（pt） |
| `--space-after` | | `2.3` | 段后间距（pt） |
| `--line-spacing` | | `10.5` | 行距（pt，固定值） |
| `--exclude` | | | 需要排除的文件或路径，可指定多个 |
| `--outfile` | `-o` | `code.docx` | 输出文件路径（.docx） |
| `--verbose` | `-v` | `false` | 打印详细日志 |

### 新增参数（Go 版）

| 参数 | 短名 | 默认值 | 说明 |
|------|------|--------|------|
| `--no-auto-exclude` | | `false` | 关闭智能排除 |
| `--dry-run` | | `false` | 预览文件数/行数/页数，不生成 docx |
| `--max-pages` | | `0`（不限制） | 最大页数，超出时按 `--page-mode` 截取 |
| `--page-mode` | | `first` | 截取策略：`first`（前N页）、`last`（后N页）、`front30back30`（前后各半） |
| `--lines-per-page` | | `50` | 每页行数（用于预估和截取） |
| `--block-comment` | `-b` | | 块注释对 `OPEN:CLOSE`（例如 `-b "/*:*/"`），可多次指定 |
| `--no-block-comment` | | `false` | 关闭默认块注释（`/* */`、`<!-- -->`、`""" """`、`''' '''`） |

### 示例

以下示例使用 [django-guardian](https://github.com/django-guardian/django-guardian) 项目演示。

#### 预览：看看会生成多少内容

```bash
swcr -i ./myproject -e py --dry-run
```

输出示例：
```
Files found : 142
Code lines  : 8450
Est. pages  : 169  (@ 50 lines/page)
```

#### 基础用法：扫描当前目录的 `.py` 文件

```bash
swcr -i ./myproject -o output.docx
```

#### 指定标题（页眉）

```bash
swcr -i ./myproject -t "MyApp V2.0" -o output.docx
```

#### 扫描多种源代码格式

```bash
swcr -i ./myproject \
    -t "MyApp V2.0" \
    -e py -e html -e js -e css \
    -o output.docx
```

#### 排除指定文件/目录

```bash
swcr -i ./myproject \
    -t "MyApp V2.0" \
    --exclude ./myproject/vendor \
    --exclude ./myproject/node_modules \
    --exclude ./myproject/.git \
    -o output.docx
```

#### 自定义注释风格

```bash
swcr -i ./myproject \
    -c '#' -c '//' -c '--' -c '/*' \
    -o output.docx
```

> **注意**：`swcr-go` 目前仅识别以指定字符串**开头**的单行注释。多行注释（如 `/* ... */`）不会被完整剔除，仅会移除以注释符开头的行。

#### 自定义字体和排版

```bash
swcr -i ./myproject \
    --font-name "Consolas" \
    --font-size 11 \
    --space-before 0 \
    --space-after 3 \
    --line-spacing 11 \
    -o output.docx
```

#### 超长项目截取（前后各 30 页）

```bash
swcr -i ./myproject -e py \
    --max-pages 60 --page-mode front30back30 \
    -o output.docx
```

#### Shell 自动补全

```bash
source <(swcr completion bash)
```

#### 详细日志

```bash
swcr -i ./myproject -v -o output.docx
```

## 如何实现每页 50 行

经测试，以下设置在 A4 纸上刚好产生每页 50 行：

| 属性 | 值 |
|------|-----|
| 字体 | 宋体 |
| 字号 | 10.5pt |
| 行距 | 10.5pt（固定值） |
| 段前间距 | 0pt |
| 段后间距 | 2.3pt |

## 开发

```bash
# 构建
make build

# 静态检查
make vet

# 测试（含竞态检测）
make test

# 格式化检查
make lint
```

### 发布

本仓库使用 [GoReleaser](https://goreleaser.com) 自动构建跨平台二进制和发布 GitHub Release。配置见 [`.goreleaser.yml`](./.goreleaser.yml)。

```bash
# 本地测试发布流程（不会真正推送）
goreleaser release --snapshot --clean
```

## License

MIT — 详见 [LICENSE](./LICENSE)。

本仓库是对 [kenley2021/swcr](https://github.com/kenley2021/swcr)（MIT 协议，Copyright (c) 2020 kenley）的 Go 重写，包含原项目的参数设计与格式参数。
