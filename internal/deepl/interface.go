package deepl

// Translatorはテキスト翻訳サービスのインターフェースを定義します。
// これにより、テスト時にAPIクライアントをモックすることができます。
type Translator interface {
	Translate(texts []string, targetLang string) ([]string, error)
}
