package main

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/your-username/translate-markdown/internal/app"
	"github.com/your-username/translate-markdown/internal/deepl"
	"github.com/your-username/translate-markdown/internal/logging"
)

var (
	configPath string
	force      bool
	parallel   int
)

// rootCmdはアプリケーションのルートコマンドを表します。
var rootCmd = &cobra.Command{
	Use:   "translate-markdown",
	Short: "A CLI tool to translate Markdown files using DeepL API.",
	Long: `translate-markdown is a command-line tool that translates Markdown files
while preserving the structure, such as code blocks and frontmatter.`,
	Run: func(cmd *cobra.Command, args []string) {
		// プロジェクトルートとログファイルのパスを設定
		projectRoot := filepath.Dir(configPath)
		logFilePath := filepath.Join(projectRoot, "translate-errors.log")

		// ログファイルを開く
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer logFile.Close()

		// ロガーを初期化
		logger := setupLogger(logFile)
		slog.SetDefault(logger)

		// 設定ファイルを読み込む
		cfg, err := app.LoadConfig(configPath)
		if err != nil {
			slog.Error("Error loading config", "error", err)
			os.Exit(1)
		}

		// APIキーを環境変数から取得
		apiKey := os.Getenv("DEEPL_AUTH_KEY")
		if apiKey == "" {
			slog.Error("DEEPL_AUTH_KEY environment variable not set.")
			os.Exit(1)
		}

		// DeepLクライアントと翻訳クライアントを初期化
		deeplClient := deepl.NewClient(apiKey, logger)
		translator, err := app.NewTranslator(deeplClient, projectRoot, force, parallel)
		if err != nil {
			slog.Error("Failed to create translator", "error", err)
			os.Exit(1)
		}

		// 全てのジョブを実行
		for _, job := range cfg.Jobs {
			slog.Info("Executing job", "source", job.Source)
			err := translator.TranslateJob(job, cfg)
			if err != nil {
				translator.Report.AddError(job.Source, err)
				slog.Warn("Error processing job", "source", job.Source, "error", err)
			}
		}

		// キャッシュを保存
		if err := translator.SaveCache(); err != nil {
			slog.Warn("Failed to save cache", "error", err)
		}

		// 完了レポートを出力
		translator.Report.Print()
	},
}

// setupLoggerはデバッグモードに応じてロガーを設定します。
func setupLogger(logFile io.Writer) *slog.Logger {
	logLevel := slog.LevelInfo
	isDebug := os.Getenv("TRANSLATE_DEBUG") == "1"
	if isDebug {
		logLevel = slog.LevelDebug
	}

	// コンソールとログファイルの両方に出力
	multiWriter := io.MultiWriter(os.Stderr, logFile)

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(multiWriter, opts)

	// デバッグモードでない場合は、エラー発生時にのみログを出力するハンドラでラップ
	if !isDebug {
		return slog.New(logging.NewFingerCrossedHandler(handler, slog.LevelWarn))
	}
	return slog.New(handler)
}

func init() {
	// フラグを定義
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.toml", "path to the configuration file")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "force translation even if the file is not modified")
	// デフォルトの並列数はCPUのコア数とする
	rootCmd.PersistentFlags().IntVar(&parallel, "parallel", runtime.NumCPU(), "number of parallel translations")
}

func main() {
	// slogを使うため、標準のlogの出力を無効化
	log.SetOutput(io.Discard)
	if err := rootCmd.Execute(); err != nil {
		// cobraのエラーはslogで出力されないため、ここで明示的に出力
		slog.Error("Command failed", "error", err)
		os.Exit(1)
	}
}
