# Integration

適用先リポジトリでの推奨手順:

1. このリポジトリで `mise trust .` を実行する
2. ローカルのツール定義を信頼する
3. `mise install` で Go、SOPS、age、Lefthook を入れる
4. `mise run install-cli` で `.bin/ai-sensitive-files` をビルドする
5. このリポジトリから `bash install.sh --target /path/to/app` を実行する
6. `.ai-sensitive-files/sensitive-files.yaml` を確認する
7. `ai-sensitive-files validate --config .ai-sensitive-files/sensitive-files.yaml` を実行する
8. `ai-sensitive-files generate --config .ai-sensitive-files/sensitive-files.yaml --out .` を実行する
9. `.gitignore.ai-sensitive-files` を確認する
10. 必要な項目を `.gitignore` に取り込む
11. Lefthook を使うリポジトリではサンプルを手動で取り込む

生成物もセキュリティに関わる設定としてレビューしてください。
単一の ignore ファイルだけを唯一の制御にしないでください。
