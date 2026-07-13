package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dengmengmian/swcr-go/internal/swcr"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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
)

var rootCmd = &cobra.Command{
	Use:   "swcr",
	Short: "软件著作权程序鉴别材料生成器",
	Long: `swcr 递归扫描源码目录，按后缀名筛选代码文件，跳过隐藏文件和
被排除的路径，逐行剔除空行和注释行，把剩下的代码行写入 .docx 文档。

页眉居中显示「软件名称+版本号」，右侧含页码；默认格式（宋体 10.5pt、
10.5pt 固定行距、段前 0、段后 2.3pt）可保证每页恰好 50 行。`,
	RunE: run,
}

func init() {
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
}

func run(cmd *cobra.Command, args []string) error {
	// Set up logging.
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	// Defaults.
	if len(indirs) == 0 {
		indirs = []string{"."}
	}
	if len(exts) == 0 {
		exts = []string{"py"}
	}
	if len(commentChars) == 0 {
		commentChars = []string{"#", "//"}
	}

	// Resolve absolute paths (like os.path.abspath in the Python code).
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
		// Strip trailing slash (del_slash in Python).
		absExcludes[i] = strings.TrimRight(a, string(filepath.Separator))
	}

	// Step 1: find code files.
	finder := swcr.NewCodeFinder(exts)
	var files []string
	for _, dir := range absIndirs {
		found, err := finder.Find(dir, absExcludes)
		if err != nil {
			return err
		}
		files = append(files, found...)
	}
	slog.Info("total code files found", "count", len(files))

	// Step 2: write to docx.
	opts := &swcr.WriterOpts{
		FontName:    fontName,
		FontSize:    fontSize,
		SpaceBefore: spaceBefore,
		SpaceAfter:  spaceAfter,
		LineSpacing: lineSpacing,
	}
	doc, err := swcr.NewDocument(outfile, opts)
	if err != nil {
		return err
	}
	writer := swcr.NewCodeWriter(commentChars, opts, doc)
	writer.WriteHeader(title)
	if err := writer.WriteFiles(files); err != nil {
		return err
	}
	if err := writer.Save(); err != nil {
		return err
	}
	slog.Info("done", "output", outfile, "files", len(files))
	return nil
}
