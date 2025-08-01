package app

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const cacheFileName = ".translation_cache.json"

// Cacheは翻訳済みファイルのハッシュを保持します。
type Cache struct {
	path string
	// キー: ファイルパス, 値: MD5ハッシュ
	Hashes map[string]string `json:"hashes"`
}

// NewCacheは新しいCacheインスタンスを作成し、既存のキャッシュファイルを読み込みます。
func NewCache(projectRoot string) (*Cache, error) {
	cachePath := filepath.Join(projectRoot, cacheFileName)
	c := &Cache{
		path:   cachePath,
		Hashes: make(map[string]string),
	}
	if err := c.Load(); err != nil {
		// ファイルが存在しない場合はエラーとしない
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return c, nil
}

// Loadはキャッシュファイルをディスクから読み込みます。
func (c *Cache) Load() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &c)
}

// Saveは現在のキャッシュの状態をディスクに保存します。
func (c *Cache) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}

// IsChangedはファイルのハッシュがキャッシュ内のものと異なるかを確認します。
// キャッシュに存在しない場合は変更ありとみなします。
func (c *Cache) IsChanged(filePath, currentHash string) bool {
	cachedHash, ok := c.Hashes[filePath]
	if !ok {
		return true // キャッシュにない場合は変更あり
	}
	return cachedHash != currentHash
}

// Updateはキャッシュ内のファイルのハッシュを更新します。
func (c *Cache) Update(filePath, newHash string) {
	c.Hashes[filePath] = newHash
}

// CalculateMD5はファイルのMD5ハッシュを計算します。
func CalculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
