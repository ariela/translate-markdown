# 技術スタック (TECH.md)
このドキュメントは、本アプリケーション開発で使用する技術スタックを定義します。

## 1. プログラミング言語
- **Go (Golang)**: バージョン 1.24
    - **理由**: コンパイル言語としての高い実行パフォーマンス、言語標準の強力な並行処理機能（Goroutine）、そして依存関係を含んだ単一バイナリとして配布できる容易さから、本CLIツールの要件に最適と判断。

## 2. 主要ライブラリ
- **CLI**: `cobra`
    - **理由**: 複雑なサブコマンドやフラグを持つ、モダンで高機能なCLIアプリケーションを容易に構築できる。Golangにおけるデファクトスタンダードであり、実績と信頼性が高い。 
- **設定ファイル (TOML)**: `BurntSushi/toml` 
    - **理由**: GolangにおけるTOML解析のデファクトスタンダード。仕様への準拠度が高く、信頼性・安定性に優れる。
- **Markdown解析**: `goldmark` 
    - **理由**: CommonMark仕様に準拠し、GFM（GitHub Flavored Markdown）にも対応。拡張性が高く、翻訳対象外とするノードを指定するようなカスタマイズに適している。
