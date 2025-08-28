# ブランチ保護ルール設定手順

このドキュメントでは、terraform-provider-rtxリポジトリのmainブランチに適用すべきブランチ保護ルールについて説明します。

## 設定手順

1. GitHubリポジトリの Settings タブに移動
2. 左サイドバーの「Branches」をクリック  
3. 「Add rule」ボタンをクリック
4. 以下の設定を適用

## 推奨設定

### 基本設定
- **Branch name pattern**: `main`
- **Require a pull request before merging**: ✅ 有効
  - **Require approvals**: 1 (1人以上の承認が必要)
  - **Dismiss stale PR approvals when new commits are pushed**: ✅ 有効
  - **Require review from code owners**: ❌ 無効 (CODEOWNERSファイルがない場合)

### ステータスチェック
- **Require status checks to pass before merging**: ✅ 有効
- **Require branches to be up to date before merging**: ✅ 有効
- **必須ステータスチェック**:
  - `Test` (CI workflow のテストジョブ)
  - `Lint` (CI workflow のlintジョブ)  
  - `Build` (CI workflow のビルドジョブ)

### 制限事項
- **Restrict pushes that create files**: ❌ 無効
- **Require linear history**: ✅ 有効
- **Allow force pushes**: ❌ 無効
- **Allow deletions**: ❌ 無効

### 管理者設定
- **Do not allow bypassing the above settings**: ✅ 有効
  - 管理者でも保護ルールをバイパスできないように設定

## 期待される効果

1. **品質保証**: 全テスト・lint・ビルドが成功しないとマージ不可
2. **コードレビュー**: Pull Request必須で最低1人の承認が必要
3. **履歴の整合性**: 線形な履歴を維持し、プロジェクトの追跡が容易
4. **事故防止**: 直接pushやforce pushを防止

## CI/CDとの連携

設定後、以下のワークフローが自動化されます：

1. Pull Request作成時に自動でCI実行
2. テスト・lint・ビルドの全てが成功するまでマージブロック
3. 承認されたPull Requestのみマージ可能
4. マージ後にmainブランチで再度CIが実行され品質確保

## 例外処理

緊急時やホットフィックスが必要な場合は、リポジトリ管理者が一時的に保護設定を無効化することも可能ですが、基本的には推奨されません。