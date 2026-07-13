# swcr-go：计算机软件著作权程序鉴别材料生成器（Go 版）

[![CI](https://github.com/dengmengmian/swcr-go/actions/workflows/ci.yml/badge.svg)](https://github.com/dengmengmian/swcr-go/actions/workflows/ci.yml)

`swcr-go` 是 [swcr](https://github.com/kenley2021/swcr)（作者 kenley，MIT 协议）的 Go 语言重写版。它与原版行为对齐：递归扫描源码目录，按后缀名筛出代码文件，跳过隐藏文件和被排除的路径，逐行剔除空行和注释行，把剩余的代码行写入一个 `.docx` 文档。页眉居中显示「软件名称 + 版本号」，右侧含页码；默认格式（宋体 10.5pt、10.5pt 固定行距、段前 0、段后 2.3pt）可保证每页恰好 50 行，满足中国版权保护中心对程序鉴别材料的格式要求。

## 与原版的关系

本项目是对 [kenley2021/swcr](https://github.com/kenley2021/swcr)（Python，MIT 协议）的 Go 重写，参数和行为完全对齐。以下为关键差异：

- **零第三方 docx 依赖**：直接用标准库 `archive/zip` + `encoding/xml` 构造 OOXML，不依赖 python-docx 或任何商业授权的 docx 库。
- **跨平台单二进制**：Go 编译为独立可执行文件，无需 Python 环境。
- **CLI 框架**：使用 `spf13/cobra`，与原版 `click` 体验一致。

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

### 示例

以下示例使用 [django-guardian](https://github.com/django-guardian/django-guardian) 项目演示。

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
go build ./...

# 静态检查
go vet ./...

# 测试（含竞态检测）
go test -race -count=1 ./...
```

## License

MIT — 详见 [LICENSE](./LICENSE)。

本仓库是对 [kenley2021/swcr](https://github.com/kenley2021/swcr)（MIT 协议，Copyright (c) 2020 kenley）的 Go 重写，包含原项目的参数设计与格式参数。
