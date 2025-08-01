package app

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Configは設定ファイル(config.toml)の構造を表します。
type Config struct {
	TargetLang string `toml:"target_lang"`
	SourceLang string `toml:"source_lang"`
	Jobs       []Job  `toml:"jobs"`
}

// Jobは個々の翻訳タスクを表します。
type Job struct {
	Source      string   `toml:"source"`
	Destination string   `toml:"destination"`
	TargetLang  string   `toml:"target_lang"`
	SourceLang  string   `toml:"source_lang"`
	Exclude     []string `toml:"exclude"`
}

// LoadConfigは指定されたパスから設定ファイルを読み込み、解析します。
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if _, err := toml.Decode(string(data), &config); err != nil {
		return nil, err
	}

	return &config, nil
}
