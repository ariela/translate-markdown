# Markdown 翻訳ツール

DeepL APIを使用して、Markdownファイルの構造を維持したまま翻訳するCLIツールです。

## 概要

このツールは、Markdownファイル内のテキスト部分を翻訳し、コードブロック、Frontmatter、インラインコードなどの要素はそのまま維持します。ファイル単位またはディレクトリ単位での翻訳に対応しています。

ファイルの更新チェック機能により、変更があったファイルのみを翻訳するため、効率的です。また、並列処理により高速な翻訳を実現し、処理完了後にはサマリーレポートが出力されます。

## 機能

- `config.toml`で指定されたファイルまたはディレクトリを翻訳します。
- ディレクトリを指定した場合、配下の`.md`ファイルを再帰的に翻訳します。
- **並列処理**: 複数のファイルを同時に翻訳し、処理時間を短縮します。
- `exclude`パターンに一致するファイルやディレクトリを翻訳対象から除外します。
- **キャッシュ機能**: ファイルのMD5ハッシュを比較し、変更がないファイルは翻訳をスキップします。
- `--force`フラグでキャッシュを無視して強制的に再翻訳できます。
- **完了レポート**: 処理完了後、成功・スキップ・失敗したファイル数や翻訳文字数を表示します。
- **レート制限対応**: APIのレート制限エラー発生時に、自動でリトライ処理を行います。
- 環境変数 `DEEPL_AUTH_KEY` からDeepL APIキーを読み取ります。

## 使い方

1.  **バイナリのダウンロード**
    [GitHub Releasesページ](https://github.com/ariela/translate-markdown/releases)から、お使いのOSに合った最新のファイルをダウンロードしてください。

2.  **APIキーの設定**
    DeepL APIキーを環境変数に設定します。

    - **macOS / Linux:**
      ```sh
      export DEEPL_AUTH_KEY="your_deepl_api_key"
      ```
    - **Windows:**
      ```sh
      set DEEPL_AUTH_KEY="your_deepl_api_key"
      ```

3.  **設定ファイルの作成**
    ダウンロードした実行ファイルと同じディレクトリに `config.toml` ファイルを作成します。設定内容の詳細は `config.toml.example` を参照してください。

4.  **実行**
    ターミナル（またはコマンドプロンプト）から以下のコマンドで実行します。

    ```sh
    ./translate-markdown --config config.toml
    ```

## 開発者向け (For Developers)

### 開発環境のセットアップ

このプロジェクトでは、開発ツールのバージョン管理に [mise](https://mise.jdx.dev/) の使用を推奨しています。

1.  **miseのインストール**
    公式サイトの案内に従って`mise`をインストールしてください。
    ```sh
    # macOS / Linux (Homebrew)
    brew install mise

    # または、以下のスクリプトでインストール
    curl https://mise.run | sh
    ```
    インストール後、シェル設定を更新するのを忘れないでください。

2.  **リポジトリのクローン**
    ```sh
    git clone https://github.com/ariela/translate-markdown.git
    cd translate-markdown
    ```

3.  **開発ツールのインストール**
    プロジェクトルートに移動すると、`mise`が`.mise.toml`を検知し、指定されたバージョンのGo言語を自動でインストール・セットアップします。手動で実行する場合は以下のコマンドを使用します。
    ```sh
    mise install
    ```

4.  **(任意) golangci-lintのインストール**
    コードの静的解析（Lint）を行う場合は、`golangci-lint` をインストールします。
    ```sh
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    ```

### よく使うコマンド (Common Commands)

プロジェクトルートに、よく使うコマンドをタスクとして定義した `.mise.toml` があります。

- **フォーマット (Formatting)**
  ```sh
  mise run fmt
  ```

- **静的解析 (Linting)**
  ```sh
  mise run lint
  ```

### ビルドと実行

リポジトリをクローンした後、以下のコマンドでソースコードから直接プログラムを実行できます。

```sh
# 通常の実行 (CPUコア数に応じた並列処理)
go run ./cmd/translate-markdown/main.go --config config.toml

# 並列数を4に指定して実行
go run ./cmd/translate-markdown/main.go --config config.toml --parallel 4

# 全てのファイルを強制的に再翻訳
go run ./cmd/translate-markdown/main.go --config config.toml --force
