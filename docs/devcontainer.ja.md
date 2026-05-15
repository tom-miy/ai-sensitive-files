# Devcontainer Runtime Secrets

AI ツールがプロジェクトのワークスペース内で動く場合、
復号後の平文ファイルをワークスペース外に置く構成が有効です。

推奨形:

- リポジトリのポリシーには暗号化済みファイルを置く
- またはシークレット参照を置く
- 復号後の平文ファイルはプロジェクト外へ書く
- 例: `/workspaces/.runtime-secrets/app.env`
- devcontainer 設定でそのファイルを環境変数ファイルとして読み込む
- `path` には `.env` のような場所を書く
- これはリポジトリ内に出てはいけない場所

ポリシー例:

```yaml
sensitive_files:
  - path: ".env"
    encrypted_path: ".env.sops.yaml"
    decrypted_path: "/workspaces/.runtime-secrets/app.env"
    crypto:
      method: "sops-age"
      decrypt_command: "umask 077; mkdir -p /workspaces/.runtime-secrets; sops --decrypt {encrypted_path} > {decrypted_path}"
      encrypt_command: "umask 077; sops --encrypt --output {encrypted_path} {decrypted_path}"
      manual_edit: "decrypted"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

devcontainer 例:

```jsonc
{
  "runArgs": [
    "--env-file",
    "/workspaces/.runtime-secrets/app.env"
  ],
  "postStartCommand": "ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml"
}
```

アプリのプロセスを起動する前に復号コマンドを実行します。
そのファイルをコンテナ実行環境に環境変数として読み込ませます。
アプリにプロジェクト内のシークレットファイルを読ませずに済みます。

このリポジトリは devcontainer やマウント構成を作りません。
意図した場所を記録します。
平文ファイルがリポジトリ側へ戻っていないかを検査します。
