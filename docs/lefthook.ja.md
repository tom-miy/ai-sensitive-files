# Lefthook

Lefthook ではコミット前のゲートとして使います。

```yaml
pre-commit:
  commands:
    ai-sensitive-files:
      run: ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml
```

このコマンドは、生成物がポリシーと同期しているかを確認します。
生成された `.gitignore` 項目が、
実際の `.gitignore` に反映されているかも確認します。
コミットブロック対象が Git 管理されていないかも確認します。
平文出力が `crypto.manual_edit` に従っているかも確認します。
コミット前に復号後ファイルが
暗号化済みファイルより新しくなっていないかも確認します。

インストーラーは既存の Lefthook 設定を編集しません。
`templates/lefthook/lefthook.example.yml` を確認して、
手動で取り込んでください。
