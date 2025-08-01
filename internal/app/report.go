package app

import (
	"fmt"
	"sync"
)

// TranslationErrorはファイルごとのエラー情報を保持します。
type TranslationError struct {
	FilePath string
	Err      error
}

// Reportは翻訳処理の結果を集計します。
type Report struct {
	mu              sync.Mutex
	SuccessCount    int
	SkippedCount    int
	FailedCount     int
	TranslatedChars int
	Errors          []TranslationError
}

// NewReportは新しいReportインスタンスを作成します。
func NewReport() *Report {
	return &Report{
		Errors: make([]TranslationError, 0),
	}
}

// IncrementSuccessは成功カウントを1増やします。
func (r *Report) IncrementSuccess() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.SuccessCount++
}

// IncrementSkippedはスキップカウントを1増やします。
func (r *Report) IncrementSkipped() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.SkippedCount++
}

// AddErrorは失敗カウントを1増やし、エラー情報を記録します。
func (r *Report) AddError(filePath string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.FailedCount++
	r.Errors = append(r.Errors, TranslationError{FilePath: filePath, Err: err})
}

// AddCharsは翻訳した文字数を加算します。
func (r *Report) AddChars(count int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.TranslatedChars += count
}

// Printは集計結果をコンソールに出力します。
func (r *Report) Print() {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Println("\n--- Translation Summary ---")
	fmt.Printf("✅ Successful: %d\n", r.SuccessCount)
	fmt.Printf("⏩ Skipped:    %d\n", r.SkippedCount)
	fmt.Printf("❌ Failed:     %d\n", r.FailedCount)
	fmt.Printf("🔤 Characters: %d\n", r.TranslatedChars)
	fmt.Println("---------------------------")

	if r.FailedCount > 0 {
		fmt.Println("\nErrors:")
		for _, e := range r.Errors {
			fmt.Printf("- File: %s\n  Error: %v\n", e.FilePath, e.Err)
		}
		fmt.Println("---------------------------")
	}
}
