package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/ariela/translate-markdown/internal/deepl"
	"github.com/ariela/translate-markdown/internal/markdown"
)

// Translatorは翻訳処理のコアロジックを管理します。
type Translator struct {
	mdParser    *markdown.Parser
	deeplClient deepl.Translator
	cache       *Cache
	Report      *Report
	force       bool
	parallel    int
}

// translationTaskは並列処理のためのタスクを表します。
type translationTask struct {
	sourcePath string
	destPath   string
	targetLang string
}

// NewTranslatorは新しいTranslatorインスタンスを作成します。
func NewTranslator(client deepl.Translator, projectRoot string, force bool, parallel int) (*Translator, error) {
	parser := markdown.NewParser()
	cache, err := NewCache(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}
	return &Translator{
		mdParser:    parser,
		deeplClient: client,
		cache:       cache,
		Report:      NewReport(),
		force:       force,
		parallel:    parallel,
	}, nil
}

// TranslateJobは単一の翻訳ジョブを処理します。
func (t *Translator) TranslateJob(job Job, cfg *Config) error {
	info, err := os.Stat(job.Source)
	if err != nil {
		return fmt.Errorf("source not found: %w", err)
	}

	targetLang := job.TargetLang
	if targetLang == "" {
		targetLang = cfg.TargetLang
	}
	if targetLang == "" {
		return fmt.Errorf("target_lang is not specified for job or globally")
	}

	if info.IsDir() {
		return t.translateDirectory(job, targetLang)
	}

	// 単一ファイルの場合も並列処理の枠組みを使う
	tasks := []translationTask{{
		sourcePath: job.Source,
		destPath:   job.Destination,
		targetLang: targetLang,
	}}
	t.runWorkers(tasks)
	return nil
}

// translateDirectoryはディレクトリ内の全てのMarkdownファイルを再帰的に翻訳します。
func (t *Translator) translateDirectory(job Job, targetLang string) error {
	var tasks []translationTask
	walkErr := filepath.WalkDir(job.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Report.AddError(path, err)
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// 除外チェック
		for _, pattern := range job.Exclude {
			match, matchErr := doublestar.Match(pattern, path)
			if matchErr != nil {
				t.Report.AddError(path, fmt.Errorf("invalid exclude pattern: %w", matchErr))
				return nil
			}
			if match {
				fmt.Printf("Skipping excluded file: %s\n", path)
				t.Report.IncrementSkipped()
				return nil
			}
		}

		relPath, relErr := filepath.Rel(job.Source, path)
		if relErr != nil {
			t.Report.AddError(path, relErr)
			return nil
		}
		destPath := filepath.Join(job.Destination, relPath)

		if mkdirErr := os.MkdirAll(filepath.Dir(destPath), 0755); mkdirErr != nil {
			t.Report.AddError(path, mkdirErr)
			return nil
		}

		tasks = append(tasks, translationTask{
			sourcePath: path,
			destPath:   destPath,
			targetLang: targetLang,
		})
		return nil
	})

	if walkErr != nil {
		return walkErr
	}

	t.runWorkers(tasks)
	return nil
}

// runWorkersはタスクをワーカーに割り当てて並列実行します。
func (t *Translator) runWorkers(tasks []translationTask) {
	taskCh := make(chan translationTask, len(tasks))
	var wg sync.WaitGroup

	// ワーカーを起動
	for i := 0; i < t.parallel; i++ {
		wg.Add(1)
		go t.worker(&wg, taskCh)
	}

	// タスクをチャネルに送信
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	// 全てのワーカーが終了するのを待つ
	wg.Wait()
}

// workerはチャネルからタスクを受け取り、翻訳処理を実行します。
func (t *Translator) worker(wg *sync.WaitGroup, tasks <-chan translationTask) {
	defer wg.Done()
	for task := range tasks {
		if err := t.translateFile(task.sourcePath, task.destPath, task.targetLang); err != nil {
			t.Report.AddError(task.sourcePath, err)
		}
	}
}

// translateFileは単一のMarkdownファイルを翻訳します。
func (t *Translator) translateFile(sourcePath, destPath, targetLang string) error {
	hash, err := CalculateMD5(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to calculate hash for %s: %w", sourcePath, err)
	}

	if !t.force && !t.cache.IsChanged(sourcePath, hash) {
		fmt.Printf("Skipping unchanged file: %s\n", sourcePath)
		t.Report.IncrementSkipped()
		return nil
	}

	fmt.Printf("Translating %s -> %s\n", sourcePath, destPath)

	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	segments, err := t.mdParser.Parse(sourceContent)
	if err != nil {
		// parserからの詳細なエラーを返す
		return fmt.Errorf("failed to parse markdown file %s: %w", sourcePath, err)
	}

	var textsToTranslate []string
	var charCount int
	for _, seg := range segments {
		if seg.IsTranslatable && strings.TrimSpace(seg.Content) != "" {
			textsToTranslate = append(textsToTranslate, seg.Content)
			charCount += utf8.RuneCountInString(seg.Content)
		}
	}

	if len(textsToTranslate) == 0 {
		fmt.Printf("No translatable text found in %s, copying file.\n", sourcePath)
		err = os.WriteFile(destPath, sourceContent, 0644)
		if err != nil {
			return err
		}
		t.cache.Update(sourcePath, hash)
		t.Report.IncrementSuccess()
		return nil
	}

	translatedTexts, err := t.deeplClient.Translate(textsToTranslate, targetLang)
	if err != nil {
		return err
	}

	translatedTextIndex := 0
	for i, seg := range segments {
		if seg.IsTranslatable && strings.TrimSpace(seg.Content) != "" {
			if translatedTextIndex < len(translatedTexts) {
				segments[i].Content = translatedTexts[translatedTextIndex]
				translatedTextIndex++
			}
		}
	}

	reconstructedContent := markdown.Reconstruct(segments)

	if err := os.WriteFile(destPath, []byte(reconstructedContent), 0644); err != nil {
		return err
	}

	t.cache.Update(sourcePath, hash)
	t.Report.IncrementSuccess()
	t.Report.AddChars(charCount)
	return nil
}

// SaveCacheはメモリ上のキャッシュをファイルに保存します。
func (t *Translator) SaveCache() error {
	return t.cache.Save()
}
