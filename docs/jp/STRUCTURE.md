# ファイル構成 (STRUCTURE.md)

このドキュメントは、プロジェクトのディレクトリとファイルの構成を定義します。この構成は、Goの標準的なプロジェクトレイアウトに従い、関心の分離とクロスコンパイルの容易さを目的としています。

```
translate-markdown/
├── cmd/
│   └── translate-markdown/
│       └── main.go         # CLIのエントリーポイント
├── internal/
│   ├── app/                # アプリケーションのコアロジック
│   │   ├── cache.go        # 翻訳キャッシュの管理
│   │   ├── config.go       # 設定ファイルの読み込み・解析
│   │   ├── report.go       # 完了レポートの管理
│   │   └── translator.go   # 翻訳処理のメインロジック
│   ├── deepl/              # DeepL APIとの連携
│   │   ├── client.go       # DeepL APIクライアントの実装
│   │   └── interface.go    # テスト容易性のためのインターフェース
│   └── markdown/           # Markdownファイルの解析
│       └── parser.go
├── .github/
│   └── workflows/
│       └── ci.yml          # CI/CDパイプライン定義
├── .goreleaser.yml         # GoReleaserの設定ファイル
├── .gitignore
├── go.mod                  # Goモジュールの定義
├── go.sum
├── README.md
└── config.toml.example     # 設定ファイルの実例
```

### ディレクトリ説明

- **`cmd/`**:
    - アプリケーションの実行可能ファイル（バイナリ）のエントリーポイントとなる`main`パッケージを配置します。
- **`internal/`**:
    - このプロジェクト内部でのみ使用されるプライベートなパッケージを配置します。
    - `internal`以下に配置されたコードは、他のプロジェクトから直接インポートできなくなり、意図しない依存関係を防ぎます。
    - 各サブディレクトリ（`app`, `deepl`, `markdown`）は、それぞれの責務に特化したロジックをカプセル化します。
- **`.github/workflows/`**:
    - GitHub Actionsを利用したCI/CD（継続的インテグレーション/継続的デリバリー）のワークフロー定義ファイルを配置します。

### 主要ファイル説明

- **`.goreleaser.yml`**:
    - クロスコンパイルとGitHubリリースの作成を自動化するツール`GoReleaser`の設定ファイルです。
- **`config.toml.example`**:
    - ユーザーが作成する設定ファイル`config.toml`のテンプレートです。
