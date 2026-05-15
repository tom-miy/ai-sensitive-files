# Policy Config

ポリシーは適用先リポジトリの
`.ai-sensitive-files/sensitive-files.yaml` に置きます。

## ファイル全体の形

ファイル全体は最上位のキー `sensitive_files` を 1 つ持ちます。
各項目が 1 つの機密ファイル場所ポリシーです。

```yaml
sensitive_files:
  - path: ".env"
    reason: "local environment secrets"
    tags: ["env", "secret"]
    action:
      aiignore: true
      gitignore: true
      encrypt: false
      commit_block: true
```

暗号化するファイルや、
シークレット管理サービスから復元するファイルでは、
`decrypted_path` と `crypto` を追加します。

```yaml
sensitive_files:
  - path: ".env"
    encrypted_path: ".env.sops.yaml"
    decrypted_path: ".env"
    reason: "local environment secrets"
    tags: ["env", "secret"]
    crypto:
      method: "sops-age"
      recipients: ["age1exampleteampublickey...", "age1examplecipublickey..."]
      encrypt_command: "sops --encrypt --output {encrypted_path} {decrypted_path}"
      decrypt_command: "sops --decrypt --output {decrypted_path} {encrypted_path}"
      manual_edit: "decrypted"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

外部シークレット管理サービスを使う場合は、
`encrypted_path` の代わりに `crypto.secret_ref` を使います。

```yaml
sensitive_files:
  - path: ".env.ci"
    decrypted_path: ".env.ci"
    reason: "CI-only environment secrets fetched from 1Password"
    crypto:
      method: "1password"
      secret_ref: "op://Engineering/App CI/.env"
      decrypt_command: "op read {secret_ref} > {decrypted_path}"
      manual_edit: "none"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

## 項目

最上位の項目:

- `sensitive_files`: 最上位のリスト

各 `sensitive_files` 項目の必須項目:

- `path`: 相対パスまたは基本的なワイルドカードパターン
- `reason`: 機密扱いする理由
- `action`: この場所をどの生成物・検査に使うかの指定
- `crypto`: `action.encrypt` が true の場合は必須

任意の場所指定項目:

- `encrypted_path`: 暗号化済みファイルの場所。例: `.env.sops.yaml`
- `decrypted_path`:
  復号後に平文ファイルが現れる場所。
  `action.encrypt` が true の場合は必須
- `crypto.secret_ref`:
  リポジトリ内の暗号化済みファイルを持たない
  1Password / Bitwarden 項目の外部シークレット参照

devcontainer などで平文ファイルをワークスペース外に置く場合、
`decrypted_path` はリポジトリ外パスでも構いません。
例: `/workspaces/.runtime-secrets/app.env`

ワークスペース外のパスは平文ファイルのずれの検査対象になります。
ただし、`.gitignore` や AI ignore ファイルには出力しません。

`action.encrypt` が true の場合の必須 `crypto` 項目:

- `method`: 暗号化方式名。例: `sops-age`
- `recipients`:
  復号を許可する age 公開鍵。
  `sops-age` では必須。
  チーム用や CI 用の公開鍵を並べる。
  公開鍵なので Git 管理してよい
- `encrypt_command`:
  `{encrypted_path}` と `{decrypted_path}` を使うコマンドテンプレート。
  リポジトリ内の暗号化済みファイルでは必須
- `decrypt_command`:
  `{encrypted_path}`、`{decrypted_path}`、
  必要なら `{secret_ref}` を使うコマンドテンプレート
- `manual_edit`: `decrypted`, `encrypted`, `none` のいずれか

`action` の各フラグ:

- `aiignore: true`:
  `.aiignore`, `.cursorignore`, `.copilotignore`,
  Claude `denyRead` などの AI/editor ignore 出力に含める
- `gitignore: true`:
  `.gitignore.ai-sensitive-files` に含める。
  `check` では実際の `.gitignore` に反映済みかも確認する
- `encrypt: true`:
  SOPS/age、1Password、Bitwarden などで保護する対象として扱う。
  暗号化計画に含め、
  `decrypted_path` の平文ファイルも検査する
- `commit_block: true`:
  `path` またはリポジトリ相対の `decrypted_path` が
  Git 管理されていたら `check` を失敗させる

各項目では、少なくとも 1 つの `action` フラグを
true にする必要があります。

## パスのルール

path は適用先リポジトリのルートからの相対パスで書きます。
区切りには `/` を使います。

`path` は `sensitive_files` の 1 項目につき 1 つだけ書きます。
`encrypted_path` と `decrypted_path` は、
同じ項目に対する別の項目です。

OK 例:

| Field | Example | 意味 |
|---|---|---|
| `path` | `.env` | AI に見せずコミットも止めたいリポジトリ内パス |
| `path` | `.agent-privacy-guard/entities.local.yaml` | 入れ子になったリポジトリ内パス |
| `path` | `secrets/**` | リポジトリ内ディレクトリの基本的なワイルドカード |
| `encrypted_path` | `.env.sops.yaml` | リポジトリ内に置く暗号化済みファイル |
| `decrypted_path` | `/workspaces/.runtime-secrets/app.env` | devcontainer が読むワークスペース外の env ファイル |

NG 例:

| Example | NG 理由 |
|---|---|
| `/Users/alice/app/.env` | ホスト上の絶対パス |
| `C:\Users\alice\.env` | バックスラッシュを使った Windows 形式のパス |
| `secrets//token.env` | 空のパス要素 |

通常のリポジトリ内ファイルはリポジトリ内パスにします。
`/workspaces/.runtime-secrets/...` は、
devcontainer などで復号後の平文ファイルを
ワークスペース外に置く場合だけ使います。
環境変数として読み込むと決めた場合の例外です。
