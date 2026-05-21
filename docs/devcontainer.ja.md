# Dev Container の実行時シークレット

このページは、復号後の平文ファイルを
プロジェクトのワークスペースに置きたくない場合の配置を決めるためのものです。

AI ツールがプロジェクトのワークスペース内で動く場合、
復号後の平文ファイルをワークスペース外に置く構成が有効です。
この構成にすると、AI ツールが通常見る作業ディレクトリには
暗号化済みファイルとポリシーだけが残ります。

推奨形:

- リポジトリのポリシーには暗号化済みファイルを置く
- またはシークレット参照を置く
- 復号後の平文ファイルはプロジェクト外へ書く
- 例: `/workspaces/.runtime-secrets/app.env`
- Dev Container 設定でそのファイルを環境変数ファイルとして読み込む
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

Dev Container 例:

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

このリポジトリは Dev Container やマウント構成を作りません。
意図した場所を記録します。
平文ファイルがリポジトリ側へ戻っていないかを検査します。

使う判断:

- 復号後の `.env` をリポジトリ内に置きたくないならこの構成を使う
- アプリには `runArgs` などで環境変数ファイルを渡す
- ポリシーには、リポジトリ内に出てはいけない場所と
  実際の復号先の両方を書く
