package main

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dengmengmian/swcr-go/internal/swcr"
)

// Build information injected via ldflags.
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func versionString() string {
	return fmt.Sprintf(
		"swcr %s (%s) %s/%s commit=%s built=%s",
		version, runtime.Version(), runtime.GOOS, runtime.GOARCH,
		commit, buildDate,
	)
}

const defaultLinesPerPage = 50

var pageModes = map[string]bool{"first": true, "last": true, "front30back30": true}

var (
	title        string
	indirs       []string
	exts         []string
	commentChars []string
	fontName     string
	fontSize     float64
	spaceBefore  float64
	spaceAfter   float64
	lineSpacing  float64
	excludes     []string
	outfile      string
	verbose      bool

	dryRun         bool
	maxPages       int
	pageMode       string
	linesPerPage   int
	blockComments  []string
	noAutoExclude  bool
	noBlockComment bool
)

var rootCmd = &cobra.Command{
	Version: versionString(),

	Use:   "swcr",
	Short: "软件著作权程序鉴别材料生成器",
	Long: `swcr 递归扫描源码目录，按后缀名筛选代码文件，跳过隐藏文件和
被排除的路径，逐行剔除空行和注释行，把剩下的代码行写入 .docx 文档。

页眉居中显示「软件名称+版本号」，右侧含页码；默认格式（宋体 10.5pt、
10.5pt 固定行距、段前 0、段后 2.3pt）可保证每页恰好 50 行。`,

	SilenceUsage: true,
	RunE:         run,
}

func init() {
	// Register completion command for bash/zsh/fish/powershell.
	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "生成 shell 自动补全脚本",
		Long: `生成指定 shell 的自动补全脚本。

  bash       source <(swcr completion bash)
  zsh        source <(swcr completion zsh)
  fish       swcr completion fish | source
  powershell swcr completion powershell | Out-String | Invoke-Expression`,
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
			case "zsh":
				return rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
			return nil
		},
	})

	rootCmd.Flags().StringVarP(&title, "title", "t",
		"软件著作权程序鉴别材料生成器V1.0",
		"软件名称+版本号，用于生成页眉")
	rootCmd.Flags().StringArrayVarP(&indirs, "indir", "i",
		nil, "源码所在文件夹，可指定多个，默认当前目录")
	rootCmd.Flags().StringArrayVarP(&exts, "ext", "e",
		nil, "源代码后缀（不含点），可指定多个，默认 py")
	rootCmd.Flags().StringArrayVarP(&commentChars, "comment-char", "c",
		nil, "注释字符串，可指定多个，默认 # //")
	rootCmd.Flags().StringVar(&fontName, "font-name", "宋体",
		"字体，默认宋体")
	rootCmd.Flags().Float64Var(&fontSize, "font-size", 10.5,
		"字号（pt），默认 10.5")
	rootCmd.Flags().Float64Var(&spaceBefore, "space-before", 0,
		"段前间距（pt），默认 0")
	rootCmd.Flags().Float64Var(&spaceAfter, "space-after", 2.3,
		"段后间距（pt），默认 2.3")
	rootCmd.Flags().Float64Var(&lineSpacing, "line-spacing", 10.5,
		"行距（pt，固定值），默认 10.5")
	rootCmd.Flags().StringArrayVar(&excludes, "exclude",
		nil, "需要排除的文件或路径，可指定多个")
	rootCmd.Flags().StringVarP(&outfile, "outfile", "o", "code.docx",
		"输出文件（.docx），默认 code.docx")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"打印详细日志")

	rootCmd.Flags().BoolVar(&noAutoExclude, "no-auto-exclude", false,
		"关闭自动排除（默认会自动跳过 node_modules、vendor、__pycache__ 等）")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"预览模式：打印文件数、总行数、预估页数，不生成 .docx")
	rootCmd.Flags().IntVar(&maxPages, "max-pages", 0,
		"最大页数限制（0 = 不限制）；超出时按 --page-mode 截取")
	rootCmd.Flags().StringVar(&pageMode, "page-mode", "first",
		"页面截取策略：first（前N页）、last（后N页）、front30back30（前后各半）")
	rootCmd.Flags().IntVar(&linesPerPage, "lines-per-page", defaultLinesPerPage,
		"每页行数，用于预估和截取（默认 50）")
	rootCmd.Flags().StringArrayVarP(&blockComments, "block-comment", "b",
		nil, "块注释对，格式 OPEN:CLOSE（例如 -b \"/*:*/\"）")
	rootCmd.Flags().BoolVar(&noBlockComment, "no-block-comment", false,
		"关闭默认块注释支持（/* */, <!-- -->, \"\"\" \"\"\", ''' '''）")
}

