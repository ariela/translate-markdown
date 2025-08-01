# コーディング規約 (CONVENTION.md)

このドキュメントは、本プロジェクトにおけるコードの品質と一貫性を保つための規約を定めます。

## 1. 静的解析 (Static Analysis)

- **ツール**: `golangci-lint`
    - **目的**: 潜在的なバグや非効率なコードを早期に発見するため、多数のリンターを統合した`golangci-lint`を使用します。
    - **運用**: CI/CDパイプラインに組み込み、すべてのプルリクエストに対して自動的にチェックを実行します。設定はリポジトリルートの`.golangci.yml`ファイルで管理します。

## 2. コードフォーマット (Code Formatting)

- **ツール**: `gofmt`
    - **目的**: Go言語の公式フォーマッターを使用し、コードスタイルを統一します。これにより、フォーマットに関する無用な議論をなくします。
    - **運用**: ファイル保存時やコミット前にエディタやIDEの機能を用いて自動的に実行することを強く推奨します。CI/CDでもフォーマットが正しいか検証します。

## 3. コミットメッセージ (Commit Messages)

- **規約**: [Conventional Commits](https://www.conventionalcommits.org/)
    - **目的**: コミット履歴を読みやすく、変更内容の追跡を容易にするため、標準化されたフォーマットに従います。
    - **例**:
        - `feat: add translation summary report`
        - `fix: correct handling of API rate limits`
        - `docs: update TECH.md with library selection`
