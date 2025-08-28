# Static Route Import Example

このディレクトリには、RTX routerの既存の静的ルートをTerraformにimportするための設定が含まれています。

## 現在のルーター設定

実際のルーター設定から以下の静的ルートが確認されています：

```
ip route 10.33.128.0/21 gateway 192.168.1.20 gateway 192.168.1.21
ip route 100.64.0.0/10 gateway 192.168.1.20 gateway 192.168.1.21
```

## セットアップ

1. `terraform.tfvars`ファイルが正しく設定されていることを確認してください：

```hcl
rtx_host           = "192.168.1.253"
rtx_username       = "shin1ohno"  
rtx_password       = "pino0108"
rtx_admin_password = "893kick!"
rtx_port           = 22
skip_host_key_check = false
```

## Import手順

1. Terraformを初期化：
```bash
terraform init
```

2. 既存の静的ルートを確認：
```bash
terraform plan
```

3. 既存ルートをimport：
```bash
./import.sh
```

または手動でimport（正しいIDフォーマットを使用）：
```bash
terraform import rtx_static_route.private_network_route "10.33.128.0/21||192.168.1.20||"
terraform import rtx_static_route.cgn_network_route "100.64.0.0/10||192.168.1.20||"
```

4. import後の確認：
```bash
terraform plan
```

## Import IDフォーマット

RTX providerでは以下のフォーマットを使用：
```
"destination||gateway||interface"
```

例：
- `"10.33.128.0/21||192.168.1.20||"` - インターフェース指定なし
- `"192.168.1.0/24||192.168.1.1||lan1"` - インターフェース指定あり

## ファイル説明

- `main.tf` - Provider設定、data source、実際の静的ルートリソース定義
- `variables.tf` - 変数定義
- `terraform.tfvars` - 実際の設定値（認証情報設定済み）
- `import.sh` - Import実行スクリプト（エラーハンドリング付き）

## トラブルシューティング

### Data Sourceが空の場合
- プロバイダーがmock modeで動作している可能性
- ルーターとの接続を確認
- プロバイダーの実装がルート読み取りに対応していない可能性

### Import失敗の場合
- ルートが実際に存在することを確認
- Import IDフォーマットが正しいことを確認
- プロバイダーがimport機能をサポートしていることを確認

## 注意事項

- Import前に`main.tf`内のリソース定義が実際のルーター設定と一致することを確認
- 複数gatewayが設定されている場合、primary gatewayのみが使用される
- Import後は`terraform plan`で差分がないことを確認