func run(_ *cobra.Command, _ []string) error {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	if maxPages > 0 && !pageModes[pageMode] {
		return fmt.Errorf("invalid --page-mode %q: must be one of first, last, front30back30", pageMode)
	}

	if len(indirs) == 0 {
		indirs = []string{"."}
	}
	if len(exts) == 0 {
		exts = []string{"py"}
	}

	absIndirs := make([]string, len(indirs))
	for i, d := range indirs {
		a, err := filepath.Abs(d)
		if err != nil {
			return err
		}
		absIndirs[i] = a
	}
	absExcludes := make([]string, len(excludes))
	for i, e := range excludes {
		a, err := filepath.Abs(e)
		if err != nil {
			return err
		}
		absExcludes[i] = strings.TrimRight(a, string(filepath.Separator))
	}

	// ── Step 1: find code files ──────────────────────────────────────────
	finder := swcr.NewCodeFinder(exts)
	finder.AutoExclude = !noAutoExclude
	var files []string
	for _, dir := range absIndirs {
		found, err := finder.Find(dir, absExcludes)
		if err != nil {
			return err
		}
		files = append(files, found...)
	}
	sort.Strings(files)
	slog.Info("total code files found", "count", len(files))

	if len(files) == 0 {
		slog.Info("no code files found; nothing to do")
		return nil
	}

	// ── Step 2: collect all code lines ────────────────────────────────────
	stripper := buildStripper()
	opts := &swcr.WriterOpts{
		FontName:    fontName,
		FontSize:    fontSize,
		SpaceBefore: spaceBefore,
		SpaceAfter:  spaceAfter,
		LineSpacing: lineSpacing,
	}
	writer := swcr.NewCodeWriter(stripper, opts, nil)
	allLines, err := writer.CollectLines(files)
	if err != nil {
		return fmt.Errorf("reading source files: %w", err)
	}
	totalLines := len(allLines)
	estPages := int(math.Ceil(float64(totalLines) / float64(linesPerPage)))

	slog.Info("code lines collected",
		"files", len(files),
		"lines", totalLines,
		"estimated_pages", estPages,
	)

	// ── Dry-run ───────────────────────────────────────────────────────────
	if dryRun {
		fmt.Printf("Files found : %d\n", len(files))
		fmt.Printf("Code lines  : %d\n", totalLines)
		fmt.Printf("Est. pages  : %d  (@ %d lines/page)\n", estPages, linesPerPage)
		if maxPages > 0 && estPages > maxPages {
			fmt.Printf("Page limit  : %d  (would apply --page-mode=%s)\n", maxPages, pageMode)
			kept := paginateLines(allLines, maxPages, linesPerPage, pageMode)
			fmt.Printf("Kept lines  : %d  (after truncation)\n", len(kept))
		}
		return nil
	}

	// ── Paginate ──────────────────────────────────────────────────────────
	if maxPages > 0 && estPages > maxPages {
		slog.Info("truncating to page limit", "max_pages", maxPages, "mode", pageMode)
		allLines = paginateLines(allLines, maxPages, linesPerPage, pageMode)
	}

	// ── Step 3: write docx ────────────────────────────────────────────────
	doc, err := swcr.NewDocument(outfile, opts)
	if err != nil {
		return err
	}
	docWriter := swcr.NewCodeWriter(stripper, opts, doc)
	docWriter.WriteHeader(title)
	docWriter.WriteLines(allLines)
	if err := docWriter.Save(); err != nil {
		return err
	}
	slog.Info("done", "output", outfile, "files", len(files), "lines", len(allLines))
	return nil
}

func buildStripper() *swcr.CommentStripper {
	if len(commentChars) == 0 {
		commentChars = []string{"#", "//"}
	}
	var bp []string
	if !noBlockComment {
		bp = swcr.DefaultBlockCommentPairs()
	}
	bp = append(bp, blockComments...)
	return swcr.NewCommentStripper(commentChars, bp)
}

func paginateLines(lines []string, maxPages, linesPerPage int, mode string) []string {
	maxLines := maxPages * linesPerPage
	if len(lines) <= maxLines {
		return lines
	}

	switch mode {
	case "last":
		start := len(lines) - maxLines
		return lines[start:]
	case "front30back30":
		frontPages := maxPages / 2
		backPages := maxPages - frontPages
		frontLines := frontPages * linesPerPage
		backLines := backPages * linesPerPage
		if frontLines+backLines >= len(lines) {
			return lines
		}
		front := lines[:frontLines]
		back := lines[len(lines)-backLines:]
		out := make([]string, 0, frontLines+backLines)
		out = append(out, front...)
		out = append(out, back...)
		return out
	default: // "first"
		return lines[:maxLines]
	}
}
